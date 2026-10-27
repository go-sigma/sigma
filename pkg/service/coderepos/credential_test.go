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
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/dal/models"
	repocoderepomocks "github.com/go-sigma/sigma/pkg/dal/repository/coderepo/mocks"
	repousermocks "github.com/go-sigma/sigma/pkg/dal/repository/user/mocks"
	credentialmocks "github.com/go-sigma/sigma/pkg/service/credential/mocks"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

func scmConfig() *config.Configuration {
	return &config.Configuration{
		Scm: config.ConfigurationScm{Providers: []config.ConfigurationScmProvider{
			{Provider: enums.ScmProviderGithub, BaseURL: "https://github.com"},
		}},
	}
}

func TestGetCodeRepositoryCredentialErrors(t *testing.T) {
	ctrl := gomock.NewController(t)

	t.Run("account is not linked", func(t *testing.T) {
		userRepository := repousermocks.NewMockUserRepository(ctrl)
		userRepository.EXPECT().
			GetUser3rdPartyByProvider(gomock.Any(), "user-1", enums.ProviderGithub).
			Return(nil, gorm.ErrRecordNotFound)

		_, err := (&service{Config: scmConfig(), RepoUser: userRepository}).
			GetCodeRepositoryCredential(t.Context(), "user-1", enums.ProviderGithub)
		require.Error(t, err)
	})

	t.Run("credential is not configured", func(t *testing.T) {
		userRepository := repousermocks.NewMockUserRepository(ctrl)
		userRepository.EXPECT().
			GetUser3rdPartyByProvider(gomock.Any(), "user-1", enums.ProviderGithub).
			Return(&models.User3rdParty{ID: "3rdparty-1"}, nil)
		codeRepository := repocoderepomocks.NewMockCodeRepositoryRepository(ctrl)
		codeRepository.EXPECT().
			GetCloneCredential(gomock.Any(), "3rdparty-1").
			Return(nil, gorm.ErrRecordNotFound)

		_, err := (&service{Config: scmConfig(), RepoUser: userRepository, RepoCode: codeRepository}).
			GetCodeRepositoryCredential(t.Context(), "user-1", enums.ProviderGithub)
		require.Error(t, err)
	})
}

func TestValidateCodeRepositoryCredential(t *testing.T) {
	t.Run("provider is not configured", func(t *testing.T) {
		svc := &service{Config: &config.Configuration{}}
		_, err := svc.ValidateCodeRepositoryCredential(t.Context(), api.UpdateCodeRepositoryCredentialRequest{
			Provider: enums.ProviderGithub,
			Type:     enums.ScmCredentialTypeToken,
			Token:    new("ghp_sigma"),
		})
		require.Error(t, err)
	})

	t.Run("missing field for the credential type", func(t *testing.T) {
		svc := &service{Config: scmConfig()}
		for _, req := range []api.UpdateCodeRepositoryCredentialRequest{
			{Provider: enums.ProviderGithub, Type: enums.ScmCredentialTypeToken},
			{Provider: enums.ProviderGithub, Type: enums.ScmCredentialTypeSsh},
			{Provider: enums.ProviderGithub, Type: enums.ScmCredentialTypeUsername, Username: new("sigma")},
			{Provider: enums.ProviderGithub, Type: enums.ScmCredentialTypeNone},
		} {
			_, err := svc.ValidateCodeRepositoryCredential(t.Context(), req)
			require.Error(t, err, "type %s must be rejected", req.Type.String())
		}
	})

	t.Run("ssh credential cannot be verified", func(t *testing.T) {
		svc := &service{Config: scmConfig()}
		resp, err := svc.ValidateCodeRepositoryCredential(t.Context(), api.UpdateCodeRepositoryCredentialRequest{
			Provider: enums.ProviderGithub,
			Type:     enums.ScmCredentialTypeSsh,
			SshKey:   new("-----BEGIN OPENSSH PRIVATE KEY-----"),
		})
		require.NoError(t, err)
		require.False(t, resp.Validated)
		require.Empty(t, resp.AccountID)
	})
}

