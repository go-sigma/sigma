// The MIT License (MIT)
//
// Copyright © 2023 Tosone <i@tosone.cn>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package handlers

import (
	"encoding/base64"
	"net/http"

	"github.com/go-playground/validator"
	"github.com/golang-jwt/jwt/v4"
	echoJwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
	echoSwagger "github.com/swaggo/echo-swagger"

	_ "github.com/ximager/ximager/pkg/handlers/apidocs"
	"github.com/ximager/ximager/pkg/handlers/user"
	"github.com/ximager/ximager/pkg/types"

	"github.com/ximager/ximager/pkg/handlers/artifact"
	"github.com/ximager/ximager/pkg/handlers/distribution"
	"github.com/ximager/ximager/pkg/handlers/namespace"
	"github.com/ximager/ximager/pkg/handlers/repository"
	"github.com/ximager/ximager/pkg/handlers/tag"
	"github.com/ximager/ximager/pkg/middlewares"
	"github.com/ximager/ximager/pkg/validators"
	"github.com/ximager/ximager/web"
)

// CustomValidator is a custom validator for echo
type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		// Optionally, you could return the error to give each route more control over the status code
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}

func Initialize(e *echo.Echo) error {
	web.RegisterHandlers(e)

	e.Any("/swagger/*", echoSwagger.WrapHandler)

	validate := validator.New()
	validators.Register(validate)
	e.Validator = &CustomValidator{validator: validate}

	e.GET("/health", func(c echo.Context) error {
		return c.String(200, "OK")
	})

	publicKeyBytes, err := base64.StdEncoding.DecodeString(viper.GetString("admin.jwt.publicKey"))
	if err != nil {
		return err
	}

	userGroup := e.Group("/user")
	userHandler := user.New()
	config := echoJwt.Config{
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return new(types.JWTClaims)
		},
		SigningKey: publicKeyBytes,
	}
	userGroup.Use(echoJwt.WithConfig(config))
	userGroup.POST("/login", userHandler.Login)

	e.GET("/service/token", func(c echo.Context) error {
		str := `{"token": "eyJ0eXAiOiJKV1QiLCJhbGciOiJFUzI1NiIsImtpZCI6IlBZWU86VEVXVTpWN0pIOjI2SlY6QVFUWjpMSkMzOlNYVko6WEdIQTozNEYyOjJMQVE6WlJNSzpaN1E2In0.eyJpc3MiOiJhdXRoLmRvY2tlci5jb20iLCJzdWIiOiJqbGhhd24iLCJhdWQiOiJyZWdpc3RyeS5kb2NrZXIuY29tIiwiZXhwIjoxNDE1Mzg3MzE1LCJuYmYiOjE0MTUzODcwMTUsImlhdCI6MTQxNTM4NzAxNSwianRpIjoidFlKQ08xYzZjbnl5N2tBbjBjN3JLUGdiVjFIMWJGd3MiLCJhY2Nlc3MiOlt7InR5cGUiOiJyZXBvc2l0b3J5IiwibmFtZSI6InNhbWFsYmEvbXktYXBwIiwiYWN0aW9ucyI6WyJwdXNoIl19XX0.QhflHPfbd6eVF4lM9bwYpFZIV0PfikbyXuLx959ykRTBpe3CYnzs6YBK8FToVb5R47920PVLrh8zuLzdCr9t3w", "expires_in": 3600,"issued_at": "2009-11-10T23:00:00Z"}`
		return c.JSONBlob(200, []byte(str))
	})

	namespaceGroup := e.Group("/namespace", middlewares.AuthWithConfig(middlewares.AuthConfig{}))
	namespaceHandler := namespace.New()
	namespaceGroup.POST("/", namespaceHandler.PostNamespace)
	namespaceGroup.PUT("/:id", func(c echo.Context) error {
		return c.String(200, "OK")
	})
	namespaceGroup.DELETE("/:id", func(c echo.Context) error {
		return c.String(200, "OK")
	})
	namespaceGroup.GET("/:id", func(c echo.Context) error {
		return c.String(200, "OK")
	})
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

	e.Any("/v2/*", distribution.All)

	return nil
}
