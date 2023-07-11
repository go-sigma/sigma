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
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"

	"github.com/ximager/ximager/pkg/consts"
	"github.com/ximager/ximager/pkg/types"
	"github.com/ximager/ximager/pkg/utils"
	"github.com/ximager/ximager/pkg/xerrors"
)

// ListNamespace handles the list namespace request
// @Summary List namespace
// @security BasicAuth
// @Tags Namespace
// @Accept json
// @Produce json
// @Router /namespaces/ [get]
// @Param limit query int64 false "limit" minimum(10) maximum(100) default(10)
// @Param last query int64 false "last" minimum(0) default(0)
// @Param sort query string false "sort field"
// @Param method query string false "sort method" Enums(asc, desc)
// @Param name query string false "search namespace with name"
// @Success 200	{object} types.CommonList{items=[]types.NamespaceItem}
// @Failure 500 {object} xerrors.ErrCode
func (h *handlers) ListNamespace(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	var req types.ListNamespaceRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, err.Error())
	}
	req.Pagination = utils.NormalizePagination(req.Pagination)

	namespaceService := h.namespaceServiceFactory.New()
	namespaces, err := namespaceService.ListNamespace(ctx, req.Name, req.Pagination, req.Sortable)
	if err != nil {
		log.Error().Err(err).Msg("List namespace from db failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, err.Error())
	}

	var resp = make([]any, 0, len(namespaces))
	for _, ns := range namespaces {
		resp = append(resp, types.NamespaceItem{
			ID:              ns.ID,
			Name:            ns.Name,
			Description:     ns.Description,
			Visibility:      ns.Visibility,
			Size:            ns.Size,
			SizeLimit:       ns.SizeLimit,
			RepositoryLimit: ns.RepositoryLimit,
			RepositoryCount: ns.RepositoryCount,
			TagLimit:        ns.TagLimit,
			TagCount:        ns.TagCount,
			CreatedAt:       ns.CreatedAt.Format(consts.DefaultTimePattern),
			UpdatedAt:       ns.UpdatedAt.Format(consts.DefaultTimePattern),
		})
	}

	total, err := namespaceService.CountNamespace(ctx, req.Name)
	if err != nil {
		log.Error().Err(err).Msg("Count namespace from db failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, err.Error())
	}

	return c.JSON(http.StatusOK, types.CommonList{Total: total, Items: resp})
}
