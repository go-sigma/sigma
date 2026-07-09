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

// GetManifest handles the get manifest request
func (h *handler) GetManifest(c *gin.Context) {
	ctx := c.Request.Context()

	user, needRet := utils.GetUserFromCtx(c, utils.UserCtxErrorDistribution)
	if needRet {
		return
	}

	uri := c.Request.URL.Path

	repository := strings.TrimPrefix(strings.TrimSuffix(uri[:strings.LastIndex(uri, "/")], "/manifests"), "/v2/")
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
	namespaceObj, err := h.ManifestSvc.GetNamespaceByName(ctx, namespace)
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
		errcode.NewDSError(c, errcode.DSErrCodeDenied)
		return
	}
	if !authChecked {
		slog.Error("auth check failed", "UserID", user.ID, "NamespaceID", namespaceObj.ID)
		errcode.NewDSError(c, errcode.DSErrCodeDenied)
		return
	}

	ref := strings.TrimPrefix(uri[strings.LastIndex(uri, "/"):], "/")
	if _, err := digest.Parse(ref); err != nil && !consts.TagRegexp.MatchString(ref) {
		slog.Error("invalid digest or tag", "err", err, "ref", ref)
		errcode.NewDSError(c, errcode.DSErrCodeTagInvalid)
		return
	}

	refs := h.parseRef(ref)

	body, contentType, tag, err := h.ManifestSvc.GetManifest(ctx, namespaceObj.ID, repository, ref)
	if err != nil {
		if e, ok := errcode.AsType[errcode.ErrCode](err); ok && e.Code == errcode.DSErrCodeManifestUnknown.Code && h.Config.Proxy.Enabled {
			h.getManifestFallbackProxy(c, refs)
			return
		}
		h.dsError(c, err)
		return
	}

	// mirror: when the proxy is enabled, also check the manifest in the remote registry
	if h.Config.Proxy.Enabled {
		h.getManifestFallbackProxy(c, refs)
		return
	}

	var dgst string
	if tag != nil {
		dgst = tag.Artifact.Digest
	} else {
		dgst = refs.Digest.String()
	}

	c.Writer.Header().Set(consts.HeaderContentType, contentType)
	c.Writer.Header().Set(consts.HeaderContentLength, strconv.FormatInt(int64(len(body)), 10))
	c.Writer.Header().Set(consts.ContentDigest, dgst)
	c.Data(http.StatusOK, contentType, body)
}

// getManifestFallbackProxy ...
func (h *handler) getManifestFallbackProxy(c *gin.Context, refs Refs) {
	statusCode, header, bodyBytes, err := h.fallbackProxy(c)
	if err != nil {
		slog.Error("fallback proxy failed", "err", err, "refs", refs, "status", statusCode)
		errcode.NewDSError(c, errcode.DSErrCodeUnknown)
		return
	}
	switch statusCode {
	case http.StatusOK:
		c.Writer.Header().Set(consts.ContentDigest, header.Get(consts.ContentDigest))
		c.Writer.Header().Set(consts.HeaderETag, header.Get(consts.HeaderETag))
		c.Data(http.StatusOK, header.Get(consts.HeaderContentType), bodyBytes)
		return
	case http.StatusNotFound:
		errcode.NewDSError(c, errcode.DSErrCodeManifestUnknown)
		return
	}
	slog.Error("fallback proxy failed", "refs", refs, "statusCode", statusCode)
	errcode.NewDSError(c, errcode.DSErrCodeUnknown)
}
