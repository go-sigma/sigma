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
	"github.com/go-sigma/sigma/pkg/modules/workq/definition"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
	"github.com/go-sigma/sigma/pkg/xerrors"
)

// DeleteNamespace handles the delete namespace request
//
//	@Summary	Delete namespace
//	@security	BasicAuth
//	@Tags		Namespace
//	@Accept		json
//	@Produce	json
//	@Router		/namespaces/{namespace_id} [delete]
//	@Param		namespace_id	path	number	true	"Namespace id"
//	@Success	204
//	@Failure	500	{object}	xerrors.ErrCode
func (h *handler) DeleteNamespace(c echo.Context) error {
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

	var req types.DeleteNamespaceRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, err.Error())
	}

	authChecked, err := h.AuthServiceFactory.New().Namespace(ptr.To(user), req.ID, enums.AuthManage)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(errors.New(utils.UnwrapJoinedErrors(err))).Int64("NamespaceID", req.ID).Msg("Resource not found")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, utils.UnwrapJoinedErrors(err))
		}
		log.Error().Err(errors.New(utils.UnwrapJoinedErrors(err))).Int64("NamespaceID", req.ID).Err(err).Msg("Get resource failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, utils.UnwrapJoinedErrors(err))
	}
	if !authChecked {
		log.Error().Int64("UserID", user.ID).Int64("NamespaceID", req.ID).Msg("Auth check failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeUnauthorized, "No permission with this api")
	}

	namespaceService := h.NamespaceServiceFactory.New()
	namespaceObj, err := namespaceService.Get(ctx, req.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Int64("id", req.ID).Msg("Namespace not found")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, fmt.Sprintf("Namespace(%d) not found", req.ID))
		}
		log.Error().Err(err).Msg("Get namespace failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Get namespace(%d) failed", req.ID))
	}

	err = query.Q.Transaction(func(tx *query.Query) error {
		namespaceService := h.NamespaceServiceFactory.New(tx)
		err = namespaceService.DeleteByID(ctx, req.ID)
		if err != nil {
			log.Error().Err(err).Msg("Delete namespace failed")
			return xerrors.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Namespace(%d) find failed: %v", req.ID, err))
		}
		auditService := h.AuditServiceFactory.New(tx)
		err = auditService.Create(ctx, &models.Audit{
			UserID:       user.ID,
			NamespaceID:  ptr.Of(req.ID),
			Action:       enums.AuditActionDelete,
			ResourceType: enums.AuditResourceTypeNamespace,
			Resource:     namespaceObj.Name,
		})
		if err != nil {
			log.Error().Err(err).Msg("Create audit for delete namespace failed")
			return xerrors.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Create audit for delete namespace failed: %v", err))
		}
		err = h.ProducerClient.Produce(ctx, enums.DaemonWebhook, types.DaemonWebhookPayload{
			NamespaceID:  ptr.Of(namespaceObj.ID),
			Action:       enums.WebhookActionDelete,
			ResourceType: enums.WebhookResourceTypeNamespace,
			Payload:      utils.MustMarshal(namespaceObj),
		}, definition.ProducerOption{Tx: tx})
		if err != nil {
			log.Error().Err(err).Msg("Webhook event produce failed")
			return xerrors.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Webhook event produce failed: %v", err))
		}
		return nil
	})
	if err != nil {
		return xerrors.NewHTTPError(c, err.(xerrors.ErrCode))
	}

	return c.NoContent(http.StatusNoContent)
}
