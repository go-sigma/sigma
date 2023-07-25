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

package users

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"

	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
	"github.com/go-sigma/sigma/pkg/xerrors"
)

// Login handles the login request
func (h *handlers) Login(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	var req types.PostUserLoginRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, err.Error())
	}

	userService := h.userServiceFactory.New()
	user, err := userService.GetByUsername(ctx, req.Username)
	if err != nil {
		log.Error().Err(err).Msg("Get user by username failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeUnauthorized, err.Error())
	}

	verify := h.passwordService.Verify(req.Password, ptr.To(user.Password))
	if !verify {
		log.Error().Err(err).Msg("Verify password failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeUnauthorized, "Invalid username or password")
	}

	refreshToken, err := h.tokenService.New(user, viper.GetDuration("auth.jwt.ttl"))
	if err != nil {
		log.Error().Err(err).Msg("Create refresh token failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, err.Error())
	}

	token, err := h.tokenService.New(user, viper.GetDuration("auth.jwt.refreshTtl"))
	if err != nil {
		log.Error().Err(err).Msg("Create token failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, err.Error())
	}

	return c.JSON(http.StatusOK, types.PostUserLoginResponse{
		RefreshToken: refreshToken,
		Token:        token,
		ID:           user.ID,
		Email:        ptr.To(user.Email),
		Username:     user.Username,
	})
}
