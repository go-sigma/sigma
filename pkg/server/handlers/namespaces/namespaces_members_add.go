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
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
	"github.com/go-sigma/sigma/pkg/xerrors"
)

// AddNamespaceMember handles the add namespace member request
//
//	@Summary	Add namespace member
//	@Tags		Namespace
//	@Accept		json
//	@Produce	json
//	@Router		/namespaces/members/ [post]
//	@Param		message	body	types.AddNamespaceMemberRequest	true	"Member object"
//	@security	BasicAuth
//	@Success	201	{object}	types.AddNamespaceMemberResponse
//	@Failure	400	{object}	xerrors.ErrCode
//	@Failure	500	{object}	xerrors.ErrCode
func (h *handler) AddNamespaceMember(c echo.Context) error {
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

	var req types.AddNamespaceMemberRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, err.Error())
	}

	authChecked, err := h.AuthServiceFactory.New().Namespace(ptr.To(user), req.NamespaceID, enums.AuthAdmin)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Int64("id", req.NamespaceID).Msg("Namespace not found")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, fmt.Sprintf("Namespace(%d) not found", req.NamespaceID))
		}
		log.Error().Err(err).Msg("Get namespace failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Get namespace(%d) failed", req.NamespaceID))
	}
	if !authChecked {
		log.Error().Int64("UserID", user.ID).Int64("NamespaceID", req.NamespaceID).Msg("Auth check failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeUnauthorized, "No permission with this api")
	}

	namespaceService := h.NamespaceServiceFactory.New()
	namespaceObj, err := namespaceService.Get(ctx, req.NamespaceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Msg("Namespace not found")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, err.Error())
		}
		log.Error().Err(err).Msg("Find namespace failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, err.Error())
	}

	roles := dal.AuthEnforcer.GetRolesForUserInDomain(fmt.Sprintf("%d", req.UserID), namespaceObj.Name)
	if len(roles) > 0 {
		log.Error().Int64("UserID", req.UserID).Int64("NamespaceID", req.NamespaceID).Msg("User already have role in namespace")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeConflict, "User already have role in namespace")
	}

	namespaceMemberService := h.NamespaceMemberServiceFactory.New()
	roleCount, err := namespaceMemberService.CountNamespaceMember(ctx, req.UserID, req.NamespaceID)
	if err != nil {
		log.Error().Int64("UserID", req.UserID).Int64("NamespaceID", req.NamespaceID).Msg("Count namespace role failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Count namespace role failed: %v", err))
	}
	if roleCount >= consts.MaxNamespaceMember {
		log.Error().Int64("UserID", req.UserID).Int64("NamespaceID", req.NamespaceID).Msg("Max namespace role quota exceeds")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, "Max namespace role quota exceeds")
	}

	var namespaceMemberObj *models.NamespaceMember
	err = query.Q.Transaction(func(tx *query.Query) error {
		namespaceMemberService := h.NamespaceMemberServiceFactory.New(tx)
		namespaceMemberObj, err = namespaceMemberService.AddNamespaceMember(ctx, req.UserID, ptr.To(namespaceObj), req.Role)
		if err != nil {
			return xerrors.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Add namespace role for user failed: %v", err))
		}
		auditService := h.AuditServiceFactory.New(tx)
		err = auditService.Create(ctx, &models.Audit{
			UserID:       user.ID,
			NamespaceID:  ptr.Of(namespaceObj.ID),
			Action:       enums.AuditActionCreate,
			ResourceType: enums.AuditResourceTypeNamespaceMember,
			Resource:     namespaceObj.Name,
			ReqRaw:       utils.MustMarshal(req),
		})
		if err != nil {
			log.Error().Err(err).Msg("Create audit failed")
			return xerrors.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Create audit failed: %v", err))
		}
		return nil
	})
	if err != nil {
		var e xerrors.ErrCode
		if errors.As(err, &e) {
			return xerrors.NewHTTPError(c, e)
		}
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError)
	}
	err = dal.AuthEnforcer.LoadPolicy()
	if err != nil {
		log.Error().Err(err).Msg("Reload policy failed")
	}
	return c.JSON(http.StatusCreated, types.AddNamespaceMemberResponse{
		ID: namespaceMemberObj.ID,
	})
}
