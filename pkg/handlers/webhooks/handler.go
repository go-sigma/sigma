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

	"github.com/go-sigma/sigma/pkg/auth"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/dao"
	"github.com/go-sigma/sigma/pkg/handlers"
	"github.com/go-sigma/sigma/pkg/middlewares"
	"github.com/go-sigma/sigma/pkg/modules/workq"
	"github.com/go-sigma/sigma/pkg/modules/workq/definition"
	"github.com/go-sigma/sigma/pkg/utils"
)

// Handler is the interface for the webhook handlers
type Handler interface {
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
	// GetWebhookPing ...
	GetWebhookPing(c echo.Context) error
	// GetWebhookLog ...
	GetWebhookLog(c echo.Context) error
	// DeleteWebhookLog ...
	DeleteWebhookLog(c echo.Context) error
	// ListWebhookLogs ...
	ListWebhookLogs(c echo.Context) error
	// GetWebhookLogResend ...
	GetWebhookLogResend(c echo.Context) error
}

var _ Handler = &handler{}

type handler struct {
	authServiceFactory      auth.AuthServiceFactory
	namespaceServiceFactory dao.NamespaceServiceFactory
	webhookServiceFactory   dao.WebhookServiceFactory
	auditServiceFactory     dao.AuditServiceFactory

	producerClient definition.WorkQueueProducer
}

type inject struct {
	authServiceFactory      auth.AuthServiceFactory
	namespaceServiceFactory dao.NamespaceServiceFactory
	webhookServiceFactory   dao.WebhookServiceFactory
	auditServiceFactory     dao.AuditServiceFactory

	producerClient definition.WorkQueueProducer
}

// handlerNew creates a new instance of the webhook handlers
func handlerNew(injects ...inject) Handler {
	namespaceServiceFactory := dao.NewNamespaceServiceFactory()
	webhookServiceFactory := dao.NewWebhookServiceFactory()
	auditServiceFactory := dao.NewAuditServiceFactory()
	authServiceFactory := auth.NewAuthServiceFactory()
	producerClient := workq.ProducerClient
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
		if ij.authServiceFactory != nil {
			authServiceFactory = ij.authServiceFactory
		}
		if ij.producerClient != nil {
			producerClient = ij.producerClient
		}
	}
	return &handler{
		authServiceFactory:      authServiceFactory,
		namespaceServiceFactory: namespaceServiceFactory,
		webhookServiceFactory:   webhookServiceFactory,
		auditServiceFactory:     auditServiceFactory,
		producerClient:          producerClient,
	}
}

type factory struct{}

// Initialize initializes the namespace handlers
func (f factory) Initialize(e *echo.Echo) error {
	webhookGroup := e.Group(consts.APIV1+"/webhooks", middlewares.AuthWithConfig(middlewares.AuthConfig{}))

	webhookHandler := handlerNew()
	webhookGroup.POST("/", webhookHandler.PostWebhook)
	webhookGroup.PUT("/:webhook_id", webhookHandler.PutWebhook)
	webhookGroup.GET("/", webhookHandler.ListWebhook)
	webhookGroup.GET("/:webhook_id", webhookHandler.GetWebhook)
	webhookGroup.DELETE("/:webhook_id", webhookHandler.DeleteWebhook)
	webhookGroup.GET("/:webhook_id/logs/", webhookHandler.ListWebhookLogs)
	webhookGroup.GET("/:webhook_id/logs/:webhook_log_id", webhookHandler.GetWebhookLog)
	webhookGroup.DELETE("/:webhook_id/logs/:webhook_log_id", webhookHandler.DeleteWebhookLog)
	webhookGroup.GET("/:webhook_id/ping", webhookHandler.GetWebhookPing)
	webhookGroup.GET("/:webhook_id/logs/:webhook_log_id/resend", webhookHandler.GetWebhookLogResend)
	return nil
}

func init() {
	utils.PanicIf(handlers.RegisterRouterFactory(path.Base(reflect.TypeOf(factory{}).PkgPath()), &factory{}))
}
