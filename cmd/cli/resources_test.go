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

package main

import (
	"encoding/json/v2"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	"github.com/go-sigma/sigma/pkg/api/enums"
)

func TestResourceParentCommands(t *testing.T) {
	opts := &options{}

	require.Contains(t, newNamespaceCmd(opts).Aliases, "ns")
	require.Contains(t, newRepositoryCmd(opts).Aliases, "repository")
	require.Len(t, newTagCmd(opts).Commands(), 3)
	require.Contains(t, newArtifactCmd(opts).Aliases, "image")
	require.Len(t, newMemberCmd(opts).Commands(), 4)
	require.Len(t, newUserCmd(opts).Commands(), 3)
}

func TestResourceCommands(t *testing.T) {
	tests := []struct {
		name        string
		command     func(*options) *cobra.Command
		args        []string
		method      string
		path        string
		query       url.Values
		body        map[string]any
		outputMatch string
	}{
		{
			name:    "namespace list",
			command: newNamespaceListCmd,
			args:    []string{"--name", "sig", "--page", "2", "--limit", "20", "--sort", "name", "--method", "asc"},
			method:  http.MethodGet,
			path:    "/api/v1/namespaces/",
			query:   url.Values{"name": []string{"sig"}, "page": []string{"2"}, "limit": []string{"20"}, "sort": []string{"name"}, "method": []string{"asc"}},
		},
		{
			name:    "namespace get",
			command: newNamespaceGetCmd,
			args:    []string{"ns-1"},
			method:  http.MethodGet,
			path:    "/api/v1/namespaces/ns-1",
		},
		{
			name:    "namespace create",
			command: newNamespaceCreateCmd,
			args:    []string{"ns", "--description", "desc", "--visibility", enums.VisibilityPrivate.String(), "--size-limit", "10", "--repository-limit", "11", "--tag-limit", "12"},
			method:  http.MethodPost,
			path:    "/api/v1/namespaces/",
			body: map[string]any{
				"name":             "ns",
				"description":      "desc",
				"visibility":       enums.VisibilityPrivate.String(),
				"size_limit":       float64(10),
				"repository_limit": float64(11),
				"tag_limit":        float64(12),
			},
		},
		{
			name:    "namespace update",
			command: newNamespaceUpdateCmd,
			args:    []string{"ns-1", "--description", "desc", "--overview", "overview", "--visibility", enums.VisibilityPublic.String(), "--size-limit", "20"},
			method:  http.MethodPut,
			path:    "/api/v1/namespaces/ns-1",
			body: map[string]any{
				"description": "desc",
				"overview":    "overview",
				"visibility":  enums.VisibilityPublic.String(),
				"size_limit":  float64(20),
			},
			outputMatch: `"status": "updated"`,
		},
		{
			name:        "namespace delete",
			command:     newNamespaceDeleteCmd,
			args:        []string{"ns-1"},
			method:      http.MethodDelete,
			path:        "/api/v1/namespaces/ns-1",
			outputMatch: `"status": "deleted"`,
		},
		{
			name:    "repository list",
			command: newRepositoryListCmd,
			args:    []string{"ns-1", "--name", "repo", "--page", "2", "--limit", "20"},
			method:  http.MethodGet,
			path:    "/api/v1/namespaces/ns-1/repositories/",
			query:   url.Values{"name": []string{"repo"}, "page": []string{"2"}, "limit": []string{"20"}},
		},
		{
			name:    "repository get",
			command: newRepositoryGetCmd,
			args:    []string{"ns-1", "repo-1"},
			method:  http.MethodGet,
			path:    "/api/v1/namespaces/ns-1/repositories/repo-1",
		},
		{
			name:    "repository create",
			command: newRepositoryCreateCmd,
			args:    []string{"ns-1", "repo", "--description", "desc", "--overview", "overview", "--visibility", enums.VisibilityPrivate.String(), "--size-limit", "10", "--tag-limit", "11"},
			method:  http.MethodPost,
			path:    "/api/v1/namespaces/ns-1/repositories/",
			body: map[string]any{
				"name":        "repo",
				"description": "desc",
				"overview":    "overview",
				"visibility":  enums.VisibilityPrivate.String(),
				"size_limit":  float64(10),
				"tag_limit":   float64(11),
			},
		},
		{
			name:    "repository update",
			command: newRepositoryUpdateCmd,
			args:    []string{"ns-1", "repo-1", "--description", "desc", "--overview", "overview", "--size-limit", "10", "--tag-limit", "11"},
			method:  http.MethodPut,
			path:    "/api/v1/namespaces/ns-1/repositories/repo-1",
			body: map[string]any{
				"description": "desc",
				"overview":    "overview",
				"size_limit":  float64(10),
				"tag_limit":   float64(11),
			},
			outputMatch: `"status": "updated"`,
		},
		{
			name:        "repository delete",
			command:     newRepositoryDeleteCmd,
			args:        []string{"ns-1", "repo-1"},
			method:      http.MethodDelete,
			path:        "/api/v1/namespaces/ns-1/repositories/repo-1",
			outputMatch: `"status": "deleted"`,
		},
		{
			name:    "tag list",
			command: newTagListCmd,
			args:    []string{"ns-1", "repo-1", "--name", "latest", "--type", enums.ArtifactTypeImage.String(), "--type", enums.ArtifactTypeChart.String()},
			method:  http.MethodGet,
			path:    "/api/v1/namespaces/ns-1/repositories/repo-1/tags/",
			query:   url.Values{"name": []string{"latest"}, "page": []string{"1"}, "limit": []string{"10"}, "type": []string{enums.ArtifactTypeImage.String(), enums.ArtifactTypeChart.String()}},
		},
		{
			name:    "tag get",
			command: newTagGetCmd,
			args:    []string{"ns-1", "repo-1", "tag-1"},
			method:  http.MethodGet,
			path:    "/api/v1/namespaces/ns-1/repositories/repo-1/tags/tag-1",
		},
		{
			name:        "tag delete",
			command:     newTagDeleteCmd,
			args:        []string{"ns-1", "repo-1", "tag-1"},
			method:      http.MethodDelete,
			path:        "/api/v1/namespaces/ns-1/repositories/repo-1/tags/tag-1",
			outputMatch: `"status": "deleted"`,
		},
		{
			name:    "artifact list",
			command: newArtifactListCmd,
			args:    []string{"ns", "repo", "--page", "2", "--limit", "20"},
			method:  http.MethodGet,
			path:    "/api/v1/namespaces/ns/artifacts/",
			query:   url.Values{"repository": []string{"repo"}, "page": []string{"2"}, "limit": []string{"20"}},
		},
		{
			name:    "artifact get",
			command: newArtifactGetCmd,
			args:    []string{"ns", "repo", "sha256:abc"},
			method:  http.MethodGet,
			path:    "/api/v1/namespaces/ns/artifacts/sha256:abc",
			query:   url.Values{"repository": []string{"repo"}},
		},
		{
			name:        "artifact delete",
			command:     newArtifactDeleteCmd,
			args:        []string{"ns", "repo", "sha256:abc"},
			method:      http.MethodDelete,
			path:        "/api/v1/namespaces/ns/artifacts/sha256:abc",
			query:       url.Values{"repository": []string{"repo"}},
			outputMatch: `"status": "deleted"`,
		},
		{
			name:    "member list",
			command: newMemberListCmd,
			args:    []string{"ns-1", "--name", "sigma"},
			method:  http.MethodGet,
			path:    "/api/v1/namespaces/ns-1/members/",
			query:   url.Values{"name": []string{"sigma"}, "page": []string{"1"}, "limit": []string{"10"}},
		},
		{
			name:    "member add",
			command: newMemberAddCmd,
			args:    []string{"ns-1", "42", "--role", enums.NamespaceRoleManager.String()},
			method:  http.MethodPost,
			path:    "/api/v1/namespaces/ns-1/members/",
			body: map[string]any{
				"user_id": float64(42),
				"role":    enums.NamespaceRoleManager.String(),
			},
		},
		{
			name:    "member update",
			command: newMemberUpdateCmd,
			args:    []string{"ns-1", "user-1", "--role", enums.NamespaceRoleReader.String()},
			method:  http.MethodPut,
			path:    "/api/v1/namespaces/ns-1/members/user-1",
			body: map[string]any{
				"role": enums.NamespaceRoleReader.String(),
			},
			outputMatch: `"status": "updated"`,
		},
		{
			name:        "member delete",
			command:     newMemberDeleteCmd,
			args:        []string{"ns-1", "user-1"},
			method:      http.MethodDelete,
			path:        "/api/v1/namespaces/ns-1/members/user-1",
			outputMatch: `"status": "deleted"`,
		},
		{
			name:    "user list",
			command: newUserListCmd,
			args:    []string{"--name", "sig", "--without-admin"},
			method:  http.MethodGet,
			path:    "/api/v1/users/",
			query:   url.Values{"name": []string{"sig"}, "page": []string{"1"}, "limit": []string{"10"}, "without_admin": []string{"true"}},
		},
		{
			name:    "user create",
			command: newUserCreateCmd,
			args:    []string{"--username", "sigma", "--password", "secret", "--email", "sigma@example.test", "--role", enums.UserRoleAdmin.String(), "--namespace-limit", "10"},
			method:  http.MethodPost,
			path:    "/api/v1/users/",
			body: map[string]any{
				"username":        "sigma",
				"password":        "secret",
				"email":           "sigma@example.test",
				"role":            enums.UserRoleAdmin.String(),
				"namespace_limit": float64(10),
			},
			outputMatch: `"status": "created"`,
		},
		{
			name:    "user update",
			command: newUserUpdateCmd,
			args:    []string{"user-1", "--username", "sigma", "--password", "secret", "--email", "sigma@example.test", "--role", enums.UserRoleAdmin.String(), "--status", enums.UserStatusInactive.String(), "--namespace-limit", "10"},
			method:  http.MethodPut,
			path:    "/api/v1/users/user-1",
			body: map[string]any{
				"username":        "sigma",
				"password":        "secret",
				"email":           "sigma@example.test",
				"role":            enums.UserRoleAdmin.String(),
				"status":          enums.UserStatusInactive.String(),
				"namespace_limit": float64(10),
			},
			outputMatch: `"status": "updated"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, tt.method, r.Method)
				require.Equal(t, tt.path, r.URL.Path)
				requireQueryContains(t, tt.query, r.URL.Query())
				username, password, ok := r.BasicAuth()
				require.True(t, ok)
				require.Equal(t, "sigma", username)
				require.Equal(t, "secret", password)
				if tt.body != nil {
					var body map[string]any
					require.NoError(t, json.UnmarshalRead(r.Body, &body))
					requireMapContains(t, tt.body, body)
				}
				w.Header().Set("Content-Type", "application/json")
				_, err := w.Write([]byte(`{}`))
				require.NoError(t, err)
			}))
			t.Cleanup(server.Close)

			opts := &options{
				client: newClient(server.URL, "sigma", "secret", time.Second, false),
			}
			output := executeCommand(t, tt.command(opts), tt.args...)
			if tt.outputMatch != "" {
				require.Contains(t, output, tt.outputMatch)
			}
		})
	}
}

func TestResourceCommandsValidateInput(t *testing.T) {
	opts := &options{client: newClient("http://example.test", "sigma", "secret", time.Second, false)}

	err := executeCommandError(newMemberAddCmd(opts), "ns-1", "not-number")
	require.Error(t, err)

	err = executeCommandError(newMemberUpdateCmd(opts), "ns-1", "user-1", "--role", "bad")
	require.Error(t, err)

	err = executeCommandError(newUserCreateCmd(opts), "--username", "sigma", "--password", "secret", "--email", "sigma@example.test", "--role", "bad")
	require.Error(t, err)

	err = executeCommandError(newUserUpdateCmd(opts), "user-1", "--role", "bad")
	require.Error(t, err)

	err = executeCommandError(newNamespaceCreateCmd(opts), "ns", "--visibility", "bad")
	require.Error(t, err)
}

func requireQueryContains(t *testing.T, expected, actual url.Values) {
	t.Helper()
	for key, values := range expected {
		require.Equal(t, values, actual[key], "query parameter %s", key)
	}
}

func requireMapContains(t *testing.T, expected, actual map[string]any) {
	t.Helper()
	for key, value := range expected {
		require.Equal(t, value, actual[key], "json field %s", key)
	}
}

func executeCommandError(command *cobra.Command, args ...string) error {
	command.SetArgs(args)
	return command.Execute()
}
