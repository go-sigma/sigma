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
	"path"
	"reflect"

	"github.com/labstack/echo/v4"
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/dao"
	"github.com/go-sigma/sigma/pkg/handlers"
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
	SelfGet(c echo.Context) error
	// SelfPut handles the self put request
	SelfPut(c echo.Context) error
	// SelfResetPassword handles the self reset request
	SelfResetPassword(c echo.Context) error
}

type handler struct {
	dig.In

	Config             configs.Configuration
	TokenService       token.Service
	PasswordService    password.Service
	UserServiceFactory dao.UserServiceFactory
}

var _ Handler = &handler{}

// handlerNew creates a new instance of the distribution handlers
func handlerNew(digCon *dig.Container) Handler {
	return ptr.Of(utils.MustGetObjFromDigCon[handler](digCon))
}

type factory struct{}

// var skipAuths = []string{"get:/api/v1/users/token", "get:/api/v1/users/signup", "get:/api/v1/users/create"}

func (f factory) Initialize(digCon *dig.Container) error {
	echo := utils.MustGetObjFromDigCon[*echo.Echo](digCon)
	userGroup := echo.Group(consts.APIV1 + "/users")
	handler := handlerNew(digCon)
	// userGroup.Use(middlewares.AuthnWithConfig(middlewares.Config{
	// 	Skipper: func(c echo.Context) bool {
	// 		authStr := strings.ToLower(fmt.Sprintf("%s:%s", c.Request().Method, c.Request().URL.Path))
	// 		return slices.Contains(skipAuths, authStr)
	// 	},
	// }))

	userGroup.GET("/", handler.List)
	userGroup.POST("/", handler.Post)
	userGroup.PUT("/:id", handler.Put)
	userGroup.POST("/login", handler.Login)
	userGroup.POST("/logout", handler.Logout)
	userGroup.GET("/signup", handler.Signup)
	userGroup.GET("/create", handler.Signup)

	userGroup.GET("/self", handler.SelfGet)
	userGroup.PUT("/self", handler.SelfPut)
	userGroup.PUT("/self/reset-password", handler.SelfResetPassword)

	userGroup.GET("/recover-password", handler.RecoverPassword)
	userGroup.PUT("/recover-password-reset/:code", handler.RecoverPasswordReset)

	userGroup.PUT("/:id/reset-password", handler.ResetPassword)

	return nil
}

func init() {
	utils.PanicIf(handlers.RegisterRouterFactory(path.Base(reflect.TypeOf(factory{}).PkgPath()), &factory{}))
}
