// Copyright 2025 sigma
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

package app

import (
	"crypto/sha256"
	"fmt"
	"hash"
	"hash/crc32"
	"log/slog"
	"reflect"
	"slices"
	"strings"

	"github.com/gin-contrib/cors"
	ginpprof "github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.uber.org/dig"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/api/enums"
	authzsvc "github.com/go-sigma/sigma/pkg/authz"
	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/consts"
	dalredis "github.com/go-sigma/sigma/pkg/dal/redis"
	repoaudit "github.com/go-sigma/sigma/pkg/dal/repository/audit"
	repouser "github.com/go-sigma/sigma/pkg/dal/repository/user"
	"github.com/go-sigma/sigma/pkg/server/middlewares"
	"github.com/go-sigma/sigma/pkg/server/middlewares/audit"
	"github.com/go-sigma/sigma/pkg/server/middlewares/authn"
	"github.com/go-sigma/sigma/pkg/server/middlewares/authz"
	"github.com/go-sigma/sigma/pkg/server/middlewares/bodylimit"
	"github.com/go-sigma/sigma/pkg/server/middlewares/etag"
	"github.com/go-sigma/sigma/pkg/server/middlewares/healthz"
	"github.com/go-sigma/sigma/pkg/server/middlewares/metrics"
	"github.com/go-sigma/sigma/pkg/service/password"
	"github.com/go-sigma/sigma/pkg/service/token"
	"github.com/go-sigma/sigma/pkg/storage"
)

// GinServerParams declares the dependencies needed to construct the gin engine.
type GinServerParams struct {
	dig.In

	Config          *config.Configuration
	TokenSvc        token.Service
	PasswordSvc     password.Service
	UserRepository  repouser.UserRepository
	AuditRepository repoaudit.AuditRepository
	Authorizer      authzsvc.Authorizer
	StorageDriver   storage.StorageDriver
	Database        *gorm.DB

	RedisClientFactory dalredis.ClientFactory `optional:"true"`
}

// NewGinServer ...
func NewGinServer(params GinServerParams) (*gin.Engine, error) {
	gin.SetMode(gin.ReleaseMode)
	e := gin.New()

	// OTel tracing — must be first so all subsequent middleware is inside the span.
	e.Use(otelgin.Middleware(consts.AppName))

	e.Use(func(c *gin.Context) {
		level := slog.LevelDebug
		path := c.Request.URL.Path
		if path == "/healthz" || path == "/readyz" || path == "/metrics" {
			level = slog.LevelDebug - 1 // trace-like
		}
		slog.Default().Log(c.Request.Context(), level, "request debugger",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"query", c.Request.URL.RawQuery,
		)
		c.Next()
	})
	e.Use(cors.New(cors.Config{
		AllowAllOrigins: true,
		AllowMethods: []string{
			"GET",
			"POST",
			"PUT",
			"PATCH",
			"DELETE",
			"HEAD",
			"OPTIONS",
		},
		AllowHeaders: []string{
			"Authorization",
			"Content-Type",
			"Origin",
			"Accept",
			"X-Requested-With",
			"X-Request-ID",
			"If-Match",
			"If-None-Match",
		},
		ExposeHeaders: []string{
			"Content-Length",
			"Content-Type",
			"ETag",
			"Location",
			"X-Request-ID",
		},
	}))
	e.Use(etag.WithEtagConfig(etag.EtagConfig{
		Skipper: func(c *gin.Context) bool {
			reqPath := c.Request.URL.Path
			if strings.HasPrefix(reqPath, "/api/v1/") {
				return true
			}
			if strings.HasPrefix(reqPath, "/v2/") {
				return true
			}
			return false
		},
		Weak: true,
		HashFn: func(config etag.EtagConfig) hash.Hash {
			if config.Weak {
				return crc32.New(crc32.MakeTable(0xD5828281))
			}
			return sha256.New()
		},
	}))
	e.Use(metrics.Middleware(consts.AppName))
	e.Use(bodylimit.BodyLimit(params.Config.HTTP.BodyLimit))
	e.GET("/metrics", gin.WrapH(promhttp.Handler()))
	e.Use(healthz.Healthz(NewReadinessResources(params)...))
	if params.Config.Log.Level == enums.LogLevelDebug || params.Config.Log.Level == enums.LogLevelTrace {
		ginpprof.Register(e, consts.PprofPath)
	}
	e.Use(middlewares.RedirectRepository(params.Config))
	e.Use(authn.AuthnWithConfig(authn.Config{
		Skipper:        genSkipper(),
		Config:         params.Config,
		TokenSvc:       params.TokenSvc,
		PasswordSvc:    params.PasswordSvc,
		UserRepository: params.UserRepository,
	}))
	e.Use(authz.AuthzWithConfig(authz.Config{
		Skipper:    genAuthzSkipper(),
		Authorizer: params.Authorizer,
	}))
	if params.Config.Audit.IsEnabled() {
		e.Use(audit.AuditWithConfig(audit.Config{
			Skipper:         genAuditSkipper(),
			RecordListGet:   params.Config.Audit.RecordListGet,
			AuditRepository: params.AuditRepository,
		}))
	}
	return e, nil
}

