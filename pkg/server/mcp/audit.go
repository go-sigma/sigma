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

package mcpserver

import (
	"context"
	"encoding/json/v2"
	"log/slog"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/go-sigma/sigma/pkg/dal/models"
	repoaudit "github.com/go-sigma/sigma/pkg/dal/repository/audit"
	"github.com/go-sigma/sigma/pkg/utils/uuid"
)

var sensitiveKeys = map[string]struct{}{
	"ak":            {},
	"authorization": {},
	"password":      {},
	"privatekey":    {},
	"scmpassword":   {},
	"scmsshkey":     {},
	"scmtoken":      {},
	"secret":        {},
	"sk":            {},
	"token":         {},
}

func auditToolCall(
	ctx context.Context,
	repo repoaudit.AuditRepository,
	path string,
	user *models.User,
	toolName string,
	request mcp.CallToolRequest,
	startedAt time.Time,
	statusCode int,
) {

	if repo == nil || user == nil {
		return
	}
	clientIP, userAgent := clientInfoFromContext(ctx)
	args := request.GetArguments()
	query, err := json.Marshal(redactSensitive(args))
	if err != nil {
		slog.Error("marshal mcp audit arguments failed", "err", err, "tool", toolName)
		query = []byte("{}")
	}
	auditObj := &models.Audit{
		ID:         uuid.NewV7String(),
		UserID:     user.ID,
		Username:   user.Username,
		UserRole:   user.Role.String(),
		Method:     "MCP",
		Path:       path,
		Route:      toolName,
		Query:      string(query),
		StatusCode: statusCode,
		ClientIP:   clientIP,
		UserAgent:  userAgent,
		LatencyMs:  time.Since(startedAt).Milliseconds(),
	}
	if namespaceID := stringArgument(args, "namespace_id", "namespaceID"); namespaceID != "" {
		auditObj.NamespaceID = &namespaceID
	}
	if e := repo.Create(ctx, auditObj); e != nil {
		slog.Error("create mcp audit failed", "err", e, "tool", toolName)
	}
}

func redactSensitive(value any) any {
	switch v := value.(type) {
	case map[string]any:
		result := make(map[string]any, len(v))
		for key, item := range v {
			if _, ok := sensitiveKeys[normalizeKey(key)]; ok {
				result[key] = "[REDACTED]"
				continue
			}
			result[key] = redactSensitive(item)
		}
		return result
	case []any:
		result := make([]any, 0, len(v))
		for _, item := range v {
			result = append(result, redactSensitive(item))
		}
		return result
	default:
		return value
	}
}

func normalizeKey(key string) string {
	replacer := strings.NewReplacer("_", "", "-", "", ".", "")
	return strings.ToLower(replacer.Replace(key))
}
