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
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	pwdvalidate "github.com/wagslane/go-password-validator"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/xerrors"
)

// RecoverPasswordReset handles the recover user's password reset
func (h *handler) RecoverPasswordReset(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	var req types.PostUserRecoverResetPasswordRequest
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

	userService := h.UserServiceFactory.New()
	userObj, err := userService.GetByRecoverCode(ctx, req.Code)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Str("code", req.Code).Msg("Recover code not found")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, "Recover code not found")
		}
		log.Error().Err(err).Str("code", req.Code).Msg("Get recover code failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, fmt.Sprintf("Get recover code failed: %v", err))
	}

	err = query.Q.Transaction(func(tx *query.Query) error {
		userService := h.UserServiceFactory.New(tx)
		err = userService.DeleteRecoverCode(ctx, userObj.ID)
		if err != nil {
			log.Error().Err(err).Str("code", req.Code).Msg("Delete recover code failed")
			return xerrors.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Delete recover code failed: %v", err))
		}
		pwdHash, err := h.PasswordService.Hash(req.Password)
		if err != nil {
			log.Error().Err(err).Msg("Hash password failed")
			return xerrors.HTTPErrCodeInternalError.Detail(err.Error())
		}
		err = userService.UpdateByID(ctx, userObj.ID, map[string]any{
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
