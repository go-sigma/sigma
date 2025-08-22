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
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/server/errcode"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

// GetTag handles the get tag request
//
//	@Summary	Get tag
//	@Tags		Tag
//	@security	BasicAuth
//	@Accept		json
//	@Produce	json
//	@Router		/namespaces/{namespace_id}/tags/{id} [get]
//	@Param		namespace_id	path		number	true	"Namespace id"
//	@Param		repository_id	path		number	false	"Repository id"
//	@Param		id				path		number	true	"Tag id"
//	@Success	200				{object}	types.TagItem
//	@Failure	404				{object}	errcode.ErrCode
//	@Failure	500				{object}	errcode.ErrCode
func (h *handler) GetTag(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	iuser := c.Get(consts.ContextUser)
	if iuser == nil {
		log.Error().Msg("Get user from header failed")
		return errcode.NewHTTPError(c, errcode.HTTPErrCodeUnauthorized)
	}
	user, ok := iuser.(*models.User)
	if !ok {
		log.Error().Msg("Convert user from header failed")
		return errcode.NewHTTPError(c, errcode.HTTPErrCodeUnauthorized)
	}

	var req types.GetTagRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return errcode.NewHTTPError(c, errcode.HTTPErrCodeBadRequest, err.Error())
	}

	authChecked, err := h.AuthServiceFactory.New().Tag(ptr.To(user), req.ID, enums.AuthRead)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(errors.New(utils.UnwrapJoinedErrors(err))).Int64("NamespaceID", req.NamespaceID).Msg("Namespace not found")
			return errcode.NewHTTPError(c, errcode.HTTPErrCodeNotFound, fmt.Sprintf("Namespace(%d) not found: %v", req.NamespaceID, err))
		}
		log.Error().Err(errors.New(utils.UnwrapJoinedErrors(err))).Int64("NamespaceID", req.NamespaceID).Msg("Namespace find failed")
		return errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError, fmt.Sprintf("Namespace(%d) find failed: %v", req.NamespaceID, err))
	}
	if !authChecked {
		log.Error().Int64("UserID", user.ID).Int64("RepositoryID", req.ID).Msg("Auth check failed")
		return errcode.NewHTTPError(c, errcode.HTTPErrCodeUnauthorized, "No permission with this api")
	}

	tagService := h.TagServiceFactory.New()
	tag, err := tagService.GetByID(ctx, req.ID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Error().Err(err).Msg("Tag not found")
			return errcode.NewHTTPError(c, errcode.HTTPErrCodeNotFound, err.Error())
		}
		log.Error().Err(err).Msg("Get tag from db failed")
		return errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError, err.Error())
	}

	var artifacts = make([]types.TagItemArtifact, 0, len(tag.Artifact.ArtifactSubs))
	for _, item := range tag.Artifact.ArtifactSubs {
		artifacts = append(artifacts, types.TagItemArtifact{
			ID:            item.ID,
			Digest:        item.Digest,
			Raw:           string(item.Raw),
			ConfigRaw:     string(item.ConfigRaw),
			Size:          item.Size,
			BlobSize:      item.BlobsSize,
			LastPull:      time.Unix(0, int64(time.Millisecond)*item.LastPull).UTC().Format(consts.DefaultTimePattern),
			PushedAt:      time.Unix(0, int64(time.Millisecond)*item.PushedAt).UTC().Format(consts.DefaultTimePattern),
			Vulnerability: string(item.Vulnerability.Result),
			Sbom:          string(item.Sbom.Result),
			CreatedAt:     time.Unix(0, int64(time.Millisecond)*item.CreatedAt).UTC().Format(consts.DefaultTimePattern),
			UpdatedAt:     time.Unix(0, int64(time.Millisecond)*item.CreatedAt).UTC().Format(consts.DefaultTimePattern),
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
			LastPull:      time.Unix(0, int64(time.Millisecond)*tag.Artifact.LastPull).UTC().Format(consts.DefaultTimePattern),
			PushedAt:      time.Unix(0, int64(time.Millisecond)*tag.Artifact.PushedAt).UTC().Format(consts.DefaultTimePattern),
			Vulnerability: string(tag.Artifact.Vulnerability.Result),
			Sbom:          string(tag.Artifact.Sbom.Result),
			CreatedAt:     time.Unix(0, int64(time.Millisecond)*tag.Artifact.CreatedAt).UTC().Format(consts.DefaultTimePattern),
			UpdatedAt:     time.Unix(0, int64(time.Millisecond)*tag.Artifact.CreatedAt).UTC().Format(consts.DefaultTimePattern),
		},
		Artifacts: artifacts,
		PushedAt:  time.Unix(0, int64(time.Millisecond)*tag.PushedAt).UTC().Format(consts.DefaultTimePattern),
		CreatedAt: time.Unix(0, int64(time.Millisecond)*tag.CreatedAt).UTC().Format(consts.DefaultTimePattern),
		UpdatedAt: time.Unix(0, int64(time.Millisecond)*tag.CreatedAt).UTC().Format(consts.DefaultTimePattern),
	})
}
