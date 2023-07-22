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

package namespaces

import (
	"path"
	"reflect"

	"github.com/labstack/echo/v4"

	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/dao"
	rhandlers "github.com/go-sigma/sigma/pkg/handlers"
	"github.com/go-sigma/sigma/pkg/middlewares"
	"github.com/go-sigma/sigma/pkg/utils"
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
	authServiceFactory       dao.AuthServiceFactory
	namespaceServiceFactory  dao.NamespaceServiceFactory
	repositoryServiceFactory dao.RepositoryServiceFactory
	tagServiceFactory        dao.TagServiceFactory
	artifactServiceFactory   dao.ArtifactServiceFactory
	auditServiceFactory      dao.AuditServiceFactory
}

type inject struct {
	authServiceFactory       dao.AuthServiceFactory
	namespaceServiceFactory  dao.NamespaceServiceFactory
	repositoryServiceFactory dao.RepositoryServiceFactory
	tagServiceFactory        dao.TagServiceFactory
	artifactServiceFactory   dao.ArtifactServiceFactory
	auditServiceFactory      dao.AuditServiceFactory
}

// handlerNew creates a new instance of the distribution handlers
func handlerNew(injects ...inject) Handlers {
	authServiceFactory := dao.NewAuthServiceFactory()
	namespaceServiceFactory := dao.NewNamespaceServiceFactory()
	repositoryServiceFactory := dao.NewRepositoryServiceFactory()
	tagServiceFactory := dao.NewTagServiceFactory()
	artifactServiceFactory := dao.NewArtifactServiceFactory()
	auditServiceFactory := dao.NewAuditServiceFactory()
	if len(injects) > 0 {
		ij := injects[0]
		if ij.authServiceFactory != nil {
			authServiceFactory = ij.authServiceFactory
		}
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
		if ij.auditServiceFactory != nil {
			auditServiceFactory = ij.auditServiceFactory
		}
	}
	return &handlers{
		authServiceFactory:       authServiceFactory,
		namespaceServiceFactory:  namespaceServiceFactory,
		repositoryServiceFactory: repositoryServiceFactory,
		tagServiceFactory:        tagServiceFactory,
		artifactServiceFactory:   artifactServiceFactory,
		auditServiceFactory:      auditServiceFactory,
	}
}

type factory struct{}

// Initialize initializes the namespace handlers
func (f factory) Initialize(e *echo.Echo) error {
	namespaceGroup := e.Group(consts.APIV1+"/namespaces", middlewares.AuthWithConfig(middlewares.AuthConfig{}))
	namespaceHandler := handlerNew()
	namespaceGroup.POST("/", namespaceHandler.PostNamespace)
	namespaceGroup.PUT("/:id", namespaceHandler.PutNamespace)
	namespaceGroup.DELETE("/:id", namespaceHandler.DeleteNamespace)
	namespaceGroup.GET("/:id", namespaceHandler.GetNamespace)
	namespaceGroup.GET("/", namespaceHandler.ListNamespace)
	return nil
}

func init() {
	utils.PanicIf(rhandlers.RegisterRouterFactory(path.Base(reflect.TypeOf(factory{}).PkgPath()), &factory{}))
}
