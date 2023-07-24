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

package users

import (
	"fmt"
	"path"
	"reflect"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
	"golang.org/x/exp/slices"

	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/dao"
	rhandlers "github.com/go-sigma/sigma/pkg/handlers"
	"github.com/go-sigma/sigma/pkg/middlewares"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/password"
	"github.com/go-sigma/sigma/pkg/utils/token"
)

// Handlers is the interface for the tag handlers
type Handlers interface {
	// Login handles the login request
	Login(c echo.Context) error
	// Logout handles the logout request
	Logout(c echo.Context) error
	// Signup handles the signup request
	Signup(c echo.Context) error
	// Self handles the self request
	Self(c echo.Context) error
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

var skipAuths = []string{"post:/api/v1/users/login", "get:/api/v1/users/token", "get:/api/v1/users/signup", "get:/api/v1/users/create"}

func (f factory) Initialize(e *echo.Echo) error {
	userGroup := e.Group(consts.APIV1 + "/users")
	userHandler, err := handlerNew()
	if err != nil {
		return err
	}
	userGroup.Use(middlewares.AuthWithConfig(middlewares.AuthConfig{
		Skipper: func(c echo.Context) bool {
			authStr := strings.ToLower(fmt.Sprintf("%s:%s", c.Request().Method, c.Request().URL.Path))
			return slices.Contains(skipAuths, authStr)
		},
	}))
	userGroup.POST("/login", userHandler.Login)
	userGroup.POST("/logout", userHandler.Logout)
	userGroup.GET("/signup", userHandler.Signup)
	userGroup.GET("/create", userHandler.Signup)
	userGroup.GET("/self", userHandler.Self)
	return nil
}

func init() {
	utils.PanicIf(rhandlers.RegisterRouterFactory(path.Base(reflect.TypeOf(factory{}).PkgPath()), &factory{}))
}
