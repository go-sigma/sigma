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
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/server/errcode"
	sigmatelemetry "github.com/go-sigma/sigma/pkg/telemetry"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

// ListWebhookLogs handles the list webhook logs request
//
//	@Summary	List webhook logs
//	@security	BasicAuth
//	@Tags		Webhook
//	@Accept		json
//	@Produce	json
//	@Router		/webhooks/{webhook_id}/logs [get]
//	@Param		webhook_id	path		int64	true	"Webhook ID"
//	@Param		limit		query		int64	false	"limit"	minimum(10)	maximum(100)	default(10)
//	@Param		page		query		int64	false	"page"	minimum(1)	default(1)
//	@Param		sort		query		string	false	"sort field"
//	@Param		method		query		string	false	"sort method"	Enums(asc, desc)
//	@Success	200			{object}	api.CommonList{items=[]api.WebhookItem}
//	@Failure	500			{object}	errcode.ErrCode
//	@Failure	401			{object}	errcode.ErrCode
func (h *handler) ListWebhookLogs(c *gin.Context, req *api.ListWebhookLogRequest) {
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

	webhookLogObjs, total, err := h.WebhookSvc.ListWebhookLogs(ctx, req.WebhookID, req.Pagination, req.Sortable)
	if err != nil {
		if e, ok := errcode.AsType[errcode.ErrCode](err); ok {
			errcode.NewHTTPError(c, e)
			return
		}
		errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError, fmt.Sprintf("List webhook logs failed: %v", err))
		return
	}
	var resp = make([]any, 0, len(webhookLogObjs))
	for _, webhookLogObj := range webhookLogObjs {
		resp = append(resp, api.WebhookLogItem{
			ID:           webhookLogObj.ID,
			ResourceType: webhookLogObj.ResourceType,
			Action:       webhookLogObj.Action,
			StatusCode:   webhookLogObj.StatusCode,
			TraceID:      sigmatelemetry.TraceIDFromCarrierBytes(webhookLogObj.TraceContext),
			TraceContext: string(webhookLogObj.TraceContext),
			ReqHeader:    string(webhookLogObj.ReqHeader),
			ReqBody:      string(webhookLogObj.ReqBody),
			RespHeader:   string(webhookLogObj.RespHeader),
			RespBody:     string(webhookLogObj.RespBody),
			CreatedAt:    time.Unix(0, int64(time.Millisecond)*webhookLogObj.CreatedAt).UTC().Format(consts.DefaultTimePattern),
			UpdatedAt:    time.Unix(0, int64(time.Millisecond)*webhookLogObj.CreatedAt).UTC().Format(consts.DefaultTimePattern),
		})
	}
	c.JSON(http.StatusOK, api.CommonList{Total: total, Items: resp})
}
