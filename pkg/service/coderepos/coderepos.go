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

package coderepos

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"go.uber.org/dig"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	repocoderepo "github.com/go-sigma/sigma/pkg/dal/repository/coderepo"
	repouser "github.com/go-sigma/sigma/pkg/dal/repository/user"
	"github.com/go-sigma/sigma/pkg/infra/workq"
	"github.com/go-sigma/sigma/pkg/server/errcode"
	svccredential "github.com/go-sigma/sigma/pkg/service/credential"
	"github.com/go-sigma/sigma/pkg/utils"
)

//go:generate go tool mockgen -mock_names Service=MockCodeRepositoryService -destination=mocks/coderepos_mocks.go -package=mocks github.com/go-sigma/sigma/pkg/service/coderepos Service

// Service encapsulates code repository related business logic.
type Service interface {
	// ListCodeRepositories lists code repositories with pagination (also returns owners for DTO mapping).
	ListCodeRepositories(ctx context.Context, userID string, provider enums.Provider, owner, name *string, pagination api.Pagination, sort api.Sortable) ([]*models.CodeRepository, []*models.CodeRepositoryOwner, int64, error)
	// GetCodeRepository gets a code repository by id (also returns owners for DTO mapping).
	GetCodeRepository(ctx context.Context, userID string, provider enums.Provider, id string) (*models.CodeRepository, []*models.CodeRepositoryOwner, error)
	// ListCodeRepositoryOwners lists code repository owners.
	ListCodeRepositoryOwners(ctx context.Context, userID string, provider enums.Provider, name *string) ([]*models.CodeRepositoryOwner, int64, error)
	// ListCodeRepositoryBranches lists the branches of a code repository, read from the provider.
	ListCodeRepositoryBranches(ctx context.Context, userID string, provider enums.Provider, codeRepositoryID string) ([]string, error)
	// GetCodeRepositoryBranch verifies that a branch exists and returns its name.
	GetCodeRepositoryBranch(ctx context.Context, userID string, provider enums.Provider, codeRepositoryID string, name string) (string, error)
	// ResyncCodeRepositories resync all code repositories for the user's 3rd party (includes: update user status + produce daemon task).
	ResyncCodeRepositories(ctx context.Context, userID string, provider enums.Provider) error
	// ListCodeRepositoryProviders lists the configured source code providers together with
	// the caller's binding, credential and last sync state.
	ListCodeRepositoryProviders(ctx context.Context, userID string) ([]api.ListCodeRepositoryProvidersResponse, error)
	// GetCodeRepositoryUser3rdParty gets the user 3rd party by provider.
	GetCodeRepositoryUser3rdParty(ctx context.Context, userID string, provider enums.Provider) (*models.User3rdParty, error)
	// GetCodeRepositoryCredential gets the SCM credential metadata of a provider.
	GetCodeRepositoryCredential(ctx context.Context, userID string, provider enums.Provider) (*models.CodeRepositoryCloneCredential, error)
	// UpsertCodeRepositoryCredential validates and stores the SCM credential of a provider,
	// then triggers a sync.
	UpsertCodeRepositoryCredential(ctx context.Context, userID string, req api.UpdateCodeRepositoryCredentialRequest) error
	// DeleteCodeRepositoryCredential clears the SCM credential of a provider.
	DeleteCodeRepositoryCredential(ctx context.Context, userID string, provider enums.Provider) error
	// ValidateCodeRepositoryCredential verifies a credential without storing it.
	ValidateCodeRepositoryCredential(ctx context.Context, req api.UpdateCodeRepositoryCredentialRequest) (*api.ValidateCodeRepositoryCredentialResponse, error)
}

type service struct {
	dig.In

	RepoCode      repocoderepo.CodeRepositoryRepository
	RepoUser      repouser.UserRepository
	Producer      workq.Producer
	Config        *config.Configuration
	SvcCredential svccredential.Service
}

func NewService(params service) Service {
	return &params
}

