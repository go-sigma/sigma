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
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"

	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/dal/dao"
)

// RedirectRepository redirect to frontend repository when request path is a docker pull path
// Note: namespace MUST be not 'api' or 'v2'
func RedirectRepository(config configs.Configuration) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := log.Logger.WithContext(c.Request().Context())

			reqPath := c.Request().URL.Path
			if !strings.Contains(strings.TrimPrefix(reqPath, "/"), "/") {
				return next(c)
			}
			if !skipRedirect(c) {
				namespace := strings.SplitN(strings.TrimPrefix(reqPath, "/"), "/", 2)[0]
				repository := strings.TrimPrefix(reqPath, "/")
				if strings.Contains(repository, ":") {
					repository = strings.SplitN(repository, ":", 2)[0]
				}
				repositoryObj, err := dao.NewRepositoryServiceFactory().New().GetByName(ctx, repository)
				if err != nil {
					log.Error().Err(err).Str("repository", repository).Msg("Get repository by name failed")
					return next(c)
				}
				return c.Redirect(http.StatusTemporaryRedirect,
					fmt.Sprintf("%s/#/namespaces/%s/repository/tags?repository=%s&repository_id=%d", config.HTTP.Endpoint, namespace, repository, repositoryObj.ID))
			}
			return next(c)
		}
	}
}

func skipRedirect(c echo.Context) bool {
	if c.Request().Method != http.MethodGet {
		return true
	}
	reqPath := c.Request().URL.Path
	if strings.HasPrefix(reqPath, "/api/v1/") {
		return true
	}
	if strings.HasPrefix(reqPath, "/v2/") {
		return true
	}
	if strings.HasPrefix(reqPath, "/assets") && (strings.HasSuffix(reqPath, ".ttf") ||
		strings.HasSuffix(reqPath, ".css") ||
		strings.HasSuffix(reqPath, ".js")) {
		return true
	}
	if strings.HasPrefix(reqPath, "/swagger") {
		return true
	}
	if strings.HasPrefix(reqPath, "/distros") &&
		(strings.HasSuffix(reqPath, ".png") ||
			strings.HasSuffix(reqPath, ".jpg") ||
			strings.HasSuffix(reqPath, ".svg")) {
		return true
	}
	if strings.HasPrefix(reqPath, "/__debug") {
		return true
	}
	return false
}
