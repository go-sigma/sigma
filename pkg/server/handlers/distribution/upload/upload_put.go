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
	"github.com/go-sigma/sigma/pkg/distribution/reference"
	"github.com/go-sigma/sigma/pkg/server/errcode"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/validators"
)

// PutUpload handles the put upload request
func (h *handler) PutUpload(c *gin.Context) {
	ctx := c.Request.Context()

	user, needRet := utils.GetUserFromCtx(c, utils.UserCtxErrorDistribution)
	if needRet {
		return
	}

	uri := c.Request.URL.Path
	c.Writer.Header().Set(consts.HeaderLocation, fmt.Sprintf("%s://%s%s", utils.Scheme(c), c.Request.Host, uri))

	uploadID := strings.TrimPrefix(uri[strings.LastIndex(uri, "/"):], "/")
	if strings.Contains(uploadID, "/") {
		errcode.NewDSError(c, errcode.DSErrCodeBlobUploadInvalid)
		return
	}

	repository := strings.TrimPrefix(strings.TrimSuffix(uri[:strings.LastIndex(uri, "/")], "/blobs"), "/v2/")
	_, namespace, _, _, err := reference.Parse(repository)
	if err != nil {
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
	if !(validators.ValidateNamespaceRaw(namespace) && validators.ValidateRepositoryRaw(repository)) { // nolint: staticcheck
		slog.Error("repository must container a valid namespace", "err", err, "repository", repository)
		errcode.NewDSError(c, errcode.DSErrCodeManifestWithNamespace)
		return
	}

	authChecked, err := h.Authorizer.Namespace(ctx, *user, namespaceObj.ID, enums.AuthManage)
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

	dgest, err := digest.Parse(c.Query("digest"))
	if err != nil {
		slog.Error("parse digest failed", "err", err, "digest", c.Query("digest"))
		errcode.NewDSError(c, errcode.DSErrCodeDigestInvalid)
		return
	}
	c.Writer.Header().Set(consts.ContentDigest, dgest.String())

	length, err := utils.GetContentLength(c.Request)
	if err != nil {
		slog.Error("get content length failed", "err", err)
		errcode.NewDSError(c, errcode.DSErrCodeUnknown)
		return
	}

	contentType := c.Request.Header.Get(consts.HeaderContentType)

	sizeBefore, sizeUploaded, err := h.UploadSvc.PutUpload(ctx, uploadID, c.Query("digest"), c.Request.Body, repository, contentType, length)
	if err != nil {
		if e, ok := errcode.AsType[errcode.ErrCode](err); ok {
			errcode.NewDSError(c, e)
			return
		}
		errcode.NewDSError(c, errcode.DSErrCodeUnknown)
		return
	}

	if sizeUploaded > 0 {
		c.Writer.Header().Set(consts.HeaderContentRange, fmt.Sprintf("%d-%d", sizeBefore, sizeBefore+sizeUploaded))
	}

	c.Status(http.StatusCreated)
}
