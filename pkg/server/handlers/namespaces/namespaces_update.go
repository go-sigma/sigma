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

package namespaces

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/modules/workq"
	"github.com/go-sigma/sigma/pkg/server/errcode"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

// PutNamespace handles the put namespace request
//
//	@Summary	Update namespace
//	@security	BasicAuth
//	@Tags		Namespace
//	@Accept		json
//	@Produce	json
//	@Router		/namespaces/{namespace_id} [put]
//	@Param		namespace_id	path	number							true	"Namespace id"
//	@Param		message			body	types.UpdateNamespaceRequest	true	"Namespace object"
//	@Success	204
func (h *handler) PutNamespace(c echo.Context) error {
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

	var req types.UpdateNamespaceRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return errcode.NewHTTPError(c, errcode.HTTPErrCodeBadRequest, fmt.Sprintf("Bind and validate request body failed: %v", err))
	}

	authChecked, err := h.AuthServiceFactory.New().Namespace(ptr.To(user), req.ID, enums.AuthAdmin)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(errors.New(utils.UnwrapJoinedErrors(err))).Int64("NamespaceID", req.ID).Msg("Resource not found")
			return errcode.NewHTTPError(c, errcode.HTTPErrCodeNotFound, utils.UnwrapJoinedErrors(err))
		}
		log.Error().Err(errors.New(utils.UnwrapJoinedErrors(err))).Int64("NamespaceID", req.ID).Err(err).Msg("Get resource failed")
		return errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError, utils.UnwrapJoinedErrors(err))
	}
	if !authChecked {
		log.Error().Int64("UserID", user.ID).Int64("NamespaceID", req.ID).Msg("Auth check failed")
		return errcode.NewHTTPError(c, errcode.HTTPErrCodeUnauthorized, "No permission with this api")
	}

	namespaceService := h.NamespaceServiceFactory.New()
	namespaceObj, err := namespaceService.Get(ctx, req.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Msg("Namespace not found")
			return errcode.NewHTTPError(c, errcode.HTTPErrCodeNotFound, err.Error())
		}
		log.Error().Err(err).Msg("Find namespace failed")
		return errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError, err.Error())
	}

	if req.SizeLimit != nil && namespaceObj.SizeLimit > ptr.To(req.SizeLimit) {
		log.Error().Err(err).Msg("Namespace quota is less than the before limit")
		return errcode.NewHTTPError(c, errcode.HTTPErrCodeBadRequest, "Namespace quota is less than the before limit")
	}

	updates := make(map[string]any, 5)
	if req.SizeLimit != nil {
		updates[query.Namespace.SizeLimit.ColumnName().String()] = ptr.To(req.SizeLimit)
	}
	if req.RepositoryLimit != nil {
		updates[query.Namespace.RepositoryLimit.ColumnName().String()] = ptr.To(req.RepositoryLimit)
	}
	if req.TagLimit != nil {
		updates[query.Namespace.TagLimit.ColumnName().String()] = ptr.To(req.TagLimit)
	}
	if req.Description != nil {
		updates[query.Namespace.Description.ColumnName().String()] = ptr.To(req.Description)
	}
	if req.Visibility != nil {
		updates[query.Namespace.Visibility.ColumnName().String()] = ptr.To(req.Visibility)
	}
	if req.Overview != nil {
		updates[query.Repository.Overview.ColumnName().String()] = []byte(ptr.To(req.Overview))
	}

	if len(updates) > 0 {
		err = query.Q.Transaction(func(tx *query.Query) error {
			namespaceService := h.NamespaceServiceFactory.New(tx)
			err = namespaceService.UpdateByID(ctx, namespaceObj.ID, updates)
			if err != nil {
				log.Error().Err(err).Msg("Update namespace failed")
				return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Update namespace failed: %v", err))
			}
			auditService := h.AuditServiceFactory.New(tx)
			err = auditService.Create(ctx, &models.Audit{
				UserID:       user.ID,
				NamespaceID:  ptr.Of(namespaceObj.ID),
				Action:       enums.AuditActionUpdate,
				ResourceType: enums.AuditResourceTypeNamespace,
				Resource:     namespaceObj.Name,
				ReqRaw:       utils.MustMarshal(req),
			})
			if err != nil {
				log.Error().Err(err).Msg("Create audit for update namespace failed")
				return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Create audit for update namespace failed: %v", err))
			}
			err = h.Producer.Produce(ctx, enums.DaemonWebhook, types.DaemonWebhookPayload{
				NamespaceID:  ptr.Of(namespaceObj.ID),
				Action:       enums.WebhookActionUpdate,
				ResourceType: enums.WebhookResourceTypeNamespace,
				Payload:      utils.MustMarshal(req),
			}, workq.ProducerOption{Tx: tx})
			if err != nil {
				log.Error().Err(err).Msg("Webhook event produce failed")
				return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Webhook event produce failed: %v", err))
			}
			return nil
		})
		if err != nil {
			var e errcode.ErrCode
			if errors.As(err, &e) {
				return errcode.NewHTTPError(c, e)
			}
			return errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError)
		}
	}

	return c.NoContent(http.StatusNoContent)
}
