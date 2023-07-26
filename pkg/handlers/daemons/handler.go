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

package daemons

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
	// Run run the specific daemon task
	Run(c echo.Context) error
	// Status get the specific daemon task status
	Status(c echo.Context) error
	// Logs get the specific daemon task logs
	Logs(c echo.Context) error
}

var _ Handlers = &handlers{}

type handlers struct {
	daemonServiceFactory dao.DaemonServiceFactory
}

type inject struct {
	daemonServiceFactory dao.DaemonServiceFactory
}

// handlerNew creates a new instance of the distribution handlers
func handlerNew(injects ...inject) Handlers {
	daemonServiceFactory := dao.NewDaemonServiceFactory()
	if len(injects) > 0 {
		ij := injects[0]
		if ij.daemonServiceFactory != nil {
			daemonServiceFactory = ij.daemonServiceFactory
		}
	}
	return &handlers{
		daemonServiceFactory: daemonServiceFactory,
	}
}

type factory struct{}

// Initialize initializes the namespace handlers
func (f factory) Initialize(e *echo.Echo) error {
	daemonGroup := e.Group(consts.APIV1+"/daemons", middlewares.AuthWithConfig(middlewares.AuthConfig{}))

	repositoryHandler := handlerNew()
	daemonGroup.POST("/:name/", repositoryHandler.Run)
	daemonGroup.GET("/:name/", repositoryHandler.Run)
	daemonGroup.GET("/:name/logs", repositoryHandler.Logs)
	return nil
}

func init() {
	utils.PanicIf(rhandlers.RegisterRouterFactory(path.Base(reflect.TypeOf(factory{}).PkgPath()), &factory{}))
}
