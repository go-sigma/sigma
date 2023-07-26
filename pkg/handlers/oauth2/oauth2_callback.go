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

package oauth2

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/go-github/v53/github"
	"github.com/labstack/echo/v4"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
	"github.com/go-sigma/sigma/pkg/xerrors"
)

// Callback ...
func (h *handlers) Callback(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	var req types.Oauth2CallbackRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, err.Error())
	}

	var conf *oauth2.Config
	switch req.Provider { // nolint: gocritic
	case enums.ProviderGithub:
		conf = &oauth2.Config{
			ClientID:     viper.GetString("auth.oauth2.github.clientId"),
			ClientSecret: viper.GetString("auth.oauth2.github.clientSecret"),
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://github.com/login/oauth/authorize",
				TokenURL: "https://github.com/login/oauth/access_token",
			},
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

	var rReq *http.Request
	switch req.Provider { // nolint: gocritic
	case enums.ProviderGithub:
		rReq, err = http.NewRequest(http.MethodGet, "https://api.github.com/user", nil)
	}
	if err != nil {
		log.Error().Err(err).Str("platform", string(req.Provider)).Str("code", req.Code).Msg("Create request failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, err.Error())
	}
	resp, err := client.Do(rReq)
	if err != nil {
		log.Error().Err(err).Str("platform", string(req.Provider)).Str("code", req.Code).Msg("Request user info failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, err.Error())
	}
	if resp.StatusCode != 200 {
		log.Error().Err(err).Str("platform", string(req.Provider)).Str("code", req.Code).Msg("Request user info failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, err.Error())
	}

	var userInfo types.Oauth2UserInfo
	switch req.Provider { // nolint: gocritic
	case enums.ProviderGithub:
		var user github.User
		err = json.NewDecoder(resp.Body).Decode(&user)
		if err != nil {
			log.Error().Err(err).Msg("Decode user info failed")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, err.Error())
		}
		userInfo = types.Oauth2UserInfo{
			Provider: req.Provider,
			ID:       strconv.FormatInt(user.GetID(), 10),
			Username: user.GetLogin(),
			Email:    user.GetEmail(),
		}
	}

	var userExist = true

	userService := h.userServiceFactory.New()
	userObj, err := userService.GetByProvider(ctx, req.Provider, userInfo.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			userExist = false
		} else {
			log.Error().Err(err).Msg("Get user by provider failed")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, err.Error())
		}
	}

	if !userExist {
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
		userObj = &models.User{
			Provider:          req.Provider,
			ProviderAccountID: ptr.Of(userInfo.ID),
			Username:          userInfo.Username,
			Email:             ptr.Of(userInfo.Email),
		}
		err = userService.Create(ctx, userObj)
		if err != nil {
			log.Error().Err(err).Msg("Create user failed")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, err.Error())
		}
	}

	refreshToken, err := h.tokenService.New(userObj, viper.GetDuration("auth.jwt.ttl"))
	if err != nil {
		log.Error().Err(err).Msg("Create refresh token failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, err.Error())
	}

	token, err := h.tokenService.New(userObj, viper.GetDuration("auth.jwt.refreshTtl"))
	if err != nil {
		log.Error().Err(err).Msg("Create token failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, err.Error())
	}

	return c.JSON(http.StatusOK, types.Oauth2CallbackResponse{
		ID:           userObj.ID,
		Username:     userObj.Username,
		Email:        ptr.To(userObj.Email),
		RefreshToken: refreshToken,
		Token:        token,
	})
}
