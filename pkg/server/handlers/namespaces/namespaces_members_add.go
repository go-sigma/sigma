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
	"fmt"
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

// AddNamespaceMember handles the add namespace member request
//
//	@Summary	Add namespace member
//	@Tags		Namespace
//	@Accept		json
//	@Produce	json
//	@Router		/namespaces/members/ [post]
//	@Param		message	body	api.AddNamespaceMemberRequest	true	"Member object"
//	@security	BasicAuth
//	@Success	201	{object}	api.AddNamespaceMemberResponse
//	@Failure	400	{object}	errcode.ErrCode
//	@Failure	500	{object}	errcode.ErrCode
func (h *handler) AddNamespaceMember(c *gin.Context, req *api.AddNamespaceMemberRequest) {
	ctx := c.Request.Context()

	user, ok := utils.GetFromCtx[*models.User](c, consts.ContextUser)
	if !ok {
		slog.Error("get user from context failed")
		errcode.NewHTTPError(c, errcode.HTTPErrCodeUnauthorized)
		return
	}

	authChecked, err := h.Authorizer.Namespace(ctx, *user, req.NamespaceID, enums.AuthAdmin)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Error("namespace not found", "err", err, "id", req.NamespaceID)
			errcode.NewHTTPError(c, errcode.HTTPErrCodeNotFound, fmt.Sprintf("Namespace(%s) not found", req.NamespaceID))
			return
		}
		slog.Error("get namespace failed", "err", err)
		errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError, fmt.Sprintf("Get namespace(%s) failed", req.NamespaceID))
		return
	}
	if !authChecked {
		slog.Error("auth check failed", "UserID", user.ID, "NamespaceID", req.NamespaceID)
		errcode.NewHTTPError(c, errcode.HTTPErrCodeUnauthorized, "No permission with this api")
		return
	}

	namespaceMemberObj, err := h.NsSvc.AddNamespaceMember(ctx, user.ID, req.NamespaceID, req.UserID, req.Role)
	if err != nil {
		if e, ok := errcode.AsType[errcode.ErrCode](err); ok {
			errcode.NewHTTPError(c, e)
			return
		}
		errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError)
		return
	}
	c.JSON(http.StatusCreated, api.AddNamespaceMemberResponse{
		ID: namespaceMemberObj.ID,
	})
}
