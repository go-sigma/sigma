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

// List list all of the code repositories
//
//	@Summary	List code repositories
//	@security	BasicAuth
//	@Tags		CodeRepository
//	@Accept		json
//	@Produce	json
//	@Router		/coderepos/{provider} [get]
//	@Param		limit		query		int64	false	"Limit size"	minimum(10)	maximum(100)	default(10)
//	@Param		page		query		int64	false	"Page number"	minimum(1)	default(1)
//	@Param		sort		query		string	false	"Sort field"
//	@Param		method		query		string	false	"Sort method"	Enums(asc, desc)
//	@Param		name		query		string	false	"Search code repository with name"
//	@Param		owner		query		string	false	"Search code repository with owner"
//	@Param		provider	path		string	true	"search code repository with provider"
//	@Success	200			{object}	api.CommonList{items=[]api.CodeRepositoryItem}
//	@Failure	500			{object}	errcode.ErrCode
func (h *handler) List(c *gin.Context, req *api.ListCodeRepositoryRequest) {
	ctx := c.Request.Context()

	user, ok := utils.GetFromCtx[*models.User](c, consts.ContextUser)
	if !ok {
		slog.Error("get user from context failed")
		errcode.NewHTTPError(c, errcode.HTTPErrCodeUnauthorized)
		return
	}

	codeRepositoryObjs, ownerObjs, total, err := h.CodeRepoSvc.ListCodeRepositories(ctx, user.ID, req.Provider, req.Owner, req.Name, req.Pagination, req.Sortable)
	if err != nil {
		if e, ok := errcode.AsType[errcode.ErrCode](err); ok {
			errcode.NewHTTPError(c, e)
			return
		}
		errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError)
		return
	}

	resp := make([]any, 0, len(codeRepositoryObjs))
	for _, codeRepositoryObj := range codeRepositoryObjs {
		resp = append(resp, api.CodeRepositoryItem{
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

	c.JSON(http.StatusOK, api.CommonList{Total: total, Items: resp})
}

func getOwnerID(ownerObjs []*models.CodeRepositoryOwner, owner string) string {
	for _, ownerObj := range ownerObjs {
		if ownerObj.Owner == owner {
			return ownerObj.ID
		}
	}
	return ""
}
