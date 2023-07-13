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
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/opencontainers/go-digest"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/xerrors"
)

// HeadManifest handles the head manifest request
func (h *handler) HeadManifest(c echo.Context) error {
	uri := c.Request().URL.Path
	ref := strings.TrimPrefix(uri[strings.LastIndex(uri, "/"):], "/")
	repository := strings.TrimPrefix(strings.TrimSuffix(uri[:strings.LastIndex(uri, "/")], "/manifests"), "/v2/")

	if _, err := digest.Parse(ref); err != nil && !consts.TagRegexp.MatchString(ref) {
		log.Error().Err(err).Str("ref", ref).Msg("Invalid digest or tag")
		return xerrors.NewDSError(c, xerrors.DSErrCodeTagInvalid)
	}

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
		tag, err := tagService.GetByName(ctx, repositoryObj.ID, ref)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) && viper.GetBool("proxy.enabled") {
				return h.headManifestFallbackProxy(c)
			}
			log.Error().Err(err).Str("ref", ref).Msg("Get artifact failed")
			return xerrors.NewDSError(c, xerrors.DSErrCodeManifestUnknown)
		}
		err = tagService.Incr(ctx, tag.ID)
		if err != nil {
			log.Error().Err(err).Str("ref", ref).Msg("Incr tag failed")
		}
		refs.Digest = digest.Digest(tag.Artifact.Digest)
	}

	artifactService := h.artifactServiceFactory.New()
	artifact, err := artifactService.GetByDigest(ctx, repositoryObj.ID, refs.Digest.String())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) && viper.GetBool("proxy.enabled") {
			return h.headManifestFallbackProxy(c)
		}
		log.Error().Err(err).Str("ref", ref).Msg("Get artifact failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, err.Error())
	}

	c.Response().Header().Set(echo.HeaderContentType, artifact.ContentType)
	c.Response().Header().Set(echo.HeaderContentLength, strconv.FormatInt(artifact.Size, 10))
	c.Response().Header().Set(consts.ContentDigest, artifact.Digest)

	return c.NoContent(http.StatusOK)
}

// headManifestFallbackProxy ...
func (h *handler) headManifestFallbackProxy(c echo.Context) error {
	statusCode, header, _, err := fallbackProxy(c)
	if err != nil {
		log.Error().Err(err).Int("status", statusCode).Msg("Fallback proxy failed")
		return xerrors.NewDSError(c, xerrors.DSErrCodeUnknown)
	}
	c.Response().Header().Set(echo.HeaderContentType, header.Get(echo.HeaderContentType))
	if statusCode == http.StatusOK || statusCode == http.StatusNotFound {
		return c.NoContent(statusCode)
	}
	log.Error().Int("statusCode", statusCode).Msg("Fallback proxy failed")
	return xerrors.NewDSError(c, xerrors.DSErrCodeUnknown)
}
