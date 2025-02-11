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
	"fmt"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog/log"
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/dal"
	"github.com/go-sigma/sigma/pkg/xerrors"
)

// Config defines the config for CasbinAuth middleware
type Config struct {
	// Skipper defines a function to skip middleware
	Skipper middleware.Skipper
	// DigCon is the dig container
	DigCon *dig.Container
}

// AuthzWithConfig returns a CasbinAuth middleware with config
func AuthzWithConfig(config Config) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				log.Trace().Msg("Skipping auth middleware, allowing request")
				return next(c)
			}

			requester := c.Request()
			requestUri := requester.RequestURI
			requestMethod := requester.Method

			var isDistribution bool
			if strings.HasPrefix(requestUri, "/v2") {
				isDistribution = true
			}

			passed, matched, err := dal.AuthEnforcer.Enforcer.EnforceEx("userid", "namespace", "", "public", requestMethod)
			if err != nil {
				if isDistribution {
					return xerrors.NewDSError(c, xerrors.DSErrCodeUnknown)
				}
				return xerrors.HTTPErrCodeInternalError.Detail(fmt.Sprintf("get scope from database failed: %v", err))
			}
			log.Debug().Strs("matched", matched).Bool("result", passed).Msg("matched")
			if !passed {
				if isDistribution {
					return xerrors.NewDSError(c, xerrors.DSErrCodeUnauthorized)
				}
				return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeUnauthorized, "Authorization failed")
			}
			return next(c)
		}
	}
}
