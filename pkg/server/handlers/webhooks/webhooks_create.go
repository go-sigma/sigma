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
	"strings"

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
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

// PostWebhook handles the post webhook request
//
//	@Summary	Create a webhook
//	@Tags		Webhook
//	@security	BasicAuth
//	@Accept		json
//	@Produce	json
//	@Router		/webhooks/ [post]
//	@Param		message	body	types.PostWebhookRequest	true	"Webhook object"
//	@Success	201
//	@Failure	400	{object}	errcode.ErrCode
//	@Failure	404	{object}	errcode.ErrCode
//	@Failure	500	{object}	errcode.ErrCode
func (h *handler) PostWebhook(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	user, needRet, err := utils.GetUserFromCtx(c)
	if err != nil {
		return err
	}
	if needRet {
		return nil
	}

	var req types.PostWebhookRequest
	err = utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return errcode.NewHTTPError(c, errcode.HTTPErrCodeBadRequest, err.Error())
	}

	if req.NamespaceID == nil {
		if !(user.Role == enums.UserRoleAdmin || user.Role == enums.UserRoleRoot) { // nolint: staticcheck
			return errcode.NewHTTPError(c, errcode.HTTPErrCodeUnauthorized, "No permission with this api")
		}
	} else {
		namespaceID := ptr.To(req.NamespaceID)
		authChecked, err := h.AuthServiceFactory.New().Namespace(ptr.To(user), namespaceID, enums.AuthManage)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				log.Error().Err(err).Int64("NamespaceID", namespaceID).Msg("Namespace not found")
				return errcode.NewHTTPError(c, errcode.HTTPErrCodeNotFound, fmt.Sprintf("Namespace(%d) not found: %v", namespaceID, err))
			}
			log.Error().Err(err).Int64("NamespaceID", namespaceID).Msg("Namespace find failed")
			return errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError, fmt.Sprintf("Namespace(%d) find failed: %v", namespaceID, err))
		}
		if !authChecked {
			log.Error().Int64("UserID", user.ID).Int64("NamespaceID", namespaceID).Msg("Auth check failed")
			return errcode.NewHTTPError(c, errcode.HTTPErrCodeUnauthorized, "No permission with this api")
		}
	}

	err = h.PostWebhookValidate(req)
	if err != nil {
		return err
	}

	webhookService := h.WebhookServiceFactory.New()
	_, total, err := webhookService.List(ctx, req.NamespaceID, types.Pagination{}, types.Sortable{})
	if err != nil {
		log.Error().Err(err).Msg("Get webhook count failed")
		return errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError, err.Error())
	}
	if total > consts.MaxWebhooks {
		log.Error().Int64("total", total).Msg("Reached the maximum webhooks")
		return errcode.NewHTTPError(c, errcode.HTTPErrCodeBadRequest, "Reached the maximum webhooks")
	}

	err = query.Q.Transaction(func(tx *query.Query) error {
		webhookService := h.WebhookServiceFactory.New(tx)
		namespaceID := req.NamespaceID
		if ptr.To(req.NamespaceID) == 0 {
			namespaceID = nil
		}
		webhookObj := &models.Webhook{
			NamespaceID:       namespaceID,
			URL:               req.URL,
			Secret:            req.Secret,
			SslVerify:         req.SslVerify,
			RetryTimes:        req.RetryTimes,
			RetryDuration:     req.RetryDuration,
			Enable:            req.Enable,
			EventNamespace:    req.EventNamespace,
			EventRepository:   req.EventRepository,
			EventTag:          req.EventTag,
			EventArtifact:     req.EventArtifact,
			EventMember:       req.EventMember,
			EventDaemonTaskGc: req.EventDaemonTaskGc,
		}
		err = webhookService.Create(ctx, webhookObj)
		if err != nil {
			log.Error().Err(err).Msg("Create webhook failed")
			return errcode.HTTPErrCodeInternalError.Detail("Create webhook failed")
		}
		auditService := h.AuditServiceFactory.New(tx)
		err = auditService.Create(ctx, &models.Audit{
			UserID:       user.ID,
			NamespaceID:  namespaceID,
			Action:       enums.AuditActionCreate,
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
	return c.NoContent(http.StatusCreated)
}

func (h *handler) PostWebhookValidate(req types.PostWebhookRequest) error {
	if !(strings.HasPrefix(req.URL, "http://") || strings.HasPrefix(req.URL, "https://")) { // nolint: staticcheck
		log.Error().Str("URL", req.URL).Msg("URL is invalid")
		return errcode.HTTPErrCodeBadRequest.Detail("URL is invalid, should start with 'http://' or 'https://'")
	}
	return nil
}
