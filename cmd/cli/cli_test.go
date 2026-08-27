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
	"context"
	"encoding/json/v2"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	"github.com/go-sigma/sigma/pkg/api/enums"
)

func TestNewRootCmd(t *testing.T) {
	t.Setenv("SIGMA_SERVER", "http://example.test")

	command := newRootCmd()

	require.Equal(t, "sigma-cli", command.Use)
	require.NotNil(t, findCommand(t, command, "login"))
	require.NotNil(t, findCommand(t, command, "namespace"))
	require.NotNil(t, findCommand(t, command, "repo"))
	require.NotNil(t, findCommand(t, command, "tag"))
	require.NotNil(t, findCommand(t, command, "artifact"))
	require.NotNil(t, findCommand(t, command, "member"))
	require.NotNil(t, findCommand(t, command, "user"))
}

func TestOptionsInitialize(t *testing.T) {
	credentialsFile := filepath.Join(t.TempDir(), "credentials.json")
	require.NoError(t, saveCredentials(credentialsFile, credentials{
		Server:   "http://saved.example.test/",
		Username: "sigma",
		Password: "secret",
	}))
	opts := &options{
		server:          "http://flag.example.test/",
		credentialsFile: credentialsFile,
		timeout:         time.Second,
	}
	command := &cobra.Command{Use: "test"}
	command.PersistentFlags().String("server", opts.server, "")

	require.NoError(t, opts.initialize(context.Background(), command))
	require.Equal(t, "http://saved.example.test", opts.client.baseURL)
	require.Equal(t, "sigma", opts.client.username)
	require.Equal(t, "secret", opts.client.password)

	require.NoError(t, command.PersistentFlags().Set("server", "http://flag.example.test/"))
	require.NoError(t, opts.initialize(context.Background(), command))
	require.Equal(t, "http://flag.example.test", opts.client.baseURL)
}

func TestOptionsInitializeRequiresCredentials(t *testing.T) {
	opts := &options{
		server:          "http://example.test",
		credentialsFile: filepath.Join(t.TempDir(), "missing.json"),
		timeout:         time.Second,
	}
	command := &cobra.Command{Use: "test"}
	command.PersistentFlags().String("server", opts.server, "")

	err := opts.initialize(context.Background(), command)

	require.ErrorContains(t, err, "basic auth credentials are required")
}

func TestNewLoginCmd(t *testing.T) {
	credentialsFile := filepath.Join(t.TempDir(), "credentials.json")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/api/v1/users/login", r.URL.Path)
		username, password, ok := r.BasicAuth()
		require.True(t, ok)
		require.Equal(t, "sigma", username)
		require.Equal(t, "secret", password)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	opts := &options{
		server:          server.URL + "/",
		credentialsFile: credentialsFile,
		timeout:         time.Second,
	}
	command := newLoginCmd(opts)
	output := executeCommand(t, command, "-u", "sigma", "-p", "secret")

	require.Contains(t, output, `"status": "logged_in"`)
	cred, err := loadCredentials(credentialsFile)
	require.NoError(t, err)
	require.Equal(t, server.URL, cred.Server)
	require.Equal(t, "sigma", cred.Username)
	require.Equal(t, "secret", cred.Password)
}

func TestClientDo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/api/v1/test", r.URL.Path)
		require.Equal(t, "value", r.URL.Query().Get("name"))
		username, password, ok := r.BasicAuth()
		require.True(t, ok)
		require.Equal(t, "override", username)
		require.Equal(t, "secret", password)
		var body map[string]string
		require.NoError(t, json.UnmarshalRead(r.Body, &body))
		require.Equal(t, "payload", body["name"])
		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte(`{"status":"ok"}`))
		require.NoError(t, err)
	}))
	t.Cleanup(server.Close)

	cli := newClient(server.URL, "default", "default-secret", time.Second, false)
	var out map[string]string
	query := url.Values{"name": []string{"value"}}
	err := cli.do(t.Context(), http.MethodPost, "/api/v1/test", query, map[string]string{"name": "payload"}, &out, basicAuth{
		username: "override",
		password: "secret",
	})

	require.NoError(t, err)
	require.Equal(t, "ok", out["status"])
}

func TestClientDoReturnsFormattedError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write([]byte(`{"message":"bad request"}`))
		require.NoError(t, err)
	}))
	t.Cleanup(server.Close)

	err := newClient(server.URL, "", "", time.Second, false).
		do(t.Context(), http.MethodGet, "/api/v1/fail", nil, nil, nil, basicAuth{})

	require.ErrorContains(t, err, "GET /api/v1/fail")
	require.ErrorContains(t, err, `"message": "bad request"`)
}

