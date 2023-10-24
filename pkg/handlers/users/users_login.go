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
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"

	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
	"github.com/go-sigma/sigma/pkg/xerrors"
)

// Login handles the login request
//
//	@Summary	Login user
//	@security	BasicAuth
//	@Tags		User
//	@Accept		json
//	@Produce	json
//	@Router		/users/login [post]
//	@Failure	500	{object}	xerrors.ErrCode
//	@Failure	401	{object}	xerrors.ErrCode
//	@Success	200	{object}	types.PostUserLoginResponse
func (h *handlers) Login(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	iuser := c.Get(consts.ContextUser)
	if iuser == nil {
		log.Error().Msg("Get user from header failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeUnauthorized)
	}
	user, ok := iuser.(*models.User)
	if !ok {
		log.Error().Msg("Convert user from header failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeUnauthorized)
	}

	userService := h.userServiceFactory.New()
	err := userService.UpdateByID(ctx, user.ID, map[string]any{
		query.User.LastLogin.ColumnName().String(): time.Now(),
	})
	if err != nil {
		log.Error().Err(err).Msg("Update user last login failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Update user last login failed: %v", err))
	}

	refreshToken, err := h.tokenService.New(user.ID, viper.GetDuration("auth.jwt.ttl"))
	if err != nil {
		log.Error().Err(err).Msg("Create refresh token failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, err.Error())
	}

	token, err := h.tokenService.New(user.ID, viper.GetDuration("auth.jwt.refreshTtl"))
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
