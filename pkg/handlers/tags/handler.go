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

package tag

import (
	"path"
	"reflect"

	"github.com/labstack/echo/v4"
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/auth"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/dao"
	"github.com/go-sigma/sigma/pkg/handlers"
	"github.com/go-sigma/sigma/pkg/middlewares"
	"github.com/go-sigma/sigma/pkg/utils"
)

// Handler is the interface for the tag handlers
type Handler interface {
	// ListTag handles the list tag request
	ListTag(c echo.Context) error
	// GetTag handles the get tag request
	GetTag(c echo.Context) error
	// DeleteTag handles the delete tag request
	DeleteTag(c echo.Context) error
}

var _ Handler = &handler{}

type handler struct {
	authServiceFactory       auth.AuthServiceFactory
	namespaceServiceFactory  dao.NamespaceServiceFactory
	repositoryServiceFactory dao.RepositoryServiceFactory
	tagServiceFactory        dao.TagServiceFactory
	artifactServiceFactory   dao.ArtifactServiceFactory
}

// handlerNew creates a new instance of the distribution handlers
func handlerNew(digCon *dig.Container) Handler {
	return &handler{
		authServiceFactory:       utils.MustGetObjFromDigCon[auth.AuthServiceFactory](digCon),
		namespaceServiceFactory:  utils.MustGetObjFromDigCon[dao.NamespaceServiceFactory](digCon),
		repositoryServiceFactory: utils.MustGetObjFromDigCon[dao.RepositoryServiceFactory](digCon),
		tagServiceFactory:        utils.MustGetObjFromDigCon[dao.TagServiceFactory](digCon),
		artifactServiceFactory:   utils.MustGetObjFromDigCon[dao.ArtifactServiceFactory](digCon),
	}
}

type factory struct{}

func (f factory) Initialize(digCon *dig.Container) error {
	e := utils.MustGetObjFromDigCon[*echo.Echo](digCon)
	tagGroup := e.Group(consts.APIV1+"/namespaces/:namespace_id/repositories/:repository_id/tags", middlewares.AuthWithConfig(middlewares.AuthConfig{}))

	tagHandler := handlerNew(digCon)

	tagGroup.GET("/", tagHandler.ListTag)
	tagGroup.GET("/:id", tagHandler.GetTag)
	tagGroup.DELETE("/:id", tagHandler.DeleteTag)

	return nil
}

func init() {
	utils.PanicIf(handlers.RegisterRouterFactory(path.Base(reflect.TypeOf(factory{}).PkgPath()), &factory{}))
}
