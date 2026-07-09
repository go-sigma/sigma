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
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/server/errcode"
)

// Logout handles the logout request
//
//	@Summary	Logout user
//	@security	BasicAuth
//	@Tags		User
//	@Accept		json
//	@Produce	json
//	@Param		message	body	api.PostUserLogoutRequest	true	"Logout user object"
//	@Router		/users/logout [post]
//	@Failure	500	{object}	errcode.ErrCode
//	@Failure	401	{object}	errcode.ErrCode
//	@Success	204
func (h *handler) Logout(c *gin.Context, req *api.PostUserLogoutRequest) {
	ctx := c.Request.Context()

	jtiVal, _ := c.Get("jti")
	jti, ok := jtiVal.(string)
	if !ok || jti == "" {
		slog.Error("get jti failed", "jti", jti)
		errcode.NewHTTPError(c, errcode.HTTPErrCodeUnauthorized, "Get jti failed")
		return
	}

	err := h.UserSvc.Logout(ctx, req.Tokens, jti)
	if err != nil {
		if e, ok := errcode.AsType[errcode.ErrCode](err); ok {
			errcode.NewHTTPError(c, e)
			return
		}
		errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError)
		return
	}

	c.Status(http.StatusNoContent)
}
