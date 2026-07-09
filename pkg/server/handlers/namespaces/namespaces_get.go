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

// GetNamespace handles the get namespace request
//
//	@Summary	Get namespace
//	@security	BasicAuth
//	@Tags		Namespace
//	@Accept		json
//	@Produce	json
//	@Router		/namespaces/{namespace_id} [get]
//	@Param		namespace_id	path		number	true	"Namespace id"
//	@Success	200				{object}	api.NamespaceItem
//	@Failure	404				{object}	errcode.ErrCode
//	@Failure	500				{object}	errcode.ErrCode
func (h *handler) GetNamespace(c *gin.Context, req *api.GetNamespaceRequest) {
	ctx := c.Request.Context()

	namespaceObj, repoCount, tagCount, err := h.NsSvc.GetNamespace(ctx, req.ID)
	if err != nil {
		if e, ok := errcode.AsType[errcode.ErrCode](err); ok {
			errcode.NewHTTPError(c, e)
			return
		}
		errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError)
		return
	}

	c.JSON(http.StatusOK, api.NamespaceItem{
		ID:              namespaceObj.ID,
		Name:            namespaceObj.Name,
		Description:     namespaceObj.Description,
		Overview:        new(string(namespaceObj.Overview)),
		Visibility:      namespaceObj.Visibility,
		Size:            namespaceObj.Size,
		SizeLimit:       namespaceObj.SizeLimit,
		RepositoryCount: repoCount,
		RepositoryLimit: namespaceObj.RepositoryLimit,
		TagCount:        tagCount,
		TagLimit:        namespaceObj.TagLimit,
		CreatedAt:       time.Unix(0, int64(time.Millisecond)*namespaceObj.CreatedAt).UTC().Format(consts.DefaultTimePattern),
		UpdatedAt:       time.Unix(0, int64(time.Millisecond)*namespaceObj.UpdatedAt).UTC().Format(consts.DefaultTimePattern),
	})
}
