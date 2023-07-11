// Copyright 2023 XImager
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
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/ximager/ximager/pkg/consts"
	"github.com/ximager/ximager/pkg/types"
	"github.com/ximager/ximager/pkg/utils"
	"github.com/ximager/ximager/pkg/xerrors"
)

// GetNamespace handles the get namespace request
// @Summary Get namespace
// @security BasicAuth
// @Tags Namespace
// @Accept json
// @Produce json
// @Router /namespaces/{id} [get]
// @Param id path string true "Namespace ID"
// @Success 200 {object} types.NamespaceItem
// @Failure 404 {object} xerrors.ErrCode
// @Failure 500 {object} xerrors.ErrCode
func (h *handlers) GetNamespace(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	var req types.GetNamespaceRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, err.Error())
	}

	namespaceService := h.namespaceServiceFactory.New()
	namespace, err := namespaceService.Get(ctx, req.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Msg("Get namespace from db failed")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, err.Error())
		}
		log.Error().Err(err).Msg("Get namespace from db failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, err.Error())
	}

	repositoryService := h.repositoryServiceFactory.New()
	repositoryMapCount, err := repositoryService.CountByNamespace(ctx, []int64{namespace.ID})
	if err != nil {
		log.Error().Err(err).Msg("Count repository failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, err.Error())
	}

	tagService := h.tagServiceFactory.New()
	tagMapCount, err := tagService.CountByNamespace(ctx, []int64{namespace.ID})
	if err != nil {
		log.Error().Err(err).Msg("Count tag failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, err.Error())
	}

	return c.JSON(http.StatusOK, types.NamespaceItem{
		ID:              namespace.ID,
		Name:            namespace.Name,
		Description:     namespace.Description,
		Visibility:      namespace.Visibility,
		Size:            namespace.Size,
		SizeLimit:       namespace.SizeLimit,
		RepositoryCount: repositoryMapCount[namespace.ID],
		RepositoryLimit: namespace.RepositoryLimit,
		TagCount:        tagMapCount[namespace.ID],
		TagLimit:        namespace.TagLimit,
		CreatedAt:       namespace.CreatedAt.Format(consts.DefaultTimePattern),
		UpdatedAt:       namespace.UpdatedAt.Format(consts.DefaultTimePattern),
	})
}
