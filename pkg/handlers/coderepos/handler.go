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

package coderepos

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

// Handler is the interface for the system handlers
type Handlers interface {
	// List list all of the code repositories
	List(c echo.Context) error
	// ListOwner list all of the code repository owner
	ListOwners(c echo.Context) error
	// Resync resync all of the code repositories
	Resync(c echo.Context) error
	// Setup setup builder for code repository
	Setup(c echo.Context) error
}

var _ Handlers = &handlers{}

type handlers struct {
	codeRepositoryServiceFactory dao.CodeRepositoryServiceFactory
	userServiceFactory           dao.UserServiceFactory
}

type inject struct {
	codeRepositoryServiceFactory dao.CodeRepositoryServiceFactory
	userServiceFactory           dao.UserServiceFactory
}

// handlerNew creates a new instance of the distribution handlers
func handlerNew(injects ...inject) Handlers {
	codeRepositoryServiceFactory := dao.NewCodeRepositoryServiceFactory()
	userServiceFactory := dao.NewUserServiceFactory()
	if len(injects) > 0 {
		ij := injects[0]
		if ij.codeRepositoryServiceFactory != nil {
			codeRepositoryServiceFactory = ij.codeRepositoryServiceFactory
		}
		if ij.userServiceFactory != nil {
			userServiceFactory = ij.userServiceFactory
		}
	}
	return &handlers{
		codeRepositoryServiceFactory: codeRepositoryServiceFactory,
		userServiceFactory:           userServiceFactory,
	}
}

type factory struct{}

// Initialize initializes the namespace handlers
func (f factory) Initialize(e *echo.Echo) error {
	codereposGroup := e.Group(consts.APIV1+"/coderepos", middlewares.AuthWithConfig(middlewares.AuthConfig{}))

	codeRepositoryHandler := handlerNew()
	codereposGroup.GET("/", codeRepositoryHandler.List)
	codereposGroup.GET("/owners", codeRepositoryHandler.ListOwners)
	codereposGroup.POST("/resync", codeRepositoryHandler.Resync)
	codereposGroup.POST("/setup", codeRepositoryHandler.Setup)
	return nil
}

func init() {
	utils.PanicIf(rhandlers.RegisterRouterFactory(path.Base(reflect.TypeOf(factory{}).PkgPath()), &factory{}))
}
