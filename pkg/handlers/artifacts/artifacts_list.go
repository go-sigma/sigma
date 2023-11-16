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

package artifact

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"

	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/xerrors"
)

// ListArtifact handles the list artifact request
func (h *handler) ListArtifact(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	var req types.ListArtifactRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, err.Error())
	}

	artifactService := h.artifactServiceFactory.New()
	artifacts, err := artifactService.ListArtifact(ctx, req)
	if err != nil {
		log.Error().Err(err).Msg("List artifact from db failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, err.Error())
	}

	var resp = make([]any, 0, len(artifacts))
	for _, artifact := range artifacts {
		resp = append(resp, types.ArtifactItem{
			ID:        artifact.ID,
			Digest:    artifact.Digest,
			Size:      artifact.Size,
			CreatedAt: artifact.CreatedAt.Format(consts.DefaultTimePattern),
			UpdatedAt: artifact.UpdatedAt.Format(consts.DefaultTimePattern),
		})
	}
	total, err := artifactService.CountArtifact(ctx, req)
	if err != nil {
		log.Error().Err(err).Msg("Count artifact from db failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, err.Error())
	}

	return c.JSON(http.StatusOK, types.CommonList{Total: total, Items: resp})
}
