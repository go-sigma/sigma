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

package webhooks

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
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

// PutWebhook handles the put webhook request
//
//	@Summary	Update a webhook
//	@Tags		Webhook
//	@security	BasicAuth
//	@Accept		json
//	@Produce	json
//	@Router		/webhook/{id} [put]
//	@Param		id		path	string					true	"Webhook id"
//	@Param		message	body	api.PutWebhookRequest	true	"Webhook object"
//	@Success	204
//	@Failure	400	{object}	errcode.ErrCode
//	@Failure	404	{object}	errcode.ErrCode
//	@Failure	500	{object}	errcode.ErrCode
func (h *handler) PutWebhook(c *gin.Context, req *api.PutWebhookRequest) {
	ctx := c.Request.Context()

	user, ok := utils.GetFromCtx[*models.User](c, consts.ContextUser)
	if !ok {
		slog.Error("get user from context failed")
		errcode.NewHTTPError(c, errcode.HTTPErrCodeUnauthorized)
		return
	}

	webhookOldObj, err := h.WebhookSvc.GetWebhook(ctx, req.ID)
	if err != nil {
		if e, ok := errcode.AsType[errcode.ErrCode](err); ok {
			errcode.NewHTTPError(c, e)
			return
		}
		errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError, fmt.Sprintf("Get webhook(%s) failed: %v", req.ID, err))
		return
	}

	if webhookOldObj.NamespaceID == nil {
		if !(user.Role == enums.UserRoleAdmin || user.Role == enums.UserRoleRoot) { // nolint: staticcheck
			errcode.NewHTTPError(c, errcode.HTTPErrCodeUnauthorized, "No permission with this api")
			return
		}
	} else {
		namespaceID := ptr.To(webhookOldObj.NamespaceID)
		authChecked, err := h.Authorizer.Namespace(ctx, *user, namespaceID, enums.AuthManage)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				slog.Error("namespace not found", "err", err, "NamespaceID", namespaceID)
				errcode.NewHTTPError(c, errcode.HTTPErrCodeNotFound, fmt.Sprintf("Namespace(%s) not found: %v", namespaceID, err))
				return
			}
			slog.Error("namespace find failed", "err", err, "NamespaceID", namespaceID)
			errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError, fmt.Sprintf("Namespace(%s) find failed: %v", namespaceID, err))
			return
		}
		if !authChecked {
			slog.Error("auth check failed", "UserID", user.ID, "NamespaceID", namespaceID)
			errcode.NewHTTPError(c, errcode.HTTPErrCodeUnauthorized, "No permission with this api")
			return
		}
	}

	err = h.WebhookSvc.UpdateWebhook(ctx, user.ID, req.ID, *req)
	if err != nil {
		if e, ok := errcode.AsType[errcode.ErrCode](err); ok {
			errcode.NewHTTPError(c, e)
			return
		}
		errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError, fmt.Sprintf("Update webhook failed: %v", err))
		return
	}
	c.Status(http.StatusNoContent)
}
