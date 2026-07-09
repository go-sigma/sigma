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

package healthz

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/go-sigma/sigma/pkg/server/errcode"
)

// Resource something that need be health checked
type Resource interface {
	HealthCheck() error // returns error if health check no passed
}

// Healthz creates health and readiness check middleware.
func Healthz(rs ...Resource) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/healthz" && c.Request.Method == http.MethodGet {
			c.String(http.StatusOK, "OK")
			c.Abort()
			return
		}
		if c.Request.URL.Path == "/readyz" && c.Request.Method == http.MethodGet {
			for _, r := range rs {
				if err := r.HealthCheck(); err != nil {
					errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError, err.Error())
					c.Abort()
					return
				}
			}
			c.String(http.StatusOK, "OK")
			c.Abort()
			return
		}
		c.Next()
	}
}
