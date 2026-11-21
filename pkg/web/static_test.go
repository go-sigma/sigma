// Copyright 2025 sigma
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

package web

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/testkit"
)

const (
	testIndexHTML   = "<html>sigma</html>"
	testSwaggerHTML = "<html>swagger</html>"
)

// newStaticTestEngine builds an engine that serves a small fake web build.
func newStaticTestEngine(t *testing.T) *gin.Engine {
	t.Helper()

	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, "assets"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(root, "swagger"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "index.html"), []byte(testIndexHTML), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(root, "assets", "app.js"), []byte("console.log('ok')"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(root, "swagger", "index.html"), []byte(testSwaggerHTML), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(root, "swagger", "swagger-ui.css"), []byte("body{}"), 0o600))

	enabled := true
	cfg := &config.Configuration{HTTP: config.ConfigurationHTTP{
		Static: config.ConfigurationHTTPStatic{Enabled: &enabled, Dir: root},
	}}
	cfg.WithDefaults()

	server := NewStatic(cfg)
	require.NotNil(t, server)

	e := testkit.NewGin()
	e.NoRoute(server.Serve)
	return e
}

func serveStaticRequest(e *gin.Engine, req *http.Request) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

func TestStaticServesIndex(t *testing.T) {
	e := newStaticTestEngine(t)

	rec := serveStaticRequest(e, httptest.NewRequest(http.MethodGet, "/", nil))
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, testIndexHTML, rec.Body.String())
	require.Equal(t, "no-cache", rec.Header().Get("Cache-Control"))
	require.NotEmpty(t, rec.Header().Get("ETag"))
}

func TestStaticServesFingerprintedAsset(t *testing.T) {
	e := newStaticTestEngine(t)

	rec := serveStaticRequest(e, httptest.NewRequest(http.MethodGet, "/assets/app.js", nil))
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "public, max-age=31536000, immutable", rec.Header().Get("Cache-Control"))
	require.Contains(t, rec.Header().Get("Content-Type"), "javascript")
}

func TestStaticConditionalRequest(t *testing.T) {
	e := newStaticTestEngine(t)

	first := serveStaticRequest(e, httptest.NewRequest(http.MethodGet, "/assets/app.js", nil))
	etag := first.Header().Get("ETag")
	require.NotEmpty(t, etag)

	req := httptest.NewRequest(http.MethodGet, "/assets/app.js", nil)
	req.Header.Set("If-None-Match", etag)
	second := serveStaticRequest(e, req)
	require.Equal(t, http.StatusNotModified, second.Code)
	require.Empty(t, second.Body.String())
}

func TestStaticRangeRequest(t *testing.T) {
	e := newStaticTestEngine(t)

	req := httptest.NewRequest(http.MethodGet, "/assets/app.js", nil)
	req.Header.Set("Range", "bytes=0-3")
	rec := serveStaticRequest(e, req)
	require.Equal(t, http.StatusPartialContent, rec.Code)
	require.Equal(t, "cons", rec.Body.String())
}

func TestStaticSpaFallback(t *testing.T) {
	e := newStaticTestEngine(t)

	rec := serveStaticRequest(e, httptest.NewRequest(http.MethodGet, "/namespaces/demo/repository", nil))
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, testIndexHTML, rec.Body.String())
	require.Equal(t, "no-cache", rec.Header().Get("Cache-Control"))
}

func TestStaticMissingAssetIsNotFound(t *testing.T) {
	e := newStaticTestEngine(t)

	rec := serveStaticRequest(e, httptest.NewRequest(http.MethodGet, "/assets/missing.js", nil))
	require.Equal(t, http.StatusNotFound, rec.Code)
	require.NotContains(t, rec.Body.String(), testIndexHTML)
}

func TestStaticReservedAPIPathIsNotFound(t *testing.T) {
	e := newStaticTestEngine(t)

	for _, target := range []string{"/api/v1/unknown", "/v2/unknown"} {
		rec := serveStaticRequest(e, httptest.NewRequest(http.MethodGet, target, nil))
		require.Equal(t, http.StatusNotFound, rec.Code, target)
		require.NotContains(t, rec.Body.String(), testIndexHTML, target)
	}
}

func TestStaticRejectsPathTraversal(t *testing.T) {
	e := newStaticTestEngine(t)

	rec := serveStaticRequest(e, httptest.NewRequest(http.MethodGet, "/../secret.txt", nil))
	require.Equal(t, http.StatusNotFound, rec.Code)
}

func TestStaticRejectsNonReadMethods(t *testing.T) {
	e := newStaticTestEngine(t)

	rec := serveStaticRequest(e, httptest.NewRequest(http.MethodPost, "/", nil))
	require.Equal(t, http.StatusNotFound, rec.Code)
}

func TestNewStaticServerDisabledOrMissingRoot(t *testing.T) {
	disabled := false
	cfg := &config.Configuration{HTTP: config.ConfigurationHTTP{
		Static: config.ConfigurationHTTPStatic{Enabled: &disabled},
	}}
	cfg.WithDefaults()
	require.Nil(t, NewStatic(cfg))

	enabled := true
	cfg = &config.Configuration{HTTP: config.ConfigurationHTTP{
		Static: config.ConfigurationHTTPStatic{Enabled: &enabled, Dir: filepath.Join(t.TempDir(), "missing")},
	}}
	cfg.WithDefaults()
	require.Nil(t, NewStatic(cfg))
}

func TestStaticServesSwaggerUI(t *testing.T) {
	e := newStaticTestEngine(t)

	rec := serveStaticRequest(e, httptest.NewRequest(http.MethodGet, "/swagger", nil))
	require.Equal(t, http.StatusMovedPermanently, rec.Code)
	require.Equal(t, "/swagger/", rec.Header().Get("Location"))

	rec = serveStaticRequest(e, httptest.NewRequest(http.MethodGet, "/swagger/", nil))
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, testSwaggerHTML, rec.Body.String())

	rec = serveStaticRequest(e, httptest.NewRequest(http.MethodGet, "/swagger/swagger-ui.css", nil))
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "body{}", rec.Body.String())
}

func TestStaticMissingSwaggerFileIsNotFound(t *testing.T) {
	e := newStaticTestEngine(t)

	for _, target := range []string{"/swagger/missing.css", "/swagger/missing-route"} {
		rec := serveStaticRequest(e, httptest.NewRequest(http.MethodGet, target, nil))
		require.Equal(t, http.StatusNotFound, rec.Code, target)
		require.NotContains(t, rec.Body.String(), testIndexHTML, target)
	}
}

func TestStaticSecurityHeaders(t *testing.T) {
	e := newStaticTestEngine(t)

	for _, target := range []string{"/", "/assets/app.js"} {
		rec := serveStaticRequest(e, httptest.NewRequest(http.MethodGet, target, nil))
		require.Equal(t, http.StatusOK, rec.Code, target)
		require.Equal(t, "DENY", rec.Header().Get("X-Frame-Options"), target)
		require.Equal(t, "frame-ancestors 'none'", rec.Header().Get("Content-Security-Policy"), target)
	}
}
