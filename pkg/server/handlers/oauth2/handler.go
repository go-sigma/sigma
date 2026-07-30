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
	"path"
	"reflect"

	"github.com/gin-gonic/gin"
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/server"
	"github.com/go-sigma/sigma/pkg/server/handlers"
	oauth2svc "github.com/go-sigma/sigma/pkg/service/oauth2"
	"github.com/go-sigma/sigma/pkg/utils"
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

type factory struct{}

// Initialize initializes the namespace handlers
func (f factory) Initialize(digCon *dig.Container) error {
	return digCon.Invoke(func(e *gin.Engine, h handler) error {
		oauth2Group := e.Group(consts.APIV1 + "/oauth2")
		oauth2Group.GET("/:provider/callback", server.WrapRequest(h.Callback))
		oauth2Group.GET("/:provider/client_id", server.WrapRequest(h.ClientID))
		oauth2Group.GET("/:provider/redirect_callback", server.WrapRequest(h.RedirectCallback))
		return nil
	})
}

func init() {
	utils.PanicIf(handlers.Routers.Register(path.Base(reflect.TypeFor[factory]().PkgPath()), &factory{}))
}
