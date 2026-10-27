// Copyright 2026 sigma
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
	"strconv"
	"time"

	"code.gitea.io/sdk/gitea"
	"github.com/google/go-github/v74/github"
	gitlab "gitlab.com/gitlab-org/api/client-go"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/server/errcode"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
	"github.com/go-sigma/sigma/pkg/utils/uuid"
)

// scmAccount is the provider account a credential resolves to.
type scmAccount struct {
	ID       string
	Username string
}

// GetCodeRepositoryCredential gets the SCM credential metadata of a provider.
//
// Secrets are not decrypted here: the response only carries the credential type, a masked
// token and "has" flags, so that the API never echoes a secret back.
func (s *service) GetCodeRepositoryCredential(ctx context.Context, userID string, provider enums.Provider) (*models.CodeRepositoryCloneCredential, error) {
	user3rdPartyObj, err := s.RepoUser.GetUser3rdPartyByProvider(ctx, userID, provider)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.HTTPErrCodeNotFound.Detail(fmt.Sprintf("No %s account is linked", provider.String()))
		}
		slog.Error("get user 3rdparty by provider failed", "err", err, "provider", provider.String())
		return nil, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Get user 3rdparty by provider failed: %v", err))
	}
	credentialObj, err := s.RepoCode.GetCloneCredential(ctx, user3rdPartyObj.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.HTTPErrCodeNotFound.Detail(fmt.Sprintf("No credential is configured for %s", provider.String()))
		}
		slog.Error("get code repository credential failed", "err", err, "provider", provider.String())
		return nil, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Get code repository credential failed: %v", err))
	}
	return credentialObj, nil
}

// UpsertCodeRepositoryCredential validates and stores the SCM credential of a provider, then triggers a sync.
func (s *service) UpsertCodeRepositoryCredential(ctx context.Context, userID string, req api.UpdateCodeRepositoryCredentialRequest) error {
	scmProvider, err := s.mustScmProvider(req.Provider)
	if err != nil {
		return err
	}
	if err = validateCredentialRequest(req); err != nil {
		return err
	}

	// A token is the only credential type that can list repositories, so it is the only
	// one verified against the provider and used to discover the linked account.
	var account *scmAccount
	if req.Type == enums.ScmCredentialTypeToken {
		account, err = fetchScmAccount(ctx, req.Provider, scmProvider.BaseURL, ptr.To(req.Token))
		if err != nil {
			return err
		}
	}

	user3rdPartyObj, err := s.ensureUser3rdParty(ctx, userID, req.Provider, account)
	if err != nil {
		return err
	}

	credentialObj := &models.CodeRepositoryCloneCredential{
		ID:             uuid.NewV7String(),
		User3rdPartyID: user3rdPartyObj.ID,
		Type:           req.Type,
		Username:       req.Username,
	}
	for _, field := range []struct {
		name  string
		plain *string
		dst   **string
	}{
		{name: "token", plain: req.Token, dst: &credentialObj.Token},
		{name: "password", plain: req.Password, dst: &credentialObj.Password},
		{name: "ssh key", plain: req.SshKey, dst: &credentialObj.SshKey},
	} {
		if field.plain == nil || *field.plain == "" {
			continue
		}
		sealed, err := s.SvcCredential.Seal(*field.plain)
		if err != nil {
			slog.Error("seal code repository credential failed", "err", err, "field", field.name)
			return errcode.HTTPErrCodeInternalError.Detail("Seal code repository credential failed")
		}
		*field.dst = &sealed
	}

	err = s.RepoCode.UpsertCredential(ctx, credentialObj)
	if err != nil {
		slog.Error("upsert code repository credential failed", "err", err, "provider", req.Provider.String())
		return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Upsert code repository credential failed: %v", err))
	}

	// Only token credentials can list repositories; SSH and username/password are clone
	// credentials, so syncing them would only produce a failure message.
	if req.Type == enums.ScmCredentialTypeToken {
		s.publishSync(ctx, user3rdPartyObj.ID)
	}
	return nil
}

