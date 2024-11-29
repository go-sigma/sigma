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
	"errors"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/xerrors"
)

// GetArtifact handles the get artifact request
func (h *handler) GetArtifact(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	var req types.GetArtifactRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, err.Error())
	}

	repositoryService := h.RepositoryServiceFactory.New()
	repositoryObj, err := repositoryService.GetByName(ctx, req.Repository)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Str("repository", req.Repository).Msg("Cannot find repository")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, err.Error())
		}
		log.Error().Err(err).Str("repository", req.Repository).Msg("Get repository failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, err.Error())
	}

	artifactService := h.ArtifactServiceFactory.New()
	artifactObj, err := artifactService.GetByDigest(ctx, repositoryObj.ID, req.Digest)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Error().Err(err).Msg("Artifact not found")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, err.Error())
		}
		log.Error().Err(err).Msg("Get artifact failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, err.Error())
	}

	return c.JSON(200, types.ArtifactItem{
		ID:        artifactObj.ID,
		Digest:    artifactObj.Digest,
		ConfigRaw: string(artifactObj.ConfigRaw),
		Size:      artifactObj.Size,
		BlobSize:  artifactObj.BlobsSize,
		PullTimes: artifactObj.PullTimes,
		LastPull:  time.Unix(0, int64(time.Millisecond)*artifactObj.LastPull).UTC().Format(consts.DefaultTimePattern),
		PushedAt:  time.Unix(0, int64(time.Millisecond)*artifactObj.PushedAt).UTC().Format(consts.DefaultTimePattern),
		CreatedAt: time.Unix(0, int64(time.Millisecond)*artifactObj.CreatedAt).UTC().Format(consts.DefaultTimePattern),
		UpdatedAt: time.Unix(0, int64(time.Millisecond)*artifactObj.CreatedAt).UTC().Format(consts.DefaultTimePattern),
	})
}
