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

package namespace

import (
	"path"
	"reflect"

	"github.com/labstack/echo/v4"

	"github.com/ximager/ximager/pkg/dal/dao"
	rhandlers "github.com/ximager/ximager/pkg/handlers"
	"github.com/ximager/ximager/pkg/middlewares"
	"github.com/ximager/ximager/pkg/utils"
)

// Handlers is the interface for the namespace handlers
type Handlers interface {
	// PostNamespace handles the post namespace request
	PostNamespace(c echo.Context) error
	// ListNamespace handles the list namespace request
	ListNamespace(c echo.Context) error
	// GetNamespace handles the get namespace request
	GetNamespace(c echo.Context) error
	// DeleteNamespace handles the delete namespace request
	DeleteNamespace(c echo.Context) error
	// PutNamespace handles the put namespace request
	PutNamespace(c echo.Context) error
}

var _ Handlers = &handlers{}

type handlers struct {
	namespaceServiceFactory dao.NamespaceServiceFactory
	artifactServiceFactory  dao.ArtifactServiceFactory
}

type inject struct {
	namespaceServiceFactory dao.NamespaceServiceFactory
	artifactServiceFactory  dao.ArtifactServiceFactory
}

// handlerNew creates a new instance of the distribution handlers
func handlerNew(injects ...inject) Handlers {
	namespaceServiceFactory := dao.NewNamespaceServiceFactory()
	artifactServiceFactory := dao.NewArtifactServiceFactory()
	if len(injects) > 0 {
		ij := injects[0]
		if ij.namespaceServiceFactory != nil {
			namespaceServiceFactory = ij.namespaceServiceFactory
		}
		if ij.artifactServiceFactory != nil {
			artifactServiceFactory = ij.artifactServiceFactory
		}
	}
	return &handlers{
		namespaceServiceFactory: namespaceServiceFactory,
		artifactServiceFactory:  artifactServiceFactory,
	}
}

type factory struct{}

// Initialize initializes the namespace handlers
func (f factory) Initialize(e *echo.Echo) error {
	namespaceGroup := e.Group("/namespace", middlewares.AuthWithConfig(middlewares.AuthConfig{}))
	namespaceHandler := handlerNew()
	namespaceGroup.POST("/", namespaceHandler.PostNamespace)
	namespaceGroup.PUT("/:id", namespaceHandler.PutNamespace)
	namespaceGroup.DELETE("/:id", namespaceHandler.DeleteNamespace)
	namespaceGroup.GET("/:id", namespaceHandler.GetNamespace)
	namespaceGroup.GET("/", namespaceHandler.ListNamespace)
	return nil
}

func init() {
	utils.PanicIf(rhandlers.RegisterRouterFactory(path.Base(reflect.TypeOf(handlers{}).PkgPath()), &factory{}))
}
