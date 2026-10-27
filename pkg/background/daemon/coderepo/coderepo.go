// Copyright 2023 sigma
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package coderepo

import (
	"context"
	"encoding/json/v2"
	"errors"
	"fmt"
	"log/slog"
	"path"
	"reflect"
	"time"

	"go.uber.org/dig"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/background/daemon"
	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	repocoderepo "github.com/go-sigma/sigma/pkg/dal/repository/coderepo"
	repouser "github.com/go-sigma/sigma/pkg/dal/repository/user"
	"github.com/go-sigma/sigma/pkg/infra/workq"
	svccredential "github.com/go-sigma/sigma/pkg/service/credential"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

const (
	perPage = 100
	// maxStatusMessageLength keeps the sync error message within the size of the
	// cr_last_update_message column.
	maxStatusMessageLength = 250
)

func init() {
	utils.PanicIf(daemon.Daemons.Register(path.Base(reflect.TypeFor[factory]().PkgPath()), &factory{}))
}

type factory struct{}

type params struct {
	dig.In

	HandlerRegistry          workq.HandlerRegistry
	UserRepository           repouser.UserRepository
	CodeRepositoryRepository repocoderepo.CodeRepositoryRepository
	Config                   *config.Configuration
	SvcCredential            svccredential.Service
}

// Initialize initializes the code repository daemon, which is used to sync the code repository.
func (f factory) Initialize(digCon *dig.Container) error {
	var p params
	if err := digCon.Invoke(func(deps params) { p = deps }); err != nil {
		return err
	}
	return p.HandlerRegistry.Register(enums.DaemonCodeRepository, workq.Consumer{
		Handler:     crRunner(p),
		Concurrency: 10,
		Timeout:     time.Minute * 10,
	})
}

func crRunner(p params) func(ctx context.Context, payload []byte) error {
	return func(ctx context.Context, payload []byte) error {
		var task api.DaemonCodeRepositoryPayload
		err := json.Unmarshal(payload, &task)
		if err != nil {
			return fmt.Errorf("code repository unmarshal payload failed: %v", err)
		}
		cr := codeRepository{
			userRepository:           p.UserRepository,
			codeRepositoryRepository: p.CodeRepositoryRepository,
			config:                   p.Config,
			svcCredential:            p.SvcCredential,
		}

		status := enums.TaskCommonStatusSuccess
		statusMessage := ""
		err = cr.runner(ctx, task)
		if err != nil {
			status = enums.TaskCommonStatusFailed
			statusMessage = truncateStatusMessage(err.Error())
		}
		userRepository := p.UserRepository
		err = userRepository.UpdateUser3rdParty(ctx, task.User3rdPartyID, map[string]any{
			query.User3rdParty.CrLastUpdateTimestamp.ColumnName().String(): time.Now().UnixMilli(),
			query.User3rdParty.CrLastUpdateStatus.ColumnName().String():    status,
			query.User3rdParty.CrLastUpdateMessage.ColumnName().String():   statusMessage,
		})
		if err != nil {
			return err
		}
		return nil
	}
}

// truncateStatusMessage keeps an error message within the column size.
func truncateStatusMessage(message string) string {
	runes := []rune(message)
	if len(runes) <= maxStatusMessageLength {
		return message
	}
	return string(runes[:maxStatusMessageLength])
}

type codeRepository struct {
	userRepository           repouser.UserRepository
	codeRepositoryRepository repocoderepo.CodeRepositoryRepository
	config                   *config.Configuration
	svcCredential            svccredential.Service
}

// scmCredential is a decrypted SCM credential used to talk to a provider.
type scmCredential struct {
	Type     enums.ScmCredentialType
	Token    string
	Username string
	Password string
	SshKey   string
}

// credential loads and decrypts the SCM credential of a third-party account.
func (cr codeRepository) credential(ctx context.Context, user3rdPartyObj *models.User3rdParty) (scmCredential, error) {
	notConfigured := fmt.Errorf("no SCM credential is configured for %s, add an access token in settings first", user3rdPartyObj.Provider.String())

	credentialObj, err := cr.codeRepositoryRepository.GetCloneCredential(ctx, user3rdPartyObj.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return scmCredential{}, notConfigured
		}
		slog.Error("get scm credential failed", "err", err, "user_3rdparty_id", user3rdPartyObj.ID)
		return scmCredential{}, fmt.Errorf("get SCM credential failed: %w", err)
	}
	if credentialObj.Type == enums.ScmCredentialTypeNone {
		return scmCredential{}, notConfigured
	}

	credential := scmCredential{Type: credentialObj.Type}
	if credentialObj.Username != nil {
		credential.Username = ptr.To(credentialObj.Username)
	}
	for _, field := range []struct {
		name string
		src  *string
		dst  *string
	}{
		{name: "token", src: credentialObj.Token, dst: &credential.Token},
		{name: "password", src: credentialObj.Password, dst: &credential.Password},
		{name: "ssh key", src: credentialObj.SshKey, dst: &credential.SshKey},
	} {
		if field.src == nil || *field.src == "" {
			continue
		}
		plaintext, err := cr.svcCredential.Open(*field.src)
		if err != nil {
			slog.Error("decrypt scm credential failed", "err", err, "user_3rdparty_id", user3rdPartyObj.ID, "field", field.name)
			return scmCredential{}, fmt.Errorf("decrypt SCM credential failed, please configure it again")
		}
		*field.dst = plaintext
	}
	return credential, nil
}

func (cr codeRepository) runner(ctx context.Context, payload api.DaemonCodeRepositoryPayload) error {
	user3rdPartyObj, err := cr.userRepository.GetUser3rdParty(ctx, payload.User3rdPartyID)
	if err != nil {
		slog.Error("get 3rdParty user failed", "err", err)
		return fmt.Errorf("get 3rdParty user failed: %v", err)
	}

	credential, err := cr.credential(ctx, user3rdPartyObj)
	if err != nil {
		return err
	}
	// Listing repositories is an API call: only token credentials can do it, SSH keys
	// and username/password are only usable for cloning.
	if credential.Type != enums.ScmCredentialTypeToken || credential.Token == "" {
		return fmt.Errorf("listing %s repositories requires a token credential, the configured %s credential cannot list them",
			user3rdPartyObj.Provider.String(), credential.Type.String())
	}

	scmProvider, ok := cr.config.ScmProvider(user3rdPartyObj.Provider)
	if !ok {
		return fmt.Errorf("scm provider %s is not configured", user3rdPartyObj.Provider.String())
	}

	switch user3rdPartyObj.Provider {
	case enums.ProviderGithub:
		return cr.github(ctx, user3rdPartyObj, credential)
	case enums.ProviderGitlab:
		return cr.gitlab(ctx, user3rdPartyObj, credential, scmProvider.BaseURL)
	case enums.ProviderGitea:
		return cr.gitea(ctx, user3rdPartyObj, credential, scmProvider.BaseURL)
	}
	return fmt.Errorf("unsupported scm provider %s", user3rdPartyObj.Provider.String())
}
