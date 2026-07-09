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
	"github.com/go-sigma/sigma/pkg/server/errcode"
)

// ListNamespaceMembers handles the list namespace members request
//
//	@Summary	List namespace members
//	@security	BasicAuth
//	@Tags		Namespace
//	@Accept		json
//	@Produce	json
//	@Router		/namespaces/{namespace_id}/members/ [get]
//	@Param		limit	query		int64	false	"Limit size"	minimum(10)	maximum(100)	default(10)
//	@Param		page	query		int64	false	"Page number"	minimum(1)	default(1)
//	@Param		sort	query		string	false	"Sort field"
//	@Param		method	query		string	false	"Sort method"	Enums(asc, desc)
//	@Param		name	query		string	false	"Search namespace namespace with name"
//	@Success	200		{object}	api.CommonList{items=[]api.NamespaceMemberItem}
//	@Failure	500		{object}	errcode.ErrCode
func (h *handler) ListNamespaceMembers(c *gin.Context, req *api.ListNamespaceMemberRequest) {
	ctx := c.Request.Context()

	namespaceMemberObjs, total, err := h.NsSvc.ListNamespaceMembers(ctx, req.NamespaceID, req.Name, req.Pagination, req.Sortable)
	if err != nil {
		if e, ok := errcode.AsType[errcode.ErrCode](err); ok {
			errcode.NewHTTPError(c, e)
			return
		}
		errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError)
		return
	}

	var resp = make([]any, 0, len(namespaceMemberObjs))
	for _, namespaceMemberObj := range namespaceMemberObjs {
		resp = append(resp, api.NamespaceMemberItem{
			ID:        namespaceMemberObj.ID,
			Username:  namespaceMemberObj.User.Username,
			UserID:    namespaceMemberObj.User.ID,
			Role:      namespaceMemberObj.Role,
			CreatedAt: time.Unix(0, int64(time.Millisecond)*namespaceMemberObj.CreatedAt).UTC().Format(consts.DefaultTimePattern),
			UpdatedAt: time.Unix(0, int64(time.Millisecond)*namespaceMemberObj.UpdatedAt).UTC().Format(consts.DefaultTimePattern),
		})
	}

	c.JSON(http.StatusOK, api.CommonList{Total: total, Items: resp})
}
