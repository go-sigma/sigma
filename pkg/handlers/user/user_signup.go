// Copyright 2023 XImager
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

package user

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	pwdvalidate "github.com/wagslane/go-password-validator"

	"github.com/ximager/ximager/pkg/consts"
	"github.com/ximager/ximager/pkg/dal/dao"
	"github.com/ximager/ximager/pkg/dal/models"
	"github.com/ximager/ximager/pkg/types"
	"github.com/ximager/ximager/pkg/utils/password"
	"github.com/ximager/ximager/pkg/utils/token"
	"github.com/ximager/ximager/pkg/xerrors"
)

// Signup handles the user signup
func (h *handlers) Signup(c echo.Context) error {
	ctx := c.Request().Context()

	var req types.PostUserSignupRequest
	err := c.Bind(&req)
	if err != nil {
		log.Error().Err(err).Msg("Bind request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, err.Error())
	}
	err = c.Validate(&req)
	if err != nil {
		log.Error().Err(err).Msg("Validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, err.Error())
	}
	err = pwdvalidate.Validate(req.Password, consts.PwdStrength)
	if err != nil {
		log.Error().Err(err).Msg("Validate password failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, err.Error())
	}

	passwordService := password.New()
	pwdHash, err := passwordService.Hash(req.Password)
	if err != nil {
		log.Error().Err(err).Msg("Hash password failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, err.Error())
	}

	userService := dao.NewUserService()
	user := &models.User{
		Username: req.Username,
		Password: pwdHash,
		Email:    req.Email,
	}
	err = userService.Create(ctx, user)
	if err != nil {
		log.Error().Err(err).Msg("Create user failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, err.Error())
	}

	tokenService, err := token.NewTokenService(viper.GetString("auth.jwt.privateKey"))
	if err != nil {
		log.Error().Err(err).Msg("Create token service failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, err.Error())
	}

	refreshToken, err := tokenService.New(user, viper.GetDuration("auth.jwt.ttl"))
	if err != nil {
		log.Error().Err(err).Msg("Create refresh token failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, err.Error())
	}

	token, err := tokenService.New(user, viper.GetDuration("auth.jwt.refreshTtl"))
	if err != nil {
		log.Error().Err(err).Msg("Create token failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, err.Error())
	}

	return c.JSON(http.StatusOK, types.PostUserLoginResponse{
		RefreshToken: refreshToken,
		Token:        token,
	})
}
