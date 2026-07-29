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
	"net/http"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/server/errcode"
)

func TestContextValues(t *testing.T) {
	ctx := context.Background()
	_, ok := userFromContext(ctx)
	require.False(t, ok)

	user := &models.User{ID: "user-1"}
	ctx = withUser(ctx, user)
	got, ok := userFromContext(ctx)
	require.True(t, ok)
	require.Same(t, user, got)

	ctx = withClientInfo(ctx, "127.0.0.1", "sigma-test")
	clientIP, userAgent := clientInfoFromContext(ctx)
	require.Equal(t, "127.0.0.1", clientIP)
	require.Equal(t, "sigma-test", userAgent)
}

func TestToolErrorMapping(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantText   string
		wantStatus int
	}{
		{name: "nil", wantText: "", wantStatus: http.StatusOK},
		{name: "unauthorized", err: errUnauthorized, wantText: "unauthorized", wantStatus: http.StatusUnauthorized},
		{name: "forbidden", err: errForbidden, wantText: "permission denied", wantStatus: http.StatusUnauthorized},
		{
			name:       "typed error",
			err:        errcode.HTTPErrCodeBadRequest.Detail("invalid input"),
			wantText:   "invalid input",
			wantStatus: http.StatusBadRequest,
		},
		{name: "internal", err: errors.New("sensitive detail"), wantText: "internal error", wantStatus: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.wantText, toolErrorMessage(tt.err))
			require.Equal(t, tt.wantStatus, statusCodeFromError(tt.err))
		})
	}
}

func TestResponseHelpers(t *testing.T) {
	page := 2
	limit := 20
	pagination := api.Pagination{Page: &page, Limit: &limit}

	response := listResponse([]string{"one"}, 1, pagination)
	require.Equal(t, []string{"one"}, response["items"])
	require.Equal(t, int64(1), response["total"])
	require.Equal(t, pagination, response["pagination"])
	require.Equal(t, map[string]bool{"ok": true}, okResponse())
}

func TestToolArgumentHelpers(t *testing.T) {
	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{"name": "sigma", "count": float64(3)},
		},
	}
	type arguments struct {
		Name  string `json:"name"`
		Count int    `json:"count"`
	}
	got, err := bindArguments[arguments](req)
	require.NoError(t, err)
	require.Equal(t, arguments{Name: "sigma", Count: 3}, got)

	invalid := mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: map[string]any{"count": "invalid"}}}
	_, err = bindArguments[arguments](invalid)
	require.Error(t, err)

	require.Equal(t, "second", stringArgument(map[string]any{
		"first":  123,
		"second": "second",
	}, "first", "second"))
	require.Empty(t, stringArgument(nil, "missing"))
}

func TestNormalizeKey(t *testing.T) {
	require.Equal(t, "authorization", normalizeKey("Authorization"))
	require.Equal(t, "privatekey", normalizeKey("private_key"))
	require.Equal(t, "scmtoken", normalizeKey("scm-token"))
}
