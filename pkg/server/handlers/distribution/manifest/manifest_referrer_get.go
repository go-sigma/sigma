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
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/opencontainers/go-digest"
	imgspecv1 "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/go-sigma/sigma/pkg/server/errcode"
)

// GetReferrer ...
func (h *handler) GetReferrer(c *gin.Context) {
	ctx := c.Request.Context()
	uri := c.Request.URL.Path
	ref := strings.TrimPrefix(uri[strings.LastIndex(uri, "/"):], "/")
	repository := strings.TrimPrefix(strings.TrimSuffix(uri[:strings.LastIndex(uri, "/")], "/referrers"), "/v2/")

	if _, err := digest.Parse(ref); err != nil {
		slog.Error("digest is invalid", "err", err, "ref", ref)
		errcode.NewDSError(c, errcode.DSErrCodeDigestInvalid)
		return
	}

	artifactType := c.Query("artifactType")
	c.Writer.Header().Set("OCI-Filters-Applied", artifactType)

	body, err := h.ManifestSvc.GetReferrer(ctx, repository, ref, strings.Split(artifactType, ","))
	if err != nil {
		h.dsError(c, err)
		return
	}

	c.Data(http.StatusOK, imgspecv1.MediaTypeImageIndex, body)
}
