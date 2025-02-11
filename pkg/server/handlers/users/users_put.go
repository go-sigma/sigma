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
	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
	"github.com/go-sigma/sigma/pkg/xerrors"
)

// Put handles the put request
//
//	@Summary	Update user
//	@Tags		User
//	@Accept		json
//	@Produce	json
//	@Router		/users/{id} [get]
//	@Param		id	path	string	true	"User id"
//	@Success	204
//	@Failure	500	{object}	xerrors.ErrCode
func (h *handler) Put(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	var req types.PutUserRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, fmt.Sprintf("Bind and validate request body failed: %v", err))
	}

	userService := h.UserServiceFactory.New()
	userObj, err := userService.Get(ctx, req.UserID)
	if err != nil {
		log.Error().Err(err).Int64("id", req.UserID).Msg("Get user failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Get user failed: %v", err))
	}

	updates := make(map[string]any, 5)
	if req.Username != nil {
		updates[query.User.Username.ColumnName().String()] = ptr.To(req.Username)
	}
	if req.Email != nil {
		updates[query.User.Email.ColumnName().String()] = ptr.To(req.Email)
	}
	if req.Status != nil {
		updates[query.User.Status.ColumnName().String()] = req.Status
	}
	if req.Password != nil {
		pwdHash, err := h.PasswordService.Hash(ptr.To(req.Password))
		if err != nil {
			log.Error().Err(err).Msg("Hash password failed")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Hash password failed: %v", err))
		}
		updates[query.User.Password.ColumnName().String()] = pwdHash
	}
	if req.NamespaceLimit != nil {
		updates[query.User.NamespaceLimit.ColumnName().String()] = req.NamespaceLimit
	}
	err = query.Q.Transaction(func(tx *query.Query) error {
		userService := h.UserServiceFactory.New(tx)
		err = userService.UpdateByID(ctx, userObj.ID, updates)
		if err != nil {
			log.Error().Err(err).Msg("Update user failed")
			return xerrors.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Update user failed: %v", err))
		}
		if userObj.Role == enums.UserRoleAdmin {
			err = userService.AddPlatformMember(ctx, userObj.ID, userObj.Role)
			if err != nil {
				log.Error().Err(err).Msg("Add role for user failed")
				return xerrors.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Add role for user failed: %v", err))
			}
		} else {
			err = userService.DeletePlatformMember(ctx, userObj.ID, userObj.Role)
			if err != nil {
				log.Error().Err(err).Msg("Delete role for user failed")
				return xerrors.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Delete role for user failed: %v", err))
			}
		}
		return nil
	})
	if err != nil {
		e, ok := err.(xerrors.ErrCode)
		if ok {
			return xerrors.NewHTTPError(c, e)
		}
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Update user failed: %v", err))
	}

	err = dal.AuthEnforcer.LoadPolicy()
	if err != nil {
		log.Error().Err(err).Msg("Reload policy failed")
	}

	return c.NoContent(http.StatusNoContent)
}
