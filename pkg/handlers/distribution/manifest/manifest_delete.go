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

package manifest

import (
	"errors"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/opencontainers/go-digest"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/xerrors"
)

// DeleteManifest handles the delete manifest request
// if reference is a tag, just delete the tag
// if reference is a digest, delete the artifact and all of the tags that reference it
func (h *handler) DeleteManifest(c echo.Context) error {
	uri := c.Request().URL.Path
	ref := strings.TrimPrefix(uri[strings.LastIndex(uri, "/"):], "/")

	if _, err := digest.Parse(ref); err != nil && !consts.TagRegexp.MatchString(ref) {
		log.Error().Err(err).Str("ref", ref).Msg("Invalid digest or tag")
		return xerrors.NewDSError(c, xerrors.DSErrCodeTagInvalid)
	}

	repository := strings.TrimPrefix(strings.TrimSuffix(uri[:strings.LastIndex(uri, "/")], "/manifests"), "/v2/")

	ctx := log.Logger.WithContext(c.Request().Context())

	repositoryService := h.repositoryServiceFactory.New()
	repositoryObj, err := repositoryService.GetByName(ctx, repository)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Str("repository", repository).Msg("Cannot find repository")
			return xerrors.NewDSError(c, xerrors.DSErrCodeNameUnknown)
		}
		log.Error().Err(err).Str("repository", repository).Msg("Get repository failed")
		return xerrors.NewDSError(c, xerrors.DSErrCodeUnknown)
	}

	refs := h.parseRef(ref)

	if refs.Tag != "" {
		tagService := h.tagServiceFactory.New()
		_, err = tagService.GetByName(ctx, repositoryObj.ID, ref)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				log.Error().Err(err).Str("repository", repository).Str("tag", ref).Msg("Cannot find tag")
				return xerrors.NewDSError(c, xerrors.DSErrCodeManifestUnknown)
			}
			log.Error().Err(err).Str("repository", repository).Str("tag", ref).Msg("Get tag failed")
			return xerrors.NewDSError(c, xerrors.DSErrCodeUnknown)
		}
		err = tagService.DeleteByName(ctx, repositoryObj.ID, ref)
		if err != nil {
			return err
		}
		return c.NoContent(http.StatusAccepted)
	}

	artifactService := h.artifactServiceFactory.New()
	artifactObj, err := artifactService.GetByDigest(ctx, repositoryObj.ID, refs.Digest.String())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Str("repository", repository).Str("artifact", refs.Digest.String()).Msg("Cannot find artifact")
			return xerrors.NewDSError(c, xerrors.DSErrCodeManifestUnknown)
		}
		log.Error().Err(err).Str("repository", repository).Str("artifact", refs.Digest.String()).Msg("Get artifact failed")
		return xerrors.NewDSError(c, xerrors.DSErrCodeUnknown)
	}
	err = query.Q.Transaction(func(tx *query.Query) error {
		tagService := h.tagServiceFactory.New(tx)
		err = tagService.DeleteByArtifactID(ctx, artifactObj.ID)
		if err != nil {
			return err
		}
		artifactService := h.artifactServiceFactory.New(tx)
		err = artifactService.DeleteByID(ctx, artifactObj.ID)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		log.Error().Err(err).Str("repository", repository).Str("artifact", refs.Digest.String()).Msg("Delete artifact failed")
		return xerrors.NewDSError(c, xerrors.DSErrCodeUnknown)
	}

	return c.NoContent(http.StatusAccepted)
}