var skipAuthns = []string{"get:/api/v1/users/token", "get:/api/v1/users/signup", "get:/api/v1/users/create"}

func genSkipper() authn.Skipper {
	var oauth2 = reflect.TypeFor[config.ConfigurationAuthOauth2]()
	for field := range oauth2.Fields() {
		skipAuthns = append(skipAuthns, fmt.Sprintf("get:/api/v1/oauth2/%s/client_id", strings.ToLower(field.Name)))
		skipAuthns = append(skipAuthns, fmt.Sprintf("get:/api/v1/oauth2/%s/callback", strings.ToLower(field.Name)))
		skipAuthns = append(skipAuthns, fmt.Sprintf("get:/api/v1/oauth2/%s/redirect_callback", strings.ToLower(field.Name)))
	}
	return func(c *gin.Context) bool {
		requestUri := c.Request.RequestURI
		requestMethod := c.Request.Method
		if !(strings.HasPrefix(requestUri, "/v2/") || strings.HasPrefix(requestUri, consts.APIV1)) { // nolint: staticcheck
			return true
		}
		return slices.Contains(skipAuthns, strings.ToLower(fmt.Sprintf("%s:%s", requestMethod, requestUri)))
	}
}

var skipAuthzs = []string{"post:/api/v1/users/login", "get:/api/v1/users/token", "get:/api/v1/users/signup", "get:/api/v1/users/create"}

func genAuthzSkipper() authz.Skipper {
	var oauth2 = reflect.TypeFor[config.ConfigurationAuthOauth2]()
	for field := range oauth2.Fields() {
		skipAuthzs = append(skipAuthzs, fmt.Sprintf("get:/api/v1/oauth2/%s/client_id", strings.ToLower(field.Name)))
		skipAuthzs = append(skipAuthzs, fmt.Sprintf("get:/api/v1/oauth2/%s/callback", strings.ToLower(field.Name)))
		skipAuthzs = append(skipAuthzs, fmt.Sprintf("get:/api/v1/oauth2/%s/redirect_callback", strings.ToLower(field.Name)))
	}
	return func(c *gin.Context) bool {
		requestUri := c.Request.RequestURI
		requestMethod := c.Request.Method
		if !(strings.HasPrefix(requestUri, "/v2/") || strings.HasPrefix(requestUri, consts.APIV1)) { // nolint: staticcheck
			return true
		}
		if config.GetConfig().MCP.Enabled && requestUri == config.GetConfig().MCP.Path {
			return true
		}
		return slices.Contains(skipAuthzs, strings.ToLower(fmt.Sprintf("%s:%s", requestMethod, requestUri)))
	}
}

func genAuditSkipper() audit.Skipper {
	return func(c *gin.Context) bool {
		requestPath := c.Request.URL.Path
		if config.GetConfig().MCP.Enabled && requestPath == config.GetConfig().MCP.Path {
			return true
		}
		return !strings.HasPrefix(requestPath, "/v2/") && requestPath != "/v2" && !strings.HasPrefix(requestPath, consts.APIV1)
	}
}
