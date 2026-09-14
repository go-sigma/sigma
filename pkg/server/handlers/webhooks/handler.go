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
	"github.com/gin-gonic/gin"
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/authz"
	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/server"
	svcwebhook "github.com/go-sigma/sigma/pkg/service/webhooks"
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
	// GetWebhookPing sends a test ping event to the webhook's configured endpoint and responds with no content.
	GetWebhookPing(c *gin.Context, req *api.GetWebhookPingRequest)
	// GetWebhookLog returns the recorded delivery log for the webhook, including request and response headers and bodies.
	GetWebhookLog(c *gin.Context, req *api.GetWebhookLogRequest)
	// DeleteWebhookLog removes the recorded webhook delivery log and responds with no content.
	DeleteWebhookLog(c *gin.Context, req *api.DeleteWebhookLogRequest)
	// ListWebhookLogs returns a paginated list of delivery logs for the webhook.
	ListWebhookLogs(c *gin.Context, req *api.ListWebhookLogRequest)
	// GetWebhookLogResend re-delivers a previously recorded webhook request and responds with no content.
	GetWebhookLogResend(c *gin.Context, req *api.GetWebhookLogResendRequest)
}

var _ Handler = &handler{}

type handler struct {
	dig.In

	WebhookSvc svcwebhook.Service
	Authorizer authz.Authorizer
	Config     *config.Configuration
}

// Initialize registers the handler routes.
func Initialize(e *gin.Engine, h handler) error {
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
}
