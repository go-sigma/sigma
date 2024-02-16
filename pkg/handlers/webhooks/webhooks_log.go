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
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/xerrors"
)

// LogWebhook ...
//
//	@Summary	Get a webhook log
//	@security	BasicAuth
//	@Tags		Webhook
//	@Accept		json
//	@Produce	json
//	@Router		/webhooks/{webhook_id}/logs/{id} [get]
//	@Param		webhook_id	path		int64	true	"Webhook id"
//	@Param		id			path		int64	true	"Webhook log id"
//	@Success	200			{object}	types.CommonList{items=[]types.WebhookLogItem}
//	@Failure	500			{object}	xerrors.ErrCode
//	@Failure	401			{object}	xerrors.ErrCode
func (h *handler) LogWebhook(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	var req types.GetWebhookLogRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, err.Error())
	}

	webhookService := h.webhookServiceFactory.New()
	_, err = webhookService.Get(ctx, req.WebhookID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Int64("id", req.WebhookID).Msg("Webhook not found")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, fmt.Sprintf("Webhook(%d) not found", req.WebhookID))
		}
		log.Error().Err(err).Int64("id", req.WebhookLogID).Msg("Get webhook failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Get webhook(%d) failed", req.WebhookID))
	}
	webhookLogObj, err := webhookService.GetLog(ctx, req.WebhookLogID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Int64("id", req.WebhookLogID).Msg("Webhook log not found")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, fmt.Sprintf("Webhook log(%d) not found", req.WebhookLogID))
		}
		log.Error().Err(err).Int64("id", req.WebhookLogID).Msg("Get webhook log failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Get webhook log(%d) failed", req.WebhookLogID))
	}
	return c.JSON(http.StatusOK, types.WebhookLogItem{
		ID:         webhookLogObj.ID,
		Event:      webhookLogObj.Event,
		Action:     webhookLogObj.Action,
		StatusCode: webhookLogObj.StatusCode,
		ReqHeader:  string(webhookLogObj.ReqHeader),
		ReqBody:    string(webhookLogObj.ReqBody),
		RespHeader: string(webhookLogObj.RespHeader),
		RespBody:   string(webhookLogObj.RespBody),
		CreatedAt:  time.Unix(0, int64(time.Millisecond)*webhookLogObj.CreatedAt).UTC().Format(consts.DefaultTimePattern),
		UpdatedAt:  time.Unix(0, int64(time.Millisecond)*webhookLogObj.CreatedAt).UTC().Format(consts.DefaultTimePattern),
	})
}
