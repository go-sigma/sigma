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

package repositories

import (
	"path"
	"reflect"

	"github.com/labstack/echo/v4"
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/auth"
	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/dao"
	"github.com/go-sigma/sigma/pkg/handlers"
	"github.com/go-sigma/sigma/pkg/middlewares"
	"github.com/go-sigma/sigma/pkg/utils"
)

// Handler is the interface for the repository handlers
type Handler interface {
	// CreateRepository handles the post repository request
	CreateRepository(c echo.Context) error
	// UpdateRepository handles the put repository request
	UpdateRepository(c echo.Context) error
	// GetRepository handles the get repository request
	GetRepository(c echo.Context) error
	// ListRepositories handles the list repository request
	ListRepositories(c echo.Context) error
	// DeleteRepository handles the delete repository request
	DeleteRepository(c echo.Context) error
}

var _ Handler = &handler{}

type handler struct {
	config                   configs.Configuration
	authServiceFactory       auth.AuthServiceFactory
	auditServiceFactory      dao.AuditServiceFactory
	namespaceServiceFactory  dao.NamespaceServiceFactory
	repositoryServiceFactory dao.RepositoryServiceFactory
	tagServiceFactory        dao.TagServiceFactory
	artifactServiceFactory   dao.ArtifactServiceFactory
	builderServiceFactory    dao.BuilderServiceFactory
}

// handlerNew creates a new instance of the distribution handlers
func handlerNew(digCon *dig.Container) Handler {
	return &handler{
		config:                   utils.MustGetObjFromDigCon[configs.Configuration](digCon),
		authServiceFactory:       utils.MustGetObjFromDigCon[auth.AuthServiceFactory](digCon),
		auditServiceFactory:      utils.MustGetObjFromDigCon[dao.AuditServiceFactory](digCon),
		namespaceServiceFactory:  utils.MustGetObjFromDigCon[dao.NamespaceServiceFactory](digCon),
		repositoryServiceFactory: utils.MustGetObjFromDigCon[dao.RepositoryServiceFactory](digCon),
		tagServiceFactory:        utils.MustGetObjFromDigCon[dao.TagServiceFactory](digCon),
		artifactServiceFactory:   utils.MustGetObjFromDigCon[dao.ArtifactServiceFactory](digCon),
		builderServiceFactory:    utils.MustGetObjFromDigCon[dao.BuilderServiceFactory](digCon),
	}
}

type factory struct{}

// Initialize initializes the namespace handlers
func (f factory) Initialize(digCon *dig.Container) error {
	e := utils.MustGetObjFromDigCon[*echo.Echo](digCon)
	repositoryGroup := e.Group(consts.APIV1+"/namespaces/:namespace_id/repositories", middlewares.AuthWithConfig(middlewares.AuthConfig{}))

	repositoryHandler := handlerNew(digCon)

	repositoryGroup.GET("/", repositoryHandler.ListRepositories)
	repositoryGroup.POST("/", repositoryHandler.CreateRepository)
	repositoryGroup.GET("/:id", repositoryHandler.GetRepository)
	repositoryGroup.PUT("/:id", repositoryHandler.UpdateRepository)
	repositoryGroup.DELETE("/:id", repositoryHandler.DeleteRepository)

	return nil
}

func init() {
	utils.PanicIf(handlers.RegisterRouterFactory(path.Base(reflect.TypeOf(factory{}).PkgPath()), &factory{}))
}
