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
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"

	"github.com/ximager/ximager/pkg/dal/dao"
	"github.com/ximager/ximager/pkg/dal/models"
	"github.com/ximager/ximager/pkg/types"
	"github.com/ximager/ximager/pkg/xerrors"
)

// PostNamespace handles the post namespace request
func (h *handlers) PostNamespace(c echo.Context) error {
	ctx := c.Request().Context()

	var req types.CreateNamespaceRequest
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
	_, err = namespaceService.Create(ctx, &models.Namespace{
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		log.Error().Err(err).Msg("Create namespace failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, err.Error())
	}

	return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeCreated)
}
