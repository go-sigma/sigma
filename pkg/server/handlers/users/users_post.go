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

	"github.com/go-sigma/sigma/pkg/dal"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
	"github.com/go-sigma/sigma/pkg/xerrors"
)

// Post handles the post request
//
//	@Summary	Create user
//	@Tags		User
//	@Accept		json
//	@Produce	json
//	@Router		/users/ [post]
//	@Param		message	body	types.PostUserRequest	true	"User object"
//	@Success	201
//	@Failure	500	{object}	xerrors.ErrCode
func (h *handler) Post(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	var req types.PostUserRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, fmt.Sprintf("Bind and validate request body failed: %v", err))
	}

	pwdHash, err := h.PasswordService.Hash(req.Password)
	if err != nil {
		log.Error().Err(err).Msg("Hash password failed")
		return xerrors.HTTPErrCodeInternalError.Detail(err.Error())
	}

	userObj := models.User{
		Username:       req.Username,
		Password:       ptr.Of(pwdHash),
		Email:          ptr.Of(req.Email),
		Role:           req.Role,
		NamespaceLimit: ptr.To(req.NamespaceLimit),
	}

	err = query.Q.Transaction(func(tx *query.Query) error {
		userService := h.UserServiceFactory.New(tx)
		if userObj.Role == enums.UserRoleAdmin {
			userObj.NamespaceLimit = 0
		}
		err = userService.Create(ctx, &userObj)
		if err != nil {
			log.Error().Err(err).Msg("Create user failed")
			return xerrors.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Create user failed: %v", err))
		}
		if userObj.Role == enums.UserRoleAdmin {
			err = userService.AddPlatformMember(ctx, userObj.ID, userObj.Role)
			if err != nil {
				log.Error().Err(err).Msg("Add platform role for user failed")
				return xerrors.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Add platform role for user failed: %v", err))
			}
		}
		return nil
	})
	if err != nil {
		e, ok := err.(xerrors.ErrCode)
		if ok {
			return xerrors.NewHTTPError(c, e)
		}
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Create user failed: %v", err))
	}

	err = dal.AuthEnforcer.LoadPolicy()
	if err != nil {
		log.Error().Err(err).Msg("Reload policy failed")
	}

	return c.NoContent(http.StatusCreated)
}
