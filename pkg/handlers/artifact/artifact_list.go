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

package artifact

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"

	"github.com/ximager/ximager/pkg/consts"
	"github.com/ximager/ximager/pkg/types"
	"github.com/ximager/ximager/pkg/utils"
	"github.com/ximager/ximager/pkg/xerrors"
)

// ListArtifact handles the list artifact request
func (h *handlers) ListArtifact(c echo.Context) error {
	ctx := c.Request().Context()

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

	var artifactIDs = make([]uint64, 0, len(artifacts))
	for _, artifact := range artifacts {
		artifactIDs = append(artifactIDs, artifact.ID)
	}
	tagService := h.tagServiceFactory.New()
	tagCountRef, err := tagService.CountByArtifact(ctx, artifactIDs)
	if err != nil {
		log.Error().Err(err).Msg("Count tag from db failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, err.Error())
	}

	var resp = make([]any, 0, len(artifacts))
	for _, artifact := range artifacts {
		tags := make([]string, 0, len(artifact.Tags))
		for _, tag := range artifact.Tags {
			tags = append(tags, tag.Name)
		}
		resp = append(resp, types.ArtifactItem{
			ID:        artifact.ID,
			Digest:    artifact.Digest,
			Size:      artifact.Size,
			Tags:      tags,
			TagCount:  tagCountRef[artifact.ID],
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
