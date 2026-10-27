// Copyright 2024 sigma
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
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/google/go-github/v74/github"
	gitlab "gitlab.com/gitlab-org/api/client-go"
	"go.uber.org/dig"
	"golang.org/x/oauth2"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	repouser "github.com/go-sigma/sigma/pkg/dal/repository/user"
	"github.com/go-sigma/sigma/pkg/server/errcode"
	"github.com/go-sigma/sigma/pkg/service/password"
	"github.com/go-sigma/sigma/pkg/service/token"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
	"github.com/go-sigma/sigma/pkg/utils/uuid"
)

//go:generate go tool mockgen -mock_names Service=MockOAuth2Service -destination=mocks/oauth2_mocks.go -package=mocks github.com/go-sigma/sigma/pkg/service/oauth2 Service

// Service encapsulates oauth2-related business logic.
type Service interface {
	// Callback handles the oauth2 callback flow: token exchange, user lookup/creation, 3rd party binding, JWT issuance.
	// Returns: user ID, username, email, accessToken, refreshToken
	Callback(ctx context.Context, provider enums.Provider, code string, endpoint string, authorization string) (string, string, string, string, string, error)
	// GetClientID returns the client ID for the given provider.
	GetClientID(ctx context.Context, provider enums.Provider) (string, error)
}

type service struct {
	dig.In

	Config      *config.Configuration
	RepoUser    repouser.UserRepository
	SvcToken    token.Service
	SvcPassword password.Service
}

func NewService(params service) Service {
	return &params
}

// GetClientID returns the client ID for the given provider.
func (s *service) GetClientID(_ context.Context, provider enums.Provider) (string, error) {
	switch provider {
	case enums.ProviderGithub:
		return s.Config.Auth.Oauth2.Github.ClientID, nil
	case enums.ProviderGitlab:
		return s.Config.Auth.Oauth2.Gitlab.ClientID, nil
	case enums.ProviderGitea:
		return s.Config.Auth.Oauth2.Gitea.ClientID, nil
	default:
		return "", errcode.HTTPErrCodeBadRequest.Detail(fmt.Sprintf("invalid provider %s", provider))
	}
}

