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
	// GetManifest ...
	GetManifest(ctx *gin.Context)
	// HeadManifest ...
	HeadManifest(ctx *gin.Context)
	// PutManifest ...
	PutManifest(ctx *gin.Context)
	// DeleteManifest ...
	DeleteManifest(ctx *gin.Context)
	// GetReferrer ...
	GetReferrer(ctx *gin.Context)
}

var _ Handler = &handler{}

type handler struct {
	dig.In

	Config      *config.Configuration
	ManifestSvc svcmanifest.DistributionManifestService
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
