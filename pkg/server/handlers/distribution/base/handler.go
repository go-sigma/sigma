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

package distribution

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/auth"
	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/dal/dao"
	"github.com/go-sigma/sigma/pkg/server/handlers/distribution"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

// Handler is the interface for the distribution handlers
type Handler interface {
	// GetHealthy handles the get healthy request
	GetHealthy(ctx echo.Context) error
	// ListTags handles the list tags request
	ListTags(ctx echo.Context) error
	// ListRepositories handles the list repositories request
	ListRepositories(ctx echo.Context) error
}

var _ Handler = &handler{}

type handler struct {
	dig.In

	Config                   configs.Configuration
	AuthServiceFactory       auth.AuthServiceFactory
	TagServiceFactory        dao.TagServiceFactory
	RepositoryServiceFactory dao.RepositoryServiceFactory
}

// New creates a new instance of the distribution handlers
func handlerNew(digCon *dig.Container) Handler {
	return ptr.Of(utils.MustGetObjFromDigCon[handler](digCon))
}

type factory struct{}

// Initialize initializes the distribution manifest handlers
func (f factory) Initialize(c echo.Context, digCon *dig.Container) error {
	method := c.Request().Method
	uri := c.Request().RequestURI
	handler := handlerNew(digCon)
	if method == http.MethodGet {
		switch {
		case uri == "/v2/":
			return handler.GetHealthy(c)
		case uri == "/v2/_catalog":
			return handler.ListRepositories(c)
		case strings.HasSuffix(uri, "/tags/list") && strings.HasPrefix(uri, "/v2/"):
			return handler.ListTags(c)
		}
	}
	return distribution.ErrNext
}

func init() {
	utils.PanicIf(distribution.RegisterRouterFactory(&factory{}, 1))
}
