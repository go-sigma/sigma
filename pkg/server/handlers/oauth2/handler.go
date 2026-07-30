// Copyright 2024 sigma
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

package oauth2

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/server"
	oauth2svc "github.com/go-sigma/sigma/pkg/service/oauth2"
)

// Handler is the interface for the oauth2 handlers
type Handler interface {
	// Callback handles the callback request
	Callback(c *gin.Context, req *api.Oauth2CallbackRequest)
	// ClientID handles the client id request
	ClientID(c *gin.Context, req *api.Oauth2ClientIDRequest)
	// RedirectCallback ...
	RedirectCallback(c *gin.Context, req *api.Oauth2CallbackRequest)
}

var _ Handler = &handler{}

type handler struct {
	dig.In

	OAuth2Svc oauth2svc.Service
	Config    *config.Configuration
}

// Initialize registers the handler routes.
func Initialize(e *gin.Engine, h handler) error {
	oauth2Group := e.Group(consts.APIV1 + "/oauth2")
	oauth2Group.GET("/:provider/callback", server.WrapRequest(h.Callback))
	oauth2Group.GET("/:provider/client_id", server.WrapRequest(h.ClientID))
	oauth2Group.GET("/:provider/redirect_callback", server.WrapRequest(h.RedirectCallback))
	return nil
}
