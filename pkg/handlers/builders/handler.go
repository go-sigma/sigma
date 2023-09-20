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

	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/dao"
	rhandlers "github.com/go-sigma/sigma/pkg/handlers"
	"github.com/go-sigma/sigma/pkg/middlewares"
	"github.com/go-sigma/sigma/pkg/utils"
)

// Handler is the interface for the builder handlers
type Handlers interface {
	// PostBuilder handles the post builder request
	PostBuilder(c echo.Context) error
	// GetBuilder handles the get builder request
	GetBuilder(c echo.Context) error
	// PutBuilder handles the put builder request
	PutBuilder(c echo.Context) error
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
}

var _ Handlers = &handlers{}

type handlers struct {
	namespaceServiceFactory  dao.NamespaceServiceFactory
	repositoryServiceFactory dao.RepositoryServiceFactory
	webhookServiceFactory    dao.WebhookServiceFactory
	auditServiceFactory      dao.AuditServiceFactory
	builderServiceFactory    dao.BuilderServiceFactory
}

type inject struct {
	namespaceServiceFactory  dao.NamespaceServiceFactory
	repositoryServiceFactory dao.RepositoryServiceFactory
	webhookServiceFactory    dao.WebhookServiceFactory
	auditServiceFactory      dao.AuditServiceFactory
	builderServiceFactory    dao.BuilderServiceFactory
}

// handlerNew creates a new instance of the builder handlers
func handlerNew(injects ...inject) Handlers {
	namespaceServiceFactory := dao.NewNamespaceServiceFactory()
	repositoryServiceFactory := dao.NewRepositoryServiceFactory()
	webhookServiceFactory := dao.NewWebhookServiceFactory()
	auditServiceFactory := dao.NewAuditServiceFactory()
	builderServiceFactory := dao.NewBuilderServiceFactory()
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
	}
	return &handlers{
		namespaceServiceFactory:  namespaceServiceFactory,
		repositoryServiceFactory: repositoryServiceFactory,
		webhookServiceFactory:    webhookServiceFactory,
		auditServiceFactory:      auditServiceFactory,
		builderServiceFactory:    builderServiceFactory,
	}
}

type factory struct{}

// Initialize initializes the namespace handlers
func (f factory) Initialize(e *echo.Echo) error {
	builderGroup := e.Group(consts.APIV1+"/repositories/:repository_id/builders", middlewares.AuthWithConfig(middlewares.AuthConfig{}))

	handler := handlerNew()
	builderGroup.POST("/", handler.PostBuilder)
	builderGroup.GET("/:id", handler.GetBuilder)
	builderGroup.PUT("/:id", handler.PutBuilder)

	builderGroup.GET("/:id/runners/", handler.ListRunners)
	builderGroup.GET("/:id/runners/run", handler.PostRunnerRun)
	builderGroup.GET("/:id/runners/:runner_id/log", handler.GetRunnerLog)
	builderGroup.GET("/:id/runners/:runner_id/stop", handler.GetRunnerStop)
	builderGroup.GET("/:id/runners/:runner_id/rerun", handler.GetRunnerStop)
	return nil
}

func init() {
	utils.PanicIf(rhandlers.RegisterRouterFactory(path.Base(reflect.TypeOf(factory{}).PkgPath()), &factory{}))
}
