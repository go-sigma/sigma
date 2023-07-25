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

package tag

import (
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/xerrors"
)

// GetTag handles the get tag request
// @Summary Get tag
// @Tags Tag
// @security BasicAuth
// @Accept json
// @Produce json
// @Router /namespaces/{namespace}/tags/{id} [get]
// @Param namespace path string true "Namespace"
// @Param id path string true "Tag ID"
// @Param repository query string false "repository"
// @Success 200 {object} types.TagItem
// @Failure 404 {object} xerrors.ErrCode
// @Failure 500 {object} xerrors.ErrCode
func (h *handlers) GetTag(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	var req types.GetTagRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, err.Error())
	}
	// TODO: check the tag namespace and repository auth

	tagService := h.tagServiceFactory.New()
	tag, err := tagService.GetByID(ctx, req.ID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Error().Err(err).Msg("Tag not found")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, err.Error())
		}
		log.Error().Err(err).Msg("Get tag from db failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, err.Error())
	}

	var artifacts = make([]types.TagItemArtifact, 0, len(tag.Artifact.ArtifactIndexes))
	for _, item := range tag.Artifact.ArtifactIndexes {
		artifacts = append(artifacts, types.TagItemArtifact{
			ID:            item.ID,
			Digest:        item.Digest,
			Raw:           string(item.Raw),
			ConfigRaw:     string(item.ConfigRaw),
			Size:          item.Size,
			BlobSize:      item.BlobsSize,
			LastPull:      item.LastPull.Time.Format(consts.DefaultTimePattern),
			PushedAt:      item.PushedAt.Format(consts.DefaultTimePattern),
			Vulnerability: string(item.Vulnerability.Result),
			Sbom:          string(item.Sbom.Result),
			CreatedAt:     item.CreatedAt.Format(consts.DefaultTimePattern),
			UpdatedAt:     item.UpdatedAt.Format(consts.DefaultTimePattern),
		})
	}

	return c.JSON(200, types.TagItem{
		ID:   tag.ID,
		Name: tag.Name,
		Artifact: types.TagItemArtifact{
			ID:            tag.Artifact.ID,
			Digest:        tag.Artifact.Digest,
			Raw:           string(tag.Artifact.Raw),
			ConfigRaw:     string(tag.Artifact.ConfigRaw),
			Size:          tag.Artifact.Size,
			BlobSize:      tag.Artifact.BlobsSize,
			LastPull:      tag.Artifact.LastPull.Time.Format(consts.DefaultTimePattern),
			PushedAt:      tag.Artifact.PushedAt.Format(consts.DefaultTimePattern),
			Vulnerability: string(tag.Artifact.Vulnerability.Result),
			Sbom:          string(tag.Artifact.Sbom.Result),
			CreatedAt:     tag.Artifact.CreatedAt.Format(consts.DefaultTimePattern),
			UpdatedAt:     tag.Artifact.UpdatedAt.Format(consts.DefaultTimePattern),
		},
		Artifacts: artifacts,
		PushedAt:  tag.PushedAt.Format(consts.DefaultTimePattern),
		CreatedAt: tag.CreatedAt.Format(consts.DefaultTimePattern),
		UpdatedAt: tag.UpdatedAt.Format(consts.DefaultTimePattern),
	})
}
