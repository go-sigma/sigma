// Copyright 2026 sigma
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

// Package audit provides HTTP request audit middleware.
package audit

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	repoaudit "github.com/go-sigma/sigma/pkg/dal/repository/audit"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/uuid"
)

// Skipper defines a function to skip middleware.
type Skipper func(c *gin.Context) bool

// Config defines audit middleware configuration.
type Config struct {
	Skipper         Skipper
	RecordListGet   bool
	AuditRepository repoaudit.AuditRepository
}

var listGetRoutes = map[string]struct{}{
	consts.APIV1 + "/namespaces/":                                                                        {},
	consts.APIV1 + "/namespaces/:namespace_id/members/":                                                  {},
	consts.APIV1 + "/namespaces/:namespace_id/repositories/":                                             {},
	consts.APIV1 + "/namespaces/:namespace_id/repositories/:repository_id/tags/":                         {},
	consts.APIV1 + "/namespaces/:namespace_id/repositories/:repository_id/artifacts/":                    {},
	consts.APIV1 + "/webhooks/":                                                                          {},
	consts.APIV1 + "/webhooks/:webhook_id/logs/":                                                         {},
	consts.APIV1 + "/users/":                                                                             {},
	consts.APIV1 + "/namespaces/:namespace_id/repositories/:repository_id/builders/:builder_id/runners/": {},
	consts.APIV1 + "/daemons/gc-repository/:namespace_id/runners/":                                       {},
	consts.APIV1 + "/daemons/gc-repository/:namespace_id/runners/:runner_id/records/":                    {},
	consts.APIV1 + "/daemons/gc-tag/:namespace_id/runners/":                                              {},
	consts.APIV1 + "/daemons/gc-tag/:namespace_id/runners/:runner_id/records/":                           {},
	consts.APIV1 + "/daemons/gc-artifact/:namespace_id/runners/":                                         {},
	consts.APIV1 + "/daemons/gc-artifact/:namespace_id/runners/:runner_id/records/":                      {},
	consts.APIV1 + "/daemons/gc-blob/:namespace_id/runners/":                                             {},
	consts.APIV1 + "/daemons/gc-blob/:namespace_id/runners/:runner_id/records/":                          {},
	consts.APIV1 + "/coderepos/:provider":                                                                {},
	consts.APIV1 + "/coderepos/:provider/owners":                                                         {},
	consts.APIV1 + "/coderepos/:provider/repos/:id/branches":                                             {},
}

// AuditWithConfig returns an HTTP audit middleware.
func AuditWithConfig(config Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		if shouldSkip(c, config) {
			c.Next()
			return
		}

		start := time.Now()
		c.Next()

		user, ok := utils.GetFromCtx[*models.User](c, consts.ContextUser)
		if !ok || user == nil {
			return
		}
		if config.AuditRepository == nil {
			slog.Error("audit repository is nil")
			return
		}

		route := c.FullPath()
		if route == "" {
			route = c.Request.URL.Path
		}
		namespaceID := namespaceIDFromContext(c)
		auditObj := &models.Audit{
			ID:          uuid.NewV7String(),
			UserID:      user.ID,
			Username:    user.Username,
			UserRole:    user.Role.String(),
			NamespaceID: namespaceID,
			Method:      c.Request.Method,
			Path:        c.Request.URL.Path,
			Route:       route,
			Query:       c.Request.URL.RawQuery,
			StatusCode:  c.Writer.Status(),
			ClientIP:    c.ClientIP(),
			UserAgent:   c.Request.UserAgent(),
			LatencyMs:   time.Since(start).Milliseconds(),
		}
		if err := config.AuditRepository.Create(c.Request.Context(), auditObj); err != nil {
			slog.Error("create audit failed", "err", err)
		}
	}
}

func shouldSkip(c *gin.Context, config Config) bool {
	if config.Skipper != nil && config.Skipper(c) {
		return true
	}
	path := c.Request.URL.Path
	if isProbeOrNonBusinessPath(path) {
		return true
	}
	if c.Request.Method == http.MethodGet && !config.RecordListGet && isListGet(c) {
		return true
	}
	return false
}

func isProbeOrNonBusinessPath(path string) bool {
	switch {
	case path == "/healthz", path == "/readyz", path == "/metrics":
		return true
	case strings.HasPrefix(path, "/debug/pprof"), strings.HasPrefix(path, consts.PprofPath):
		return true
	case strings.HasPrefix(path, consts.APIV1), strings.HasPrefix(path, "/v2/"), path == "/v2":
		return false
	default:
		return true
	}
}

func isListGet(c *gin.Context) bool {
	path := c.Request.URL.Path
	if path == "/v2/_catalog" {
		return true
	}
	if strings.HasPrefix(path, "/v2/") && strings.HasSuffix(path, "/tags/list") {
		return true
	}
	_, ok := listGetRoutes[c.FullPath()]
	return ok
}

func namespaceIDFromContext(c *gin.Context) *string {
	namespaceID := c.Param("namespace_id")
	if namespaceID == "" {
		return nil
	}
	return &namespaceID
}
