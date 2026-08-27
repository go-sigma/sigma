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
	"log/slog"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	mcpsdk "github.com/mark3labs/mcp-go/server"

	"github.com/go-sigma/sigma/pkg/server/errcode"
)

type toolFunc func(context.Context, mcp.CallToolRequest) (any, error)

func (s *Server) addTool(mcpServer *mcpsdk.MCPServer, name, description string, write bool, handler toolFunc) {
	tool := mcp.NewTool(name, mcp.WithDescription(description), mcp.WithSchemaAdditionalProperties(true))
	tool.Annotations = mcp.ToolAnnotation{
		ReadOnlyHint:    mcp.ToBoolPtr(!write),
		DestructiveHint: mcp.ToBoolPtr(write),
		OpenWorldHint:   mcp.ToBoolPtr(false),
	}
	mcpServer.AddTool(tool, s.wrapTool(name, write, handler))
}

func (s *Server) wrapTool(name string, write bool, handler toolFunc) mcpsdk.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		startedAt := time.Now()
		user, ok := userFromContext(ctx)
		if !ok {
			recordToolCall(name, toolResultError, startedAt)
			return mcp.NewToolResultError(errUnauthorized.Error()), nil
		}

		ctx, cancel := context.WithTimeout(ctx, s.config.MCP.ToolTimeout)
		defer cancel()

		data, err := handler(ctx, req)
		statusCode := statusCodeFromError(err)
		if write && s.config.MCP.AuditWritesEnabled() {
			auditToolCall(ctx, s.auditRepository, s.config.MCP.Path, user, name, req, startedAt, statusCode)
		}
		if err != nil {
			slog.Error("mcp tool failed", "tool", name, "user_id", user.ID, "err", err)
			recordToolCall(name, toolResultError, startedAt)
			return mcp.NewToolResultError(toolErrorMessage(err)), nil
		}

		recordToolCall(name, toolResultOK, startedAt)
		return mcp.NewToolResultStructuredOnly(map[string]any{"data": data}), nil
	}
}

func bindArguments[T any](req mcp.CallToolRequest) (T, error) {
	var target T
	if e := req.BindArguments(&target); e != nil {
		return target, errcode.HTTPErrCodeBadRequest.Detail(e.Error())
	}
	return target, nil
}

func stringArgument(args map[string]any, names ...string) string {
	for _, name := range names {
		if val, ok := args[name].(string); ok {
			return val
		}
	}
	return ""
}
