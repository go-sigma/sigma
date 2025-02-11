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
	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/server/validators"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/imagerefs"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
	"github.com/go-sigma/sigma/pkg/xerrors"
)

// DeleteManifest handles the delete manifest request
// if reference is a tag, just delete the tag
// if reference is a digest, delete the artifact and all of the tags that reference it
func (h *handler) DeleteManifest(c echo.Context) error {
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

	authChecked, err := h.AuthServiceFactory.New().Namespace(ptr.To(user), namespaceObj.ID, enums.AuthManage)
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
			log.Error().Err(err).Str("Tag", ref).Msg("Delete tag failed")
			return xerrors.NewDSError(c, xerrors.DSErrCodeUnknown)
		}
		return c.NoContent(http.StatusAccepted)
	}

	artifactService := h.ArtifactServiceFactory.New()
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
		tagService := h.TagServiceFactory.New(tx)
		err = tagService.DeleteByArtifactID(ctx, artifactObj.ID)
		if err != nil {
			log.Error().Err(err).Int64("ArtifactID", artifactObj.ID).Msg("Delete tag by artifact id failed")
			return xerrors.DSErrCodeUnknown
		}
		artifactService := h.ArtifactServiceFactory.New(tx)
		err = artifactService.DeleteByID(ctx, artifactObj.ID)
		if err != nil {
			log.Error().Err(err).Int64("ArtifactID", artifactObj.ID).Msg("Delete artifact by id failed")
			return xerrors.DSErrCodeUnknown
		}
		return nil
	})
	if err != nil {
		var e xerrors.ErrCode
		if errors.As(err, &e) {
			return xerrors.NewDSError(c, e)
		}
		return xerrors.NewDSError(c, xerrors.DSErrCodeUnknown)
	}

	return c.NoContent(http.StatusAccepted)
}
