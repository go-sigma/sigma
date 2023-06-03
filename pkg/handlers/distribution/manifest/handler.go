// Copyright 2023 XImager
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

	"github.com/labstack/echo/v4"

	"github.com/ximager/ximager/pkg/dal/dao"
	"github.com/ximager/ximager/pkg/handlers/distribution"
	"github.com/ximager/ximager/pkg/utils"
)

// Handlers is the interface for the distribution manifest handlers
type Handlers interface {
	// GetManifest ...
	GetManifest(ctx echo.Context) error
	// HeadManifest ...
	HeadManifest(ctx echo.Context) error
	// PutManifest ...
	PutManifest(ctx echo.Context) error
	// DeleteManifest ...
	DeleteManifest(ctx echo.Context) error
}

var _ Handlers = &handler{}

type handler struct {
	artifactServiceFactory dao.ArtifactServiceFactory
}

type inject struct {
	artifactServiceFactory dao.ArtifactServiceFactory
}

// New creates a new instance of the distribution manifest handlers
func handlerNew(injects ...inject) Handlers {
	artifactServiceFactory := dao.NewArtifactServiceFactory()
	if len(injects) > 0 {
		ij := injects[0]
		if ij.artifactServiceFactory != nil {
			artifactServiceFactory = ij.artifactServiceFactory
		}
	}
	return &handler{
		artifactServiceFactory: artifactServiceFactory,
	}
}

type factory struct{}

// Initialize initializes the distribution manifest handlers
func (f factory) Initialize(c echo.Context) error {
	method := c.Request().Method
	uri := c.Request().RequestURI
	urix := uri[:strings.LastIndex(uri, "/")]
	manifestHandler := handlerNew()
	if strings.HasSuffix(urix, "/manifests") {
		switch method {
		case http.MethodGet:
			return manifestHandler.GetManifest(c)
		case http.MethodHead:
			return manifestHandler.HeadManifest(c)
		case http.MethodPut:
			return manifestHandler.PutManifest(c)
		case http.MethodDelete:
			return manifestHandler.DeleteManifest(c)
		default:
			return c.String(http.StatusMethodNotAllowed, "Method Not Allowed")
		}
	}
	return distribution.ErrNext
}

func init() {
	utils.PanicIf(distribution.RegisterRouterFactory(&factory{}, 4))
}
