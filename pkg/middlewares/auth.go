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

package middlewares

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"gorm.io/gorm"

	"github.com/ximager/ximager/pkg/consts"
	"github.com/ximager/ximager/pkg/dal/dao"
	"github.com/ximager/ximager/pkg/utils/password"
	"github.com/ximager/ximager/pkg/utils/ptr"
	"github.com/ximager/ximager/pkg/utils/token"
	"github.com/ximager/ximager/pkg/xerrors"
)

// AuthConfig is the configuration for the Auth middleware.
type AuthConfig struct {
	// Skipper defines a function to skip middleware.
	Skipper middleware.Skipper
	// DS is distribution service or not.
	DS bool
}

// AuthWithConfig returns a middleware which authenticates requests.
func AuthWithConfig(config AuthConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper != nil && config.Skipper(c) {
				log.Trace().Msg("Skipping auth middleware, allowing request")
				return next(c)
			}

			privateKey := viper.GetString("auth.jwt.privateKey")
			tokenService, err := token.NewTokenService(privateKey)
			if err != nil {
				log.Error().Err(err).Msg("Create token service failed")
				if config.DS {
					return xerrors.NewDSError(c, xerrors.DSErrCodeUnknown)
				}
				return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, err.Error())
			}

			req := c.Request()
			ctx := log.Logger.WithContext(req.Context())
			authorization := req.Header.Get("Authorization")

			var username, jti string

			switch {
			case strings.HasPrefix(authorization, "Basic"):
				var pwd string
				var ok bool
				username, pwd, ok = c.Request().BasicAuth()
				if !ok {
					log.Error().Str("Authorization", c.Request().Header.Get("Authorization")).Msg("Basic auth failed")
					if config.DS {
						return xerrors.NewDSError(c, xerrors.DSErrCodeUnauthorized)
					}
					return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeUnauthorized, "Basic auth failed")
				}

				userServiceFactory := dao.NewUserServiceFactory()
				userService := userServiceFactory.New()
				user, err := userService.GetByUsername(ctx, username)
				if err != nil {
					log.Error().Err(err).Msg("Get user by username failed")
					c.Response().Header().Set("WWW-Authenticate", genWwwAuthenticate(req.Host, c.Scheme()))
					if config.DS {
						return xerrors.NewDSError(c, xerrors.DSErrCodeUnauthorized)
					}
					return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeUnauthorized, err.Error())
				}

				passwordService := password.New()
				verify := passwordService.Verify(pwd, ptr.To(user.Password))
				if !verify {
					log.Error().Err(err).Msg("Verify password failed")
					c.Response().Header().Set("WWW-Authenticate", genWwwAuthenticate(req.Host, c.Scheme()))
					if config.DS {
						return xerrors.NewDSError(c, xerrors.DSErrCodeUnauthorized)
					}
					return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeUnauthorized)
				}
				jti = uuid.New().String()
			case strings.HasPrefix(authorization, "Bearer"):
				jti, username, err = tokenService.Validate(ctx, strings.TrimSpace(strings.TrimPrefix(authorization, "Bearer")))
				if err != nil {
					log.Error().Err(err).Msg("Validate token failed")
					c.Response().Header().Set("WWW-Authenticate", genWwwAuthenticate(req.Host, c.Scheme()))
					if config.DS {
						return xerrors.NewDSError(c, xerrors.DSErrCodeUnauthorized)
					}
					return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeUnauthorized, err.Error())
				}
			default:
				if config.DS {
					c.Response().Header().Set("WWW-Authenticate", genWwwAuthenticate(req.Host, c.Scheme()))
					return xerrors.NewDSError(c, xerrors.DSErrCodeUnauthorized)
				}
				return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeUnauthorized)
			}

			userServiceFactory := dao.NewUserServiceFactory()
			userService := userServiceFactory.New()
			user, err := userService.GetByUsername(ctx, username)
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					log.Error().Err(err).Msg("User not found")
					if config.DS {
						return xerrors.NewDSError(c, xerrors.DSErrCodeUnknown)
					}
					return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeUnauthorized, err.Error())
				}
				log.Error().Err(err).Msg("Get user failed")
				if config.DS {
					return xerrors.NewDSError(c, xerrors.DSErrCodeUnknown)
				}
				return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, err.Error())
			}

			c.Set(consts.ContextUser, user)
			c.Set(consts.ContextJti, jti)

			return next(c)
		}
	}
}

func genWwwAuthenticate(host, schema string) string {
	realm := fmt.Sprintf("%s://%s%s/tokens", schema, host, consts.APIV1)
	rRealm := viper.GetString("auth.token.realm")
	if rRealm != "" {
		realm = rRealm
	}
	service := consts.AppName
	rService := viper.GetString("auth.token.service")
	if rService != "" {
		service = rService
	}
	return fmt.Sprintf("Bearer realm=\"%s\",service=\"%s\"", realm, service)
}
