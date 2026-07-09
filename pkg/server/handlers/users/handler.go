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

	"github.com/gin-gonic/gin"
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/server"
	"github.com/go-sigma/sigma/pkg/server/handlers"
	"github.com/go-sigma/sigma/pkg/service/users"
	"github.com/go-sigma/sigma/pkg/utils"
)

// Handler is the interface for the tag handlers
type Handler interface {
	// Login handles the login request
	Login(c *gin.Context)
	// Logout handles the logout request
	Logout(c *gin.Context, req *api.PostUserLogoutRequest)
	// Signup handles the signup request
	Signup(c *gin.Context, req *api.PostUserSignupRequest)
	// ResetPassword handles the reset request
	ResetPassword(c *gin.Context, req *api.PostUserResetPasswordPasswordRequest)
	// List handles the list user request
	List(c *gin.Context, req *api.GetUserListRequest)
	// Put handles the put request
	Put(c *gin.Context, req *api.PutUserRequest)
	// Post handles the post request
	Post(c *gin.Context, req *api.PostUserRequest)

	// RecoverPassword handles the recover user's password
	RecoverPassword(c *gin.Context, req *api.PostUserRecoverPasswordRequest)
	// RecoverPasswordReset handles the recover user's password reset
	RecoverPasswordReset(c *gin.Context, req *api.PostUserRecoverResetPasswordRequest)

	// Self handles the self request
	SelfGet(c *gin.Context)
	// SelfPut handles the self put request
	SelfPut(c *gin.Context, req *api.PutUserSelfRequest)
	// SelfResetPassword handles the self reset request
	SelfResetPassword(c *gin.Context, req *api.PutUserSelfResetPasswordRequest)
}

type handler struct {
	dig.In

	Config  *config.Configuration
	UserSvc users.UserService
}

var _ Handler = &handler{}

type factory struct{}

// Initialize ...
func (f factory) Initialize(digCon *dig.Container) error {
	return digCon.Invoke(func(e *gin.Engine, h handler) error {
		group := e.Group(consts.APIV1 + "/users")
		group.GET("/", server.WrapRequest(h.List))
		group.POST("/", server.WrapRequest(h.Post))
		group.PUT("/:id", server.WrapRequest(h.Put))
		group.POST("/login", server.Wrap(h.Login))
		group.POST("/logout", server.WrapRequest(h.Logout))
		group.GET("/signup", server.WrapRequest(h.Signup))
		group.GET("/create", server.WrapRequest(h.Signup))
		group.GET("/self", server.Wrap(h.SelfGet))
		group.PUT("/self", server.WrapRequest(h.SelfPut))
		group.PUT("/self/reset-password", server.WrapRequest(h.SelfResetPassword))
		group.GET("/recover-password", server.WrapRequest(h.RecoverPassword))
		group.PUT("/recover-password-reset/:code", server.WrapRequest(h.RecoverPasswordReset))
		group.PUT("/:id/reset-password", server.WrapRequest(h.ResetPassword))
		return nil
	})
}

func init() {
	utils.PanicIf(handlers.Routers.Register(path.Base(reflect.TypeFor[factory]().PkgPath()), &factory{}))
}
