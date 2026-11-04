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

package upload

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/opencontainers/go-digest"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/distribution/reference"
	"github.com/go-sigma/sigma/pkg/server/errcode"
	"github.com/go-sigma/sigma/pkg/server/validators"
	"github.com/go-sigma/sigma/pkg/utils"
)

// PostUpload creates a new upload.
func (h *handler) PostUpload(c *gin.Context) {
	ctx := c.Request.Context()

	user, needRet := utils.GetUserFromCtx(c, utils.UserCtxErrorDistribution)
	if needRet {
		return
	}

	host := c.Request.Host
	uri := c.Request.URL.Path
	protocol := utils.Scheme(c)

	repository := strings.TrimPrefix(strings.TrimSuffix(uri[:strings.LastIndex(uri, "/")], "/blobs"), "/v2/")
	_, namespace, _, _, err := reference.Parse(repository)
	if err != nil {
		slog.Error("repository must container a valid namespace", "err", err, "repository", repository)
		errcode.NewDSError(c, errcode.DSErrCodeManifestWithNamespace)
		return
	}
	if !(validators.ValidateNamespaceRaw(namespace) && validators.ValidateRepositoryRaw(repository)) { // nolint: staticcheck
		slog.Error("repository must container a valid namespace", "err", err, "repository", repository)
		errcode.NewDSError(c, errcode.DSErrCodeManifestWithNamespace)
		return
	}
	namespaceObj, err := h.UploadSvc.GetNamespaceByName(ctx, namespace)
	if err != nil {
		slog.Error("get repository by name failed", "err", err, "repository", repository)
		errcode.NewDSError(c, errcode.DSErrCodeBlobUnknown)
		return
	}

	authChecked, err := h.Authorizer.Namespace(ctx, *user, namespaceObj.ID, enums.AuthManage)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Error("resource not found", "err", err)
			errcode.NewDSError(c, errcode.GenDSErrCodeResourceNotFound(err))
			return
		}
		errcode.NewDSError(c, errcode.DSErrCodeUnknown)
		return
	}
	if !authChecked {
		slog.Error("auth check failed", "UserID", user.ID, "NamespaceID", namespaceObj.ID)
		errcode.NewDSError(c, errcode.DSErrCodeDenied)
		return
	}

	// Cross-repository blob mount: when the blob already exists in the source
	// repository, answer 201 Created instead of starting an upload session. Any reason
	// not to mount falls back to a regular upload, which is what the distribution spec
	// expects from a registry that cannot serve the mount.
	if mount, from := c.Query("mount"), c.Query("from"); mount != "" && from != "" {
		if h.mountBlob(c, ctx, user, strings.TrimSuffix(repository, "/blobs/uploads"), mount, from) {
			return
		}
	}

	digestStr := c.Query("digest")
	if digestStr != "" {
		dgest, err := digest.Parse(digestStr)
		if err != nil {
			slog.Error("parse digest failed", "err", err, "digest", digestStr)
			errcode.NewDSError(c, errcode.DSErrCodeBlobUploadInvalid)
			return
		}
		c.Writer.Header().Set(consts.ContentDigest, dgest.String())
	}

	contentType := c.Request.Header.Get(consts.HeaderContentType)

	uploadID, err := h.UploadSvc.PostUpload(ctx, repository, digestStr, c.Request.Body, contentType)
	if err != nil {
		if e, ok := errcode.AsType[errcode.ErrCode](err); ok {
			errcode.NewDSError(c, e)
			return
		}
		errcode.NewDSError(c, errcode.DSErrCodeUnknown)
		return
	}

	c.Writer.Header().Set("Docker-Upload-UUID", uploadID)
	c.Writer.Header().Set(consts.HeaderLocation, fmt.Sprintf("%s://%s%s%s", protocol, host, uri, uploadID))
	c.Writer.Header().Set(consts.HeaderRange, "0-0")

	c.Status(http.StatusAccepted)
}

// mountBlob serves a cross-repository blob mount request. It returns true after writing a
// 201 Created response; otherwise it returns false and the caller starts a regular upload
// session instead. Mounting requires read access to the source namespace and an artifact
// in the source repository that references the blob.
func (h *handler) mountBlob(c *gin.Context, ctx context.Context, user *models.User, targetRepository, mount, from string) bool {
	dgest, err := digest.Parse(mount)
	if err != nil {
		slog.Debug("ignore mount request with invalid digest", "err", err, "digest", mount)
		return false
	}

	fromRepository := strings.Trim(strings.TrimSpace(from), "/")
	_, fromNamespace, _, _, err := reference.Parse(fromRepository)
	if err != nil {
		slog.Debug("ignore mount request with invalid source repository", "err", err, "from", from)
		return false
	}
	if !(validators.ValidateNamespaceRaw(fromNamespace) && validators.ValidateRepositoryRaw(fromRepository)) { // nolint: staticcheck
		return false
	}

	fromNamespaceObj, err := h.UploadSvc.GetNamespaceByName(ctx, fromNamespace)
	if err != nil {
		slog.Debug("ignore mount request with unknown source namespace", "err", err, "namespace", fromNamespace)
		return false
	}
	authChecked, err := h.Authorizer.Namespace(ctx, *user, fromNamespaceObj.ID, enums.AuthRead)
	if err != nil || !authChecked {
		slog.Debug("ignore mount request without read access to source namespace", "err", err, "namespace", fromNamespace, "user", user.ID)
		return false
	}

	exists, err := h.UploadSvc.CanMountBlob(ctx, fromRepository, dgest.String())
	if err != nil {
		slog.Error("check blob in source repository failed", "err", err, "from", fromRepository, "digest", dgest.String())
		return false
	}
	if !exists {
		slog.Debug("ignore mount request, blob not in source repository", "from", fromRepository, "digest", dgest.String())
		return false
	}

	c.Writer.Header().Set(consts.HeaderLocation, fmt.Sprintf("%s://%s/v2/%s/blobs/%s", utils.Scheme(c), c.Request.Host, targetRepository, dgest.String()))
	c.Writer.Header().Set(consts.ContentDigest, dgest.String())
	c.Status(http.StatusCreated)
	return true
}
