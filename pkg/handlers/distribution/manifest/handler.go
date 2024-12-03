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

	"github.com/labstack/echo/v4"
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/auth"
	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/dal/dao"
	"github.com/go-sigma/sigma/pkg/handlers/distribution"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

// Handler is the interface for the distribution manifest handlers
type Handler interface {
	// GetManifest ...
	GetManifest(ctx echo.Context) error
	// HeadManifest ...
	HeadManifest(ctx echo.Context) error
	// PutManifest ...
	PutManifest(ctx echo.Context) error
	// DeleteManifest ...
	DeleteManifest(ctx echo.Context) error
	// GetReferrer ...
	GetReferrer(ctx echo.Context) error
}

var _ Handler = &handler{}

type handler struct {
	dig.In

	Config                   configs.Configuration
	AuthServiceFactory       auth.AuthServiceFactory
	AuditServiceFactory      dao.AuditServiceFactory
	NamespaceServiceFactory  dao.NamespaceServiceFactory
	RepositoryServiceFactory dao.RepositoryServiceFactory
	TagServiceFactory        dao.TagServiceFactory
	ArtifactServiceFactory   dao.ArtifactServiceFactory
	BlobServiceFactory       dao.BlobServiceFactory
}

// New creates a new instance of the distribution manifest handlers
func handlerNew(digCon *dig.Container) Handler {
	return ptr.Of(utils.MustGetObjFromDigCon[handler](digCon))
}

type factory struct{}

// Initialize initializes the distribution manifest handlers
func (f factory) Initialize(c echo.Context, digCon *dig.Container) error {
	method := c.Request().Method
	uri := c.Request().RequestURI
	urix := uri[:strings.LastIndex(uri, "/")]
	manifestHandler := handlerNew(digCon)
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
	} else if strings.HasSuffix(urix, "/referrers") && method == http.MethodGet {
		return manifestHandler.GetReferrer(c)
	}
	return distribution.ErrNext
}

func init() {
	utils.PanicIf(distribution.RegisterRouterFactory(&factory{}, 4))
}
