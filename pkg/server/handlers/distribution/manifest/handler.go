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
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/authz"
	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/server/handlers/distribution"
	svcmanifest "github.com/go-sigma/sigma/pkg/service/distribution/manifest"
	"github.com/go-sigma/sigma/pkg/utils"
)

// Handler is the interface for the distribution manifest handlers
type Handler interface {
	// GetManifest returns the raw manifest stored for the given tag or digest together with its stored content type and digest headers, records a pull event, and falls back to the upstream registry proxy when mirroring is enabled.
	GetManifest(ctx *gin.Context)
	// HeadManifest returns the manifest metadata (content type, content length and digest headers) for the given tag or digest without a body and without recording a pull event.
	HeadManifest(ctx *gin.Context)
	// PutManifest stores the request body as a manifest for the given tag or digest, handling image manifests, image indexes and artifacts, advances the tag, and responds 201 Created with Docker-Content-Digest and Location headers.
	PutManifest(ctx *gin.Context)
	// DeleteManifest deletes the manifest referenced by the path tag or digest, removing only the tag when a tag is given and the artifact plus every tag pointing to it when a digest is given, and responds 202 Accepted.
	DeleteManifest(ctx *gin.Context)
	// GetReferrer implements the OCI referrers API, returning an image index of the manifests whose subject is the path digest, filtered by the optional artifactType query parameter echoed back in OCI-Filters-Applied.
	GetReferrer(ctx *gin.Context)
}

var _ Handler = &handler{}

type handler struct {
	dig.In

	Config      *config.Configuration
	ManifestSvc svcmanifest.Service
	Authorizer  authz.Authorizer
}

type factory struct{}

// Initialize initializes the distribution manifest handlers
func (f factory) Initialize(c *gin.Context, digCon *dig.Container) error {
	method := c.Request.Method
	uri := c.Request.RequestURI
	urix := uri[:strings.LastIndex(uri, "/")]
	return digCon.Invoke(func(h handler) error {
		if strings.HasSuffix(urix, "/manifests") {
			switch method {
			case http.MethodGet:
				h.GetManifest(c)
				return nil
			case http.MethodHead:
				h.HeadManifest(c)
				return nil
			case http.MethodPut:
				h.PutManifest(c)
				return nil
			case http.MethodDelete:
				h.DeleteManifest(c)
				return nil
			default:
				c.String(http.StatusMethodNotAllowed, "Method Not Allowed")
				return nil
			}
		} else if strings.HasSuffix(urix, "/referrers") && method == http.MethodGet {
			h.GetReferrer(c)
			return nil
		}
		return distribution.ErrNext
	})
}

func init() {
	utils.PanicIf(distribution.RegisterRouterFactory(&factory{}, 4))
}
