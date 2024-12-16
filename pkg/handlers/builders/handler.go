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

package builders

import (
	"path"
	"reflect"

	"github.com/labstack/echo/v4"
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/dao"
	"github.com/go-sigma/sigma/pkg/handlers"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

// Handler is the interface for the builder handlers
type Handler interface {
	// CreateBuilder handles the create builder request
	CreateBuilder(c echo.Context) error
	// UpdateBuilder handles the update builder request
	UpdateBuilder(c echo.Context) error
	// ListRunners handles the list builder runners request
	ListRunners(c echo.Context) error
	// PostRunnerRun ...
	PostRunnerRun(c echo.Context) error
	// GetRunnerRerun ...
	GetRunnerRerun(c echo.Context) error
	// GetRunnerStop ...
	GetRunnerStop(c echo.Context) error
	// GetRunnerLog ...
	GetRunnerLog(c echo.Context) error
	// GetRunner ...
	GetRunner(c echo.Context) error
}

var _ Handler = &handler{}

type handler struct {
	dig.In

	NamespaceServiceFactory      dao.NamespaceServiceFactory
	RepositoryServiceFactory     dao.RepositoryServiceFactory
	WebhookServiceFactory        dao.WebhookServiceFactory
	AuditServiceFactory          dao.AuditServiceFactory
	BuilderServiceFactory        dao.BuilderServiceFactory
	UserServiceFactory           dao.UserServiceFactory
	CodeRepositoryServiceFactory dao.CodeRepositoryServiceFactory
}

// handlerNew creates a new instance of the builder handlers
func handlerNew(digCon *dig.Container) Handler {
	return ptr.Of(utils.MustGetObjFromDigCon[handler](digCon))
}

type factory struct{}

// Initialize initializes the namespace handlers
func (f factory) Initialize(digCon *dig.Container) error {
	e := utils.MustGetObjFromDigCon[*echo.Echo](digCon)
	handler := handlerNew(digCon)

	config := configs.GetConfiguration() // TODO: use dig
	if config.Daemon.Builder.Enabled {
		builderGroup := e.Group(consts.APIV1 + "/namespaces/:namespace_id/repositories/:repository_id/builders")
		builderGroup.POST("/", handler.CreateBuilder)
		builderGroup.PUT("/:builder_id", handler.UpdateBuilder)
		builderGroup.GET("/:builder_id/runners/", handler.ListRunners)
		builderGroup.POST("/:builder_id/runners/run", handler.PostRunnerRun)
		builderGroup.GET("/:builder_id/runners/:runner_id", handler.GetRunner)
		builderGroup.GET("/:builder_id/runners/:runner_id/stop", handler.GetRunnerStop)
		builderGroup.GET("/:builder_id/runners/:runner_id/rerun", handler.GetRunnerRerun)
		builderGroup.GET("/:builder_id/runners/:runner_id/log", handler.GetRunnerLog)
	}
	return nil
}

func init() {
	utils.PanicIf(handlers.RegisterRouterFactory(path.Base(reflect.TypeOf(factory{}).PkgPath()), &factory{}))
}
