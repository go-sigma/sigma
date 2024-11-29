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
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
	"github.com/go-sigma/sigma/pkg/xerrors"
)

// ListTag handles the list tag request
//
//	@Summary	List tag
//	@Tags		Tag
//	@security	BasicAuth
//	@Accept		json
//	@Produce	json
//	@Router		/namespaces/{namespace_id}/tags/ [get]
//	@Param		namespace_id	path		number		true	"Namespace id"
//	@Param		repository_id	path		number		false	"Repository id"
//	@Param		limit			query		int64		false	"Limit size"	minimum(10)	maximum(100)	default(10)
//	@Param		page			query		int64		false	"Page number"	minimum(1)	default(1)
//	@Param		sort			query		string		false	"Sort field"
//	@Param		method			query		string		false	"Sort method"	Enums(asc, desc)
//	@Param		name			query		string		false	"search tag with name"
//	@Param		type			query		[]string	false	"search tag with type"	Enums(Image, ImageIndex, Chart, Cnab, Cosign, Wasm, Provenance, Unknown)	collectionFormat(multi)
//	@Success	200				{object}	types.CommonList{items=[]types.TagItem}
//	@Failure	404				{object}	xerrors.ErrCode
//	@Failure	500				{object}	xerrors.ErrCode
func (h *handler) ListTag(c echo.Context) error {
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

	var req types.ListTagRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, err.Error())
	}

	authChecked, err := h.AuthServiceFactory.New().Repository(ptr.To(user), req.RepositoryID, enums.AuthRead)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Int64("NamespaceID", req.NamespaceID).Msg("Namespace not found")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, fmt.Sprintf("Namespace(%d) not found: %v", req.NamespaceID, err))
		}
		log.Error().Err(err).Int64("NamespaceID", req.NamespaceID).Msg("Namespace find failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Namespace(%d) find failed: %v", req.NamespaceID, err))
	}
	if !authChecked {
		log.Error().Int64("UserID", user.ID).Int64("RepositoryID", req.RepositoryID).Msg("Auth check failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeUnauthorized, "No permission with this api")
	}

	namespaceService := h.NamespaceServiceFactory.New()
	namespaceObj, err := namespaceService.Get(ctx, req.NamespaceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Int64("NamespaceID", req.NamespaceID).Msg("Namespace not found")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, fmt.Sprintf("Namespace(%d) not found: %v", req.NamespaceID, err))
		}
		log.Error().Err(err).Int64("NamespaceID", req.NamespaceID).Msg("Namespace find failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Namespace(%d) find failed: %v", req.NamespaceID, err))
	}

	repositoryService := h.RepositoryServiceFactory.New()
	repositoryObj, err := repositoryService.Get(ctx, req.RepositoryID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Int64("RepositoryID", req.RepositoryID).Msg("Repository not found")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, fmt.Sprintf("Repository(%d) not found: %v", req.RepositoryID, err))
		}
		log.Error().Err(err).Int64("RepositoryID", req.RepositoryID).Msg("Repository find failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Repository(%d) find failed: %v", req.RepositoryID, err))
	}
	if repositoryObj.NamespaceID != namespaceObj.ID {
		log.Error().Interface("repositoryObj", repositoryObj).Interface("namespaceObj", namespaceObj).Msg("Repository's namespace ref id not equal namespace id")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound)
	}

	tagService := h.TagServiceFactory.New()
	tags, total, err := tagService.ListTag(ctx, repositoryObj.ID, req.Name, req.Type, req.Pagination, req.Sortable)
	if err != nil {
		log.Error().Err(err).Msg("List tag from db failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, err.Error())
	}

	var resp = make([]any, 0, len(tags))
	for _, tag := range tags {
		if tag.Artifact == nil {
			log.Error().Str("image", fmt.Sprintf("%s:%s", repositoryObj.Name, tag.Name)).Msg("Some tag's artifact reference invalid")
			continue
		}
		var artifacts []types.TagItemArtifact
		for _, item := range tag.Artifact.ArtifactSubs {
			artifacts = append(artifacts, types.TagItemArtifact{
				ID:              item.ID,
				Digest:          item.Digest,
				MediaType:       item.ContentType,
				Raw:             string(item.Raw),
				ConfigMediaType: ptr.To(item.ConfigMediaType),
				ConfigRaw:       string(item.ConfigRaw),
				Type:            string(item.Type),
				Size:            item.Size,
				BlobSize:        item.BlobsSize,
				LastPull:        time.Unix(0, int64(time.Millisecond)*item.LastPull).UTC().Format(consts.DefaultTimePattern),
				PushedAt:        time.Unix(0, int64(time.Millisecond)*item.PushedAt).UTC().Format(consts.DefaultTimePattern),
				Vulnerability:   string(item.Vulnerability.Result),
				Sbom:            string(item.Sbom.Result),
				CreatedAt:       time.Unix(0, int64(time.Millisecond)*item.CreatedAt).UTC().Format(consts.DefaultTimePattern),
				UpdatedAt:       time.Unix(0, int64(time.Millisecond)*item.CreatedAt).UTC().Format(consts.DefaultTimePattern),
			})
		}
		resp = append(resp, types.TagItem{
			ID:   tag.ID,
			Name: tag.Name,
			Artifact: types.TagItemArtifact{
				ID:              tag.Artifact.ID,
				Digest:          tag.Artifact.Digest,
				MediaType:       tag.Artifact.ContentType,
				Raw:             string(tag.Artifact.Raw),
				ConfigMediaType: ptr.To(tag.Artifact.ConfigMediaType),
				ConfigRaw:       string(tag.Artifact.ConfigRaw),
				Type:            string(tag.Artifact.Type),
				Size:            tag.Artifact.Size,
				BlobSize:        tag.Artifact.BlobsSize,
				LastPull:        time.Unix(0, int64(time.Millisecond)*tag.Artifact.LastPull).UTC().Format(consts.DefaultTimePattern),
				PushedAt:        time.Unix(0, int64(time.Millisecond)*tag.Artifact.PushedAt).UTC().Format(consts.DefaultTimePattern),
				Vulnerability:   string(tag.Artifact.Vulnerability.Result),
				Sbom:            string(tag.Artifact.Sbom.Result),
				CreatedAt:       time.Unix(0, int64(time.Millisecond)*tag.Artifact.CreatedAt).UTC().Format(consts.DefaultTimePattern),
				UpdatedAt:       time.Unix(0, int64(time.Millisecond)*tag.Artifact.CreatedAt).UTC().Format(consts.DefaultTimePattern),
			},
			Artifacts: artifacts,
			PullTimes: tag.PullTimes,
			PushedAt:  time.Unix(0, int64(time.Millisecond)*tag.PushedAt).UTC().Format(consts.DefaultTimePattern),
			CreatedAt: time.Unix(0, int64(time.Millisecond)*tag.CreatedAt).UTC().Format(consts.DefaultTimePattern),
			UpdatedAt: time.Unix(0, int64(time.Millisecond)*tag.CreatedAt).UTC().Format(consts.DefaultTimePattern),
		})
	}

	return c.JSON(http.StatusOK, types.CommonList{Total: total, Items: resp})
}
