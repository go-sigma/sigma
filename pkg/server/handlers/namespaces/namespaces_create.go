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

package namespaces

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/server/errcode"
	"github.com/go-sigma/sigma/pkg/utils"
)

// PostNamespace handles the post namespace request
//
//	@Summary	Create namespace
//	@Tags		Namespace
//	@security	BasicAuth
//	@Accept		json
//	@Produce	json
//	@Router		/namespaces/ [post]
//	@Param		message	body		api.PostNamespaceRequest	true	"Namespace object"
//	@Success	201		{object}	api.PostNamespaceResponse
//	@Failure	400		{object}	errcode.ErrCode
//	@Failure	500		{object}	errcode.ErrCode
func (h *handler) PostNamespace(c *gin.Context, req *api.PostNamespaceRequest) {
	ctx := c.Request.Context()

	user, ok := utils.GetFromCtx[*models.User](c, consts.ContextUser)
	if !ok {
		slog.Error("get user from context failed")
		errcode.NewHTTPError(c, errcode.HTTPErrCodeUnauthorized)
		return
	}

	namespaceObj, err := h.NsSvc.CreateNamespace(ctx, user.ID, *req)
	if err != nil {
		if e, ok := errcode.AsType[errcode.ErrCode](err); ok {
			errcode.NewHTTPError(c, e)
			return
		}
		errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError)
		return
	}

	c.JSON(http.StatusCreated, api.PostNamespaceResponse{ID: namespaceObj.ID})
}
