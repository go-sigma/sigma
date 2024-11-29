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

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
	"github.com/go-sigma/sigma/pkg/xerrors"
)

// RecoverPassword handles the recover user's password
func (h *handler) RecoverPassword(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	var req types.PostUserRecoverPasswordRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, err.Error())
	}

	userService := h.UserServiceFactory.New()
	user, err := userService.GetByUsername(ctx, req.Username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Str("username", req.Username).Msg("Username not found")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, "User or email not found")
		}
		log.Error().Err(err).Str("username", req.Username).Msg("Username find failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, fmt.Sprintf("Username find failed: %v", err))
	}
	if ptr.To(user.Email) != req.Email {
		log.Error().Err(err).Str("username", req.Username).Str("realEmail", ptr.To(user.Email)).Str("email", req.Email).Msg("Email not equal to real")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, "User or email not found")
	}

	_, err = userService.GetRecoverCodeByUserID(ctx, user.ID)
	if err == nil {
		log.Error().Err(err).Str("username", req.Username).Msg("Recover code already exists")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeConflict, "Recover code already exists")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Error().Err(err).Str("username", req.Username).Msg("Get recover code failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, fmt.Sprintf("Get recover code failed: %v", err))
	}

	err = query.Q.Transaction(func(tx *query.Query) error {
		userService := h.UserServiceFactory.New(tx)
		err = userService.CreateRecoverCode(ctx, &models.UserRecoverCode{
			UserID: user.ID,
			Code:   uuid.NewString(),
		})
		if err != nil {
			log.Error().Err(err).Str("username", req.Username).Msg("Create recover code failed")
			return xerrors.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Create recover code failed: %v", err))
		}
		return nil
	})
	if err != nil {
		return xerrors.NewHTTPError(c, err.(xerrors.ErrCode))
	}
	return c.NoContent(http.StatusCreated)
}
