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

package users

import (
	"fmt"
	"path"
	"reflect"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
	"golang.org/x/exp/slices"

	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/dao"
	"github.com/go-sigma/sigma/pkg/handlers"
	"github.com/go-sigma/sigma/pkg/middlewares"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/password"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
	"github.com/go-sigma/sigma/pkg/utils/token"
)

// Handler is the interface for the tag handlers
type Handler interface {
	// Login handles the login request
	Login(c echo.Context) error
	// Logout handles the logout request
	Logout(c echo.Context) error
	// Signup handles the signup request
	Signup(c echo.Context) error
	// ResetPassword handles the reset request
	ResetPassword(c echo.Context) error
	// List handles the list user request
	List(c echo.Context) error
	// Put handles the put request
	Put(c echo.Context) error
	// Post handles the post request
	Post(c echo.Context) error

	// RecoverPassword handles the recover user's password
	RecoverPassword(c echo.Context) error
	// RecoverPasswordReset handles the recover user's password reset
	RecoverPasswordReset(c echo.Context) error

	// Self handles the self request
	Self(c echo.Context) error
	// SelfPut handles the self put request
	SelfPut(c echo.Context) error
	// SelfResetPassword handles the self reset request
	SelfResetPassword(c echo.Context) error
}

type handler struct {
	config             configs.Configuration
	tokenService       token.TokenService
	passwordService    password.Password
	userServiceFactory dao.UserServiceFactory
}

var _ Handler = &handler{}

type inject struct {
	tokenService       token.TokenService
	passwordService    password.Password
	userServiceFactory dao.UserServiceFactory
}

// handlerNew creates a new instance of the distribution handlers
func handlerNew(injects ...inject) (Handler, error) {
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
	return &handler{
		config:             ptr.To(configs.GetConfiguration()),
		tokenService:       tokenService,
		passwordService:    passwordService,
		userServiceFactory: userServiceFactory,
	}, nil
}

type factory struct{}

// "post:/api/v1/users/login",
var skipAuths = []string{"get:/api/v1/users/token", "get:/api/v1/users/signup", "get:/api/v1/users/create"}

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

	userGroup.GET("/", userHandler.List)
	userGroup.POST("/", userHandler.Post)
	userGroup.PUT("/:id", userHandler.Put)
	userGroup.POST("/login", userHandler.Login)
	userGroup.POST("/logout", userHandler.Logout)
	userGroup.GET("/signup", userHandler.Signup)
	userGroup.GET("/create", userHandler.Signup)

	userGroup.GET("/self", userHandler.Self)
	userGroup.PUT("/self", userHandler.SelfPut)
	userGroup.PUT("/self/reset-password", userHandler.SelfResetPassword)

	userGroup.GET("/recover-password", userHandler.RecoverPassword)
	userGroup.PUT("/recover-password-reset/:code", userHandler.RecoverPasswordReset)

	userGroup.PUT("/:id/reset-password", userHandler.ResetPassword)

	return nil
}

func init() {
	utils.PanicIf(handlers.RegisterRouterFactory(path.Base(reflect.TypeOf(factory{}).PkgPath()), &factory{}))
}
