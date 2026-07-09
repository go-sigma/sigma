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

package tags

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/server/errcode"
	"github.com/go-sigma/sigma/pkg/utils"
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
//	@Success	200				{object}	api.TagItem
//	@Failure	404				{object}	errcode.ErrCode
//	@Failure	500				{object}	errcode.ErrCode
func (h *handler) GetTag(c *gin.Context, req *api.GetTagRequest) {
	ctx := c.Request.Context()

	user, ok := utils.GetFromCtx[*models.User](c, consts.ContextUser)
	if !ok {
		slog.Error("get user from context failed")
		errcode.NewHTTPError(c, errcode.HTTPErrCodeUnauthorized)
		return
	}

	authChecked, err := h.Authorizer.Tag(ctx, *user, req.ID, enums.AuthRead)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Error("namespace not found", "err", errors.New(utils.UnwrapJoinedErrors(err)), "NamespaceID", req.NamespaceID)
			errcode.NewHTTPError(c, errcode.HTTPErrCodeNotFound, fmt.Sprintf("Namespace(%s) not found: %v", req.NamespaceID, err))
			return
		}
		slog.Error("namespace find failed", "err", errors.New(utils.UnwrapJoinedErrors(err)), "NamespaceID", req.NamespaceID)
		errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError, fmt.Sprintf("Namespace(%s) find failed: %v", req.NamespaceID, err))
		return
	}
	if !authChecked {
		slog.Error("auth check failed", "UserID", user.ID, "RepositoryID", req.ID)
		errcode.NewHTTPError(c, errcode.HTTPErrCodeUnauthorized, "No permission with this api")
		return
	}

	tagSvc := h.TagSvc
	tag, err := tagSvc.GetTag(ctx, req.ID)
	if err != nil {
		if e, ok := errcode.AsType[errcode.ErrCode](err); ok {
			errcode.NewHTTPError(c, e)
			return
		}
		errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError, err.Error())
		return
	}

	var artifacts = make([]api.TagItemArtifact, 0, len(tag.Artifact.ArtifactSubs))
	for _, item := range tag.Artifact.ArtifactSubs {
		raw, err := tagSvc.GetArtifactRaw(ctx, item.Digest)
		if err != nil {
			if e, ok := errcode.AsType[errcode.ErrCode](err); ok {
				errcode.NewHTTPError(c, e)
				return
			}
			errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError, err.Error())
			return
		}
		artifacts = append(artifacts, api.TagItemArtifact{
			ID:            item.ID,
			Digest:        item.Digest,
			Raw:           string(raw),
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

	raw, err := tagSvc.GetArtifactRaw(ctx, tag.Artifact.Digest)
	if err != nil {
		if e, ok := errcode.AsType[errcode.ErrCode](err); ok {
			errcode.NewHTTPError(c, e)
			return
		}
		errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError, err.Error())
		return
	}

	c.JSON(200, api.TagItem{
		ID:   tag.ID,
		Name: tag.Name,
		Artifact: api.TagItemArtifact{
			ID:            tag.Artifact.ID,
			Digest:        tag.Artifact.Digest,
			Raw:           string(raw),
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
