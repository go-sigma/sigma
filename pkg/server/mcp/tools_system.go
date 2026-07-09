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

	"github.com/mark3labs/mcp-go/mcp"
	mcpsdk "github.com/mark3labs/mcp-go/server"
)

func (s *Server) registerSystemTools(mcpServer *mcpsdk.MCPServer) {
	s.addTool(mcpServer, "system-version-get", "Get Sigma build version information", false, s.systemVersionGet)
	s.addTool(mcpServer, "system-endpoint-get", "Get Sigma public HTTP endpoint", false, s.systemEndpointGet)
	s.addTool(mcpServer, "system-config-summary-get", "Get a redacted Sigma configuration summary", false, s.systemConfigSummaryGet)
}

func (s *Server) systemVersionGet(ctx context.Context, _ mcp.CallToolRequest) (any, error) {
	return s.systemsSvc.GetVersion(ctx)
}

func (s *Server) systemEndpointGet(ctx context.Context, _ mcp.CallToolRequest) (any, error) {
	endpoint, err := s.systemsSvc.GetEndpoint(ctx)
	if err != nil {
		return nil, err
	}
	return map[string]string{"endpoint": endpoint}, nil
}

func (s *Server) systemConfigSummaryGet(_ context.Context, _ mcp.CallToolRequest) (any, error) {
	cfg := s.config
	return map[string]any{
		"database": map[string]any{
			"type": cfg.Database.Type,
		},
		"storage": map[string]any{
			"type":     cfg.Storage.Type,
			"redirect": cfg.Storage.Redirect,
		},
		"proxy": map[string]any{
			"enabled": cfg.Proxy.Enabled,
		},
		"daemon": map[string]any{
			"builder": cfg.Daemon.Builder.Enabled,
		},
		"audit": map[string]any{
			"enabled":         cfg.Audit.IsEnabled(),
			"record_list_get": cfg.Audit.RecordListGet,
		},
		"mcp": map[string]any{
			"enabled":          cfg.MCP.Enabled,
			"path":             cfg.MCP.Path,
			"transport":        cfg.MCP.Transport,
			"audit_writes":     cfg.MCP.AuditWritesEnabled(),
			"tool_timeout":     cfg.MCP.ToolTimeout.String(),
			"max_request_body": cfg.MCP.MaxRequestBody,
		},
	}, nil
}
