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
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"
	mcpsdk "github.com/mark3labs/mcp-go/server"
)

func (s *Server) registerResources(mcpServer *mcpsdk.MCPServer) {
	mcpServer.AddResource(
		mcp.NewResource(
			"sigma://system/version",
			"Sigma Version",
			mcp.WithMIMEType("application/json"),
			mcp.WithResourceDescription("Sigma build version information"),
		),
		s.systemVersionResource,
	)
	mcpServer.AddResource(
		mcp.NewResource(
			"sigma://system/capabilities",
			"Sigma MCP Capabilities",
			mcp.WithMIMEType("application/json"),
			mcp.WithResourceDescription("The MCP tools exposed by Sigma"),
		),
		s.capabilitiesResource,
	)
}

func (s *Server) systemVersionResource(ctx context.Context, _ mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	version, err := s.systemsSvc.GetVersion(ctx)
	if err != nil {
		return nil, err
	}
	return jsonResource("sigma://system/version", version)
}

func (s *Server) capabilitiesResource(_ context.Context, _ mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	capabilities := map[string]any{
		"tools": []string{
			"system-version-get",
			"system-endpoint-get",
			"system-config-summary-get",
			"namespace-list",
			"namespace-get",
			"namespace-create",
			"namespace-update",
			"namespace-delete",
			"namespace-member-list",
			"namespace-member-add",
			"namespace-member-update",
			"namespace-member-delete",
			"namespace-member-self-get",
			"repository-list",
			"repository-get",
			"repository-create",
			"repository-update",
			"repository-delete",
			"tag-list",
			"tag-get",
			"tag-delete",
			"tag-manifest-raw-get",
			"artifact-list",
			"artifact-get",
			"artifact-delete",
		},
		"auth": map[string]any{
			"type": "basic",
		},
	}
	return jsonResource("sigma://system/capabilities", capabilities)
}

func jsonResource(uri string, value any) ([]mcp.ResourceContents, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      uri,
			MIMEType: "application/json",
			Text:     string(data),
		},
	}, nil
}
