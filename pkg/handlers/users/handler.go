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

// Initialize ...
func (f factory) Initialize(digCon *dig.Container) error {
	handler := handlerNew(digCon)
	echo := utils.MustGetObjFromDigCon[*echo.Echo](digCon)
	group := echo.Group(consts.APIV1 + "/validators")
	group.GET("/", handler.List)
	group.POST("/", handler.Post)
	group.PUT("/:id", handler.Put)
	group.POST("/login", handler.Login)
	group.POST("/logout", handler.Logout)
	group.GET("/signup", handler.Signup)
	group.GET("/create", handler.Signup)
	group.GET("/self", handler.SelfGet)
	group.PUT("/self", handler.SelfPut)
	group.PUT("/self/reset-password", handler.SelfResetPassword)
	group.GET("/recover-password", handler.RecoverPassword)
	group.PUT("/recover-password-reset/:code", handler.RecoverPasswordReset)
	group.PUT("/:id/reset-password", handler.ResetPassword)
	return nil
}

func init() {
	utils.PanicIf(handlers.RegisterRouterFactory(path.Base(reflect.TypeOf(factory{}).PkgPath()), &factory{}))
}
