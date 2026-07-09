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

// GetNamespaceMemberSelf handles the get self namespace member info request
//
//	@Summary	Get self namespace member info
//	@security	BasicAuth
//	@Tags		Namespace
//	@Accept		json
//	@Produce	json
//	@Router		/namespaces/{namespace_id}/members/self [get]
//	@Param		namespace_id	path		number	true	"Namespace id"
//	@Success	200				{object}	api.NamespaceMemberItem
//	@Failure	404				{object}	errcode.ErrCode
//	@Failure	500				{object}	errcode.ErrCode
func (h *handler) GetNamespaceMemberSelf(c *gin.Context, req *api.GetNamespaceMemberSelfRequest) {
	ctx := c.Request.Context()

	user, ok := utils.GetFromCtx[*models.User](c, consts.ContextUser)
	if !ok {
		slog.Error("get user from context failed")
		errcode.NewHTTPError(c, errcode.HTTPErrCodeUnauthorized)
		return
	}

	namespaceMemberObj, err := h.NsSvc.GetNamespaceMember(ctx, req.NamespaceID, user.ID)
	if err != nil {
		if e, ok := errcode.AsType[errcode.ErrCode](err); ok {
			errcode.NewHTTPError(c, e)
			return
		}
		errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError)
		return
	}

	c.JSON(http.StatusOK, api.NamespaceMemberItem{
		ID:        namespaceMemberObj.ID,
		Username:  user.Username,
		UserID:    user.ID,
		Role:      namespaceMemberObj.Role,
		CreatedAt: time.Unix(0, int64(time.Millisecond)*namespaceMemberObj.CreatedAt).UTC().Format(consts.DefaultTimePattern),
		UpdatedAt: time.Unix(0, int64(time.Millisecond)*namespaceMemberObj.UpdatedAt).UTC().Format(consts.DefaultTimePattern),
	})
}
