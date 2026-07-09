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
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/server/errcode"
	"github.com/go-sigma/sigma/pkg/utils"
)

// ListNamespaces handles the list namespace request
//
//	@Summary	List namespace
//	@security	BasicAuth
//	@Tags		Namespace
//	@Accept		json
//	@Produce	json
//	@Router		/namespaces/ [get]
//	@Param		limit	query		int64	false	"Limit size"	minimum(10)	maximum(100)	default(10)
//	@Param		page	query		int64	false	"Page number"	minimum(1)	default(1)
//	@Param		sort	query		string	false	"Sort field"
//	@Param		method	query		string	false	"Sort method"	Enums(asc, desc)
//	@Param		name	query		string	false	"Search namespace with name"
//	@Success	200		{object}	api.CommonList{items=[]api.NamespaceItem}
//	@Failure	500		{object}	errcode.ErrCode
func (h *handler) ListNamespaces(c *gin.Context, req *api.ListNamespaceRequest) {
	ctx := c.Request.Context()

	user, ok := utils.GetFromCtx[*models.User](c, consts.ContextUser)
	if !ok {
		user = &models.User{}
	}

	namespaceObjs, total, err := h.NsSvc.ListNamespaces(ctx, user.ID, req.Name, req.Pagination, req.Sortable)
	if err != nil {
		if e, ok := errcode.AsType[errcode.ErrCode](err); ok {
			errcode.NewHTTPError(c, e)
			return
		}
		errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError)
		return
	}

	var resp = make([]any, 0, len(namespaceObjs))
	for _, namespaceObj := range namespaceObjs {
		resp = append(resp, api.NamespaceItem{
			ID:              namespaceObj.ID,
			Name:            namespaceObj.Name,
			Description:     namespaceObj.Description,
			Visibility:      namespaceObj.Visibility,
			Size:            namespaceObj.Size,
			SizeLimit:       namespaceObj.SizeLimit,
			RepositoryLimit: namespaceObj.RepositoryLimit,
			RepositoryCount: namespaceObj.RepositoryCount,
			TagLimit:        namespaceObj.TagLimit,
			TagCount:        namespaceObj.TagCount,
			CreatedAt:       time.Unix(0, int64(time.Millisecond)*namespaceObj.CreatedAt).UTC().Format(consts.DefaultTimePattern),
			UpdatedAt:       time.Unix(0, int64(time.Millisecond)*namespaceObj.UpdatedAt).UTC().Format(consts.DefaultTimePattern),
		})
	}

	c.JSON(http.StatusOK, api.CommonList{Total: total, Items: resp})
}
