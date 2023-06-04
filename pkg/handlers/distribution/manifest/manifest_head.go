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
	"fmt"
	"net/http"
	"strings"

	"github.com/distribution/distribution/v3/reference"
	"github.com/labstack/echo/v4"
	"github.com/opencontainers/go-digest"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"gorm.io/gorm"

	"github.com/ximager/ximager/pkg/dal/dao"
	"github.com/ximager/ximager/pkg/xerrors"
)

// HeadManifest handles the head manifest request
func (h *handler) HeadManifest(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())
	uri := c.Request().URL.Path
	ref := strings.TrimPrefix(uri[strings.LastIndex(uri, "/"):], "/")
	repository := strings.TrimPrefix(strings.TrimSuffix(uri[:strings.LastIndex(uri, "/")], "/manifests"), "/v2/")

	if _, err := digest.Parse(ref); err != nil && !reference.TagRegexp.MatchString(ref) {
		log.Error().Err(err).Str("ref", ref).Msg("not valid digest or tag")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, fmt.Sprintf("reference %s not valid", ref))
	}

	referenceServiceFactory := dao.NewReferenceServiceFactory()
	referenceService := referenceServiceFactory.New()
	reference, err := referenceService.Get(ctx, repository, ref)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) && viper.GetBool("proxy.enabled") {
			statusCode, header, _, err := fallbackProxy(c)
			if err != nil {
				log.Error().Err(err).Str("ref", ref).Int("status", statusCode).Msg("Fallback proxy")
				return xerrors.NewDSError(c, xerrors.DSErrCodeUnknown)
			}
			c.Response().Header().Set("Content-Type", header.Get("Content-Type"))
			if statusCode == http.StatusOK || statusCode == http.StatusNotFound {
				return c.NoContent(statusCode)
			}
			return xerrors.NewDSError(c, xerrors.DSErrCodeUnknown)
		}
		log.Error().Err(err).Str("ref", ref).Msg("Get local reference failed")
		return xerrors.NewDSError(c, xerrors.DSErrCodeManifestUnknown)
	}

	if reference.Artifact == nil {
		log.Error().Err(err).Str("ref", ref).Msg("Artifact not found")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, "Artifact not found")
	}

	artifactService := h.artifactServiceFactory.New()
	artifact, err := artifactService.GetByDigest(ctx, repository, reference.Artifact.Digest)
	if err != nil {
		log.Error().Err(err).Str("ref", ref).Msg("Get artifact failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, err.Error())
	}

	contentType := artifact.ContentType
	if contentType == "" {
		contentType = "application/vnd.docker.distribution.manifest.v2+json"
	}
	c.Response().Header().Set("Content-Type", contentType)

	return c.NoContent(http.StatusOK)
}
