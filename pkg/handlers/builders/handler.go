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

	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/dao"
	"github.com/go-sigma/sigma/pkg/handlers"
	"github.com/go-sigma/sigma/pkg/middlewares"
	"github.com/go-sigma/sigma/pkg/utils"
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
	namespaceServiceFactory      dao.NamespaceServiceFactory
	repositoryServiceFactory     dao.RepositoryServiceFactory
	webhookServiceFactory        dao.WebhookServiceFactory
	auditServiceFactory          dao.AuditServiceFactory
	builderServiceFactory        dao.BuilderServiceFactory
	userServiceFactory           dao.UserServiceFactory
	codeRepositoryServiceFactory dao.CodeRepositoryServiceFactory
}

type inject struct {
	namespaceServiceFactory      dao.NamespaceServiceFactory
	repositoryServiceFactory     dao.RepositoryServiceFactory
	webhookServiceFactory        dao.WebhookServiceFactory
	auditServiceFactory          dao.AuditServiceFactory
	builderServiceFactory        dao.BuilderServiceFactory
	userServiceFactory           dao.UserServiceFactory
	codeRepositoryServiceFactory dao.CodeRepositoryServiceFactory
}

// handlerNew creates a new instance of the builder handlers
func handlerNew(injects ...inject) Handler {
	namespaceServiceFactory := dao.NewNamespaceServiceFactory()
	repositoryServiceFactory := dao.NewRepositoryServiceFactory()
	webhookServiceFactory := dao.NewWebhookServiceFactory()
	auditServiceFactory := dao.NewAuditServiceFactory()
	builderServiceFactory := dao.NewBuilderServiceFactory()
	userServiceFactory := dao.NewUserServiceFactory()
	codeRepositoryServiceFactory := dao.NewCodeRepositoryServiceFactory()
	if len(injects) > 0 {
		ij := injects[0]
		if ij.namespaceServiceFactory != nil {
			namespaceServiceFactory = ij.namespaceServiceFactory
		}
		if ij.webhookServiceFactory != nil {
			webhookServiceFactory = ij.webhookServiceFactory
		}
		if ij.auditServiceFactory != nil {
			auditServiceFactory = ij.auditServiceFactory
		}
		if ij.builderServiceFactory != nil {
			builderServiceFactory = ij.builderServiceFactory
		}
		if ij.repositoryServiceFactory != nil {
			repositoryServiceFactory = ij.repositoryServiceFactory
		}
		if ij.userServiceFactory != nil {
			userServiceFactory = ij.userServiceFactory
		}
		if ij.codeRepositoryServiceFactory != nil {
			codeRepositoryServiceFactory = ij.codeRepositoryServiceFactory
		}
	}
	return &handler{
		namespaceServiceFactory:      namespaceServiceFactory,
		repositoryServiceFactory:     repositoryServiceFactory,
		webhookServiceFactory:        webhookServiceFactory,
		auditServiceFactory:          auditServiceFactory,
		builderServiceFactory:        builderServiceFactory,
		userServiceFactory:           userServiceFactory,
		codeRepositoryServiceFactory: codeRepositoryServiceFactory,
	}
}

type factory struct{}

// Initialize initializes the namespace handlers
func (f factory) Initialize(e *echo.Echo) error {
	handler := handlerNew()

	config := configs.GetConfiguration()
	if config.Daemon.Builder.Enabled {
		builderGroup := e.Group(consts.APIV1+"/namespaces/:namespace_id/repositories/:repository_id/builders",
			middlewares.AuthWithConfig(middlewares.AuthConfig{}))
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
