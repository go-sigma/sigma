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
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/opencontainers/go-digest"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/server/validators"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/imagerefs"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
	"github.com/go-sigma/sigma/pkg/xerrors"
)

// HeadManifest handles the head manifest request
func (h *handler) HeadManifest(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	user, needRet, err := utils.GetUserFromCtxForDs(c)
	if err != nil {
		return err
	}
	if needRet {
		return nil
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
	namespaceObj, err := h.NamespaceServiceFactory.New().GetByName(ctx, namespace)
	if err != nil {
		log.Error().Err(err).Str("Name", repository).Msg("Get repository by name failed")
		return xerrors.NewDSError(c, xerrors.DSErrCodeBlobUnknown)
	}

	authChecked, err := h.AuthServiceFactory.New().Namespace(ptr.To(user), namespaceObj.ID, enums.AuthRead)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(errors.New(utils.UnwrapJoinedErrors(err))).Msg("Resource not found")
			return xerrors.GenDSErrCodeResourceNotFound(err)
		}
		return xerrors.NewDSError(c, xerrors.DSErrCodeUnknown)
	}
	if !authChecked {
		log.Error().Int64("UserID", user.ID).Int64("NamespaceID", namespaceObj.ID).Msg("Auth check failed")
		return xerrors.NewDSError(c, xerrors.DSErrCodeDenied)
	}

	ref := strings.TrimPrefix(uri[strings.LastIndex(uri, "/"):], "/")
	if _, err := digest.Parse(ref); err != nil && !consts.TagRegexp.MatchString(ref) {
		log.Error().Err(err).Str("ref", ref).Msg("Invalid digest or tag")
		return xerrors.NewDSError(c, xerrors.DSErrCodeTagInvalid)
	}

	repositoryService := h.RepositoryServiceFactory.New()
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
		tagService := h.TagServiceFactory.New()
		tag, err := tagService.GetByName(ctx, repositoryObj.ID, ref)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) && h.Config.Proxy.Enabled {
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

	artifactService := h.ArtifactServiceFactory.New()
	artifact, err := artifactService.GetByDigest(ctx, repositoryObj.ID, refs.Digest.String())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if h.Config.Proxy.Enabled {
				return h.headManifestFallbackProxy(c)
			} else {
				log.Error().Err(err).Str("ref", ref).Msg("Artifact not found")
				return xerrors.NewDSError(c, xerrors.DSErrCodeManifestUnknown)
			}
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
	statusCode, header, _, err := h.fallbackProxy(c)
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
