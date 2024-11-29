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
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	pwdvalidate "github.com/wagslane/go-password-validator"

	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
	"github.com/go-sigma/sigma/pkg/xerrors"
)

// Signup handles the user signup
func (h *handler) Signup(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	var req types.PostUserSignupRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, err.Error())
	}
	err = pwdvalidate.Validate(req.Password, consts.PwdStrength)
	if err != nil {
		log.Error().Err(err).Msg("Validate password failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, err.Error())
	}

	pwdHash, err := h.PasswordService.Hash(req.Password)
	if err != nil {
		log.Error().Err(err).Msg("Hash password failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, err.Error())
	}

	userService := h.UserServiceFactory.New()
	_, err = userService.GetByUsername(ctx, req.Username)
	if err == nil {
		log.Error().Msg("Username already exists")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeConflict, fmt.Errorf("username already exists").Error())
	}

	user := &models.User{
		Username: req.Username,
		Password: ptr.Of(pwdHash),
		Email:    ptr.Of(req.Email),
	}
	err = userService.Create(ctx, user)
	if err != nil {
		log.Error().Err(err).Msg("Create user failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, err.Error())
	}

	refreshToken, err := h.TokenService.New(user.ID, h.Config.Auth.Jwt.Ttl)
	if err != nil {
		log.Error().Err(err).Msg("Create refresh token failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, err.Error())
	}

	token, err := h.TokenService.New(user.ID, h.Config.Auth.Jwt.RefreshTtl)
	if err != nil {
		log.Error().Err(err).Msg("Create token failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, err.Error())
	}

	return c.JSON(http.StatusOK, types.PostUserLoginResponse{
		RefreshToken: refreshToken,
		Token:        token,
	})
}
