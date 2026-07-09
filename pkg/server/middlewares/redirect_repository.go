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
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/go-sigma/sigma/pkg/config"
	reporegistry "github.com/go-sigma/sigma/pkg/dal/repository/registry"
)

// RedirectRepository redirect to frontend repository when request path is a docker pull path
// Note: namespace MUST be not 'api' or 'v2'
func RedirectRepository(config *config.Configuration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		reqPath := c.Request.URL.Path
		if !strings.Contains(strings.TrimPrefix(reqPath, "/"), "/") {
			c.Next()
			return
		}
		if !skipRedirect(c) {
			namespace, _, _ := strings.Cut(strings.TrimPrefix(reqPath, "/"), "/")
			repoName := strings.TrimPrefix(reqPath, "/")
			if strings.Contains(repoName, ":") {
				repoName = strings.SplitN(repoName, ":", 2)[0]
			}
			repositoryObj, err := reporegistry.NewRepositoryRepository().GetByName(ctx, repoName)
			if err != nil {
				slog.Error("get repository by name failed", "err", err, "repository", repoName)
				c.Next()
				return
			}
			c.Redirect(http.StatusTemporaryRedirect,
				fmt.Sprintf("%s/#/namespaces/%s/repository/tags?repository=%s&repository_id=%s", config.HTTP.Endpoint, namespace, repoName, repositoryObj.ID))
			c.Abort()
			return
		}
		c.Next()
	}
}

func skipRedirect(c *gin.Context) bool {
	if c.Request.Method != http.MethodGet {
		return true
	}
	reqPath := c.Request.URL.Path
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
