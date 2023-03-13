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
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"gorm.io/gorm"

	"github.com/ximager/ximager/pkg/consts"
	"github.com/ximager/ximager/pkg/dal/dao"
	"github.com/ximager/ximager/pkg/utils/token"
	"github.com/ximager/ximager/pkg/xerrors"
)

// AuthDSConfig is the configuration for the Auth middleware.
type AuthDSConfig struct {
	Skipper middleware.Skipper
}

// AuthDSWithConfig returns a middleware which authenticates requests.
func AuthDSWithConfig(config AuthDSConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper != nil && config.Skipper(c) {
				log.Trace().Msg("Skipping auth middleware, allowing request")
				return next(c)
			}

			tokenService, err := token.NewTokenService(viper.GetString("auth.jwt.privateKey"), viper.GetString("auth.jwt.publicKey"))
			if err != nil {
				log.Error().Err(err).Msg("Create token service failed")
				return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, err.Error())
			}

			ctx := c.Request().Context()
			authorization := c.Request().Header.Get("Authorization")

			jti, username, err := tokenService.Validate(ctx, strings.TrimSpace(strings.TrimPrefix(authorization, "Bearer")))
			if err != nil {
				log.Error().Err(err).Msg("Validate token failed")
				c.Response().Header().Set("WWW-Authenticate", "Bearer realm=\"http://10.82.47.25:3000/user/token\"")
				return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeUnauthorized, err.Error())
			}

			userService := dao.NewUserService()
			user, err := userService.GetByUsername(ctx, username)
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					log.Error().Err(err).Msg("User not found")
					return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeUnauthorized, err.Error())
				}
				log.Error().Err(err).Msg("Get user failed")
				return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, err.Error())
			}

			c.Set(consts.ContextUser, user)
			c.Set(consts.ContextJti, jti)

			return next(c)
		}
	}
}
