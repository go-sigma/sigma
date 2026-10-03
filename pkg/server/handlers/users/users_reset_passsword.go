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

	"github.com/gin-gonic/gin"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/server/errcode"
)

// ResetPassword handles the reset request
//
//	@Summary	Reset user password
//	@security	BasicAuth
//	@Tags		User
//	@Accept		json
//	@Produce	json
//	@Router		/users/{user_id}/reset-password [put]
//	@Param		user_id	path	string										true	"User ID"
//	@Param		message	body	api.PostUserResetPasswordPasswordRequest	true	"Password object"
//	@Success	204
//	@Failure	400	{object}	errcode.ErrCode
//	@Failure	401	{object}	errcode.ErrCode
//	@Failure	404	{object}	errcode.ErrCode
//	@Failure	500	{object}	errcode.ErrCode
func (h *handler) ResetPassword(c *gin.Context, req *api.PostUserResetPasswordPasswordRequest) {
	ctx := c.Request.Context()

	err := h.UserSvc.ResetPassword(ctx, req.ID, req.Password)
	if err != nil {
		if e, ok := errcode.AsType[errcode.ErrCode](err); ok {
			errcode.NewHTTPError(c, e)
			return
		}
		errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError, fmt.Sprintf("Reset password failed: %v", err))
		return
	}

	c.Status(http.StatusNoContent)
}
