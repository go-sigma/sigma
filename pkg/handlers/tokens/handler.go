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

package token

import (
	"fmt"
	"path"
	"reflect"

	"github.com/labstack/echo/v4"
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/handlers"
	"github.com/go-sigma/sigma/pkg/middlewares"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/password"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
	"github.com/go-sigma/sigma/pkg/utils/token"
)

// Handler is the interface for the tag handlers
type Handler interface {
	// Token handles the token request
	Token(c echo.Context) error
}

type handler struct {
	dig.In

	Config          configs.Configuration
	TokenService    token.TokenService
	PasswordService password.Password
}

var _ Handler = &handler{}

// handlerNew creates a new instance of the distribution handlers
func handlerNew(c *dig.Container) (Handler, error) {
	var h *handler
	err := c.Invoke(func(handler handler) { h = ptr.Of(handler) })
	if err != nil {
		return nil, fmt.Errorf("failed to build handler: %v", err)
	}
	return h, nil
}

type factory struct{}

func (f factory) Initialize(e *echo.Echo, c *dig.Container) error {
	userHandler, err := handlerNew(c)
	if err != nil {
		return err
	}
	tokenGroup := e.Group(consts.APIV1, middlewares.AuthWithConfig(middlewares.AuthConfig{}))
	tokenGroup.GET("/tokens", userHandler.Token)
	return nil
}

func init() {
	utils.PanicIf(handlers.RegisterRouterFactory(path.Base(reflect.TypeOf(factory{}).PkgPath()), &factory{}))
}
