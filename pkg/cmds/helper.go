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

package cmds

import (
	"crypto/sha256"
	"fmt"
	"hash"
	"hash/crc32"
	"reflect"
	"slices"
	"strings"

	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo-contrib/pprof"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog/log"
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/server/middlewares"
	"github.com/go-sigma/sigma/pkg/server/middlewares/authn"
	"github.com/go-sigma/sigma/pkg/server/middlewares/authz"
	"github.com/go-sigma/sigma/pkg/server/middlewares/etag"
	"github.com/go-sigma/sigma/pkg/server/middlewares/healthz"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/serializer"
)

// NewEchoServer ...
func NewEchoServer(digCon *dig.Container) (*echo.Echo, error) {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Use(echo.MiddlewareFunc(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			var logger = log.Debug()
			var path = c.Request().URL.Path
			if path == "/healthz" || path == "/metrics" {
				logger = log.Trace()
			}
			logger.
				Str("method", c.Request().Method).
				Str("path", c.Request().URL.Path).
				Str("query", c.Request().URL.RawQuery).
				Msg("request debugger")
			return next(c)
		}
	}))
	e.Use(middleware.CORS())
	e.Use(etag.WithEtagConfig(etag.EtagConfig{
		Skipper: func(c echo.Context) bool {
			reqPath := c.Request().URL.Path
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
	e.Use(echoprometheus.NewMiddleware(consts.AppName))
	e.GET("/metrics", echoprometheus.NewHandler())
	e.Use(healthz.Healthz())
	e.JSONSerializer = new(serializer.DefaultJSONSerializer)
	config := utils.MustGetObjFromDigCon[configs.Configuration](digCon)
	if config.Log.Level == enums.LogLevelDebug || config.Log.Level == enums.LogLevelTrace {
		pprof.Register(e, consts.PprofPath)
	}
	e.Use(middlewares.RedirectRepository(config))
	e.Use(authn.AuthnWithConfig(authn.Config{
		DigCon:  digCon,
		Skipper: genSkipper(),
	}))
	e.Use(authz.AuthzWithConfig(authz.Config{
		DigCon:  digCon,
		Skipper: genAuthzSkipper(),
	}))
	return e, nil
}

var skipAuthns = []string{"get:/api/v1/users/token", "get:/api/v1/users/signup", "get:/api/v1/users/create"}

func genSkipper() middleware.Skipper {
	var oauth2 = reflect.TypeOf(configs.ConfigurationAuthOauth2{})
	for key := range oauth2.NumField() {
		skipAuthns = append(skipAuthns, fmt.Sprintf("get:/api/v1/oauth2/%s/client_id", strings.ToLower(oauth2.Field(key).Name)))
		skipAuthns = append(skipAuthns, fmt.Sprintf("get:/api/v1/oauth2/%s/callback", strings.ToLower(oauth2.Field(key).Name)))
		skipAuthns = append(skipAuthns, fmt.Sprintf("get:/api/v1/oauth2/%s/redirect_callback", strings.ToLower(oauth2.Field(key).Name)))
	}
	return func(c echo.Context) bool {
		requestUri := c.Request().RequestURI
		requestMethod := c.Request().Method
		if !(strings.HasPrefix(requestUri, "/v2/") || strings.HasPrefix(requestUri, consts.APIV1)) { // nolint: staticcheck
			return true
		}
		return slices.Contains(skipAuthns, strings.ToLower(fmt.Sprintf("%s:%s", requestMethod, requestUri)))
	}
}

var skipAuthzs = []string{"post:/api/v1/users/login", "get:/api/v1/users/token", "get:/api/v1/users/signup", "get:/api/v1/users/create"}

func genAuthzSkipper() middleware.Skipper {
	var oauth2 = reflect.TypeOf(configs.ConfigurationAuthOauth2{})
	for key := range oauth2.NumField() {
		skipAuthzs = append(skipAuthzs, fmt.Sprintf("get:/api/v1/oauth2/%s/client_id", strings.ToLower(oauth2.Field(key).Name)))
		skipAuthzs = append(skipAuthzs, fmt.Sprintf("get:/api/v1/oauth2/%s/callback", strings.ToLower(oauth2.Field(key).Name)))
		skipAuthzs = append(skipAuthzs, fmt.Sprintf("get:/api/v1/oauth2/%s/redirect_callback", strings.ToLower(oauth2.Field(key).Name)))
	}
	return func(c echo.Context) bool {
		requestUri := c.Request().RequestURI
		requestMethod := c.Request().Method
		if !(strings.HasPrefix(requestUri, "/v2/") || strings.HasPrefix(requestUri, consts.APIV1)) { // nolint: staticcheck
			return true
		}
		return slices.Contains(skipAuthzs, strings.ToLower(fmt.Sprintf("%s:%s", requestMethod, requestUri)))
	}
}
