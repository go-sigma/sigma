// Copyright 2024 sigma
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

// DeleteWebhookLog handles the delete webhook log request
//
//	@Summary	Delete a webhook log
//	@security	BasicAuth
//	@Tags		Webhook
//	@Accept		json
//	@Produce	json
//	@Router		/webhooks/{webhook_id}/logs/{webhook_log_id} [delete]
//	@Param		webhook_id		path	int64	true	"Webhook id"
//	@Param		webhook_log_id	path	int64	true	"Webhook log id"
//	@Success	204
//	@Failure	500	{object}	errcode.ErrCode
//	@Failure	401	{object}	errcode.ErrCode
func (h *handler) DeleteWebhookLog(c *gin.Context, req *api.DeleteWebhookLogRequest) {
	ctx := c.Request.Context()

	user, ok := utils.GetFromCtx[*models.User](c, consts.ContextUser)
	if !ok {
		slog.Error("get user from context failed")
		errcode.NewHTTPError(c, errcode.HTTPErrCodeUnauthorized)
		return
	}

	webhookObj, err := h.WebhookSvc.GetWebhook(ctx, req.WebhookID)
	if err != nil {
		if e, ok := errcode.AsType[errcode.ErrCode](err); ok {
			errcode.NewHTTPError(c, e)
			return
		}
		errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError, fmt.Sprintf("Get webhook(%s) failed: %v", req.WebhookID, err))
		return
	}

	if webhookObj.NamespaceID == nil {
		if !(user.Role == enums.UserRoleAdmin || user.Role == enums.UserRoleRoot) { // nolint: staticcheck
			errcode.NewHTTPError(c, errcode.HTTPErrCodeUnauthorized, "No permission with this api")
			return
		}
	} else {
		namespaceID := ptr.To(webhookObj.NamespaceID)
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

	err = h.WebhookSvc.DeleteWebhookLog(ctx, user.ID, req.WebhookID, req.WebhookLogID)
	if err != nil {
		if e, ok := errcode.AsType[errcode.ErrCode](err); ok {
			errcode.NewHTTPError(c, e)
			return
		}
		errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError, fmt.Sprintf("Delete webhook log failed: %v", err))
		return
	}
	c.Status(http.StatusNoContent)
}
