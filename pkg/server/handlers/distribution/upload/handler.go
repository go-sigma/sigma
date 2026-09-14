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
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/authz"
	"github.com/go-sigma/sigma/pkg/server/handlers/distribution"
	svcupload "github.com/go-sigma/sigma/pkg/service/distribution/upload"
	"github.com/go-sigma/sigma/pkg/utils"
)

// Handler is the interface for the distribution blob handlers
type Handler interface {
	// DeleteUpload cancels an upload session, discarding its partial content and responding 204 No Content.
	DeleteUpload(ctx *gin.Context)
	// GetUpload reports the status of an upload session, responding 204 No Content with the upload UUID and Location headers, or a distribution error when the session is unknown.
	GetUpload(ctx *gin.Context)
	// PatchUpload appends the request body as a chunk to an existing upload session and responds 202 Accepted with an inclusive Range header covering the bytes received so far.
	PatchUpload(ctx *gin.Context)
	// PostUpload starts a blob upload session, responding 202 Accepted with a Docker-Upload-UUID and Location header, or performs a single-shot upload that stores the blob directly when a digest query parameter is present; cross-repository blob mounting is not supported.
	PostUpload(ctx *gin.Context)
	// PutUpload finalises an upload session under the digest query parameter, appending any remaining request body and responding 201 Created with a Docker-Content-Digest header.
	PutUpload(ctx *gin.Context)
}

var _ Handler = &handler{}

type handler struct {
	dig.In

	UploadSvc  svcupload.Service
	Authorizer authz.Authorizer
}

type factory struct{}

// Initialize initializes the distribution manifest handlers
func (f factory) Initialize(c *gin.Context, digCon *dig.Container) error {
	method := c.Request.Method
	uri := c.Request.RequestURI

	return digCon.Invoke(func(h handler) error {
		if method == http.MethodPost && strings.HasSuffix(uri, "blobs/uploads/") {
			h.PostUpload(c)
			return nil
		}

		urix := uri[:strings.LastIndex(uri, "/")]
		if strings.HasSuffix(urix, "/blobs/uploads") {
			switch method {
			case http.MethodGet:
				h.GetUpload(c)
				return nil
			case http.MethodPatch:
				h.PatchUpload(c)
				return nil
			case http.MethodPut:
				h.PutUpload(c)
				return nil
			case http.MethodDelete:
				h.DeleteUpload(c)
				return nil
			default:
				c.String(http.StatusMethodNotAllowed, "method Not Allowed")
				return nil
			}
		}
		return distribution.ErrNext
	})
}

func init() {
	utils.PanicIf(distribution.RegisterRouterFactory(&factory{}, 2))
}
