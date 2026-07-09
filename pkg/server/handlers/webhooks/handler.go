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

	"github.com/gin-gonic/gin"
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/authz"
	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/server"
	"github.com/go-sigma/sigma/pkg/server/handlers"
	svcwebhook "github.com/go-sigma/sigma/pkg/service/webhooks"
	"github.com/go-sigma/sigma/pkg/utils"
)

// Handler is the interface for the webhook handlers
type Handler interface {
	// PostWebhook handles the post webhook request
	PostWebhook(c *gin.Context, req *api.PostWebhookRequest)
	// ListWebhook handles the list webhook request
	ListWebhook(c *gin.Context, req *api.ListWebhookRequest)
	// GetWebhook handles the get webhook request
	GetWebhook(c *gin.Context, req *api.GetWebhookRequest)
	// DeleteWebhook handles the delete webhook request
	DeleteWebhook(c *gin.Context, req *api.DeleteWebhookRequest)
	// PutWebhook handles the put webhook request
	PutWebhook(c *gin.Context, req *api.PutWebhookRequest)
	// GetWebhookPing ...
	GetWebhookPing(c *gin.Context, req *api.GetWebhookPingRequest)
	// GetWebhookLog ...
	GetWebhookLog(c *gin.Context, req *api.GetWebhookLogRequest)
	// DeleteWebhookLog ...
	DeleteWebhookLog(c *gin.Context, req *api.DeleteWebhookLogRequest)
	// ListWebhookLogs ...
	ListWebhookLogs(c *gin.Context, req *api.ListWebhookLogRequest)
	// GetWebhookLogResend ...
	GetWebhookLogResend(c *gin.Context, req *api.GetWebhookLogResendRequest)
}

var _ Handler = &handler{}

type handler struct {
	dig.In

	WebhookSvc svcwebhook.WebhookService
	Authorizer authz.Authorizer
	Config     *config.Configuration
}

func init() {
	utils.PanicIf(handlers.Routers.Register(path.Base(reflect.TypeFor[factory]().PkgPath()), &factory{}))
}

type factory struct{}

// Initialize initializes the namespace handlers
func (f factory) Initialize(digCon *dig.Container) error {
	return digCon.Invoke(func(e *gin.Engine, h handler) error {
		webhookGroup := e.Group(consts.APIV1 + "/webhooks")
		webhookGroup.POST("/", server.WrapRequest(h.PostWebhook))
		webhookGroup.PUT("/:webhook_id", server.WrapRequest(h.PutWebhook))
		webhookGroup.GET("/", server.WrapRequest(h.ListWebhook))
		webhookGroup.GET("/:webhook_id", server.WrapRequest(h.GetWebhook))
		webhookGroup.DELETE("/:webhook_id", server.WrapRequest(h.DeleteWebhook))
		webhookGroup.GET("/:webhook_id/logs/", server.WrapRequest(h.ListWebhookLogs))
		webhookGroup.GET("/:webhook_id/logs/:webhook_log_id", server.WrapRequest(h.GetWebhookLog))
		webhookGroup.DELETE("/:webhook_id/logs/:webhook_log_id", server.WrapRequest(h.DeleteWebhookLog))
		webhookGroup.GET("/:webhook_id/ping", server.WrapRequest(h.GetWebhookPing))
		webhookGroup.GET("/:webhook_id/logs/:webhook_log_id/resend", server.WrapRequest(h.GetWebhookLogResend))
		return nil
	})
}