func TestCredentialsFileLifecycle(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing.json")
	cred, err := loadCredentials(missing)
	require.NoError(t, err)
	require.Empty(t, cred)

	path := filepath.Join(t.TempDir(), "nested", "credentials.json")
	expected := credentials{
		Server:   "http://example.test",
		Username: "sigma",
		Password: "secret",
	}
	require.NoError(t, saveCredentials(path, expected))

	info, err := os.Stat(path)
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0o600), info.Mode().Perm())

	actual, err := loadCredentials(path)
	require.NoError(t, err)
	require.Equal(t, expected, actual)

	invalid := filepath.Join(t.TempDir(), "invalid.json")
	require.NoError(t, os.WriteFile(invalid, []byte("{"), 0o600))
	_, err = loadCredentials(invalid)
	require.ErrorContains(t, err, "decode credentials file")
}

func TestEnvAndCredentialsFileDefaults(t *testing.T) {
	t.Setenv("SIGMA_TEST_ENV", "  configured  ")
	require.Equal(t, "configured", envOrDefault("SIGMA_TEST_ENV", "fallback"))
	t.Setenv("SIGMA_TEST_ENV", "  ")
	require.Equal(t, "fallback", envOrDefault("SIGMA_TEST_ENV", "fallback"))

	t.Setenv("SIGMA_CREDENTIALS_FILE", "/tmp/sigma-creds.json")
	require.Equal(t, "/tmp/sigma-creds.json", defaultCredentialsFile())
}

func TestCommonListQuery(t *testing.T) {
	query := commonListQuery("sigma", 2, 50, "created_at", "desc")

	require.Equal(t, "sigma", query.Get("name"))
	require.Equal(t, "2", query.Get("page"))
	require.Equal(t, "50", query.Get("limit"))
	require.Equal(t, "created_at", query.Get("sort"))
	require.Equal(t, "desc", query.Get("method"))

	require.Empty(t, commonListQuery("", 0, 0, "", ""))
}

func TestOptionalFlagValues(t *testing.T) {
	command := &cobra.Command{Use: "test"}
	addNamespaceWriteFlags(command)
	require.NoError(t, command.Flags().Set("description", "desc"))
	require.NoError(t, command.Flags().Set("size-limit", "1024"))
	require.NoError(t, command.Flags().Set("visibility", enums.VisibilityPrivate.String()))

	require.Equal(t, "desc", *optionalString(command, "description"))
	require.Equal(t, int64(1024), *optionalInt64(command, "size-limit"))
	visibility, err := optionalVisibility(command)
	require.NoError(t, err)
	require.Equal(t, enums.VisibilityPrivate, *visibility)

	require.Nil(t, optionalString(command, "overview"))

	invalid := &cobra.Command{Use: "invalid"}
	invalid.Flags().String("visibility", "", "")
	require.NoError(t, invalid.Flags().Set("visibility", "bad"))
	_, err = optionalVisibility(invalid)
	require.Error(t, err)
}

func TestFormatErrorBodyAndEndpoint(t *testing.T) {
	require.Equal(t, "{\n  \"message\": \"bad\"\n}", formatErrorBody([]byte(`{"message":"bad"}`)))
	require.Equal(t, "plain error", formatErrorBody([]byte(" plain error ")))
	require.Equal(t, "/api/v1/namespaces/ns", endpoint("/api/v1/namespaces/%s", "ns"))
}

func findCommand(t *testing.T, root *cobra.Command, name string) *cobra.Command {
	t.Helper()
	command, _, err := root.Find([]string{name})
	require.NoError(t, err)
	return command
}

func executeCommand(t *testing.T, command *cobra.Command, args ...string) string {
	t.Helper()

	oldStdout := os.Stdout
	readFile, writeFile, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = writeFile

	command.SetArgs(args)
	command.SetOut(writeFile)
	command.SetErr(writeFile)
	err = command.Execute()
	require.NoError(t, err)
	require.NoError(t, writeFile.Close())
	os.Stdout = oldStdout
	t.Cleanup(func() {
		os.Stdout = oldStdout
	})

	data, err := io.ReadAll(readFile)
	require.NoError(t, err)
	require.NoError(t, readFile.Close())
	return string(data)
}
