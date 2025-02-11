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
	pwdvalidate "github.com/wagslane/go-password-validator"

	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/xerrors"
)

// ResetPassword handles the reset request
func (h *handler) ResetPassword(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	var req types.PostUserResetPasswordPasswordRequest
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
		return xerrors.HTTPErrCodeInternalError.Detail(err.Error())
	}
	err = query.Q.Transaction(func(tx *query.Query) error {
		userService := h.UserServiceFactory.New(tx)
		err = userService.UpdateByID(ctx, req.ID, map[string]any{
			query.User.Password.ColumnName().String(): pwdHash,
		})
		if err != nil {
			log.Error().Err(err).Msg("Update user failed")
			return xerrors.HTTPErrCodeInternalError.Detail(err.Error())
		}
		return nil
	})
	if err != nil {
		return xerrors.NewHTTPError(c, err.(xerrors.ErrCode))
	}
	return c.NoContent(http.StatusAccepted)
}
