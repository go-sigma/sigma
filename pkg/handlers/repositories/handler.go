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

package repositories

import (
	"path"
	"reflect"

	"github.com/labstack/echo/v4"

	"github.com/ximager/ximager/pkg/consts"
	"github.com/ximager/ximager/pkg/dal/dao"
	rhandlers "github.com/ximager/ximager/pkg/handlers"
	"github.com/ximager/ximager/pkg/middlewares"
	"github.com/ximager/ximager/pkg/utils"
)

// Handlers is the interface for the repository handlers
type Handlers interface {
	// PostRepository handles the post repository request
	PostRepository(c echo.Context) error
	// PutRepository handles the put repository request
	PutRepository(c echo.Context) error
	// GetRepository handles the get repository request
	GetRepository(c echo.Context) error
	// ListNamespace handles the list repository request
	ListRepository(c echo.Context) error
	// DeleteRepository handles the delete repository request
	DeleteRepository(c echo.Context) error
}

var _ Handlers = &handlers{}

type handlers struct {
	namespaceServiceFactory  dao.NamespaceServiceFactory
	repositoryServiceFactory dao.RepositoryServiceFactory
	tagServiceFactory        dao.TagServiceFactory
	artifactServiceFactory   dao.ArtifactServiceFactory
}

type inject struct {
	namespaceServiceFactory  dao.NamespaceServiceFactory
	repositoryServiceFactory dao.RepositoryServiceFactory
	tagServiceFactory        dao.TagServiceFactory
	artifactServiceFactory   dao.ArtifactServiceFactory
}

// handlerNew creates a new instance of the distribution handlers
// nolint: unparam
func handlerNew(injects ...inject) Handlers {
	namespaceServiceFactory := dao.NewNamespaceServiceFactory()
	repositoryServiceFactory := dao.NewRepositoryServiceFactory()
	tagServiceFactory := dao.NewTagServiceFactory()
	artifactServiceFactory := dao.NewArtifactServiceFactory()
	if len(injects) > 0 {
		ij := injects[0]
		if ij.namespaceServiceFactory != nil {
			namespaceServiceFactory = ij.namespaceServiceFactory
		}
		if ij.repositoryServiceFactory != nil {
			repositoryServiceFactory = ij.repositoryServiceFactory
		}
		if ij.tagServiceFactory != nil {
			tagServiceFactory = ij.tagServiceFactory
		}
		if ij.artifactServiceFactory != nil {
			artifactServiceFactory = ij.artifactServiceFactory
		}
	}
	return &handlers{
		namespaceServiceFactory:  namespaceServiceFactory,
		repositoryServiceFactory: repositoryServiceFactory,
		tagServiceFactory:        tagServiceFactory,
		artifactServiceFactory:   artifactServiceFactory,
	}
}

type factory struct{}

// Initialize initializes the namespace handlers
func (f factory) Initialize(e *echo.Echo) error {
	repositoryGroup := e.Group(consts.APIV1+"/namespaces/:namespace/repositories", middlewares.AuthWithConfig(middlewares.AuthConfig{}))
	repositoryHandler := handlerNew()
	repositoryGroup.GET("/", repositoryHandler.ListRepository)
	repositoryGroup.GET("/:id", repositoryHandler.GetRepository)
	repositoryGroup.DELETE("/:id", repositoryHandler.DeleteRepository)
	return nil
}

func init() {
	utils.PanicIf(rhandlers.RegisterRouterFactory(path.Base(reflect.TypeOf(factory{}).PkgPath()), &factory{}))
}
