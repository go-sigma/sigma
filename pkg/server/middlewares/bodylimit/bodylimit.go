// Copyright 2026 sigma
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

package bodylimit

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/server/errcode"
)

// BodyLimit limits request bodies according to the configured route class.
func BodyLimit(cfg config.ConfigurationHTTPBodyLimit) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !cfg.IsEnabled() {
			c.Next()
			return
		}

		limit := limitForRequest(cfg, c.Request)
		if limit <= 0 {
			c.Next()
			return
		}

		if c.Request.ContentLength > limit {
			writePayloadTooLarge(c, limit)
			c.Abort()
			return
		}

		if c.Request.Body != nil {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, limit)
		}
		c.Next()
	}
}

func limitForRequest(cfg config.ConfigurationHTTPBodyLimit, req *http.Request) int64 {
	if req == nil || req.URL == nil {
		return cfg.DefaultLimit()
	}
	if isBlobUpload(req) {
		return cfg.BlobLimit()
	}
	if isManifestUpload(req) {
		return cfg.ManifestLimit()
	}
	return cfg.DefaultLimit()
}

func isBlobUpload(req *http.Request) bool {
	path := req.URL.Path
	switch req.Method {
	case http.MethodPost:
		return strings.HasSuffix(path, "/blobs/uploads/")
	case http.MethodPatch, http.MethodPut:
		lastSlash := strings.LastIndex(path, "/")
		return lastSlash > 0 && strings.HasSuffix(path[:lastSlash], "/blobs/uploads")
	default:
		return false
	}
}

func isManifestUpload(req *http.Request) bool {
	if req.Method != http.MethodPut {
		return false
	}
	path := req.URL.Path
	lastSlash := strings.LastIndex(path, "/")
	return lastSlash > 0 && strings.HasSuffix(path[:lastSlash], "/manifests")
}

func writePayloadTooLarge(c *gin.Context, limit int64) {
	message := fmt.Sprintf("request body exceeds the configured limit of %d bytes", limit)
	path := c.Request.URL.Path
	if strings.HasPrefix(path, "/v2/") || path == "/v2" {
		code := errcode.DSErrCodeDenied
		code.HTTPStatusCode = http.StatusRequestEntityTooLarge
		code.Description = message
		errcode.NewDSError(c, code)
		return
	}

	errcode.NewHTTPError(c, errcode.ErrCode{
		Code:           "PAYLOAD_TOO_LARGE",
		Title:          "payload too large",
		Description:    message,
		HTTPStatusCode: http.StatusRequestEntityTooLarge,
	})
}
