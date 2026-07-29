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
	"errors"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/server/errcode"
	svcsystems "github.com/go-sigma/sigma/pkg/service/systems"
)

func TestWrapTool(t *testing.T) {
	server := &Server{config: &config.Configuration{MCP: config.ConfigurationMCP{ToolTimeout: time.Second}}}
	request := mcp.CallToolRequest{}

	t.Run("missing user", func(t *testing.T) {
		result, err := server.wrapTool("test", false, func(context.Context, mcp.CallToolRequest) (any, error) {
			t.Fatal("handler should not be called")
			return nil, nil
		})(t.Context(), request)
		require.NoError(t, err)
		require.True(t, result.IsError)
		require.Equal(t, "unauthorized", toolResultText(t, result))
	})

	t.Run("success", func(t *testing.T) {
		ctx := withUser(t.Context(), &models.User{ID: "user-1"})
		result, err := server.wrapTool("test", false, func(context.Context, mcp.CallToolRequest) (any, error) {
			return map[string]string{"status": "ok"}, nil
		})(ctx, request)
		require.NoError(t, err)
		require.False(t, result.IsError)
		require.Equal(t, map[string]any{"data": map[string]string{"status": "ok"}}, result.StructuredContent)
	})

	t.Run("handler error", func(t *testing.T) {
		ctx := withUser(t.Context(), &models.User{ID: "user-1"})
		result, err := server.wrapTool("test", false, func(context.Context, mcp.CallToolRequest) (any, error) {
			return nil, errcode.HTTPErrCodeBadRequest.Detail("invalid input")
		})(ctx, request)
		require.NoError(t, err)
		require.True(t, result.IsError)
		require.Equal(t, "invalid input", toolResultText(t, result))
	})
}

func TestResources(t *testing.T) {
	resources, err := jsonResource("sigma://test", map[string]string{"status": "ok"})
	require.NoError(t, err)
	require.Len(t, resources, 1)
	content, ok := resources[0].(mcp.TextResourceContents)
	require.True(t, ok)
	require.Equal(t, "sigma://test", content.URI)
	require.Equal(t, "application/json", content.MIMEType)
	require.JSONEq(t, `{"status":"ok"}`, content.Text)

	server := &Server{}
	resources, err = server.capabilitiesResource(t.Context(), mcp.ReadResourceRequest{})
	require.NoError(t, err)
	content, ok = resources[0].(mcp.TextResourceContents)
	require.True(t, ok)
	require.Contains(t, content.Text, `"namespace-list"`)
	require.Contains(t, content.Text, `"artifact-delete"`)
	require.Contains(t, content.Text, `"type":"basic"`)
}

func TestNewHTTPHandler(t *testing.T) {
	server := &Server{
		config: &config.Configuration{
			MCP: config.ConfigurationMCP{ToolTimeout: time.Second},
		},
	}

	require.NotNil(t, server.newHTTPHandler())
}

func TestSystemTools(t *testing.T) {
	ctrl := gomock.NewController(t)
	service := svcsystems.NewMockSystemsService(ctrl)
	service.EXPECT().GetVersion(gomock.Any()).Return(api.GetSystemVersionResponse{Version: "1.0.0"}, nil)
	service.EXPECT().GetEndpoint(gomock.Any()).Return("https://sigma.example.test", nil)
	server := &Server{
		systemsSvc: service,
		config: &config.Configuration{
			Database: config.ConfigurationDatabase{Type: enums.DatabasePostgresql},
			Storage:  config.ConfigurationStorage{Type: enums.StorageTypeS3, Redirect: true},
			Proxy:    config.ConfigurationProxy{Enabled: true},
			MCP: config.ConfigurationMCP{
				Enabled:        true,
				Path:           "/api/v1/mcp",
				Transport:      "streamable_http",
				ToolTimeout:    time.Second,
				MaxRequestBody: 1024,
			},
		},
	}

	version, err := server.systemVersionGet(t.Context(), mcp.CallToolRequest{})
	require.NoError(t, err)
	require.Equal(t, api.GetSystemVersionResponse{Version: "1.0.0"}, version)

	endpoint, err := server.systemEndpointGet(t.Context(), mcp.CallToolRequest{})
	require.NoError(t, err)
	require.Equal(t, map[string]string{"endpoint": "https://sigma.example.test"}, endpoint)

	summary, err := server.systemConfigSummaryGet(t.Context(), mcp.CallToolRequest{})
	require.NoError(t, err)
	summaryMap, ok := summary.(map[string]any)
	require.True(t, ok)
	require.Contains(t, summaryMap, "database")
	require.Contains(t, summaryMap, "storage")
	require.Contains(t, summaryMap, "mcp")
}

func TestSystemEndpointGetError(t *testing.T) {
	ctrl := gomock.NewController(t)
	service := svcsystems.NewMockSystemsService(ctrl)
	expectedErr := errors.New("failed")
	service.EXPECT().GetEndpoint(gomock.Any()).Return("", expectedErr)

	_, err := (&Server{systemsSvc: service}).systemEndpointGet(t.Context(), mcp.CallToolRequest{})
	require.ErrorIs(t, err, expectedErr)
}

func toolResultText(t *testing.T, result *mcp.CallToolResult) string {
	t.Helper()
	require.NotEmpty(t, result.Content)
	content, ok := mcp.AsTextContent(result.Content[0])
	require.True(t, ok)
	return content.Text
}
