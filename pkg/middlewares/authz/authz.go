// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: © 2017 LabStack and Echo contributors

// Copyright 2024 sigma
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

package authz

import (
	"strconv"
	"strings"

	"github.com/casbin/casbin/v2"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog/log"

	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/xerrors"
)

type (
	// Config defines the config for CasbinAuth middleware
	Config struct {
		// Skipper defines a function to skip middleware
		Skipper middleware.Skipper
		// Enforcer CasbinAuth main rule
		Enforcer *casbin.SyncedEnforcer
	}
)

// AuthzWithConfig returns a CasbinAuth middleware with config
func AuthzWithConfig(config Config) echo.MiddlewareFunc {
	if config.Enforcer == nil {
		panic("casbin middleware Enforcer field must be set")
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				log.Trace().Msg("Skipping auth middleware, allowing request")
				return next(c)
			}

			var request = c.Request()

			var isDistribution bool
			requestUri := request.RequestURI
			if strings.HasPrefix(requestUri, "/v2") {
				isDistribution = true
			}

			iUser := c.Get(consts.ContextUser)
			if iUser == nil {
				log.Error().Msg("Get user from header failed")
				if isDistribution {
					return xerrors.NewDSError(c, xerrors.DSErrCodeUnauthorized)
				}
				return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeUnauthorized, "Authorization failed")
			}
			user, ok := iUser.(*models.User)
			if !ok {
				log.Error().Msg("Convert user from header failed")
				if isDistribution {
					return xerrors.NewDSError(c, xerrors.DSErrCodeUnauthorized)
				}
				return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeUnauthorized, "Authorization failed")
			}

			pass, err := config.Enforcer.Enforce(strconv.FormatInt(user.ID, 10), "library", "/v2/library/busybox/manifests/latest", "public", strings.ToUpper(request.Method))
			if err != nil {
				if isDistribution {
					return xerrors.NewDSError(c, xerrors.DSErrCodeUnknown)
				}
				return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, "Internal server error")
			}
			if !pass {
				if isDistribution {
					return xerrors.NewDSError(c, xerrors.DSErrCodeUnauthorized)
				}
				return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeUnauthorized, "Authorization failed")
			}
			return next(c)
		}
	}
}
