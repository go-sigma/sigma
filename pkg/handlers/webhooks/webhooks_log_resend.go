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
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/modules/workq/definition"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
	"github.com/go-sigma/sigma/pkg/xerrors"
)

// GetWebhookLogResend handles the resend webhook log request
//
//	@Summary	Resend a webhook log
//	@security	BasicAuth
//	@Tags		Webhook
//	@Accept		json
//	@Produce	json
//	@Router		/webhooks/{webhook_id}/logs/{webhook_log_id}/resend [get]
//	@Param		webhook_id		path	int64	true	"Webhook id"
//	@Param		webhook_log_id	path	int64	true	"Webhook log id"
//	@Success	204
//	@Failure	500	{object}	xerrors.ErrCode
//	@Failure	401	{object}	xerrors.ErrCode
func (h *handler) GetWebhookLogResend(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	iuser := c.Get(consts.ContextUser)
	if iuser == nil {
		log.Error().Msg("Get user from header failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeUnauthorized)
	}
	user, ok := iuser.(*models.User)
	if !ok {
		log.Error().Msg("Convert user from header failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeUnauthorized)
	}

	var req types.GetWebhookLogResendRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, err.Error())
	}

	webhookService := h.webhookServiceFactory.New()
	webhookLogObj, err := webhookService.GetLog(ctx, req.WebhookLogID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Int64("WebhookID", req.WebhookID).Int64("WebhookLogID", req.WebhookLogID).Msg("Webhook not found")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, fmt.Sprintf("Webhook log(%d) not found", req.WebhookLogID))
		}
		log.Error().Err(err).Int64("WebhookID", req.WebhookID).Int64("WebhookLogID", req.WebhookLogID).Msg("Get webhook failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Get webhook log(%d) failed", req.WebhookLogID))
	}

	if webhookLogObj.Webhook.NamespaceID == nil {
		if !(user.Role == enums.UserRoleAdmin || user.Role == enums.UserRoleRoot) {
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeUnauthorized, "No permission with this api")
		}
	} else {
		namespaceID := ptr.To(webhookLogObj.Webhook.NamespaceID)
		authChecked, err := h.authServiceFactory.New().Namespace(ptr.To(user), namespaceID, enums.AuthManage)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				log.Error().Err(err).Int64("NamespaceID", namespaceID).Msg("Namespace not found")
				return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, fmt.Sprintf("Namespace(%d) not found: %v", namespaceID, err))
			}
			log.Error().Err(err).Int64("NamespaceID", namespaceID).Msg("Namespace find failed")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Namespace(%d) find failed: %v", namespaceID, err))
		}
		if !authChecked {
			log.Error().Int64("UserID", user.ID).Int64("NamespaceID", namespaceID).Msg("Auth check failed")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeUnauthorized, "No permission with this api")
		}
	}

	err = query.Q.Transaction(func(tx *query.Query) error {
		err := h.producerClient.Produce(ctx, enums.DaemonWebhook.String(), types.DaemonWebhookPayload{
			NamespaceID:  webhookLogObj.Webhook.NamespaceID,
			WebhookID:    webhookLogObj.Webhook.ID,
			WebhookLogID: ptr.Of(req.WebhookLogID),
			Type:         enums.WebhookTypeResend,
		}, definition.ProducerOption{Tx: tx})
		if err != nil {
			log.Error().Err(err).Msg("Webhook event produce failed")
			return xerrors.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Webhook event produce failed: %v", err))
		}
		auditService := h.auditServiceFactory.New(tx)
		err = auditService.Create(ctx, &models.Audit{
			UserID:       user.ID,
			NamespaceID:  webhookLogObj.Webhook.NamespaceID,
			Action:       enums.AuditActionDelete,
			ResourceType: enums.AuditResourceTypeWebhook,
			Resource:     strconv.FormatInt(webhookLogObj.Webhook.ID, 10),
			ReqRaw:       utils.MustMarshal(webhookLogObj.Webhook),
		})
		if err != nil {
			log.Error().Err(err).Msg("Create audit failed")
			return xerrors.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Create audit failed: %v", err))
		}
		return nil
	})
	if err != nil {
		return xerrors.NewHTTPError(c, err.(xerrors.ErrCode))
	}
	return c.NoContent(http.StatusNoContent)
}
