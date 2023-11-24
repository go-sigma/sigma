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

// GetWebhook handles the get webhook request
//
//	@Summary	Get a webhook
//	@security	BasicAuth
//	@Tags		Webhook
//	@Accept		json
//	@Produce	json
//	@Router		/webhooks/{id} [get]
//	@Param		id	path		string	true	"Webhook id"
//	@Success	200	{object}	types.CommonList{items=[]types.GetWebhookResponse}
//	@Failure	500	{object}	xerrors.ErrCode
//	@Failure	401	{object}	xerrors.ErrCode
func (h *handler) GetWebhook(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	var req types.GetWebhookRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, err.Error())
	}

	webhookService := h.webhookServiceFactory.New()
	webhookObj, err := webhookService.Get(ctx, req.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Int64("id", req.ID).Msg("Webhook not found")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, fmt.Sprintf("Webhook(%d) not found", req.ID))
		}
		log.Error().Err(err).Msg("Get webhook failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Get namespace(%d) failed", req.ID))
	}

	return c.JSON(http.StatusOK, types.GetWebhookResponse{
		ID:              webhookObj.ID,
		NamespaceID:     webhookObj.NamespaceID,
		Url:             webhookObj.Url,
		Secret:          webhookObj.Secret,
		SslVerify:       webhookObj.SslVerify,
		RetryTimes:      webhookObj.RetryTimes,
		RetryDuration:   webhookObj.RetryDuration,
		Enable:          webhookObj.Enable,
		EventNamespace:  webhookObj.EventNamespace,
		EventRepository: webhookObj.EventRepository,
		EventTag:        webhookObj.EventTag,
		EventArtifact:   webhookObj.EventArtifact,
		EventMember:     webhookObj.EventMember,
		CreatedAt:       time.Unix(0, int64(time.Millisecond)*webhookObj.CreatedAt).UTC().Format(consts.DefaultTimePattern),
		UpdatedAt:       time.Unix(0, int64(time.Millisecond)*webhookObj.CreatedAt).UTC().Format(consts.DefaultTimePattern),
	})
}
