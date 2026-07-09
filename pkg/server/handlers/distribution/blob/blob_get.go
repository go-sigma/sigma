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

package blob

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/opencontainers/go-digest"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/distribution/reference"
	"github.com/go-sigma/sigma/pkg/server/errcode"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/validators"
)

// GetBlob returns the blob's size and digest.
func (h *handler) GetBlob(c *gin.Context) {
	ctx := c.Request.Context()

	user, needRet := utils.GetUserFromCtx(c, utils.UserCtxErrorDistribution)
	if needRet {
		return
	}

	uri := c.Request.URL.Path

	repository := strings.TrimPrefix(strings.TrimSuffix(uri[:strings.LastIndex(uri, "/")], "/blobs"), "/v2/")
	_, namespace, _, _, err := reference.Parse(repository)
	if err != nil {
		slog.Error("repository must container a valid namespace", "err", err, "Repository", repository)
		errcode.NewDSError(c, errcode.DSErrCodeManifestWithNamespace)
		return
	}
	if !(validators.ValidateNamespaceRaw(namespace) && validators.ValidateRepositoryRaw(repository)) { // nolint: staticcheck
		slog.Error("repository must container a valid namespace", "err", err, "Repository", repository)
		errcode.NewDSError(c, errcode.DSErrCodeManifestWithNamespace)
		return
	}
	namespaceObj, err := h.BlobSvc.GetNamespaceByName(ctx, namespace)
	if err != nil {
		slog.Error("get repository by name failed", "err", err, "Name", repository)
		errcode.NewDSError(c, errcode.DSErrCodeBlobUnknown)
		return
	}

	authChecked, err := h.Authorizer.Namespace(ctx, *user, namespaceObj.ID, enums.AuthRead)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Error("resource not found", "err", errors.New(utils.UnwrapJoinedErrors(err)))
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

	dgest, err := digest.Parse(strings.TrimPrefix(uri[strings.LastIndex(uri, "/"):], "/"))
	if err != nil {
		slog.Error("parse digest failed", "err", err, "digest", c.Query("digest"))
		errcode.NewDSError(c, errcode.DSErrCodeDigestInvalid)
		return
	}

	reader, blob, redirectURL, err := h.BlobSvc.GetBlob(ctx, namespaceObj.ID, dgest.String(), c.Request.Method, c.Request.URL.Path)
	if err != nil {
		h.dsError(c, err)
		return
	}
	if redirectURL != "" {
		c.Redirect(http.StatusMovedPermanently, redirectURL)
		return
	}
	defer func() {
		if reader != nil {
			_ = reader.Close()
		}
	}()

	c.Writer.Header().Set(consts.HeaderContentLength, strconv.FormatInt(blob.Size, 10))
	c.Writer.Header().Set(consts.HeaderContentType, blob.ContentType)
	c.Writer.Header().Set(consts.ContentDigest, blob.Digest)
	c.Writer.WriteHeader(http.StatusOK)
	_, err = io.Copy(c.Writer, reader)
	if err != nil {
		slog.Error("stream blob failed", "err", err, "digest", dgest.String())
		errcode.NewDSError(c, errcode.DSErrCodeUnknown)
		return
	}
}
