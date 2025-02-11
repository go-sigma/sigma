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
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
	"github.com/go-sigma/sigma/pkg/xerrors"
)

// GetNamespace handles the get namespace request
//
//	@Summary	Get namespace
//	@security	BasicAuth
//	@Tags		Namespace
//	@Accept		json
//	@Produce	json
//	@Router		/namespaces/{namespace_id} [get]
//	@Param		namespace_id	path		number	true	"Namespace id"
//	@Success	200				{object}	types.NamespaceItem
//	@Failure	404				{object}	xerrors.ErrCode
//	@Failure	500				{object}	xerrors.ErrCode
func (h *handler) GetNamespace(c echo.Context) error {
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

	var req types.GetNamespaceRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, err.Error())
	}

	namespaceService := h.NamespaceServiceFactory.New()
	namespaceObj, err := namespaceService.Get(ctx, req.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Msg("Get namespace from db failed")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, err.Error())
		}
		log.Error().Err(err).Msg("Get namespace from db failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, err.Error())
	}

	authService := h.AuthServiceFactory.New()
	authChecked, err := authService.Namespace(ptr.To(user), req.ID, enums.AuthRead)
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

	namespaceRole, err := authService.NamespaceRole(ptr.To(user), req.ID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Get user namespace role failed: %v", err))
	}

	repositoryService := h.RepositoryServiceFactory.New()
	repositoryMapCount, err := repositoryService.CountByNamespace(ctx, []int64{namespaceObj.ID})
	if err != nil {
		log.Error().Err(err).Msg("Count repository failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, err.Error())
	}

	tagService := h.TagServiceFactory.New()
	tagMapCount, err := tagService.CountByNamespace(ctx, []int64{namespaceObj.ID})
	if err != nil {
		log.Error().Err(err).Msg("Count tag failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, err.Error())
	}

	return c.JSON(http.StatusOK, types.NamespaceItem{
		ID:              namespaceObj.ID,
		Name:            namespaceObj.Name,
		Description:     namespaceObj.Description,
		Overview:        ptr.Of(string(namespaceObj.Overview)),
		Visibility:      namespaceObj.Visibility,
		Role:            namespaceRole,
		Size:            namespaceObj.Size,
		SizeLimit:       namespaceObj.SizeLimit,
		RepositoryCount: repositoryMapCount[namespaceObj.ID],
		RepositoryLimit: namespaceObj.RepositoryLimit,
		TagCount:        tagMapCount[namespaceObj.ID],
		TagLimit:        namespaceObj.TagLimit,
		CreatedAt:       time.Unix(0, int64(time.Millisecond)*namespaceObj.CreatedAt).UTC().Format(consts.DefaultTimePattern),
		UpdatedAt:       time.Unix(0, int64(time.Millisecond)*namespaceObj.CreatedAt).UTC().Format(consts.DefaultTimePattern),
	})
}
