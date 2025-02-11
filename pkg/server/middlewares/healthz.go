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

package middlewares

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/go-sigma/sigma/pkg/xerrors"
)

// Resource something that need be health checked
type Resource interface {
	HealthCheck() error // returns error if health check no passed
}

// Healthz create a health check middleware
func Healthz(rs ...Resource) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if c.Request().URL.Path == "/healthz" && c.Request().Method == http.MethodGet {
				for _, r := range rs {
					if err := r.HealthCheck(); err != nil {
						return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, err.Error())
					}
				}
				return c.String(http.StatusOK, "OK")
			}
			return next(c)
		}
	}
}
