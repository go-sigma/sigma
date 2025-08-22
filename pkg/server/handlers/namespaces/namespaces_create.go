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
	"github.com/go-sigma/sigma/pkg/server/errcode"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

// PostNamespace handles the post namespace request
//
//	@Summary	Create namespace
//	@Tags		Namespace
//	@security	BasicAuth
//	@Accept		json
//	@Produce	json
//	@Router		/namespaces/ [post]
//	@Param		message	body		types.PostNamespaceRequest	true	"Namespace object"
//	@Success	201		{object}	types.PostNamespaceResponse
//	@Failure	400		{object}	errcode.ErrCode
//	@Failure	500		{object}	errcode.ErrCode
func (h *handler) PostNamespace(c echo.Context) error {
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

	var req types.PostNamespaceRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return errcode.NewHTTPError(c, errcode.HTTPErrCodeBadRequest, fmt.Sprintf("Bind and validate request body failed: %v", err))
	}

	namespaceService := h.NamespaceServiceFactory.New()
	_, err = namespaceService.GetByName(ctx, req.Name)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Msg("Get namespace by name failed")
			return errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError, "Get namespace by name failed")
		}
	}
	if err == nil {
		return errcode.NewHTTPError(c, errcode.HTTPErrCodeConflict, "Namespace already exists")
	}

	namespaceObj := &models.Namespace{
		Name:        req.Name,
		Description: req.Description,
	}
	if req.Visibility != nil {
		namespaceObj.Visibility = ptr.To(req.Visibility)
	}
	if ptr.To(req.SizeLimit) > 0 {
		namespaceObj.SizeLimit = ptr.To(req.SizeLimit)
	}
	if ptr.To(req.RepositoryLimit) > 0 {
		namespaceObj.RepositoryLimit = ptr.To(req.RepositoryLimit)
	}
	if ptr.To(req.TagLimit) > 0 {
		namespaceObj.TagLimit = ptr.To(req.TagLimit)
	}
	err = query.Q.Transaction(func(tx *query.Query) error {
		namespaceService := h.NamespaceServiceFactory.New(tx)
		err = namespaceService.Create(ctx, namespaceObj)
		if err != nil {
			log.Error().Err(err).Msg("Create namespace failed")
			return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Create namespace failed: %v", err))
		}
		namespaceMemberService := h.NamespaceMemberServiceFactory.New(tx)
		_, err = namespaceMemberService.AddNamespaceMember(ctx, user.ID, ptr.To(namespaceObj), enums.NamespaceRoleAdmin)
		if err != nil {
			log.Error().Err(err).Msg("Add namespace member failed")
			return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Add namespace member failed: %v", err))
		}
		auditService := h.AuditServiceFactory.New(tx)
		err = auditService.Create(ctx, &models.Audit{
			UserID:       user.ID,
			NamespaceID:  ptr.Of(namespaceObj.ID),
			Action:       enums.AuditActionCreate,
			ResourceType: enums.AuditResourceTypeNamespace,
			Resource:     namespaceObj.Name,
		})
		if err != nil {
			log.Error().Err(err).Msg("Create audit failed")
			return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Create audit failed: %v", err))
		}
		err = h.ProducerClient.Produce(ctx, enums.DaemonWebhook, types.DaemonWebhookPayload{
			NamespaceID:  ptr.Of(namespaceObj.ID),
			Action:       enums.WebhookActionCreate,
			ResourceType: enums.WebhookResourceTypeNamespace,
			Payload:      utils.MustMarshal(namespaceObj),
		}, definition.ProducerOption{Tx: tx})
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

	return c.JSON(http.StatusCreated, types.PostNamespaceResponse{ID: namespaceObj.ID})
}