func TestUpsertCodeRepositoryCredentialSealsSecret(t *testing.T) {
	ctrl := gomock.NewController(t)
	userRepository := repousermocks.NewMockUserRepository(ctrl)
	codeRepository := repocoderepomocks.NewMockCodeRepositoryRepository(ctrl)
	svcCredential := credentialmocks.NewMockCredentialService(ctrl)

	userRepository.EXPECT().
		GetUser3rdPartyByProvider(gomock.Any(), "user-1", enums.ProviderGithub).
		Return(&models.User3rdParty{ID: "3rdparty-1", UserID: "user-1", Provider: enums.ProviderGithub}, nil)
	svcCredential.EXPECT().
		Seal("-----BEGIN OPENSSH PRIVATE KEY-----").
		Return("sealed-ssh-key", nil)
	codeRepository.EXPECT().
		UpsertCredential(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ any, credential *models.CodeRepositoryCloneCredential) error {
			require.Equal(t, "3rdparty-1", credential.User3rdPartyID)
			require.Equal(t, enums.ScmCredentialTypeSsh, credential.Type)
			require.Equal(t, "sealed-ssh-key", ptr.To(credential.SshKey))
			require.Nil(t, credential.Token)
			return nil
		})

	// An SSH credential cannot list repositories, so no sync task is queued here.
	svc := &service{Config: scmConfig(), RepoUser: userRepository, RepoCode: codeRepository, SvcCredential: svcCredential}
	require.NoError(t, svc.UpsertCodeRepositoryCredential(t.Context(), "user-1", api.UpdateCodeRepositoryCredentialRequest{
		Provider: enums.ProviderGithub,
		Type:     enums.ScmCredentialTypeSsh,
		SshKey:   new("-----BEGIN OPENSSH PRIVATE KEY-----"),
	}))
}

func TestEnsureUser3rdPartyRejectsForeignAccount(t *testing.T) {
	ctrl := gomock.NewController(t)
	userRepository := repousermocks.NewMockUserRepository(ctrl)

	userRepository.EXPECT().
		GetUser3rdPartyByProvider(gomock.Any(), "user-1", enums.ProviderGithub).
		Return(nil, gorm.ErrRecordNotFound)
	userRepository.EXPECT().
		GetUser3rdPartyByAccountID(gomock.Any(), enums.ProviderGithub, "account-1").
		Return(&models.User3rdParty{ID: "3rdparty-1", UserID: "another-user"}, nil)

	_, err := (&service{Config: scmConfig(), RepoUser: userRepository}).
		ensureUser3rdParty(t.Context(), "user-1", enums.ProviderGithub, &scmAccount{ID: "account-1"})
	require.Error(t, err)
}

func TestEnsureUser3rdPartyRequiresAccountForNewBinding(t *testing.T) {
	ctrl := gomock.NewController(t)
	userRepository := repousermocks.NewMockUserRepository(ctrl)
	userRepository.EXPECT().
		GetUser3rdPartyByProvider(gomock.Any(), "user-1", enums.ProviderGithub).
		Return(nil, gorm.ErrRecordNotFound)

	_, err := (&service{Config: scmConfig(), RepoUser: userRepository}).
		ensureUser3rdParty(t.Context(), "user-1", enums.ProviderGithub, nil)
	require.Error(t, err)
}

func TestEnsureUser3rdPartyCreatesBinding(t *testing.T) {
	ctrl := gomock.NewController(t)
	userRepository := repousermocks.NewMockUserRepository(ctrl)

	userRepository.EXPECT().
		GetUser3rdPartyByProvider(gomock.Any(), "user-1", enums.ProviderGithub).
		Return(nil, gorm.ErrRecordNotFound)
	userRepository.EXPECT().
		GetUser3rdPartyByAccountID(gomock.Any(), enums.ProviderGithub, "account-1").
		Return(nil, gorm.ErrRecordNotFound)
	userRepository.EXPECT().
		UpsertUser3rdPartyByAccountID(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ any, user3rdParty *models.User3rdParty) (*models.User3rdParty, error) {
			require.Equal(t, "user-1", user3rdParty.UserID)
			require.Equal(t, "account-1", ptr.To(user3rdParty.AccountID))
			return user3rdParty, nil
		})

	got, err := (&service{Config: scmConfig(), RepoUser: userRepository}).
		ensureUser3rdParty(t.Context(), "user-1", enums.ProviderGithub, &scmAccount{ID: "account-1"})
	require.NoError(t, err)
	require.Equal(t, "user-1", got.UserID)
}

