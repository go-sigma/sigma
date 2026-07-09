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
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/server/errcode"
	"github.com/go-sigma/sigma/pkg/utils"
)

// Get get code repository by id
//
//	@Summary	Get code repository by id
//	@security	BasicAuth
//	@Tags		CodeRepository
//	@Accept		json
//	@Produce	json
//	@Router		/coderepos/{provider}/repos/{id} [get]
//	@Param		provider	path		string	true	"Search code repository with provider"
//	@Param		id			path		string	true	"Code repository id"
//	@Success	200			{object}	api.CodeRepositoryItem
//	@Failure	500			{object}	errcode.ErrCode
func (h *handler) Get(c *gin.Context, req *api.GetCodeRepositoryRequest) {
	ctx := c.Request.Context()

	user, ok := utils.GetFromCtx[*models.User](c, consts.ContextUser)
	if !ok {
		slog.Error("get user from context failed")
		errcode.NewHTTPError(c, errcode.HTTPErrCodeUnauthorized)
		return
	}

	codeRepositoryObj, ownerObjs, err := h.CodeRepoSvc.GetCodeRepository(ctx, user.ID, req.Provider, req.ID)
	if err != nil {
		if e, ok := errcode.AsType[errcode.ErrCode](err); ok {
			errcode.NewHTTPError(c, e)
			return
		}
		errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError)
		return
	}

	c.JSON(http.StatusOK, api.CodeRepositoryItem{
		ID:           codeRepositoryObj.ID,
		RepositoryID: codeRepositoryObj.RepositoryID,
		Provider:     enums.ScmProvider(req.Provider),
		Name:         codeRepositoryObj.Name,
		OwnerID:      getOwnerID(ownerObjs, codeRepositoryObj.Owner),
		Owner:        codeRepositoryObj.Owner,
		IsOrg:        codeRepositoryObj.IsOrg,
		CloneUrl:     codeRepositoryObj.CloneUrl,
		SshUrl:       codeRepositoryObj.SshUrl,
		OciRepoCount: codeRepositoryObj.OciRepoCount,
		CreatedAt:    time.Unix(0, int64(time.Millisecond)*codeRepositoryObj.CreatedAt).UTC().Format(consts.DefaultTimePattern),
		UpdatedAt:    time.Unix(0, int64(time.Millisecond)*codeRepositoryObj.CreatedAt).UTC().Format(consts.DefaultTimePattern),
	})
}