func (s *service) ListCodeRepositories(ctx context.Context, userID string, provider enums.Provider, owner, name *string, pagination api.Pagination, sort api.Sortable) ([]*models.CodeRepository, []*models.CodeRepositoryOwner, int64, error) {
	pagination = utils.NormalizePagination(pagination)
	user3rdPartyObj, err := s.RepoUser.GetUser3rdPartyByProvider(ctx, userID, provider)
	if err != nil {
		slog.Error("get user 3rdParty by provider failed", "err", err, "Provider", provider.String())
		return nil, nil, 0, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Get user 3rdParty by provider failed: %v", err))
	}

	ownerObjs, err := s.RepoCode.ListOwnersAll(ctx, user3rdPartyObj.ID)
	if err != nil {
		slog.Error("list all owners failed", "err", err)
		return nil, nil, 0, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("List all owners failed: %v", err))
	}

	codeRepositoryObjs, total, err := s.RepoCode.ListWithPagination(ctx, userID, provider, owner, name, pagination, sort)
	if err != nil {
		slog.Error("list code repositories failed", "err", err)
		return nil, nil, 0, errcode.HTTPErrCodeInternalError.Detail(err.Error())
	}
	return codeRepositoryObjs, ownerObjs, total, nil
}

func (s *service) GetCodeRepository(ctx context.Context, userID string, provider enums.Provider, id string) (*models.CodeRepository, []*models.CodeRepositoryOwner, error) {
	user3rdPartyObj, err := s.RepoUser.GetUser3rdPartyByProvider(ctx, userID, provider)
	if err != nil {
		slog.Error("get user 3rdParty by provider failed", "err", err, "Provider", provider.String())
		return nil, nil, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Get user 3rdParty by provider failed: %v", err))
	}

	ownerObjs, err := s.RepoCode.ListOwnersAll(ctx, user3rdPartyObj.ID)
	if err != nil {
		slog.Error("list all owners failed", "err", err)
		return nil, nil, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("List all owners failed: %v", err))
	}

	codeRepositoryObj, err := s.RepoCode.Get(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Error("code repository not found", "err", err, "provider", provider.String(), "id", id)
			return nil, nil, errcode.HTTPErrCodeNotFound.Detail(fmt.Sprintf("Code repository(%s) not found: %s", id, err))
		}
		slog.Error("get code repository failed", "err", err, "repositoryID", id, "id", id)
		return nil, nil, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Code repository(%s) not found: %s", id, err))
	}
	if codeRepositoryObj.User3rdParty.Provider != provider {
		slog.Error("code repository not found", "err", err, "provider", provider.String(), "id", id)
		return nil, nil, errcode.HTTPErrCodeNotFound.Detail(fmt.Sprintf("Code repository(%s) not found", id))
	}
	return codeRepositoryObj, ownerObjs, nil
}

func (s *service) ListCodeRepositoryOwners(ctx context.Context, userID string, provider enums.Provider, name *string) ([]*models.CodeRepositoryOwner, int64, error) {
	codeRepositoryOwnerObjs, total, err := s.RepoCode.ListOwnerWithoutPagination(ctx, userID, provider, name)
	if err != nil {
		slog.Error("list code repository owners failed", "err", err)
		return nil, 0, errcode.HTTPErrCodeInternalError.Detail(err.Error())
	}
	return codeRepositoryOwnerObjs, total, nil
}

