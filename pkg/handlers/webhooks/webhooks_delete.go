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
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/xerrors"
)

// DeleteWebhook handles the delete webhook request
// @Summary Delete a webhook
// @security BasicAuth
// @Tags Webhook
// @Accept json
// @Produce json
// @Router /webhooks/{id} [get]
// @Param id path string true "Webhook id"
// @Success 204
// @Failure 500 {object} xerrors.ErrCode
// @Failure 401 {object} xerrors.ErrCode
func (h *handlers) DeleteWebhook(c echo.Context) error {
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

	var req types.DeleteWebhookRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, err.Error())
	}

	webhookService := h.webhookServiceFactory.New()
	webhookOldObj, err := webhookService.Get(ctx, req.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Int64("id", req.ID).Msg("Webhook not found")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, fmt.Sprintf("Webhook(%d) not found", req.ID))
		}
		log.Error().Err(err).Msg("Get webhook failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Get namespace(%d) failed", req.ID))
	}

	err = query.Q.Transaction(func(tx *query.Query) error {
		webhookService := h.webhookServiceFactory.New(tx)
		err = webhookService.DeleteByID(ctx, req.ID)
		if err != nil {
			log.Error().Err(err).Msg("Delete webhook failed")
			return xerrors.HTTPErrCodeInternalError.Detail("Delete webhook failed")
		}
		auditService := h.auditServiceFactory.New(tx)
		err = auditService.Create(ctx, &models.Audit{
			UserID:       user.ID,
			NamespaceID:  webhookOldObj.NamespaceID,
			Action:       enums.AuditActionCreate,
			ResourceType: enums.AuditResourceTypeWebhook,
			Resource:     strconv.FormatInt(webhookOldObj.ID, 10),
			BeforeRaw:    utils.MustMarshal(webhookOldObj),
			ReqRaw:       utils.MustMarshal(req),
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
