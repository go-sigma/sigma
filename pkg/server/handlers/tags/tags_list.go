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
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/server/errcode"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
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
//	@Success	200				{object}	api.CommonList{items=[]api.TagItem}
//	@Failure	404				{object}	errcode.ErrCode
//	@Failure	500				{object}	errcode.ErrCode
func (h *handler) ListTag(c *gin.Context, req *api.ListTagRequest) {
	ctx := c.Request.Context()

	user, ok := utils.GetFromCtx[*models.User](c, consts.ContextUser)
	if !ok {
		slog.Error("get user from context failed")
		errcode.NewHTTPError(c, errcode.HTTPErrCodeUnauthorized)
		return
	}

	authChecked, err := h.Authorizer.Repository(ctx, *user, req.RepositoryID, enums.AuthRead)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Error("namespace not found", "err", err, "NamespaceID", req.NamespaceID)
			errcode.NewHTTPError(c, errcode.HTTPErrCodeNotFound, fmt.Sprintf("Namespace(%s) not found: %v", req.NamespaceID, err))
			return
		}
		slog.Error("namespace find failed", "err", err, "NamespaceID", req.NamespaceID)
		errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError, fmt.Sprintf("Namespace(%s) find failed: %v", req.NamespaceID, err))
		return
	}
	if !authChecked {
		slog.Error("auth check failed", "UserID", user.ID, "RepositoryID", req.RepositoryID)
		errcode.NewHTTPError(c, errcode.HTTPErrCodeUnauthorized, "No permission with this api")
		return
	}

	tagSvc := h.TagSvc
	tags, total, err := tagSvc.ListTags(ctx, req.NamespaceID, req.RepositoryID, req.Name, req.Type, req.Pagination, req.Sortable)
	if err != nil {
		if e, ok := errcode.AsType[errcode.ErrCode](err); ok {
			errcode.NewHTTPError(c, e)
			return
		}
		errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError, err.Error())
		return
	}

	var resp = make([]any, 0, len(tags))
	for _, tag := range tags {
		if tag.Artifact == nil {
			slog.Error("some tag's artifact reference invalid", "repository_id", req.RepositoryID, "tag", tag.Name)
			continue
		}
		var artifacts []api.TagItemArtifact
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
				ID:              item.ID,
				Digest:          item.Digest,
				MediaType:       item.ContentType,
				Raw:             string(raw),
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
		raw, err := tagSvc.GetArtifactRaw(ctx, tag.Artifact.Digest)
		if err != nil {
			if e, ok := errcode.AsType[errcode.ErrCode](err); ok {
				errcode.NewHTTPError(c, e)
				return
			}
			errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError, err.Error())
			return
		}
		resp = append(resp, api.TagItem{
			ID:   tag.ID,
			Name: tag.Name,
			Artifact: api.TagItemArtifact{
				ID:              tag.Artifact.ID,
				Digest:          tag.Artifact.Digest,
				MediaType:       tag.Artifact.ContentType,
				Raw:             string(raw),
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

	c.JSON(http.StatusOK, api.CommonList{Total: total, Items: resp})
}
