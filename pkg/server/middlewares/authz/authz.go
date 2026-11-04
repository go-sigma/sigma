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
	"log/slog"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/authz"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/logger"
	"github.com/go-sigma/sigma/pkg/server/errcode"
	"github.com/go-sigma/sigma/pkg/utils"
)

// Skipper defines a function to skip middleware.
type Skipper func(c *gin.Context) bool

// Config defines the config for the authorization middleware.
type Config struct {
	// Skipper defines a function to skip middleware
	Skipper Skipper
	// Authorizer checks whether the user can access a resource.
	Authorizer authz.Authorizer
}

// selfScopedPaths are the API paths any signed in user may call: they act on the caller's own
// account, validate input or expose read-only metadata.
var selfScopedPaths = []string{
	"/api/v1/overview",
	"/api/v1/activity/daily",
	"/api/v1/builds/daily",
	"/api/v1/users/logout",
}

// selfScopedPrefixes cover the routes below a self scoped path.
var selfScopedPrefixes = []string{
	"/api/v1/users/self",
	"/api/v1/validators/",
	"/api/v1/systems/",
}

// isSelfScoped reports whether the path is reachable without a namespace permission check.
// The handlers of these routes still authorize the caller themselves when it matters.
func isSelfScoped(path string) bool {
	if slices.Contains(selfScopedPaths, path) {
		return true
	}
	for _, prefix := range selfScopedPrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	// The activity heatmap authorizes the profile owner (or an administrator) in its handler.
	return strings.HasPrefix(path, "/api/v1/users/") && strings.HasSuffix(path, "/activity/heatmap")
}

// AuthzWithConfig returns an authorization middleware with config.
func AuthzWithConfig(config Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		if config.Skipper != nil && config.Skipper(c) {
			slog.Debug("skipping auth middleware, allowing request")
			c.Next()
			return
		}

		requester := c.Request
		requestUri := strings.TrimSpace(requester.RequestURI)
		requestMethod := requester.Method

		var isDistribution bool
		if strings.HasPrefix(requestUri, "/v2") {
			isDistribution = true
		}

		user, ok := utils.GetFromCtx[*models.User](c, consts.ContextUser)
		if !ok {
			slog.Error("get user from header failed")
			errcode.NewHTTPError(c, errcode.HTTPErrCodeUnauthorized)
			c.Abort()
			return
		}
		// admin or root can access all resources
		if user.Role == enums.UserRoleAdmin || user.Role == enums.UserRoleRoot {
			slog.Log(c.Request.Context(), logger.LevelTrace, "skipping auth middleware, allowing admin and root can access all of resources")
			c.Next()
			return
		}

		// /api/v1/namespaces/ (list endpoint) is allowed for any authenticated
		// user; the handler applies its own visibility-based filtering.
		// Compare the path, not the request URI: the list always carries pagination parameters.
		if isSelfScoped(requester.URL.Path) {
			c.Next()
			return
		}
		if requester.URL.Path == fmt.Sprintf("%s/namespaces/", consts.APIV1) {
			c.Next()
			return
		}

		// Only namespace-scoped /api/v1 and /v2 paths are authorized here.
		// Other /api/v1 paths are denied; non-/api, non-/v2 paths are allowed.
		isNamespacedAPI := strings.HasPrefix(requestUri, fmt.Sprintf("%s/namespaces/", consts.APIV1))
		if !isDistribution && !isNamespacedAPI {
			if strings.HasPrefix(requestUri, consts.APIV1) {
				errcode.NewHTTPError(c, errcode.HTTPErrCodeUnauthorized, "Authorization failed")
				c.Abort()
				return
			}
			c.Next()
			return
		}

		isAnonymous := user.Role == enums.UserRoleAnonymous
		passed, err := config.Authorizer.Authorize(c.Request.Context(), user.ID, isAnonymous, requestUri, requestMethod)
		if err != nil {
			slog.Error("authorize failed", "err", err)
			if isDistribution {
				errcode.NewDSError(c, errcode.DSErrCodeUnknown)
			} else {
				errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("authorize failed: %v", err)))
			}
			c.Abort()
			return
		}
		if !passed {
			if isDistribution {
				errcode.NewDSError(c, errcode.DSErrCodeUnauthorized)
			} else {
				errcode.NewHTTPError(c, errcode.HTTPErrCodeUnauthorized, "Authorization failed")
			}
			c.Abort()
			return
		}
		c.Next()
	}
}
