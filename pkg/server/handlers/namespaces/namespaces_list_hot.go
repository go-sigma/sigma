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
	"time"

	"github.com/gin-gonic/gin"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/server/errcode"
	"github.com/go-sigma/sigma/pkg/utils"
)

// HotNamespace handles the hot namespace request
//
//	@Summary	Hot namespace
//	@security	BasicAuth
//	@Tags		Namespace
//	@Accept		json
//	@Produce	json
//	@Router		/namespaces/hot [get]
//	@Success	200	{object}	api.CommonList{items=[]api.NamespaceItem}
//	@Failure	500	{object}	errcode.ErrCode
//	@Failure	401	{object}	errcode.ErrCode
func (h *handler) HotNamespace(c *gin.Context) {
	ctx := c.Request.Context()

	user, ok := utils.GetFromCtx[*models.User](c, consts.ContextUser)
	if !ok {
		slog.Error("get user from context failed")
		errcode.NewHTTPError(c, errcode.HTTPErrCodeUnauthorized)
		return
	}

	namespaceObjs, err := h.NsSvc.HotNamespaces(ctx, user.ID)
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
			CreatedAt:       time.Unix(namespaceObj.CreatedAt, 0).UTC().Format(consts.DefaultTimePattern),
			UpdatedAt:       time.Unix(namespaceObj.UpdatedAt, 0).UTC().Format(consts.DefaultTimePattern),
		})
	}

	c.JSON(http.StatusOK, api.CommonList{Total: int64(len(namespaceObjs)), Items: resp})
}
