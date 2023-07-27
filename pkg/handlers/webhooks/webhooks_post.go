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
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"

	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/xerrors"
)

// PostWebhook handles the post webhook request
// @Summary Create a webhook
// @Tags Webhook
// @security BasicAuth
// @Accept json
// @Produce json
// @Router /webhooks [post]
// @Param namespace_id query int64 false "create webhook for namespace"
// @Param message body types.PostWebhookRequestSwagger true "Webhook object"
// @Success 201
// @Failure 400 {object} xerrors.ErrCode
// @Failure 404 {object} xerrors.ErrCode
// @Failure 500 {object} xerrors.ErrCode
func (h *handlers) PostWebhook(c echo.Context) error {
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

	var req types.PostWebhookRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, err.Error())
	}

	webhookService := h.webhookServiceFactory.New()
	_, total, err := webhookService.List(ctx, req.NamespaceID, types.Pagination{}, types.Sortable{})
	if err != nil {
		log.Error().Err(err).Msg("Get webhook count failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, err.Error())
	}
	if total > 5 {
		log.Error().Int64("total", total).Msg("Reached the maximum webhooks")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, "Reached the maximum webhooks")
	}

	err = query.Q.Transaction(func(tx *query.Query) error {
		webhookService := h.webhookServiceFactory.New(tx)
		webhookObj := &models.Webhook{
			NamespaceID:     req.NamespaceID,
			Url:             req.Url,
			Secret:          req.Secret,
			SslVerify:       req.SslVerify,
			RetryTimes:      req.RetryTimes,
			RetryDuration:   req.RetryDuration,
			Enable:          req.Enable,
			EventNamespace:  req.EventNamespace,
			EventRepository: req.EventRepository,
			EventTag:        req.EventTag,
			EventMember:     req.EventMember,
			EventArtifact:   req.EventArtifact,
		}
		err = webhookService.Create(ctx, webhookObj)
		if err != nil {
			log.Error().Err(err).Msg("Create webhook failed")
			return xerrors.HTTPErrCodeInternalError.Detail("Create webhook failed")
		}
		auditService := h.auditServiceFactory.New(tx)
		err = auditService.Create(ctx, &models.Audit{
			UserID:       user.ID,
			NamespaceID:  req.NamespaceID,
			Action:       enums.AuditActionCreate,
			ResourceType: enums.AuditResourceTypeWebhook,
			Resource:     strconv.FormatInt(webhookObj.ID, 10),
			ReqRaw:       utils.MustMarshal(webhookObj),
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
	return c.NoContent(http.StatusCreated)
}
