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

// Handlers is the interface for the gc handlers
type Handlers interface {
	// UpdateGcTagRule ...
	UpdateGcTagRule(c echo.Context) error
	// GetGcTagRule ...
	GetGcTagRule(c echo.Context) error
	// GetGcTagLatestRunner ...
	GetGcTagLatestRunner(c echo.Context) error
	// CreateGcTagRunner ...
	CreateGcTagRunner(c echo.Context) error
	// ListGcTagRunners ...
	ListGcTagRunners(c echo.Context) error
	// GetGcTagRunner ...
	GetGcTagRunner(c echo.Context) error
	// ListGcTagRecords ...
	ListGcTagRecords(c echo.Context) error
	// GetGcTagRecord ...
	GetGcTagRecord(c echo.Context) error

	// UpdateGcRepositoryRule ...
	UpdateGcRepositoryRule(c echo.Context) error
	// GetGcRepositoryRule ...
	GetGcRepositoryRule(c echo.Context) error
	// GetGcRepositoryLatestRunner ...
	GetGcRepositoryLatestRunner(c echo.Context) error
	// CreateGcRepositoryRunner ...
	CreateGcRepositoryRunner(c echo.Context) error
	// ListGcRepositoryRunners ...
	ListGcRepositoryRunners(c echo.Context) error
	// GetGcRepositoryRunner ...
	GetGcRepositoryRunner(c echo.Context) error
	// ListGcRepositoryRecords ...
	ListGcRepositoryRecords(c echo.Context) error
	// GetGcRepositoryRecord ...
	GetGcRepositoryRecord(c echo.Context) error

	// UpdateGcArtifactRule ...
	UpdateGcArtifactRule(c echo.Context) error
	// GetGcArtifactRule ...
	GetGcArtifactRule(c echo.Context) error
	// GetGcArtifactLatestRunner ...
	GetGcArtifactLatestRunner(c echo.Context) error
	// CreateGcArtifactRunner ...
	CreateGcArtifactRunner(c echo.Context) error
	// ListGcArtifactRunners ...
	ListGcArtifactRunners(c echo.Context) error
	// GetGcArtifactRunner ...
	GetGcArtifactRunner(c echo.Context) error
	// ListGcArtifactRecords ...
	ListGcArtifactRecords(c echo.Context) error
	// GetGcArtifactRecord ...
	GetGcArtifactRecord(c echo.Context) error

	// UpdateGcBlobRule ...
	UpdateGcBlobRule(c echo.Context) error
	// GetGcBlobRule ...
	GetGcBlobRule(c echo.Context) error
	// GetGcBlobLatestRunner ...
	GetGcBlobLatestRunner(c echo.Context) error
	// CreateGcBlobRunner ...
	CreateGcBlobRunner(c echo.Context) error
	// ListGcBlobRunners ...
	ListGcBlobRunners(c echo.Context) error
	// GetGcBlobRunner ...
	GetGcBlobRunner(c echo.Context) error
	// ListGcBlobRecords ...
	ListGcBlobRecords(c echo.Context) error
	// GetGcBlobRecord ...
	GetGcBlobRecord(c echo.Context) error
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

	daemonHandler := handlerNew()

	daemonGroup.PUT("/gc-repository/:namespace_id/", daemonHandler.UpdateGcRepositoryRule)
	daemonGroup.GET("/gc-repository/:namespace_id/", daemonHandler.GetGcRepositoryRule)
	daemonGroup.GET("/gc-repository/:namespace_id/runners/latest", daemonHandler.GetGcRepositoryLatestRunner)
	daemonGroup.POST("/gc-repository/:namespace_id/runners/", daemonHandler.CreateGcRepositoryRunner)
	daemonGroup.GET("/gc-repository/:namespace_id/runners/", daemonHandler.ListGcRepositoryRunners)
	daemonGroup.GET("/gc-repository/:namespace_id/runners/:runner_id", daemonHandler.GetGcRepositoryRunner)
	daemonGroup.GET("/gc-repository/:namespace_id/runners/:runner_id/records/", daemonHandler.ListGcRepositoryRecords)
	daemonGroup.GET("/gc-repository/:namespace_id/runners/:runner_id/records/:record_id", daemonHandler.GetGcRepositoryRecord)

	daemonGroup.PUT("/gc-tag/:namespace_id/", daemonHandler.UpdateGcTagRule)
	daemonGroup.GET("/gc-tag/:namespace_id/", daemonHandler.GetGcTagRule)
	daemonGroup.GET("/gc-tag/:namespace_id/runners/latest", daemonHandler.GetGcTagLatestRunner)
	daemonGroup.POST("/gc-tag/:namespace_id/runners/", daemonHandler.CreateGcTagRunner)
	daemonGroup.GET("/gc-tag/:namespace_id/runners/", daemonHandler.ListGcTagRunners)
	daemonGroup.GET("/gc-tag/:namespace_id/runners/:runner_id", daemonHandler.GetGcTagRunner)
	daemonGroup.GET("/gc-tag/:namespace_id/runners/:runner_id/records/", daemonHandler.ListGcTagRecords)
	daemonGroup.GET("/gc-tag/:namespace_id/runners/:runner_id/records/:record_id", daemonHandler.GetGcTagRecord)

	daemonGroup.PUT("/gc-artifact/:namespace_id/", daemonHandler.UpdateGcArtifactRule)
	daemonGroup.GET("/gc-artifact/:namespace_id/", daemonHandler.GetGcArtifactRule)
	daemonGroup.GET("/gc-artifact/:namespace_id/runners/latest", daemonHandler.GetGcArtifactLatestRunner)
	daemonGroup.POST("/gc-artifact/:namespace_id/runners/", daemonHandler.CreateGcArtifactRunner)
	daemonGroup.GET("/gc-artifact/:namespace_id/runners/", daemonHandler.ListGcArtifactRunners)
	daemonGroup.GET("/gc-artifact/:namespace_id/runners/:runner_id", daemonHandler.GetGcArtifactRunner)
	daemonGroup.GET("/gc-artifact/:namespace_id/runners/:runner_id/records/", daemonHandler.ListGcArtifactRecords)
	daemonGroup.GET("/gc-artifact/:namespace_id/runners/:runner_id/records/:record_id", daemonHandler.GetGcArtifactRecord)

	daemonGroup.PUT("/gc-blob/:namespace_id/", daemonHandler.UpdateGcBlobRule)
	daemonGroup.GET("/gc-blob/:namespace_id/", daemonHandler.GetGcBlobRule)
	daemonGroup.GET("/gc-blob/:namespace_id/runners/latest", daemonHandler.GetGcBlobLatestRunner)
	daemonGroup.POST("/gc-blob/:namespace_id/runners/", daemonHandler.CreateGcBlobRunner)
	daemonGroup.GET("/gc-blob/:namespace_id/runners/", daemonHandler.ListGcBlobRunners)
	daemonGroup.GET("/gc-blob/:namespace_id/runners/:runner_id", daemonHandler.GetGcBlobRunner)
	daemonGroup.GET("/gc-blob/:namespace_id/runners/:runner_id/records/", daemonHandler.ListGcBlobRecords)
	daemonGroup.GET("/gc-blob/:namespace_id/runners/:runner_id/records/:record_id", daemonHandler.GetGcBlobRecord)

	return nil
}

func init() {
	utils.PanicIf(rhandlers.RegisterRouterFactory(path.Base(reflect.TypeOf(factory{}).PkgPath()), &factory{}))
}
