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
	"path"
	"reflect"

	"github.com/labstack/echo/v4"

	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/dao"
	"github.com/go-sigma/sigma/pkg/handlers"
	"github.com/go-sigma/sigma/pkg/middlewares"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/password"
	"github.com/go-sigma/sigma/pkg/utils/token"
)

// Handler is the interface for the tag handlers
type Handler interface {
	// Token handles the token request
	Token(c echo.Context) error
}

type handler struct {
	config             *configs.Configuration
	tokenService       token.TokenService
	passwordService    password.Password
	userServiceFactory dao.UserServiceFactory
}

var _ Handler = &handler{}

type inject struct {
	config             *configs.Configuration
	tokenService       token.TokenService
	passwordService    password.Password
	userServiceFactory dao.UserServiceFactory
}

// handlerNew creates a new instance of the distribution handlers
func handlerNew(injects ...inject) (Handler, error) {
	var tokenService token.TokenService
	passwordService := password.New()
	userServiceFactory := dao.NewUserServiceFactory()
	config := configs.GetConfiguration()
	if len(injects) > 0 {
		ij := injects[0]
		if ij.tokenService != nil {
			tokenService = ij.tokenService
		}
		if ij.passwordService != nil {
			passwordService = ij.passwordService
		}
		if ij.userServiceFactory != nil {
			userServiceFactory = ij.userServiceFactory
		}
		if ij.config != nil {
			config = ij.config
		}
	} else {
		var err error
		tokenService, err = token.NewTokenService(config.Auth.Jwt.PrivateKey)
		if err != nil {
			return nil, err
		}
	}
	return &handler{
		config:             config,
		tokenService:       tokenService,
		passwordService:    passwordService,
		userServiceFactory: userServiceFactory,
	}, nil
}

type factory struct{}

func (f factory) Initialize(e *echo.Echo) error {
	userHandler, err := handlerNew()
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
