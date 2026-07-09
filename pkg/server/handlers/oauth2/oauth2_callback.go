// Copyright 2024 sigma
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

package oauth2

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/server/errcode"
)

// Callback handles the oauth2 callback request
//
//	@Summary	OAuth2 callback
//	@security	BasicAuth
//	@Tags		OAuth2
//	@Accept		json
//	@Produce	json
//	@Router		/oauth2/{provider}/callback [get]
//	@Param		provider	path		string	true	"oauth2 provider"
//	@Param		code		query		string	true	"code"
//	@Param		endpoint	query		string	false	"endpoint"
//	@Success	200			{object}	api.Oauth2ClientIDResponse
//	@Failure	500			{object}	errcode.ErrCode
func (h *handler) Callback(c *gin.Context, req *api.Oauth2CallbackRequest) {
	ctx := c.Request.Context()

	userID, username, email, accessToken, refreshToken, err := h.OAuth2Svc.Callback(ctx, req.Provider, req.Code, req.Endpoint, c.Request.Header.Get("Authorization"))
	if err != nil {
		if e, ok := errcode.AsType[errcode.ErrCode](err); ok {
			errcode.NewHTTPError(c, e)
			return
		}
		errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError, fmt.Sprintf("OAuth2 callback failed: %v", err))
		return
	}

	c.JSON(http.StatusOK, api.Oauth2CallbackResponse{
		ID:           userID,
		Username:     username,
		Email:        email,
		RefreshToken: refreshToken,
		Token:        accessToken,
	})
}
