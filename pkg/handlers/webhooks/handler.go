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

package webhooks

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
	// PostWebhook handles the post webhook request
	PostWebhook(c echo.Context) error
	// ListWebhook handles the list webhook request
	ListWebhook(c echo.Context) error
	// GetWebhook handles the get webhook request
	GetWebhook(c echo.Context) error
	// DeleteWebhook handles the delete webhook request
	DeleteWebhook(c echo.Context) error
	// PutWebhook handles the put webhook request
	PutWebhook(c echo.Context) error
	// PingWebhook ...
	PingWebhook(c echo.Context) error
	// LogWebhook ...
	LogWebhook(c echo.Context) error
	// LogsWebhook ...
	LogsWebhook(c echo.Context) error
	// Resend ...
	Resend(c echo.Context) error
}

var _ Handlers = &handlers{}

type handlers struct {
	namespaceServiceFactory dao.NamespaceServiceFactory
	webhookServiceFactory   dao.WebhookServiceFactory
	auditServiceFactory     dao.AuditServiceFactory
}

type inject struct {
	namespaceServiceFactory dao.NamespaceServiceFactory
	webhookServiceFactory   dao.WebhookServiceFactory
	auditServiceFactory     dao.AuditServiceFactory
}

// handlerNew creates a new instance of the distribution handlers
func handlerNew(injects ...inject) Handlers {
	namespaceServiceFactory := dao.NewNamespaceServiceFactory()
	webhookServiceFactory := dao.NewWebhookServiceFactory()
	auditServiceFactory := dao.NewAuditServiceFactory()
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
	}
	return &handlers{
		namespaceServiceFactory: namespaceServiceFactory,
		webhookServiceFactory:   webhookServiceFactory,
		auditServiceFactory:     auditServiceFactory,
	}
}

type factory struct{}

// Initialize initializes the namespace handlers
func (f factory) Initialize(e *echo.Echo) error {
	webhookGroup := e.Group(consts.APIV1+"/webhooks", middlewares.AuthWithConfig(middlewares.AuthConfig{}))

	webhookHandler := handlerNew()
	webhookGroup.POST("/", webhookHandler.PostWebhook)
	webhookGroup.PUT("/:id", webhookHandler.PutWebhook)
	webhookGroup.GET("/", webhookHandler.ListWebhook)
	webhookGroup.GET("/:id", webhookHandler.GetWebhook)
	webhookGroup.DELETE("/:id", webhookHandler.DeleteWebhook)
	webhookGroup.GET("/:webhook_id/logs", webhookHandler.LogsWebhook)
	webhookGroup.GET("/:webhook_id/logs/:log_id", webhookHandler.LogWebhook)
	webhookGroup.GET("/:id/ping", webhookHandler.PingWebhook)
	webhookGroup.GET("/:id/resend", webhookHandler.Resend)
	return nil
}

func init() {
	utils.PanicIf(rhandlers.RegisterRouterFactory(path.Base(reflect.TypeOf(factory{}).PkgPath()), &factory{}))
}
