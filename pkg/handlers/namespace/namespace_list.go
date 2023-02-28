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

package namespace

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"

	"github.com/ximager/ximager/pkg/consts"
	"github.com/ximager/ximager/pkg/dal/dao"
	"github.com/ximager/ximager/pkg/types"
	"github.com/ximager/ximager/pkg/xerrors"
)

// ListNamespace handles the list namespace request
// @Summary List namespace
// @Accept json
// @Produce json
// @Router /namespace/ [get]
// @Param page_size query int64 true "page size" minimum(10) maximum(100) default(10)
// @Param page_num query int64 true "page number" minimum(1) default(1)
// @Param name query string false "search namespace with name"
// @Success 200	{object} types.ListNamespaceResponse
func (h *handlers) ListNamespace(c echo.Context) error {
	ctx := c.Request().Context()

	var req types.ListNamespaceRequest
	err := c.Bind(&req)
	if err != nil {
		log.Error().Err(err).Msg("Bind request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, err.Error())
	}
	err = c.Validate(&req)
	if err != nil {
		log.Error().Err(err).Msg("Validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, err.Error())
	}

	namespaceService := dao.NewNamespaceService()
	namespaces, err := namespaceService.ListNamespace(ctx, req)
	if err != nil {
		log.Error().Err(err).Msg("List namespace from db failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, err.Error())
	}

	var namespaceIDs []uint64
	for _, ns := range namespaces {
		namespaceIDs = append(namespaceIDs, ns.ID)
	}
	artifactService := dao.NewArtifactService()
	artifactCountRef, err := artifactService.CountByNamespace(ctx, namespaceIDs)
	if err != nil {
		log.Error().Err(err).Msg("Count artifact from db failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, err.Error())
	}

	var resp []any
	for _, ns := range namespaces {
		resp = append(resp, types.NamespaceItem{
			ID:            ns.ID,
			Name:          ns.Name,
			Description:   ns.Description,
			ArtifactCount: artifactCountRef[ns.ID],
			CreatedAt:     ns.CreatedAt.Format(consts.DefaultTimePattern),
			UpdatedAt:     ns.UpdatedAt.Format(consts.DefaultTimePattern),
		})
	}

	total, err := namespaceService.CountNamespace(ctx, req)
	if err != nil {
		log.Error().Err(err).Msg("Count namespace from db failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, err.Error())
	}

	return c.JSON(http.StatusOK, types.CommonList{Total: total, Items: resp})
}