func (s *service) ResyncCodeRepositories(ctx context.Context, userID string, provider enums.Provider) error {
	user3rdPartyObj, err := s.RepoUser.GetUser3rdPartyByProvider(ctx, userID, provider)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Error("code repository not found", "err", err, "userID", userID, "provider", provider.String())
			return errcode.HTTPErrCodeNotFound.Detail("Code repository not found")
		}
		slog.Error("code repository find failed", "err", err, "userID", userID, "provider", provider.String())
		return errcode.HTTPErrCodeNotFound.Detail("Code repository find failed")
	}
	if user3rdPartyObj.CrLastUpdateStatus == enums.TaskCommonStatusDoing {
		slog.Error("code repository status already is syncing", "provider", provider.String())
		return errcode.HTTPErrCodeConflict.Detail(fmt.Sprintf("Code repository(%s) status already is syncing", provider.String()))
	}
	err = query.Q.Transaction(func(tx *query.Query) error {
		repoUser := repouser.NewUserRepository(tx)
		err = repoUser.UpdateUser3rdParty(ctx, user3rdPartyObj.ID, map[string]any{
			query.User3rdParty.CrLastUpdateTimestamp.ColumnName().String(): time.Now().UnixMilli(),
			query.User3rdParty.CrLastUpdateStatus.ColumnName().String():    enums.TaskCommonStatusDoing,
			query.User3rdParty.CrLastUpdateMessage.ColumnName().String():   "",
		})
		if err != nil {
			return errcode.HTTPErrCodeInternalError.Detail("Update user status failed")
		}
		err = s.Producer.Produce(ctx, enums.DaemonCodeRepository,
			api.DaemonCodeRepositoryPayload{User3rdPartyID: user3rdPartyObj.ID})
		if err != nil {
			slog.Error("publish sync code repository failed", "err", err, "user_id", user3rdPartyObj.UserID)
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

// ListCodeRepositoryProviders lists the configured source code providers together with the
// caller's binding, credential and last sync state.
//
// Providers come from the configuration (not from the user's bindings), so a provider that
// is configured but not linked yet is still listed: the web UI needs it to offer the
// "configure access token" action.
func (s *service) ListCodeRepositoryProviders(ctx context.Context, userID string) ([]api.ListCodeRepositoryProvidersResponse, error) {
	user3rdPartyObjs, err := s.RepoUser.ListUser3rdParty(ctx, userID)
	if err != nil {
		slog.Error("list providers failed", "err", err)
		return nil, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("List providers failed: %v", err))
	}
	bindings := make(map[enums.Provider]*models.User3rdParty, len(user3rdPartyObjs))
	for _, user3rdPartyObj := range user3rdPartyObjs {
		bindings[user3rdPartyObj.Provider] = user3rdPartyObj
	}

	scmProviders := s.Config.ResolveScmProviders()
	resp := make([]api.ListCodeRepositoryProvidersResponse, 0, len(scmProviders))
	for _, scmProvider := range scmProviders {
		provider := enums.Provider(scmProvider.Provider.String())
		item := api.ListCodeRepositoryProvidersResponse{
			Provider:              provider,
			DisplayName:           scmProviderDisplayName(scmProvider.Provider),
			BaseURL:               scmProvider.BaseURL,
			OauthEnabled:          s.oauthEnabled(provider),
			CredentialType:        enums.ScmCredentialTypeNone,
			CrLastUpdateStatus:    enums.TaskCommonStatusPending,
			CrLastUpdateTimestamp: time.Unix(0, 0).UTC().Format(consts.DefaultTimePattern),
		}

		user3rdPartyObj, ok := bindings[provider]
		if !ok {
			resp = append(resp, item)
			continue
		}
		item.Linked = true
		item.AccountID = user3rdPartyObj.AccountID
		item.CrLastUpdateTimestamp = time.UnixMilli(user3rdPartyObj.CrLastUpdateTimestamp).UTC().Format(consts.DefaultTimePattern)
		item.CrLastUpdateStatus = user3rdPartyObj.CrLastUpdateStatus
		item.CrLastUpdateMessage = user3rdPartyObj.CrLastUpdateMessage

		credentialObj, err := s.RepoCode.GetCloneCredential(ctx, user3rdPartyObj.ID)
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				slog.Error("get code repository credential failed", "err", err, "user_3rdparty_id", user3rdPartyObj.ID)
				return nil, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Get code repository credential failed: %v", err))
			}
			resp = append(resp, item)
			continue
		}
		if credentialObj.Type != enums.ScmCredentialTypeNone {
			item.CredentialType = credentialObj.Type
			item.HasCredential = true
		}
		resp = append(resp, item)
	}
	return resp, nil
}

// oauthEnabled reports whether the provider can be used to sign in.
func (s *service) oauthEnabled(provider enums.Provider) bool {
	switch provider {
	case enums.ProviderGithub:
		return s.Config.Auth.Oauth2.Github.Enabled
	case enums.ProviderGitlab:
		return s.Config.Auth.Oauth2.Gitlab.Enabled
	case enums.ProviderGitea:
		return s.Config.Auth.Oauth2.Gitea.Enabled
	}
	return false
}

// scmProviderDisplayName returns the human readable name of a source code provider.
func scmProviderDisplayName(provider enums.ScmProvider) string {
	switch provider {
	case enums.ScmProviderGithub:
		return "GitHub"
	case enums.ScmProviderGitlab:
		return "GitLab"
	case enums.ScmProviderGitea:
		return "Gitea"
	}
	return provider.String()
}

func (s *service) GetCodeRepositoryUser3rdParty(ctx context.Context, userID string, provider enums.Provider) (*models.User3rdParty, error) {
	user3rdPartyObj, err := s.RepoUser.GetUser3rdPartyByProvider(ctx, userID, provider)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Error("code repository not found", "err", err, "userID", userID, "provider", provider.String())
			return nil, errcode.HTTPErrCodeNotFound.Detail("Code repository not found")
		}
		slog.Error("code repository find failed", "err", err, "userID", userID, "provider", provider.String())
		return nil, errcode.HTTPErrCodeNotFound.Detail("Code repository find failed")
	}
	return user3rdPartyObj, nil
}
