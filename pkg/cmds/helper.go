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
			if c.Request().URL.Path == "/healthz" ||
				c.Request().URL.Path == "/metrics" {
				log.Trace().
					Str("method", c.Request().Method).
					Str("path", c.Request().URL.Path).
					Str("query", c.Request().URL.RawQuery).
					Msg("Request debugger")
			} else {
				log.Debug().
					Str("method", c.Request().Method).
					Str("path", c.Request().URL.Path).
					Str("query", c.Request().URL.RawQuery).
					Msg("Request debugger")
			}
			reqPath := c.Request().URL.Path
			if strings.HasPrefix(reqPath, "/assets/") {
				if strings.HasSuffix(c.Request().URL.Path, ".js") || // TODO: test content type
					strings.HasSuffix(c.Request().URL.Path, ".map") ||
					strings.HasSuffix(c.Request().URL.Path, ".css") ||
					strings.HasSuffix(c.Request().URL.Path, ".svg") ||
					strings.HasSuffix(c.Request().URL.Path, ".png") ||
					strings.HasSuffix(c.Request().URL.Path, ".ttf") ||
					strings.HasSuffix(c.Request().URL.Path, ".json") ||
					strings.HasSuffix(c.Request().URL.Path, ".yaml") {
					c.Response().Header().Add("Cache-Control", "max-age=3600")
				}
			}
			n := next(c)
			return n
		}
	}))
	e.Use(middleware.CORS())
	e.Use(middlewares.WithEtagConfig(middlewares.EtagConfig{
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
		HashFn: func(config middlewares.EtagConfig) hash.Hash {
			if config.Weak {
				return crc32.New(crc32.MakeTable(0xD5828281))
			}
			return sha256.New()
		},
	}))
	e.Use(echoprometheus.NewMiddleware(consts.AppName))
	e.GET("/metrics", echoprometheus.NewHandler())
	e.Use(middlewares.Healthz())
	e.JSONSerializer = new(serializer.DefaultJSONSerializer)
	config := utils.MustGetObjFromDigCon[configs.Configuration](digCon)
	if config.Log.Level == enums.LogLevelDebug || config.Log.Level == enums.LogLevelTrace {
		pprof.Register(e, consts.PprofPath)
	}
	e.Use(middlewares.RedirectRepository(config))
	e.Use(authn.AuthnWithConfig(authn.Config{
		Skipper: genSkipper(),
	}))
	e.Use(authz.AuthzWithConfig(authz.Config{
		Skipper: genSkipper(),
	}))
	return e, nil
}

var skipAuths = []string{"get:/api/v1/users/token", "get:/api/v1/users/signup", "get:/api/v1/users/create"}

func genSkipper() middleware.Skipper {
	var oauth2 = reflect.TypeOf(configs.ConfigurationAuthOauth2{})
	for key := range oauth2.NumField() {
		skipAuths = append(skipAuths, fmt.Sprintf("get:/api/v1/oauth2/%s/client_id", strings.ToLower(oauth2.Field(key).Name)))
		skipAuths = append(skipAuths, fmt.Sprintf("get:/api/v1/oauth2/%s/callback", strings.ToLower(oauth2.Field(key).Name)))
		skipAuths = append(skipAuths, fmt.Sprintf("get:/api/v1/oauth2/%s/redirect_callback", strings.ToLower(oauth2.Field(key).Name)))
	}
	return func(c echo.Context) bool {
		requestUri := c.Request().RequestURI
		requestMethod := c.Request().Method
		return slices.Contains(skipAuths, strings.ToLower(fmt.Sprintf("%s:%s", requestMethod, requestUri)))
	}
}
