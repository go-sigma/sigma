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

package authn

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog/log"
	"go.uber.org/dig"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/dao"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/password"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
	"github.com/go-sigma/sigma/pkg/utils/token"
	"github.com/go-sigma/sigma/pkg/xerrors"
)

// Config is the configuration for the Auth middleware
type Config struct {
	// Skipper defines a function to skip middleware
	Skipper middleware.Skipper
	// DigCon is the dig container
	DigCon *dig.Container
}

// AuthnConfig ...
type AuthnConfig struct {
	Skip bool
}

// AuthMapper ...
var AuthMapper = make(map[*regexp.Regexp]*AuthnConfig)

// AuthnWithConfig returns a middleware which authenticates requests.
func AuthnWithConfig(config Config) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper != nil && config.Skipper(c) {
				log.Trace().Msg("Skipping auth middleware, allowing request")
				return next(c)
			}

			requestUri := c.Request().RequestURI
			requestMethod := c.Request().Method

			var isDistribution bool
			if strings.HasPrefix(requestUri, "/v2/") {
				isDistribution = true
			}

			echo := utils.MustGetObjFromDigCon[*echo.Echo](config.DigCon)
			authConfig := authMatch(echo, requestMethod, requestUri)
			if authConfig == nil || authConfig.Skip {
				return next(c)
			}

			tokenService, err := token.New(config.DigCon)
			if err != nil {
				log.Error().Err(err).Msg("Create token service failed")
				if isDistribution {
					return xerrors.NewDSError(c, xerrors.DSErrCodeUnknown)
				}
				return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, err.Error())
			}

			req := c.Request()
			ctx := log.Logger.WithContext(req.Context())
			authorization := req.Header.Get("Authorization")

			var uid int64
			var jti = uuid.New().String()

			userService := utils.MustGetObjFromDigCon[dao.UserServiceFactory](config.DigCon).New()

			switch {
			case strings.HasPrefix(authorization, "Basic"):
				var username string
				var pwd string
				var ok bool
				username, pwd, ok = c.Request().BasicAuth()
				if !ok {
					log.Error().Str("Authorization", c.Request().Header.Get("Authorization")).Msg("Basic auth failed")
					if isDistribution {
						return xerrors.NewDSError(c, xerrors.DSErrCodeUnauthorized)
					}
					return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeUnauthorized, "Basic auth failed")
				}

				userServiceFactory := dao.NewUserServiceFactory()
				userService := userServiceFactory.New()
				user, err := userService.GetByUsername(ctx, username)
				if err != nil {
					log.Error().Err(err).Msg("Get user by username failed")
					c.Response().Header().Set("WWW-Authenticate", genWwwAuthenticate(config.DigCon, req.Host, c.Scheme()))
					if isDistribution {
						return xerrors.NewDSError(c, xerrors.DSErrCodeUnauthorized)
					}
					return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeUnauthorized, "Username or password is not correct")
				}
				uid = user.ID

				passwordService := password.New()
				verify := passwordService.Verify(pwd, ptr.To(user.Password))
				if !verify {
					log.Error().Err(err).Msg("Verify password failed")
					c.Response().Header().Set("WWW-Authenticate", genWwwAuthenticate(config.DigCon, req.Host, c.Scheme()))
					if isDistribution {
						return xerrors.NewDSError(c, xerrors.DSErrCodeUnauthorized)
					}
					return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeUnauthorized, "Username or password is not correct")
				}
			case strings.HasPrefix(authorization, "Bearer"):
				jti, uid, err = tokenService.Validate(ctx, strings.TrimSpace(strings.TrimPrefix(authorization, "Bearer")))
				if err != nil {
					log.Error().Err(err).Msg("Validate token failed")
					c.Response().Header().Set("WWW-Authenticate", genWwwAuthenticate(config.DigCon, req.Host, c.Scheme()))
					if isDistribution {
						return xerrors.NewDSError(c, xerrors.DSErrCodeUnauthorized)
					}
					return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeUnauthorized, err.Error())
				}
			default:
				uri := c.Request().URL.Path
				if strings.HasPrefix(uri, "/v2") || uri == "/api/v1/users/self" {
					c.Response().Header().Set("WWW-Authenticate", genWwwAuthenticate(config.DigCon, req.Host, c.Scheme()))
					if isDistribution {
						return xerrors.NewDSError(c, xerrors.DSErrCodeUnauthorized)
					}
					return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeUnauthorized)
				}
				userObj, err := userService.GetByUsername(ctx, consts.UserAnonymous)
				if err != nil {
					log.Error().Err(err).Msg("Get anonymous user failed")
					c.Response().Header().Set("WWW-Authenticate", genWwwAuthenticate(config.DigCon, req.Host, c.Scheme()))
					if isDistribution {
						return xerrors.NewDSError(c, xerrors.DSErrCodeUnauthorized)
					}
					return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeUnauthorized)
				}
				uid = userObj.ID
			}

			userObj, err := userService.Get(ctx, uid)
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					log.Error().Err(err).Msg("User not found")
					if isDistribution {
						return xerrors.NewDSError(c, xerrors.DSErrCodeUnknown)
					}
					return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeUnauthorized, err.Error())
				}
				log.Error().Err(err).Msg("Get user failed")
				if isDistribution {
					return xerrors.NewDSError(c, xerrors.DSErrCodeUnknown)
				}
				return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, err.Error())
			}

			c.Set(consts.ContextUser, userObj)
			c.Set(consts.ContextJti, jti)

			return next(c)
		}
	}
}

func genWwwAuthenticate(digCon *dig.Container, host, schema string) string {
	config := utils.MustGetObjFromDigCon[configs.Configuration](digCon)
	realm := fmt.Sprintf("%s://%s%s/tokens", schema, host, consts.APIV1)
	if config.Auth.Token.Realm != "" {
		realm = config.Auth.Token.Realm
	}
	service := consts.AppName
	if config.Auth.Token.Service != "" {
		service = config.Auth.Token.Service
	}
	return fmt.Sprintf("Bearer realm=\"%s\",service=\"%s\"", realm, service)
}

func authMatch(echo *echo.Echo, method, uri string) *AuthnConfig {
	if strings.HasPrefix(uri, "/api/v1/") {
		ctx := echo.AcquireContext()
		defer echo.ReleaseContext(ctx)
		echo.Router().Find(method, uri, ctx)
		matchedPath := ctx.Path()
		if matchedPath == "" {
			return nil
		}
		return authMapperMatcher(uri)
	} else if strings.HasPrefix(uri, "/v2/") {
		return authMapperMatcher(uri)
	}
	return nil
}

func authMapperMatcher(uri string) *AuthnConfig {
	for reg, config := range AuthMapper {
		if !reg.MatchString(uri) {
			continue
		}
		if config == nil {
			continue
		}
		return config
	}
	return nil
}