func TestDeleteCodeRepositoryCredential(t *testing.T) {
	ctrl := gomock.NewController(t)
	userRepository := repousermocks.NewMockUserRepository(ctrl)
	codeRepository := repocoderepomocks.NewMockCodeRepositoryRepository(ctrl)

	svc := &service{Config: scmConfig(), RepoUser: userRepository, RepoCode: codeRepository}

	// A provider that is not configured is rejected before touching the database.
	require.Error(t, svc.DeleteCodeRepositoryCredential(t.Context(), "user-1", enums.ProviderGitlab))

	userRepository.EXPECT().
		GetUser3rdPartyByProvider(gomock.Any(), "user-1", enums.ProviderGithub).
		Return(&models.User3rdParty{ID: "3rdparty-1"}, nil)
	codeRepository.EXPECT().
		DeleteCredential(gomock.Any(), "3rdparty-1").
		Return(nil)

	require.NoError(t, svc.DeleteCodeRepositoryCredential(t.Context(), "user-1", enums.ProviderGithub))
}

func TestListCodeRepositoryProviders(t *testing.T) {
	ctrl := gomock.NewController(t)
	userRepository := repousermocks.NewMockUserRepository(ctrl)
	codeRepository := repocoderepomocks.NewMockCodeRepositoryRepository(ctrl)

	cfg := &config.Configuration{
		Auth: config.ConfigurationAuth{Oauth2: config.ConfigurationAuthOauth2{
			Github: config.ConfigurationAuthOauth2Github{Enabled: true},
		}},
		Scm: config.ConfigurationScm{Providers: []config.ConfigurationScmProvider{
			{Provider: enums.ScmProviderGithub, BaseURL: "https://github.com"},
			{Provider: enums.ScmProviderGitea, BaseURL: "https://gitea.example.com"},
		}},
	}

	userRepository.EXPECT().
		ListUser3rdParty(gomock.Any(), "user-1").
		Return([]*models.User3rdParty{{
			ID:                 "3rdparty-1",
			UserID:             "user-1",
			Provider:           enums.ProviderGithub,
			AccountID:          new("account-1"),
			CrLastUpdateStatus: enums.TaskCommonStatusFailed,
		}}, nil)
	codeRepository.EXPECT().
		GetCloneCredential(gomock.Any(), "3rdparty-1").
		Return(&models.CodeRepositoryCloneCredential{Type: enums.ScmCredentialTypeToken}, nil)

	providers, err := (&service{Config: cfg, RepoUser: userRepository, RepoCode: codeRepository}).
		ListCodeRepositoryProviders(t.Context(), "user-1")
	require.NoError(t, err)
	require.Len(t, providers, 2)

	// GitHub is linked and has a credential.
	require.Equal(t, enums.ProviderGithub, providers[0].Provider)
	require.Equal(t, "GitHub", providers[0].DisplayName)
	require.True(t, providers[0].OauthEnabled)
	require.True(t, providers[0].Linked)
	require.True(t, providers[0].HasCredential)
	require.Equal(t, enums.ScmCredentialTypeToken, providers[0].CredentialType)
	require.Equal(t, enums.TaskCommonStatusFailed, providers[0].CrLastUpdateStatus)

	// Gitea is configured but neither linked nor configured, and it is not an OAuth provider.
	require.Equal(t, enums.ProviderGitea, providers[1].Provider)
	require.False(t, providers[1].OauthEnabled)
	require.False(t, providers[1].Linked)
	require.False(t, providers[1].HasCredential)
	require.Equal(t, enums.ScmCredentialTypeNone, providers[1].CredentialType)
}
