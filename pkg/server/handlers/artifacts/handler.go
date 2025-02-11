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

package artifact

import (
	"path"
	"reflect"

	"github.com/labstack/echo/v4"
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/dao"
	"github.com/go-sigma/sigma/pkg/server/handlers"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

// Handler is the interface for the artifact handlers
type Handler interface {
	// ListArtifact handles the list artifact request
	ListArtifact(c echo.Context) error
	// GetArtifact handles the get artifact request
	GetArtifact(c echo.Context) error
	// DeleteArtifact handles the delete artifact request
	DeleteArtifact(c echo.Context) error
}

var _ Handler = &handler{}

type handler struct {
	dig.In

	NamespaceServiceFactory  dao.NamespaceServiceFactory
	RepositoryServiceFactory dao.RepositoryServiceFactory
	ArtifactServiceFactory   dao.ArtifactServiceFactory
	TagServiceFactory        dao.TagServiceFactory
}

// handlerNew creates a new instance of the distribution handlers
func handlerNew(digCon *dig.Container) Handler {
	return ptr.Of(utils.MustGetObjFromDigCon[handler](digCon))
}

type factory struct{}

// Initialize initializes the namespace handlers
func (f factory) Initialize(digCon *dig.Container) error {
	e := utils.MustGetObjFromDigCon[*echo.Echo](digCon)
	artifactHandler := handlerNew(digCon)
	artifactGroup := e.Group(consts.APIV1 + "/namespaces/:namespace/artifacts")
	artifactGroup.GET("/", artifactHandler.ListArtifact)
	artifactGroup.GET("/:digest", artifactHandler.GetArtifact)
	artifactGroup.DELETE("/:digest", artifactHandler.DeleteArtifact)
	return nil
}

func init() {
	utils.PanicIf(handlers.RegisterRouterFactory(path.Base(reflect.TypeOf(factory{}).PkgPath()), &factory{}))
}
