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
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/server/errcode"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils"
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
func (h *handler) DeleteWebhookLog(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	iuser := c.Get(consts.ContextUser)
	if iuser == nil {
		log.Error().Msg("Get user from header failed")
		return errcode.NewHTTPError(c, errcode.HTTPErrCodeUnauthorized)
	}
	user, ok := iuser.(*models.User)
	if !ok {
		log.Error().Msg("Convert user from header failed")
		return errcode.NewHTTPError(c, errcode.HTTPErrCodeUnauthorized)
	}

	var req types.DeleteWebhookLogRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return errcode.NewHTTPError(c, errcode.HTTPErrCodeBadRequest, err.Error())
	}

	webhookService := h.WebhookServiceFactory.New()
	webhookObj, err := webhookService.Get(ctx, req.WebhookID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Int64("WebhookID", req.WebhookID).Int64("WebhookLogID", req.WebhookLogID).Msg("Webhook not found")
			return errcode.NewHTTPError(c, errcode.HTTPErrCodeNotFound, fmt.Sprintf("Webhook(%d) not found", req.WebhookID))
		}
		log.Error().Err(err).Int64("WebhookID", req.WebhookID).Int64("WebhookLogID", req.WebhookLogID).Msg("Get webhook failed")
		return errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError, fmt.Sprintf("Get webhook(%d) failed", req.WebhookID))
	}

	if err := h.checkWebhookAuth(c, user, webhookObj, enums.AuthManage); err != nil {
		return err
	}

	err = query.Q.Transaction(func(tx *query.Query) error {
		webhookService := h.WebhookServiceFactory.New(tx)
		err = webhookService.DeleteByID(ctx, req.WebhookLogID)
		if err != nil {
			log.Error().Err(err).Msg("Create webhook failed")
			return errcode.HTTPErrCodeInternalError.Detail("Create webhook failed")
		}
		auditService := h.AuditServiceFactory.New(tx)
		err = auditService.Create(ctx, &models.Audit{
			UserID:       user.ID,
			NamespaceID:  webhookObj.NamespaceID,
			Action:       enums.AuditActionDelete,
			ResourceType: enums.AuditResourceTypeWebhook,
			Resource:     strconv.FormatInt(webhookObj.ID, 10),
			ReqRaw:       utils.MustMarshal(webhookObj),
		})
		if err != nil {
			log.Error().Err(err).Msg("Create audit failed")
			return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Create audit failed: %v", err))
		}
		return nil
	})
	if err != nil {
		return errcode.NewHTTPError(c, err.(errcode.ErrCode))
	}
	return c.NoContent(http.StatusNoContent)
}
