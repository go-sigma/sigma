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

package artifact

import (
	"path"
	"reflect"

	"github.com/labstack/echo/v4"

	"github.com/ximager/ximager/pkg/dal/dao"
	rhandlers "github.com/ximager/ximager/pkg/handlers"
	"github.com/ximager/ximager/pkg/middlewares"
	"github.com/ximager/ximager/pkg/utils"
)

// Handlers is the interface for the artifact handlers
type Handlers interface {
	// ListArtifact handles the list artifact request
	ListArtifact(c echo.Context) error
	// GetArtifact handles the get artifact request
	GetArtifact(c echo.Context) error
	// DeleteArtifact handles the delete artifact request
	DeleteArtifact(c echo.Context) error
}

var _ Handlers = &handlers{}

type handlers struct {
	namespaceServiceFactory dao.NamespaceServiceFactory
	artifactServiceFactory  dao.ArtifactServiceFactory
	tagServiceFactory       dao.TagServiceFactory
}

type inject struct {
	namespaceServiceFactory dao.NamespaceServiceFactory
	artifactServiceFactory  dao.ArtifactServiceFactory
	tagServiceFactory       dao.TagServiceFactory
}

// handlerNew creates a new instance of the distribution handlers
func handlerNew(injects ...inject) Handlers {
	namespaceServiceFactory := dao.NewNamespaceServiceFactory()
	tagServiceFactory := dao.NewTagServiceFactory()
	artifactServiceFactory := dao.NewArtifactServiceFactory()
	if len(injects) > 0 {
		ij := injects[0]
		if ij.namespaceServiceFactory != nil {
			namespaceServiceFactory = ij.namespaceServiceFactory
		}
		if ij.tagServiceFactory != nil {
			tagServiceFactory = ij.tagServiceFactory
		}
		if ij.artifactServiceFactory != nil {
			artifactServiceFactory = ij.artifactServiceFactory
		}
	}
	return &handlers{
		namespaceServiceFactory: namespaceServiceFactory,
		tagServiceFactory:       tagServiceFactory,
		artifactServiceFactory:  artifactServiceFactory,
	}
}

type factory struct{}

// Initialize initializes the namespace handlers
func (f factory) Initialize(e *echo.Echo) error {
	artifactGroup := e.Group("/namespace/:namespace/artifact", middlewares.AuthWithConfig(middlewares.AuthConfig{}))
	artifactHandler := handlerNew()
	artifactGroup.GET("/", artifactHandler.ListArtifact)
	artifactGroup.GET("/:id", artifactHandler.GetArtifact)
	artifactGroup.DELETE("/:id", artifactHandler.DeleteArtifact)
	return nil
}

func init() {
	utils.PanicIf(rhandlers.RegisterRouterFactory(path.Base(reflect.TypeOf(factory{}).PkgPath()), &factory{}))
}
