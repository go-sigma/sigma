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

package repositories

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
	"github.com/go-sigma/sigma/pkg/xerrors"
)

// PutRepository handles the put repository request
// @Summary Update repository
// @Tags Repository
// @security BasicAuth
// @Accept json
// @Produce json
// @Router /namespaces/{namespace}/repositories/{id} [put]
// @Param namespace path string true "Namespace name"
// @Param id path string true "Repository id"
// @Param message body types.PutRepositoryRequestSwagger true "Repository object"
// @Success 204
// @Failure 400 {object} xerrors.ErrCode
// @Failure 404 {object} xerrors.ErrCode
// @Failure 500 {object} xerrors.ErrCode
func (h *handlers) PutRepository(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	var req types.PutRepositoryRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, err.Error())
	}

	namespaceService := h.namespaceServiceFactory.New()
	namespaceObj, err := namespaceService.GetByName(ctx, req.Namespace)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Str("namespace", namespaceObj.Name).Msg("Namespace not found")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, err.Error())
		}
		log.Error().Err(err).Msg("Namespace find failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, err.Error())
	}

	repositoryService := h.repositoryServiceFactory.New()
	repositoryObj, err := repositoryService.Get(ctx, req.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Int64("id", req.ID).Msg("Repository not found")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, err.Error())
		}
		log.Error().Err(err).Msg("Repository find failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, err.Error())
	}
	if repositoryObj.NamespaceID != namespaceObj.ID {
		log.Error().Interface("repositoryObj", repositoryObj).Interface("namespaceObj", namespaceObj).Msg("Repository's namespace ref id not equal namespace id")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound)
	}

	updates := make(map[string]interface{}, 5)
	if req.SizeLimit != nil {
		updates[query.Namespace.SizeLimit.ColumnName().String()] = ptr.To(req.SizeLimit)
	}
	if req.TagLimit != nil {
		updates[query.Namespace.TagLimit.ColumnName().String()] = ptr.To(req.TagLimit)
	}
	if req.Description != nil {
		updates[query.Namespace.Description.ColumnName().String()] = ptr.To(req.Description)
	}
	if req.Overview != nil {
		updates[query.Repository.Overview.ColumnName().String()] = []byte(ptr.To(req.Overview))
	}
	if req.Visibility != nil {
		updates[query.Repository.Visibility.ColumnName().String()] = ptr.To(req.Visibility)
	}

	if len(updates) > 0 {
		err = repositoryService.UpdateRepository(ctx, repositoryObj.ID, updates)
		if err != nil {
			log.Error().Err(err).Int64("id", repositoryObj.ID).Msg("Repository update failed")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, err.Error())
		}
	}

	return c.NoContent(http.StatusNoContent)
}