// Callback handles the oauth2 callback flow: token exchange, user lookup/creation, 3rd party binding, JWT issuance.
// Returns: user ID, username, email, accessToken, refreshToken
func (s *service) Callback(ctx context.Context, provider enums.Provider, code string, endpoint string, authorization string) (string, string, string, string, string, error) {
	userSignedObj, err := s.tryGetUser(ctx, authorization)
	if err != nil {
		slog.Error("get user failed", "err", err)
		return "", "", "", "", "", errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Get user failed: %v", err))
	}

	var conf *oauth2.Config
	switch provider {
	case enums.ProviderGithub:
		conf = &oauth2.Config{
			ClientID:     s.Config.Auth.Oauth2.Github.ClientID,
			ClientSecret: s.Config.Auth.Oauth2.Github.ClientSecret,
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://github.com/login/oauth/authorize",    // #nosec G101 -- public OAuth endpoint, not a credential
				TokenURL: "https://github.com/login/oauth/access_token", // #nosec G101 -- public OAuth endpoint, not a credential
			},
		}
	case enums.ProviderGitlab:
		baseURL := s.scmBaseURL(provider)
		conf = &oauth2.Config{
			ClientID:     s.Config.Auth.Oauth2.Gitlab.ClientID,
			ClientSecret: s.Config.Auth.Oauth2.Gitlab.ClientSecret,
			Endpoint: oauth2.Endpoint{
				AuthURL:  baseURL + "/oauth/authorize",
				TokenURL: baseURL + "/oauth/token",
			},
			RedirectURL: fmt.Sprintf("%s/api/v1/oauth2/%s/redirect_callback?endpoint=%s",
				s.Config.HTTP.Endpoint, enums.ProviderGitlab.String(), url.QueryEscape(endpoint)),
			Scopes: []string{"api", "read_api", "read_user", "read_repository"},
		}
	case enums.ProviderGitea:
		baseURL := s.scmBaseURL(provider)
		conf = &oauth2.Config{
			ClientID:     s.Config.Auth.Oauth2.Gitea.ClientID,
			ClientSecret: s.Config.Auth.Oauth2.Gitea.ClientSecret,
			Endpoint: oauth2.Endpoint{
				AuthURL:  baseURL + "/login/oauth/authorize",
				TokenURL: baseURL + "/login/oauth/access_token",
			},
			RedirectURL: fmt.Sprintf("%s/api/v1/oauth2/%s/redirect_callback?endpoint=%s",
				s.Config.HTTP.Endpoint, enums.ProviderGitea.String(), url.QueryEscape(endpoint)),
		}
	}

	oauth2Token, err := conf.Exchange(ctx, code)
	if err != nil {
		if strings.Contains(err.Error(), "bad_verification_code") {
			slog.Error("verification code invalid", "err", err, "platform", string(provider), "code", code)
			return "", "", "", "", "", errcode.HTTPErrCodeVerificationCodeInvalid.Detail(err.Error())
		}
		slog.Error("request token failed", "err", err, "platform", string(provider), "code", code)
		return "", "", "", "", "", errcode.HTTPErrCodeInternalError.Detail(err.Error())
	}

	client := conf.Client(ctx, oauth2Token)

	var userInfo api.Oauth2UserInfo

	switch provider {
	case enums.ProviderGithub:
		user, _, err := github.NewClient(client).Users.Get(ctx, "")
		if err != nil {
			slog.Error("get user info failed", "err", err)
			return "", "", "", "", "", errcode.HTTPErrCodeInternalError.Detail(err.Error())
		}
		userInfo = api.Oauth2UserInfo{
			Provider: provider,
			ID:       strconv.FormatInt(user.GetID(), 10),
			Username: user.GetLogin(),
			Email:    user.GetEmail(),
		}
	case enums.ProviderGitlab:
		glClient, err := gitlab.NewOAuthClient(oauth2Token.AccessToken) // nolint: staticcheck
		if err != nil {
			slog.Error("create gitlab client failed", "err", err)
			return "", "", "", "", "", errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Create gitlab client failed: %v", err))
		}
		user, _, err := glClient.Users.CurrentUser()
		if err != nil {
			slog.Error("get user info failed", "err", err)
			return "", "", "", "", "", errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Get user info failed: %v", err))
		}
		userInfo = api.Oauth2UserInfo{
			Provider: provider,
			ID:       strconv.FormatInt(user.ID, 10),
			Username: user.Name,
			Email:    user.Email,
		}
	case enums.ProviderGitea:
		// gitea.NewClient("", gitea.SetHTTPClient(client))
	}

	var userExist = true

	user3rdPartyObj, err := s.RepoUser.GetUser3rdPartyByAccountID(ctx, provider, userInfo.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			userExist = false
		} else {
			slog.Error("get user by provider failed", "err", err)
			return "", "", "", "", "", errcode.HTTPErrCodeInternalError.Detail(err.Error())
		}
	}

	if user3rdPartyObj != nil && userSignedObj != nil && user3rdPartyObj.UserID != userSignedObj.ID {
		slog.Error("user already bound to another account", "user_id", user3rdPartyObj.UserID, "signed", userSignedObj.ID)
		return "", "", "", "", "", errcode.HTTPErrCodeConflict.Detail("User already bound to another account")
	}

	// OAuth tokens are no longer persisted: SCM credentials are configured
	// explicitly per provider (see pkg/service/coderepos) because OAuth access
	// tokens expire and refresh tokens were never used.

	if !userExist {
		if userSignedObj == nil {
			var usernameExist = true
			_, err = s.RepoUser.GetByUsername(ctx, userInfo.Username)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					usernameExist = false
				} else {
					slog.Error("get user by username failed", "err", err)
					return "", "", "", "", "", errcode.HTTPErrCodeInternalError.Detail(err.Error())
				}
			}
			if usernameExist {
				userInfo.Username = fmt.Sprintf("%s-%s", userInfo.Username, uuid.NewV7String())
			}
			err = query.Q.Transaction(func(tx *query.Query) error {
				txUserRepository := repouser.NewUserRepository(tx)
				userSignedObj = &models.User{
					ID:       uuid.NewV7String(),
					Username: userInfo.Username,
					Email:    new(userInfo.Email),
				}
				err = txUserRepository.Create(ctx, userSignedObj)
				if err != nil {
					slog.Error("create user failed", "err", err)
					return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Create user failed: %v", err))
				}
				user3rdPartyObj = &models.User3rdParty{
					ID:        uuid.NewV7String(),
					Provider:  provider,
					AccountID: new(userInfo.ID),
					UserID:    userSignedObj.ID,
				}
				err = txUserRepository.CreateUser3rdParty(ctx, user3rdPartyObj)
				if err != nil {
					slog.Error("create user failed", "err", err)
					return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Create user failed: %v", err))
				}
				user3rdPartyObj.User = ptr.To(userSignedObj)
				return nil
			})
			if err != nil {
				return "", "", "", "", "", err
			}
		} else {
			err = query.Q.Transaction(func(tx *query.Query) error {
				txUserRepository := repouser.NewUserRepository(tx)
				user3rdPartyObj = &models.User3rdParty{
					ID:        uuid.NewV7String(),
					Provider:  provider,
					AccountID: new(userInfo.ID),
					UserID:    userSignedObj.ID,
				}
				err = txUserRepository.CreateUser3rdParty(ctx, user3rdPartyObj)
				if err != nil {
					slog.Error("create user failed", "err", err)
					return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Create user failed: %v", err))
				}
				user3rdPartyObj.User = ptr.To(userSignedObj)
				return nil
			})
			if err != nil {
				return "", "", "", "", "", err
			}
		}
	}

	refreshToken, err := s.SvcToken.New(user3rdPartyObj.User.ID, s.Config.Auth.Jwt.Ttl)
	if err != nil {
		slog.Error("create refresh token failed", "err", err)
		return "", "", "", "", "", errcode.HTTPErrCodeInternalError.Detail(err.Error())
	}

	accessToken, err := s.SvcToken.New(user3rdPartyObj.User.ID, s.Config.Auth.Jwt.RefreshTTL)
	if err != nil {
		slog.Error("create token failed", "err", err)
		return "", "", "", "", "", errcode.HTTPErrCodeInternalError.Detail(err.Error())
	}

	return user3rdPartyObj.User.ID, user3rdPartyObj.User.Username, ptr.To(user3rdPartyObj.User.Email), accessToken, refreshToken, nil
}

