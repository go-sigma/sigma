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

package token

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/server/errcode"
	"github.com/go-sigma/sigma/pkg/utils"
)

// Token generate token for docker client
//
//	@Summary	Generate token
//	@Tags		Token
//	@security	BasicAuth
//	@Accept		json
//	@Produce	json
//	@Router		/tokens [get]
//	@Success	200	{object}	api.PostUserTokenResponse
//	@Failure	401	{object}	errcode.ErrCode
//	@Failure	500	{object}	errcode.ErrCode
func (h *handler) Token(c *gin.Context) {
	user, ok := utils.GetFromCtx[*models.User](c, consts.ContextUser)
	if !ok {
		slog.Error("get user from context failed")
		errcode.NewHTTPError(c, errcode.HTTPErrCodeUnauthorized)
		return
	}

	token, err := h.TokenSvc.New(user.ID, h.Config.Auth.Jwt.Ttl)
	if err != nil {
		slog.Error("create token failed", "err", err)
		errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError, err.Error())
		return
	}

	c.JSON(http.StatusOK, api.PostUserTokenResponse{
		Token:     token,
		ExpiresIn: int(h.Config.Auth.Jwt.Ttl.Seconds()),
		IssuedAt:  time.Now().Format(time.RFC3339),
	})
}
