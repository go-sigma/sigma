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

	"github.com/gin-gonic/gin"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/server/errcode"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

// Login handles the login request
//
//	@Summary	Login user
//	@security	BasicAuth
//	@Tags		User
//	@Accept		json
//	@Produce	json
//	@Router		/users/login [post]
//	@Param		message	body		api.PostUserLoginRequest	true	"User login object"
//	@Failure	500		{object}	errcode.ErrCode
//	@Failure	401		{object}	errcode.ErrCode
//	@Success	200		{object}	api.PostUserLoginResponse
func (h *handler) Login(c *gin.Context) {
	ctx := c.Request.Context()

	user, needRet := utils.GetUserFromCtx(c, utils.UserCtxErrorHTTP)
	if needRet {
		return
	}

	accessToken, refreshToken, err := h.UserSvc.Login(ctx, user.ID, h.Config.Auth.Jwt.Ttl, h.Config.Auth.Jwt.RefreshTTL)
	if err != nil {
		if e, ok := errcode.AsType[errcode.ErrCode](err); ok {
			errcode.NewHTTPError(c, e)
			return
		}
		errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError)
		return
	}

	c.JSON(http.StatusOK, api.PostUserLoginResponse{
		RefreshToken: refreshToken,
		Token:        accessToken,
		ID:           user.ID,
		Email:        ptr.To(user.Email),
		Username:     user.Username,
	})
}
