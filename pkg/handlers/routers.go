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

package handlers

import (
	"fmt"
	"strings"

	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"
	"golang.org/x/exp/slices"

	"github.com/ximager/ximager/pkg/handlers/artifact"
	"github.com/ximager/ximager/pkg/handlers/distribution"
	"github.com/ximager/ximager/pkg/handlers/namespace"
	"github.com/ximager/ximager/pkg/handlers/repository"
	"github.com/ximager/ximager/pkg/handlers/tag"
	"github.com/ximager/ximager/pkg/handlers/user"
	"github.com/ximager/ximager/pkg/middlewares"
	"github.com/ximager/ximager/pkg/validators"
	"github.com/ximager/ximager/web"

	_ "github.com/ximager/ximager/pkg/handlers/apidocs"
)

var skipAuths = []string{"post:/user/login", "get:/user/token", "get:/user/signup", "get:/user/create"}

var userHandler user.Handlers

func Initialize(e *echo.Echo) error {
	web.RegisterHandlers(e)

	e.Any("/swagger/*", echoSwagger.WrapHandler)

	validators.Initialize(e)

	userGroup := e.Group("/user")
	userHandler = user.New()
	userGroup.Use(middlewares.AuthWithConfig(middlewares.AuthConfig{
		Skipper: func(c echo.Context) bool {
			authStr := strings.ToLower(fmt.Sprintf("%s:%s", c.Request().Method, c.Request().URL.Path))
			return slices.Contains(skipAuths, authStr)
		},
	}))
	userGroup.POST("/login", userHandler.Login)
	userGroup.GET("/logout", userHandler.Logout)
	userGroup.GET("/token", userHandler.Token)
	userGroup.GET("/signup", userHandler.Signup)
	userGroup.GET("/create", userHandler.Signup)

	namespaceGroup := e.Group("/namespace", middlewares.AuthWithConfig(middlewares.AuthConfig{}))
	namespaceHandler := namespace.New()
	namespaceGroup.POST("/", namespaceHandler.PostNamespace)
	namespaceGroup.PUT("/:id", namespaceHandler.PutNamespace)
	namespaceGroup.DELETE("/:id", namespaceHandler.DeleteNamespace)
	// namespaceGroup.GET("/:id",name)
	namespaceGroup.GET("/", namespaceHandler.ListNamespace)

	repositoryGroup := namespaceGroup.Group("/:namespace/repository")
	repositoryHandler := repository.New()
	repositoryGroup.GET("/", repositoryHandler.ListRepository)
	repositoryGroup.GET("/:id", repositoryHandler.GetRepository)
	repositoryGroup.DELETE("/:id", repositoryHandler.DeleteRepository)

	artifactGroup := namespaceGroup.Group("/:namespace/artifact")
	artifactHandler := artifact.New()
	artifactGroup.GET("/", artifactHandler.ListArtifact)
	artifactGroup.GET("/:id", artifactHandler.GetArtifact)
	artifactGroup.DELETE("/:id", artifactHandler.DeleteArtifact)

	tagGroup := namespaceGroup.Group("/:namespace/tag")
	tagHandler := tag.New()
	tagGroup.GET("/", tagHandler.ListTag)
	tagGroup.GET("/:id", tagHandler.GetTag)
	tagGroup.DELETE("/:id", tagHandler.DeleteTag)

	e.Any("/v2/*", distribution.All, middlewares.AuthWithConfig(middlewares.AuthConfig{DS: true}))

	return nil
}
