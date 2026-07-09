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

package builders

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

// CreateBuilder handles the create builder request
//
//	@Summary	Create a builder for repository
//	@Tags		Builder
//	@security	BasicAuth
//	@Accept		json
//	@Produce	json
//	@Router		/namespaces/{namespace_id}/repositories/{repository_id}/builders/ [post]
//	@Param		namespace_id	path	string						true	"Namespace ID"
//	@Param		repository_id	path	string						true	"Repository ID"
//	@Param		message			body	api.CreateBuilderRequest	true	"Builder object"
//	@Success	201
//	@Failure	400	{object}	errcode.ErrCode
//	@Failure	404	{object}	errcode.ErrCode
//	@Failure	500	{object}	errcode.ErrCode
func (h *handler) CreateBuilder(c *gin.Context, req *api.CreateBuilderRequest) {
	ctx := c.Request.Context()

	user, ok := utils.GetFromCtx[*models.User](c, consts.ContextUser)
	if !ok {
		slog.Error("get user from context failed")
		errcode.NewHTTPError(c, errcode.HTTPErrCodeUnauthorized)
		return
	}

	err := h.BuilderSvc.CreateBuilder(ctx, user.ID, *req)
	if err != nil {
		if e, ok := errcode.AsType[errcode.ErrCode](err); ok {
			errcode.NewHTTPError(c, e)
			return
		}
		errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError, fmt.Sprintf("Create builder failed: %v", err))
		return
	}
	c.Status(http.StatusCreated)
}
