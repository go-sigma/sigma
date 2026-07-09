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
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/server/errcode"
)

// ListBranches list all of the branches
//
//	@Summary	List code repository branches
//	@security	BasicAuth
//	@Tags		CodeRepository
//	@Accept		json
//	@Produce	json
//	@Router		/{provider}/repos/coderepos/{id}/branches [get]
//	@Param		provider	path		string	true	"code repository provider"
//	@Param		id			path		string	true	"Code repository id"
//	@Success	200			{object}	api.CommonList{items=[]api.CodeRepositoryBranchItem}
//	@Failure	500			{object}	errcode.ErrCode
func (h *handler) ListBranches(c *gin.Context, req *api.ListCodeRepositoryBranchesRequest) {
	ctx := c.Request.Context()

	branchObjs, total, err := h.CodeRepoSvc.ListCodeRepositoryBranches(ctx, req.ID)
	if err != nil {
		if e, ok := errcode.AsType[errcode.ErrCode](err); ok {
			errcode.NewHTTPError(c, e)
			return
		}
		errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError, fmt.Sprintf("List branches failed: %v", err))
		return
	}
	resp := make([]any, 0, len(branchObjs))
	for _, branchObj := range branchObjs {
		resp = append(resp, api.CodeRepositoryBranchItem{
			ID:        branchObj.ID,
			Name:      branchObj.Name,
			CreatedAt: time.Unix(0, int64(time.Millisecond)*branchObj.CreatedAt).UTC().Format(consts.DefaultTimePattern),
			UpdatedAt: time.Unix(0, int64(time.Millisecond)*branchObj.CreatedAt).UTC().Format(consts.DefaultTimePattern),
		})
	}

	c.JSON(http.StatusOK, api.CommonList{Total: total, Items: resp})
}
