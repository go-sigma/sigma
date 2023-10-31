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

	// GcRepositoryRun ...
	GcRepositoryRun(c echo.Context) error
	// GcRepositoryRunners ...
	GcRepositoryRunners(c echo.Context) error
	// GcRepositoryGet ...
	GcRepositoryGet(c echo.Context) error
	// GcRepositoryRecords ...
	GcRepositoryRecords(c echo.Context) error
	// GcRepositoryRecord ...
	GcRepositoryRecord(c echo.Context) error

	// GcArtifactRun ...
	GcArtifactRun(c echo.Context) error
	// GcArtifactGet ...
	GcArtifactGet(c echo.Context) error
	// GcArtifactRecords ...
	GcArtifactRecords(c echo.Context) error
	// GcArtifactRecord ...
	GcArtifactRecord(c echo.Context) error

	// GcBlobRun ...
	GcBlobRun(c echo.Context) error
	// GcBlobGet ...
	GcBlobGet(c echo.Context) error
	// GcBlobRecords ...
	GcBlobRecords(c echo.Context) error
	// GcBlobRecord ...
	GcBlobRecord(c echo.Context) error
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

	daemonGroup.POST("/gc-repository/", repositoryHandler.GcRepositoryRun)
	daemonGroup.GET("/gc-repository/", repositoryHandler.GcRepositoryRunners)
	daemonGroup.GET("/gc-repository/detail", repositoryHandler.GcRepositoryGet)
	daemonGroup.GET("/gc-repository/:runner_id/", repositoryHandler.GcRepositoryRecords)
	daemonGroup.GET("/gc-repository/:runner_id/records/:record_id", repositoryHandler.GcRepositoryRecord)

	daemonGroup.POST("/gc-artifact/", repositoryHandler.GcArtifactRun)
	daemonGroup.GET("/gc-artifact/", repositoryHandler.GcArtifactRecords)
	daemonGroup.GET("/gc-artifact/detail", repositoryHandler.GcArtifactGet)
	daemonGroup.GET("/gc-artifact/:id", repositoryHandler.GcArtifactRecord)

	daemonGroup.POST("/gc-blob/", repositoryHandler.GcBlobRun)
	daemonGroup.GET("/gc-blob/", repositoryHandler.GcBlobRecords)
	daemonGroup.GET("/gc-blob/detail", repositoryHandler.GcBlobGet)
	daemonGroup.GET("/gc-blob/:id", repositoryHandler.GcBlobRecord)

	return nil
}

func init() {
	utils.PanicIf(rhandlers.RegisterRouterFactory(path.Base(reflect.TypeOf(factory{}).PkgPath()), &factory{}))
}
