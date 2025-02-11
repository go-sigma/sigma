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
	"github.com/go-sigma/sigma/pkg/utils/ptr"
	"github.com/go-sigma/sigma/pkg/xerrors"
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
//	@Param		message	body	types.PutWebhookRequest	true	"Webhook object"
//	@Success	204
//	@Failure	400	{object}	xerrors.ErrCode
//	@Failure	404	{object}	xerrors.ErrCode
//	@Failure	500	{object}	xerrors.ErrCode
func (h *handler) PutWebhook(c echo.Context) error {
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

	var req types.PutWebhookRequest
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

	if webhookOldObj.NamespaceID == nil {
		if !(user.Role == enums.UserRoleAdmin || user.Role == enums.UserRoleRoot) {
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeUnauthorized, "No permission with this api")
		}
	} else {
		namespaceID := ptr.To(webhookOldObj.NamespaceID)
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

	updates := make(map[string]any, 15)
	if req.Url != nil {
		updates[query.Webhook.URL.ColumnName().String()] = ptr.To(req.Url)
	}
	if req.Secret != nil {
		updates[query.Webhook.Secret.ColumnName().String()] = ptr.To(req.Secret)
	}
	if req.SslVerify != nil {
		updates[query.Webhook.SslVerify.ColumnName().String()] = ptr.To(req.SslVerify)
	}
	if req.RetryTimes != nil {
		updates[query.Webhook.RetryTimes.ColumnName().String()] = ptr.To(req.RetryTimes)
	}
	if req.RetryDuration != nil {
		updates[query.Webhook.RetryDuration.ColumnName().String()] = ptr.To(req.RetryDuration)
	}
	if req.Enable != nil {
		updates[query.Webhook.Enable.ColumnName().String()] = ptr.To(req.Enable)
	}
	if req.EventNamespace != nil {
		updates[query.Webhook.EventNamespace.ColumnName().String()] = ptr.To(req.EventNamespace)
	}
	if req.EventRepository != nil {
		updates[query.Webhook.EventRepository.ColumnName().String()] = ptr.To(req.EventRepository)
	}
	if req.EventTag != nil {
		updates[query.Webhook.EventTag.ColumnName().String()] = ptr.To(req.EventTag)
	}
	if req.EventArtifact != nil {
		updates[query.Webhook.EventArtifact.ColumnName().String()] = ptr.To(req.EventArtifact)
	}
	if req.EventMember != nil {
		updates[query.Webhook.EventMember.ColumnName().String()] = ptr.To(req.EventMember)
	}
	if req.EventMember != nil {
		updates[query.Webhook.EventDaemonTaskGc.ColumnName().String()] = ptr.To(req.EventDaemonTaskGc)
	}

	err = query.Q.Transaction(func(tx *query.Query) error {
		webhookService := h.webhookServiceFactory.New(tx)
		err = webhookService.UpdateByID(ctx, req.ID, updates)
		if err != nil {
			log.Error().Err(err).Msg("Update webhook failed")
			return xerrors.HTTPErrCodeInternalError.Detail("Update webhook failed")
		}
		auditService := h.auditServiceFactory.New(tx)
		err = auditService.Create(ctx, &models.Audit{
			UserID:       user.ID,
			NamespaceID:  webhookOldObj.NamespaceID,
			Action:       enums.AuditActionUpdate,
			ResourceType: enums.AuditResourceTypeWebhook,
			Resource:     strconv.FormatInt(webhookOldObj.ID, 10),
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