// DeleteCodeRepositoryCredential clears the SCM credential of a provider.
func (s *service) DeleteCodeRepositoryCredential(ctx context.Context, userID string, provider enums.Provider) error {
	if _, err := s.mustScmProvider(provider); err != nil {
		return err
	}
	user3rdPartyObj, err := s.RepoUser.GetUser3rdPartyByProvider(ctx, userID, provider)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.HTTPErrCodeNotFound.Detail(fmt.Sprintf("No %s account is linked", provider.String()))
		}
		slog.Error("get user 3rdparty by provider failed", "err", err, "provider", provider.String())
		return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Get user 3rdparty by provider failed: %v", err))
	}
	err = s.RepoCode.DeleteCredential(ctx, user3rdPartyObj.ID)
	if err != nil {
		slog.Error("delete code repository credential failed", "err", err, "provider", provider.String())
		return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Delete code repository credential failed: %v", err))
	}
	return nil
}

// ValidateCodeRepositoryCredential verifies a credential without storing it.
//
// Only token credentials can be verified: SSH keys and username/password credentials are
// used by the builder when cloning and are never sent to a provider API here.
func (s *service) ValidateCodeRepositoryCredential(ctx context.Context, req api.UpdateCodeRepositoryCredentialRequest) (*api.ValidateCodeRepositoryCredentialResponse, error) {
	scmProvider, err := s.mustScmProvider(req.Provider)
	if err != nil {
		return nil, err
	}
	if err = validateCredentialRequest(req); err != nil {
		return nil, err
	}
	resp := &api.ValidateCodeRepositoryCredentialResponse{
		Provider: req.Provider,
		Type:     req.Type,
	}
	if req.Type != enums.ScmCredentialTypeToken {
		return resp, nil
	}
	account, err := fetchScmAccount(ctx, req.Provider, scmProvider.BaseURL, ptr.To(req.Token))
	if err != nil {
		return nil, err
	}
	resp.AccountID = account.ID
	resp.Username = account.Username
	resp.Validated = true
	return resp, nil
}

// mustScmProvider returns the configuration of a provider, rejecting providers that the
// administrator did not enable.
func (s *service) mustScmProvider(provider enums.Provider) (config.ConfigurationScmProvider, error) {
	scmProvider, ok := s.Config.ScmProvider(provider)
	if !ok {
		return config.ConfigurationScmProvider{}, errcode.HTTPErrCodeBadRequest.Detail(fmt.Sprintf("scm provider %s is not configured", provider.String()))
	}
	return scmProvider, nil
}

// Maximum sizes of the credential fields. They are enforced here rather than with the
// max validator tag, which panics on a nil pointer field when combined with required_if.
const (
	maxTokenLength    = 1024
	maxSshKeyLength   = 8192
	maxUsernameLength = 256
	maxPasswordLength = 512
)

// validateCredentialRequest checks that the required fields of a credential type are
// present and within their size limits.
func validateCredentialRequest(req api.UpdateCodeRepositoryCredentialRequest) error {
	switch req.Type {
	case enums.ScmCredentialTypeToken:
		if err := checkSecret("token", req.Token, maxTokenLength); err != nil {
			return err
		}
	case enums.ScmCredentialTypeSsh:
		if err := checkSecret("ssh_key", req.SshKey, maxSshKeyLength); err != nil {
			return err
		}
	case enums.ScmCredentialTypeUsername:
		if err := checkSecret("username", req.Username, maxUsernameLength); err != nil {
			return err
		}
		if err := checkSecret("password", req.Password, maxPasswordLength); err != nil {
			return err
		}
	case enums.ScmCredentialTypeNone:
		return errcode.HTTPErrCodeBadRequest.Detail("type none is not a credential, delete the credential instead")
	default:
		return errcode.HTTPErrCodeBadRequest.Detail(fmt.Sprintf("invalid credential type %s", req.Type.String()))
	}
	return nil
}

// checkSecret reports whether a credential field is present and not oversized.
func checkSecret(name string, value *string, maxLength int) error {
	if value == nil || *value == "" {
		return errcode.HTTPErrCodeBadRequest.Detail(fmt.Sprintf("%s is required", name))
	}
	if len(*value) > maxLength {
		return errcode.HTTPErrCodeBadRequest.Detail(fmt.Sprintf("%s must be at most %d bytes", name, maxLength))
	}
	return nil
}

