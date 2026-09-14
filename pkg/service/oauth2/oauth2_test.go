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

package oauth2

import (
	"encoding/base64"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/dal/models"
	repouser "github.com/go-sigma/sigma/pkg/dal/repository/user"
	"github.com/go-sigma/sigma/pkg/server/errcode"
	"github.com/go-sigma/sigma/pkg/service/password"
	"github.com/go-sigma/sigma/pkg/service/token"
)

func TestGetClientID(t *testing.T) {
	service := &service{Config: &config.Configuration{
		Auth: config.ConfigurationAuth{
			Oauth2: config.ConfigurationAuthOauth2{
				Github: config.ConfigurationAuthOauth2Github{ClientID: "github-client"},
				Gitlab: config.ConfigurationAuthOauth2Gitlab{ClientID: "gitlab-client"},
				Gitea:  config.ConfigurationAuthOauth2Gitea{ClientID: "gitea-client"},
			},
		},
	}}

	tests := []struct {
		name     string
		provider enums.Provider
		want     string
		wantErr  bool
	}{
		{name: "github", provider: enums.ProviderGithub, want: "github-client"},
		{name: "gitlab", provider: enums.ProviderGitlab, want: "gitlab-client"},
		{name: "gitea", provider: enums.ProviderGitea, want: "gitea-client"},
		{name: "invalid", provider: enums.Provider("invalid"), wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := service.GetClientID(t.Context(), tt.provider)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func requireErrCode(t *testing.T, err error, want string) {
	t.Helper()
	code, ok := errcode.AsType[errcode.ErrCode](err)
	require.True(t, ok)
	require.Equal(t, want, code.Code)
}

var (
	hashedPassword = "hashed"
	sigmaBasicAuth = "Basic " + base64.StdEncoding.EncodeToString([]byte("sigma:password"))
)

func TestTryGetUser(t *testing.T) {
	ctx := t.Context()

	tests := []struct {
		name          string
		authorization string
		setup         func(*gomock.Controller) *service
		wantID        string
		wantErrCode   string
	}{
		{
			name:          "no_authorization",
			authorization: "",
			setup: func(ctrl *gomock.Controller) *service {
				return &service{}
			},
		},
		{
			name:          "basic_success",
			authorization: sigmaBasicAuth,
			setup: func(ctrl *gomock.Controller) *service {
				userRepo := repouser.NewMockUserRepository(ctrl)
				passwordSvc := password.NewMockService(ctrl)
				user := &models.User{ID: "user-1", Username: "sigma", Password: &hashedPassword}
				userRepo.EXPECT().GetByUsername(gomock.Any(), "sigma").Return(user, nil)
				passwordSvc.EXPECT().Verify("password", "hashed").Return(true)
				userRepo.EXPECT().Get(gomock.Any(), "user-1").Return(user, nil)
				return &service{RepoUser: userRepo, SvcPassword: passwordSvc}
			},
			wantID: "user-1",
		},
		{
			name:          "basic_bad_header",
			authorization: "Basic not-base64",
			setup: func(ctrl *gomock.Controller) *service {
				return &service{}
			},
		},
		{
			name:          "basic_get_username_error",
			authorization: sigmaBasicAuth,
			setup: func(ctrl *gomock.Controller) *service {
				userRepo := repouser.NewMockUserRepository(ctrl)
				userRepo.EXPECT().GetByUsername(gomock.Any(), "sigma").Return(nil, errors.New("db error"))
				return &service{RepoUser: userRepo}
			},
			wantErrCode: errcode.HTTPErrCodeInternalError.Code,
		},
		{
			name:          "basic_verify_failed",
			authorization: sigmaBasicAuth,
			setup: func(ctrl *gomock.Controller) *service {
				userRepo := repouser.NewMockUserRepository(ctrl)
				passwordSvc := password.NewMockService(ctrl)
				user := &models.User{ID: "user-1", Password: &hashedPassword}
				userRepo.EXPECT().GetByUsername(gomock.Any(), "sigma").Return(user, nil)
				passwordSvc.EXPECT().Verify("password", "hashed").Return(false)
				return &service{RepoUser: userRepo, SvcPassword: passwordSvc}
			},
			wantErrCode: errcode.HTTPErrCodeUnauthorized.Code,
		},
		{
			name:          "basic_get_user_error",
			authorization: sigmaBasicAuth,
			setup: func(ctrl *gomock.Controller) *service {
				userRepo := repouser.NewMockUserRepository(ctrl)
				passwordSvc := password.NewMockService(ctrl)
				user := &models.User{ID: "user-1", Password: &hashedPassword}
				userRepo.EXPECT().GetByUsername(gomock.Any(), "sigma").Return(user, nil)
				passwordSvc.EXPECT().Verify("password", "hashed").Return(true)
				userRepo.EXPECT().Get(gomock.Any(), "user-1").Return(nil, errors.New("db error"))
				return &service{RepoUser: userRepo, SvcPassword: passwordSvc}
			},
			wantErrCode: errcode.HTTPErrCodeInternalError.Code,
		},
		{
			name:          "bearer_success",
			authorization: "Bearer token-123",
			setup: func(ctrl *gomock.Controller) *service {
				userRepo := repouser.NewMockUserRepository(ctrl)
				tokenSvc := token.NewMockService(ctrl)
				user := &models.User{ID: "user-1", Username: "sigma"}
				tokenSvc.EXPECT().Validate(gomock.Any(), "token-123").Return("jti", "user-1", nil)
				userRepo.EXPECT().Get(gomock.Any(), "user-1").Return(user, nil)
				return &service{RepoUser: userRepo, SvcToken: tokenSvc}
			},
			wantID: "user-1",
		},
		{
			name:          "bearer_validate_error",
			authorization: "Bearer token-123",
			setup: func(ctrl *gomock.Controller) *service {
				tokenSvc := token.NewMockService(ctrl)
				tokenSvc.EXPECT().Validate(gomock.Any(), "token-123").Return("", "", errors.New("invalid token"))
				return &service{SvcToken: tokenSvc}
			},
			wantErrCode: errcode.HTTPErrCodeUnauthorized.Code,
		},
		{
			name:          "bearer_get_user_error",
			authorization: "Bearer token-123",
			setup: func(ctrl *gomock.Controller) *service {
				userRepo := repouser.NewMockUserRepository(ctrl)
				tokenSvc := token.NewMockService(ctrl)
				tokenSvc.EXPECT().Validate(gomock.Any(), "token-123").Return("jti", "user-1", nil)
				userRepo.EXPECT().Get(gomock.Any(), "user-1").Return(nil, errors.New("db error"))
				return &service{RepoUser: userRepo, SvcToken: tokenSvc}
			},
			wantErrCode: errcode.HTTPErrCodeInternalError.Code,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			got, err := tt.setup(ctrl).tryGetUser(ctx, tt.authorization)
			if tt.wantErrCode != "" {
				requireErrCode(t, err, tt.wantErrCode)
				return
			}
			require.NoError(t, err)
			if tt.wantID == "" {
				require.Nil(t, got)
			} else {
				require.Equal(t, tt.wantID, got.ID)
			}
		})
	}
}

func TestCallbackRejectsUserLookupFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	userRepo := repouser.NewMockUserRepository(ctrl)
	userRepo.EXPECT().GetByUsername(gomock.Any(), "sigma").Return(nil, errors.New("db error"))
	svc := &service{RepoUser: userRepo}

	_, _, _, _, _, err := svc.Callback(t.Context(), enums.ProviderGithub, "code", "endpoint", sigmaBasicAuth)
	requireErrCode(t, err, errcode.HTTPErrCodeInternalError.Code)
}
