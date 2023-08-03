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
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"

	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/xerrors"
)

// ListWebhook handles the list webhook request
// @Summary List webhooks
// @security BasicAuth
// @Tags Webhook
// @Accept json
// @Produce json
// @Router /webhooks/ [get]
// @Param limit query int64 false "limit" minimum(10) maximum(100) default(10)
// @Param page query int64 false "page" minimum(1) default(1)
// @Param sort query string false "sort field"
// @Param method query string false "sort method" Enums(asc, desc)
// @Param namespace_id query int64 false "filter by namespace id"
// @Success 200	{object} types.CommonList{items=[]types.WebhookItem}
// @Failure 500 {object} xerrors.ErrCode
// @Failure 401 {object} xerrors.ErrCode
func (h *handlers) ListWebhook(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	var req types.ListWebhookRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, err.Error())
	}

	webhookService := h.webhookServiceFactory.New()
	webhookObjs, total, err := webhookService.List(ctx, req.NamespaceID, req.Pagination, req.Sortable)
	if err != nil {
		log.Error().Err(err).Msg("List webhook failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, "List webhook failed")
	}
	var resp = make([]any, 0, len(webhookObjs))
	for _, webhookObj := range webhookObjs {
		resp = append(resp, types.WebhookItem{
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
			CreatedAt:       webhookObj.CreatedAt.Format(consts.DefaultTimePattern),
			UpdatedAt:       webhookObj.UpdatedAt.Format(consts.DefaultTimePattern),
		})
	}
	return c.JSON(http.StatusOK, types.CommonList{Total: total, Items: resp})
}