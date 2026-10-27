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

package coderepos

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

// GetBranch get branch by name
//
//	@Summary	Get specific name code repository branch
//	@security	BasicAuth
//	@Tags		CodeRepository
//	@Accept		json
//	@Produce	json
//	@Router		/coderepos/{provider}/repos/{repo_id}/branches/{name} [get]
//	@Param		provider	path		string	true	"code repository provider"
//	@Param		repo_id		path		string	true	"Code repository ID"
//	@Param		name		path		string	true	"Branch name"
//	@Success	200			{object}	api.CodeRepositoryBranchItem
//	@Failure	500			{object}	errcode.ErrCode
func (h *handler) GetBranch(c *gin.Context, req *api.GetCodeRepositoryBranchRequest) {
	ctx := c.Request.Context()

	user, ok := utils.GetFromCtx[*models.User](c, consts.ContextUser)
	if !ok {
		slog.Error("get user from context failed")
		errcode.NewHTTPError(c, errcode.HTTPErrCodeUnauthorized)
		return
	}

	branch, err := h.CodeRepoSvc.GetCodeRepositoryBranch(ctx, user.ID, req.Provider, req.ID, req.Name)
	if err != nil {
		if e, ok := errcode.AsType[errcode.ErrCode](err); ok {
			errcode.NewHTTPError(c, e)
			return
		}
		errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError, fmt.Sprintf("Get branch failed: %v", err))
		return
	}
	c.JSON(http.StatusOK, api.CodeRepositoryBranchItem{Name: branch})
}