// ensureUser3rdParty returns the binding of the provider for the user, creating it from the
// account resolved by the credential when the user never signed in with OAuth.
func (s *service) ensureUser3rdParty(ctx context.Context, userID string, provider enums.Provider, account *scmAccount) (*models.User3rdParty, error) {
	user3rdPartyObj, err := s.RepoUser.GetUser3rdPartyByProvider(ctx, userID, provider)
	if err == nil {
		return user3rdPartyObj, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		slog.Error("get user 3rdparty by provider failed", "err", err, "provider", provider.String())
		return nil, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Get user 3rdparty by provider failed: %v", err))
	}

	if account == nil || account.ID == "" {
		return nil, errcode.HTTPErrCodeBadRequest.Detail(fmt.Sprintf(
			"No %s account is linked yet, use a token credential to link it", provider.String()))
	}

	if existing, err := s.RepoUser.GetUser3rdPartyByAccountID(ctx, provider, account.ID); err == nil {
		if existing.UserID != userID {
			slog.Error("account already bound to another user", "err", err, "provider", provider.String())
			return nil, errcode.HTTPErrCodeConflict.Detail("This account is already bound to another user")
		}
		return existing, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		slog.Error("get user 3rdparty by account id failed", "err", err, "provider", provider.String())
		return nil, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Get user 3rdparty by account id failed: %v", err))
	}

	created, err := s.RepoUser.UpsertUser3rdPartyByAccountID(ctx, &models.User3rdParty{
		ID:        uuid.NewV7String(),
		UserID:    userID,
		Provider:  provider,
		AccountID: &account.ID,
	})
	if err != nil {
		slog.Error("create user 3rdparty failed", "err", err, "provider", provider.String())
		return nil, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Create user 3rdparty failed: %v", err))
	}
	if created.UserID != userID {
		return nil, errcode.HTTPErrCodeConflict.Detail("This account is already bound to another user")
	}
	return created, nil
}

// publishSync marks the binding as syncing and queues a sync task.
func (s *service) publishSync(ctx context.Context, user3rdPartyID string) {
	err := s.RepoUser.UpdateUser3rdParty(ctx, user3rdPartyID, map[string]any{
		query.User3rdParty.CrLastUpdateTimestamp.ColumnName().String(): time.Now().UnixMilli(),
		query.User3rdParty.CrLastUpdateStatus.ColumnName().String():    enums.TaskCommonStatusDoing,
		query.User3rdParty.CrLastUpdateMessage.ColumnName().String():   "",
	})
	if err != nil {
		slog.Error("update user 3rdparty sync status failed", "err", err, "user_3rdparty_id", user3rdPartyID)
	}
	err = s.Producer.Produce(ctx, enums.DaemonCodeRepository,
		api.DaemonCodeRepositoryPayload{User3rdPartyID: user3rdPartyID})
	if err != nil {
		slog.Error("publish sync code repository failed", "err", err, "user_3rdparty_id", user3rdPartyID)
	}
}

// fetchScmAccount resolves the account a token belongs to.
func fetchScmAccount(ctx context.Context, provider enums.Provider, baseURL, token string) (*scmAccount, error) {
	invalid := func(err error) error {
		slog.Error("validate scm credential failed", "err", err, "provider", provider.String())
		return errcode.HTTPErrCodeBadRequest.Detail(fmt.Sprintf("Credential is invalid: %v", err))
	}
	switch provider {
	case enums.ProviderGithub:
		userObj, _, err := github.NewTokenClient(ctx, token).Users.Get(ctx, "")
		if err != nil {
			return nil, invalid(err)
		}
		return &scmAccount{ID: strconv.FormatInt(userObj.GetID(), 10), Username: userObj.GetLogin()}, nil
	case enums.ProviderGitlab:
		client, err := gitlab.NewClient(token, gitlab.WithBaseURL(baseURL))
		if err != nil {
			return nil, invalid(err)
		}
		userObj, _, err := client.Users.CurrentUser()
		if err != nil {
			return nil, invalid(err)
		}
		return &scmAccount{ID: strconv.FormatInt(userObj.ID, 10), Username: userObj.Username}, nil
	case enums.ProviderGitea:
		client, err := gitea.NewClient(baseURL, gitea.SetToken(token))
		if err != nil {
			return nil, invalid(err)
		}
		userObj, _, err := client.GetMyUserInfo()
		if err != nil {
			return nil, invalid(err)
		}
		return &scmAccount{ID: strconv.FormatInt(userObj.ID, 10), Username: userObj.UserName}, nil
	}
	return nil, errcode.HTTPErrCodeBadRequest.Detail(fmt.Sprintf("unsupported scm provider %s", provider.String()))
}