// scmBaseURL returns the base URL of a provider from the scm configuration, falling back to
// the hosted endpoint when the provider is not configured.
func (s *service) scmBaseURL(provider enums.Provider) string {
	if scmProvider, ok := s.Config.ScmProvider(provider); ok && scmProvider.BaseURL != "" {
		return scmProvider.BaseURL
	}
	switch provider {
	case enums.ProviderGithub:
		return "https://github.com"
	case enums.ProviderGitlab:
		return "https://gitlab.com"
	}
	return ""
}

// tryGetUser resolves the user from the Authorization header.
// For Basic auth: lookup by username + verify password.
// For Bearer auth: validate the JWT token.
// Returns the resolved user, or nil if no Authorization header is present.
func (s *service) tryGetUser(ctx context.Context, authorization string) (*models.User, error) {
	var uid string
	var err error

	switch {
	case strings.HasPrefix(authorization, "Basic"):
		req := &http.Request{Header: http.Header{}}
		req.Header.Set(consts.HeaderAuthorization, authorization)
		username, pwd, ok := req.BasicAuth()
		if !ok {
			return nil, nil
		}

		user, err := s.RepoUser.GetByUsername(ctx, username)
		if err != nil {
			slog.Error("get user by username failed", "err", err)
			return nil, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Get user by username failed: %v", err))
		}
		uid = user.ID

		verify := s.SvcPassword.Verify(pwd, ptr.To(user.Password))
		if !verify {
			slog.Error("verify password failed")
			return nil, errcode.HTTPErrCodeUnauthorized.Detail("Verify password failed")
		}
	case strings.HasPrefix(authorization, "Bearer"):
		_, uid, err = s.SvcToken.Validate(ctx, strings.TrimSpace(strings.TrimPrefix(authorization, "Bearer")))
		if err != nil {
			slog.Error("validate token failed", "err", err)
			return nil, errcode.HTTPErrCodeUnauthorized.Detail(fmt.Sprintf("Validate token failed: %v", err))
		}
	default:
		return nil, nil
	}

	userObj, err := s.RepoUser.Get(ctx, uid)
	if err != nil {
		slog.Error("get user failed", "err", err)
		return nil, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Get user failed: %v", err))
	}

	return userObj, nil
}
