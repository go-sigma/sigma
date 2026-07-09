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
	"encoding/json"
	"fmt"
	"log/slog"
	"path"
	"reflect"
	"time"

	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/background/daemon"
	"github.com/go-sigma/sigma/pkg/dal/query"
	repocoderepo "github.com/go-sigma/sigma/pkg/dal/repository/coderepo"
	repouser "github.com/go-sigma/sigma/pkg/dal/repository/user"
	"github.com/go-sigma/sigma/pkg/infra/workq"
	"github.com/go-sigma/sigma/pkg/utils"
)

const (
	perPage = 100
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
		}

		status := enums.TaskCommonStatusSuccess
		statusMessage := ""
		err = cr.runner(ctx, task)
		if err != nil {
			status = enums.TaskCommonStatusFailed
			statusMessage = err.Error()
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

type codeRepository struct {
	userRepository           repouser.UserRepository
	codeRepositoryRepository repocoderepo.CodeRepositoryRepository
}

func (cr codeRepository) runner(ctx context.Context, payload api.DaemonCodeRepositoryPayload) error {
	userRepository := cr.userRepository
	// TODO: fix get user 3rdparty
	user3rdPartyObj, err := userRepository.GetUser3rdParty(ctx, payload.User3rdPartyID)
	if err != nil {
		slog.Error("get 3rdParty user failed", "err", err)
		return fmt.Errorf("get 3rdParty user failed: %v", err)
	}
	switch user3rdPartyObj.Provider {
	case enums.ProviderGithub:
		return cr.github(ctx, user3rdPartyObj)
	case enums.ProviderGitlab:
		return cr.gitlab(ctx, user3rdPartyObj)
	case enums.ProviderGitea:
		return cr.gitea(ctx, user3rdPartyObj)
	}
	return nil
}
