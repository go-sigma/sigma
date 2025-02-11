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
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/google/go-github/v53/github"
	"github.com/labstack/echo/v4"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/rs/zerolog/log"
	gitlab "gitlab.com/gitlab-org/api/client-go"
	"go.uber.org/dig"
	"golang.org/x/oauth2"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/dal/dao"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/modules/workq"
	"github.com/go-sigma/sigma/pkg/modules/workq/definition"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/password"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
	"github.com/go-sigma/sigma/pkg/utils/token"
	"github.com/go-sigma/sigma/pkg/xerrors"
)

// Callback handles the oauth2 callback request
//
//	@Summary	OAuth2 callback
//	@security	BasicAuth
//	@Tags		OAuth2
//	@Accept		json
//	@Produce	json
//	@Router		/oauth2/{provider}/callback [get]
//	@Param		provider	path		string	true	"oauth2 provider"
//	@Param		code		query		string	true	"code"
//	@Param		endpoint	query		string	false	"endpoint"
//	@Success	200			{object}	types.Oauth2ClientIDResponse
//	@Failure	500			{object}	xerrors.ErrCode
func (h *handler) Callback(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	userSignedObj, err := h.tryGetUser(c)
	if err != nil {
		log.Error().Err(err).Msg("Get user failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Get user failed: %v", err))
	}

	var req types.Oauth2CallbackRequest
	err = utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, err.Error())
	}

	var conf *oauth2.Config
	switch req.Provider {
	case enums.ProviderGithub:
		conf = &oauth2.Config{
			ClientID:     h.Config.Auth.Oauth2.Github.ClientID,
			ClientSecret: h.Config.Auth.Oauth2.Github.ClientSecret,
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://github.com/login/oauth/authorize",
				TokenURL: "https://github.com/login/oauth/access_token",
			},
		}
	case enums.ProviderGitlab:
		conf = &oauth2.Config{
			ClientID:     h.Config.Auth.Oauth2.Gitlab.ClientID,
			ClientSecret: h.Config.Auth.Oauth2.Gitlab.ClientSecret,
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://gitlab.com/oauth/authorize",
				TokenURL: "https://gitlab.com/oauth/token",
			},
			RedirectURL: fmt.Sprintf("%s/api/v1/oauth2/%s/redirect_callback?endpoint=%s",
				h.Config.HTTP.Endpoint, enums.ProviderGitlab.String(), url.QueryEscape(req.Endpoint)),
			Scopes: []string{"api", "read_api", "read_user", "read_repository"},
		}
	case enums.ProviderGitea:
		conf = &oauth2.Config{
			ClientID:     h.Config.Auth.Oauth2.Gitea.ClientID,
			ClientSecret: h.Config.Auth.Oauth2.Gitea.ClientSecret,
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://gitlab.com/oauth/authorize",
				TokenURL: "https://gitlab.com/oauth/token",
			},
			RedirectURL: "http://localhost:3000/api/v1/oauth2/github/redirect_callback",
		}
	}

	oauth2Token, err := conf.Exchange(ctx, req.Code)
	if err != nil {
		if strings.Contains(err.Error(), "bad_verification_code") {
			log.Error().Err(err).Str("platform", string(req.Provider)).Str("code", req.Code).Msg("Verification code invalid")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeVerificationCodeInvalid, err.Error())
		}
		log.Error().Err(err).Str("platform", string(req.Provider)).Str("code", req.Code).Msg("Request token failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, err.Error())
	}

	client := conf.Client(ctx, oauth2Token)

	var userInfo types.Oauth2UserInfo

	switch req.Provider {
	case enums.ProviderGithub:
		user, _, err := github.NewClient(client).Users.Get(ctx, "")
		if err != nil {
			log.Error().Err(err).Msg("Get user info failed")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, err.Error())
		}
		userInfo = types.Oauth2UserInfo{
			Provider:     req.Provider,
			ID:           strconv.FormatInt(user.GetID(), 10),
			Username:     user.GetLogin(),
			Email:        user.GetEmail(),
			Token:        oauth2Token.AccessToken,
			RefreshToken: oauth2Token.RefreshToken,
		}
	case enums.ProviderGitlab:
		client, err := gitlab.NewOAuthClient(oauth2Token.AccessToken)
		if err != nil {
			log.Error().Err(err).Msg("Create gitlab client failed")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Create gitlab client failed: %v", err))
		}
		user, _, err := client.Users.CurrentUser()
		if err != nil {
			log.Error().Err(err).Msg("Get user info failed")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Get user info failed: %v", err))
		}
		userInfo = types.Oauth2UserInfo{
			Provider:     req.Provider,
			ID:           strconv.FormatInt(int64(user.ID), 10),
			Username:     user.Name,
			Email:        user.Email,
			Token:        oauth2Token.AccessToken,
			RefreshToken: oauth2Token.RefreshToken,
		}
	case enums.ProviderGitea:
		// gitea.NewClient("", gitea.SetHTTPClient(client))
	}

	var userExist = true

	userService := h.UserServiceFactory.New()
	user3rdPartyObj, err := userService.GetUser3rdPartyByAccountID(ctx, req.Provider, userInfo.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			userExist = false
		} else {
			log.Error().Err(err).Msg("Get user by provider failed")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, err.Error())
		}
	}

	if user3rdPartyObj != nil && userSignedObj != nil && user3rdPartyObj.UserID != userSignedObj.ID {
		log.Error().Int64("user_id", user3rdPartyObj.UserID).Int64("signed", userSignedObj.ID).Msg("User already bound to another account")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeConflict, "User already bound to another account")
	}

	if userExist {
		err = userService.UpdateUser3rdParty(ctx, user3rdPartyObj.ID, map[string]any{
			query.User3rdParty.Token.ColumnName().String():        oauth2Token.AccessToken,
			query.User3rdParty.RefreshToken.ColumnName().String(): oauth2Token.RefreshToken,
		})
		if err != nil {
			log.Error().Err(err).Msg("Update user 3rdparty failed")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, err.Error())
		}
	}

	if !userExist {
		if userSignedObj == nil {
			var usernameExist = true
			_, err := userService.GetByUsername(ctx, userInfo.Username)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					usernameExist = false
				} else {
					log.Error().Err(err).Msg("Get user by username failed")
					return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, err.Error())
				}
			}
			if usernameExist {
				userInfo.Username = fmt.Sprintf("%s-%s", userInfo.Username, gonanoid.Must(6))
			}
			err = query.Q.Transaction(func(tx *query.Query) error {
				userService := dao.NewUserServiceFactory().New(tx)
				userSignedObj = &models.User{
					Username: userInfo.Username,
					Email:    ptr.Of(userInfo.Email),
				}
				err = userService.Create(ctx, userSignedObj)
				if err != nil {
					log.Error().Err(err).Msg("Create user failed")
					return xerrors.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Create user failed: %v", err))
				}
				user3rdPartyObj = &models.User3rdParty{
					Provider:     req.Provider,
					AccountID:    ptr.Of(userInfo.ID),
					Token:        ptr.Of(userInfo.Token),
					RefreshToken: ptr.Of(userInfo.RefreshToken),
					UserID:       userSignedObj.ID,
				}
				err = userService.CreateUser3rdParty(ctx, user3rdPartyObj)
				if err != nil {
					log.Error().Err(err).Msg("Create user failed")
					return xerrors.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Create user failed: %v", err))
				}
				err = workq.ProducerClient.Produce(ctx, enums.DaemonCodeRepository,
					types.DaemonCodeRepositoryPayload{User3rdPartyID: user3rdPartyObj.ID}, definition.ProducerOption{Tx: tx})
				if err != nil {
					log.Error().Err(err).Int64("user_id", user3rdPartyObj.UserID).Msg("Publish sync code repository failed")
					return xerrors.HTTPErrCodeInternalError.Detail("Publish sync code repository failed")
				}
				user3rdPartyObj.User = ptr.To(userSignedObj)
				return nil
			})
			if err != nil {
				return xerrors.NewHTTPError(c, err.(xerrors.ErrCode))
			}
		} else {
			err = query.Q.Transaction(func(tx *query.Query) error {
				userService := dao.NewUserServiceFactory().New(tx)
				user3rdPartyObj = &models.User3rdParty{
					Provider:     req.Provider,
					AccountID:    ptr.Of(userInfo.ID),
					Token:        ptr.Of(userInfo.Token),
					RefreshToken: ptr.Of(userInfo.RefreshToken),
					UserID:       userSignedObj.ID,
				}
				err = userService.CreateUser3rdParty(ctx, user3rdPartyObj)
				if err != nil {
					log.Error().Err(err).Msg("Create user failed")
					return xerrors.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Create user failed: %v", err))
				}
				err = workq.ProducerClient.Produce(ctx, enums.DaemonCodeRepository,
					types.DaemonCodeRepositoryPayload{User3rdPartyID: user3rdPartyObj.ID}, definition.ProducerOption{Tx: tx})
				if err != nil {
					log.Error().Err(err).Int64("user_id", user3rdPartyObj.UserID).Msg("Publish sync code repository failed")
					return xerrors.HTTPErrCodeInternalError.Detail("Publish sync code repository failed")
				}
				user3rdPartyObj.User = ptr.To(userSignedObj)
				return nil
			})
			if err != nil {
				return xerrors.NewHTTPError(c, err.(xerrors.ErrCode))
			}
		}
	}

	refreshToken, err := h.TokenService.New(user3rdPartyObj.User.ID, h.Config.Auth.Jwt.Ttl)
	if err != nil {
		log.Error().Err(err).Msg("Create refresh token failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, err.Error())
	}

	token, err := h.TokenService.New(user3rdPartyObj.User.ID, h.Config.Auth.Jwt.RefreshTtl)
	if err != nil {
		log.Error().Err(err).Msg("Create token failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, err.Error())
	}

	return c.JSON(http.StatusOK, types.Oauth2CallbackResponse{
		ID:           user3rdPartyObj.User.ID,
		Username:     user3rdPartyObj.User.Username,
		Email:        ptr.To(user3rdPartyObj.User.Email),
		RefreshToken: refreshToken,
		Token:        token,
	})
}

func (h *handler) tryGetUser(c echo.Context) (*models.User, error) {
	req := c.Request()
	ctx := log.Logger.WithContext(req.Context())
	authorization := req.Header.Get("Authorization")

	var uid int64

	userService := h.UserServiceFactory.New()

	switch {
	case strings.HasPrefix(authorization, "Basic"):
		var username string
		var pwd string
		var ok bool
		username, pwd, ok = c.Request().BasicAuth()
		if !ok {
			return nil, nil
		}

		user, err := userService.GetByUsername(ctx, username)
		if err != nil {
			log.Error().Err(err).Msg("Get user by username failed")
			return nil, xerrors.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Get user by username failed: %v", err))
		}
		uid = user.ID

		passwordService := password.New()
		verify := passwordService.Verify(pwd, ptr.To(user.Password))
		if !verify {
			log.Error().Err(err).Msg("Verify password failed")
			return nil, xerrors.HTTPErrCodeUnauthorized.Detail(fmt.Sprintf("Verify password failed: %v", err))
		}
	case strings.HasPrefix(authorization, "Bearer"):
		tokenService, err := token.New(dig.New()) // TODO: dig
		if err != nil {
			log.Error().Err(err).Msg("Create token service failed")
			return nil, xerrors.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Create token service failed: %v", err))
		}
		_, uid, err = tokenService.Validate(ctx, strings.TrimSpace(strings.TrimPrefix(authorization, "Bearer")))
		if err != nil {
			log.Error().Err(err).Msg("Validate token failed")
			return nil, xerrors.HTTPErrCodeUnauthorized.Detail(fmt.Sprintf("Validate token failed: %v", err))
		}
	default:
		return nil, nil
	}

	userObj, err := userService.Get(ctx, uid)
	if err != nil {
		log.Error().Err(err).Msg("Get user failed")
		return nil, xerrors.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Get user failed: %v", err))
	}

	return userObj, nil
}
