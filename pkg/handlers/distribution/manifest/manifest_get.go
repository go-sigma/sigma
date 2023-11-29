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
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils/imagerefs"
	"github.com/go-sigma/sigma/pkg/validators"
	"github.com/go-sigma/sigma/pkg/xerrors"
)

// GetManifest handles the get manifest request
func (h *handler) GetManifest(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	iuser := c.Get(consts.ContextUser)
	if iuser == nil {
		log.Error().Msg("Get user from header failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeUnauthorized)
	}
	user, ok := iuser.(*models.User)
	if !ok {
		log.Error().Msg("Convert user from header failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeUnauthorized)
	}

	uri := c.Request().URL.Path

	repository := strings.TrimPrefix(strings.TrimSuffix(uri[:strings.LastIndex(uri, "/")], "/manifests"), "/v2/")
	_, namespace, _, _, err := imagerefs.Parse(repository)
	if err != nil {
		log.Error().Err(err).Str("Repository", repository).Msg("Repository must container a valid namespace")
		return xerrors.NewDSError(c, xerrors.DSErrCodeManifestWithNamespace)
	}
	if !(validators.ValidateNamespaceRaw(namespace) && validators.ValidateRepositoryRaw(repository)) {
		log.Error().Err(err).Str("Repository", repository).Msg("Repository must container a valid namespace")
		return xerrors.NewDSError(c, xerrors.DSErrCodeManifestWithNamespace)
	}
	namespaceObj, err := h.namespaceServiceFactory.New().GetByName(ctx, namespace)
	if err != nil {
		log.Error().Err(err).Str("Name", repository).Msg("Get repository by name failed")
		return xerrors.NewDSError(c, xerrors.DSErrCodeBlobUnknown)
	}
	if !h.authServiceFactory.New().Namespace(c, namespaceObj.ID, enums.AuthRead) {
		log.Error().Int64("UserID", user.ID).Int64("NamespaceID", namespaceObj.ID).Msg("Auth check failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeUnauthorized, "No permission with this api")
	}

	ref := strings.TrimPrefix(uri[strings.LastIndex(uri, "/"):], "/")
	if _, err := digest.Parse(ref); err != nil && !consts.TagRegexp.MatchString(ref) {
		log.Error().Err(err).Str("ref", ref).Msg("Invalid digest or tag")
		return xerrors.NewDSError(c, xerrors.DSErrCodeTagInvalid)
	}

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

	var refs = h.parseRef(ref)

	if refs.Tag != "" {
		tagService := h.tagServiceFactory.New()
		tag, err := tagService.GetByName(ctx, repositoryObj.ID, ref)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) && h.config.Proxy.Enabled {
				return h.getManifestFallbackProxy(c, refs)
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
		if errors.Is(err, gorm.ErrRecordNotFound) && h.config.Proxy.Enabled {
			return h.getManifestFallbackProxy(c, refs)
		}
		log.Error().Err(err).Str("ref", ref).Msg("Get artifact failed")
		return xerrors.NewDSError(c, xerrors.DSErrCodeManifestUnknown)
	}

	return c.Blob(http.StatusOK, artifact.ContentType, artifact.Raw)
}

// getManifestFallbackProxy ...
func (h *handler) getManifestFallbackProxy(c echo.Context, refs Refs) error {
	statusCode, header, bodyBytes, err := h.fallbackProxy(c)
	if err != nil {
		log.Error().Err(err).Interface("refs", refs).Int("status", statusCode).Msg("Fallback proxy failed")
		return xerrors.NewDSError(c, xerrors.DSErrCodeUnknown)
	}
	if statusCode == http.StatusOK {
		c.Response().Header().Set(consts.ContentDigest, header.Get(consts.ContentDigest))
		c.Response().Header().Set("ETag", header.Get("ETag"))
		return c.Blob(http.StatusOK, header.Get(echo.HeaderContentType), bodyBytes)
	} else if statusCode == http.StatusNotFound {
		return xerrors.NewDSError(c, xerrors.DSErrCodeManifestUnknown)
	}
	log.Error().Interface("refs", refs).Int("statusCode", statusCode).Msg("Fallback proxy failed")
	return xerrors.NewDSError(c, xerrors.DSErrCodeUnknown)
}
