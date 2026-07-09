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
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/server/errcode"
	"github.com/go-sigma/sigma/pkg/utils"
)

// DeleteNamespace handles the delete namespace request
//
//	@Summary	Delete namespace
//	@security	BasicAuth
//	@Tags		Namespace
//	@Accept		json
//	@Produce	json
//	@Router		/namespaces/{namespace_id} [delete]
//	@Param		namespace_id	path	number	true	"Namespace id"
//	@Success	204
//	@Failure	500	{object}	errcode.ErrCode
func (h *handler) DeleteNamespace(c *gin.Context, req *api.DeleteNamespaceRequest) {
	ctx := c.Request.Context()

	user, ok := utils.GetFromCtx[*models.User](c, consts.ContextUser)
	if !ok {
		slog.Error("get user from context failed")
		errcode.NewHTTPError(c, errcode.HTTPErrCodeUnauthorized)
		return
	}

	authChecked, err := h.Authorizer.Namespace(ctx, *user, req.ID, enums.AuthManage)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Error("resource not found", "err", errors.New(utils.UnwrapJoinedErrors(err)), "NamespaceID", req.ID)
			errcode.NewHTTPError(c, errcode.HTTPErrCodeNotFound, utils.UnwrapJoinedErrors(err))
			return
		}
		slog.Error("get resource failed", "err", errors.New(utils.UnwrapJoinedErrors(err)), "NamespaceID", req.ID, "err", err)
		errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError, utils.UnwrapJoinedErrors(err))
		return
	}
	if !authChecked {
		slog.Error("auth check failed", "UserID", user.ID, "NamespaceID", req.ID)
		errcode.NewHTTPError(c, errcode.HTTPErrCodeUnauthorized, "No permission with this api")
		return
	}

	err = h.NsSvc.DeleteNamespace(ctx, user.ID, req.ID)
	if err != nil {
		if e, ok := errcode.AsType[errcode.ErrCode](err); ok {
			errcode.NewHTTPError(c, e)
			return
		}
		errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError)
		return
	}

	c.Status(http.StatusNoContent)
}
