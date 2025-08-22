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
	"github.com/go-sigma/sigma/pkg/dal"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/server/errcode"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

// UpdateNamespaceMember handles the update namespace member request
//
//	@Summary	Update namespace member
//	@security	BasicAuth
//	@Tags		Namespace
//	@Accept		json
//	@Produce	json
//	@Router		/namespaces/{namespace_id}/members/{user_id} [put]
//	@Param		namespace_id	path	number								true	"Namespace id"
//	@Param		user_id			path	number								true	"User id"
//	@Param		message			body	types.UpdateNamespaceMemberRequest	true	"Namespace member object"
//	@Success	204
func (h *handler) UpdateNamespaceMember(c echo.Context) error {
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

	var req types.UpdateNamespaceMemberRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return errcode.NewHTTPError(c, errcode.HTTPErrCodeBadRequest, fmt.Sprintf("Bind and validate request body failed: %v", err))
	}

	namespaceService := h.NamespaceServiceFactory.New()
	namespaceObj, err := namespaceService.Get(ctx, req.NamespaceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Msg("Namespace not found")
			return errcode.NewHTTPError(c, errcode.HTTPErrCodeNotFound, fmt.Sprintf("Namespace not found: %v", err))
		}
		log.Error().Err(err).Msg("Find namespace failed")
		return errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError, fmt.Sprintf("Find namespace failed: %v", err))
	}

	roles := dal.AuthEnforcer.GetRolesForUserInDomain(fmt.Sprintf("%d", req.UserID), namespaceObj.Name)
	if len(roles) != 1 {
		log.Error().Int64("UserID", req.UserID).Int64("NamespaceID", req.NamespaceID).Msg("User not have role in namespace")
		return errcode.NewHTTPError(c, errcode.HTTPErrCodeNotFound, "User not have role in namespace")
	}

	role := roles[0]

	if req.Role.String() == role {
		log.Info().Int64("UserID", req.UserID).Int64("NamespaceID", req.NamespaceID).Str("Role", req.Role.String()).Msg("User added to namespace already")
		return c.NoContent(http.StatusNoContent)
	}

	err = query.Q.Transaction(func(tx *query.Query) error {
		namespaceMemberService := h.NamespaceMemberServiceFactory.New(tx)
		err = namespaceMemberService.UpdateNamespaceMember(ctx, req.UserID, ptr.To(namespaceObj), req.Role)
		if err != nil {
			return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Update namespace role for user failed: %v", err))
		}
		auditService := h.AuditServiceFactory.New(tx)
		err = auditService.Create(ctx, &models.Audit{
			UserID:       user.ID,
			NamespaceID:  ptr.Of(namespaceObj.ID),
			Action:       enums.AuditActionUpdate,
			ResourceType: enums.AuditResourceTypeNamespaceMember,
			Resource:     namespaceObj.Name,
			ReqRaw:       utils.MustMarshal(req),
		})
		if err != nil {
			log.Error().Err(err).Msg("Create audit failed")
			return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Create audit failed: %v", err))
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
	err = dal.AuthEnforcer.LoadPolicy()
	if err != nil {
		log.Error().Err(err).Msg("Reload policy failed")
	}
	return c.NoContent(http.StatusNoContent)
}
