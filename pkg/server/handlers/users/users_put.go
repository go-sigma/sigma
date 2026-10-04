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
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/server/errcode"
	"github.com/go-sigma/sigma/pkg/utils"
)

// Put handles the put request
//
//	@Summary	Update user
//	@Tags		User
//	@Accept		json
//	@Produce	json
//	@Router		/users/{user_id} [put]
//	@Param		user_id	path	string				true	"User ID"
//	@Param		message	body	api.PutUserRequest	true	"User object"
//	@Success	204
//	@Failure	500	{object}	errcode.ErrCode
func (h *handler) Put(c *gin.Context, req *api.PutUserRequest) {
	ctx := c.Request.Context()

	user, ok := utils.GetFromCtx[*models.User](c, consts.ContextUser)
	if !ok {
		slog.Error("get user from context failed")
		errcode.NewHTTPError(c, errcode.HTTPErrCodeUnauthorized)
		return
	}
	// A user must not be able to change their own status or role.
	if user.ID == req.UserID {
		if req.Status != nil {
			errcode.NewHTTPError(c, errcode.HTTPErrCodeForbidden, "Cannot update your own status")
			return
		}
		if req.Role != nil {
			errcode.NewHTTPError(c, errcode.HTTPErrCodeForbidden, "Cannot update your own role")
			return
		}
	}

	err := h.UserSvc.UpdateUser(ctx, req.UserID, *req)
	if err != nil {
		if e, ok := errcode.AsType[errcode.ErrCode](err); ok {
			errcode.NewHTTPError(c, e)
			return
		}
		errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError, fmt.Sprintf("Update user failed: %v", err))
		return
	}

	c.Status(http.StatusNoContent)
}
