// Copyright 2023 XImager
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
	"github.com/spf13/viper"

	"github.com/ximager/ximager/pkg/dal/dao"
	rhandlers "github.com/ximager/ximager/pkg/handlers"
	"github.com/ximager/ximager/pkg/utils"
	"github.com/ximager/ximager/pkg/utils/password"
	"github.com/ximager/ximager/pkg/utils/token"
)

// Handlers is the interface for the tag handlers
type Handlers interface {
	// Token handles the token request
	Token(c echo.Context) error
}

type handlers struct {
	tokenService       token.TokenService
	passwordService    password.Password
	userServiceFactory dao.UserServiceFactory
}

var _ Handlers = &handlers{}

type inject struct {
	tokenService       token.TokenService
	passwordService    password.Password
	userServiceFactory dao.UserServiceFactory
}

// handlerNew creates a new instance of the distribution handlers
func handlerNew(injects ...inject) (Handlers, error) {
	tokenService, err := token.NewTokenService(viper.GetString("auth.jwt.privateKey"))
	if err != nil {
		return nil, err
	}
	passwordService := password.New()
	userServiceFactory := dao.NewUserServiceFactory()
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
	}
	return &handlers{
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
	e.GET("/token", userHandler.Token)
	return nil
}

func init() {
	utils.PanicIf(rhandlers.RegisterRouterFactory(path.Base(reflect.TypeOf(handlers{}).PkgPath()), &factory{}))
}
