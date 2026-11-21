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

package app

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/go-sigma/sigma/pkg/config"
)

const staticIndexFile = "index.html"

// staticServer serves the built web UI. It replaces the nginx static file
// server: files are streamed from disk with sendfile, responses carry a
// metadata ETag so conditional requests get a 304, and http.ServeContent
// handles Range requests.
type staticServer struct {
	root  string
	index string
	cfg   config.ConfigurationHTTPStatic
}

// newStaticServer returns nil when static serving is disabled or the web root
// is unusable, so a server without a web build keeps running.
func newStaticServer(cfg *config.Configuration) *staticServer {
	staticCfg := cfg.HTTP.Static
	if !staticCfg.IsEnabled() {
		return nil
	}

	root, err := filepath.Abs(staticCfg.Dir)
	if err != nil {
		slog.Warn("static: resolving web root failed, web UI will not be served", "dir", staticCfg.Dir, "err", err)
		return nil
	}
	info, err := os.Stat(root)
	if err != nil || !info.IsDir() {
		slog.Warn("static: web root is not a directory, web UI will not be served", "dir", staticCfg.Dir)
		return nil
	}
	index := filepath.Join(root, staticIndexFile)
	if info, err = os.Stat(index); err != nil || info.IsDir() {
		slog.Warn("static: index.html is missing, web UI will not be served", "dir", staticCfg.Dir)
		return nil
	}

	slog.Info("static: serving web UI", "dir", root, "spa", staticCfg.SpaEnabled())
	return &staticServer{root: root, index: index, cfg: staticCfg}
}

// Register installs the static handler as the router fallback. Routes that the
// server owns keep precedence, and unknown API paths stay 404s.
func (s *staticServer) Register(e *gin.Engine) {
	e.NoRoute(s.handle)
}

func (s *staticServer) handle(c *gin.Context) {
	if c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead {
		http.NotFound(c.Writer, c.Request)
		return
	}

	urlPath := c.Request.URL.Path
	if isReservedPath(urlPath) {
		http.NotFound(c.Writer, c.Request)
		return
	}

	if s.serveFile(c, urlPath) {
		return
	}

	// A single-page application resolves its own routes on the client, so an
	// unknown extensionless path must return index.html instead of a 404. A
	// missing asset keeps returning 404 so broken references stay visible.
	if s.cfg.SpaEnabled() && path.Ext(urlPath) == "" && s.serveFile(c, "/"+staticIndexFile) {
		return
	}

	http.NotFound(c.Writer, c.Request)
}

// serveFile serves a root-relative URL path and reports whether it exists.
func (s *staticServer) serveFile(c *gin.Context, urlPath string) bool {
	webPath := cleanWebPath(urlPath)
	full := filepath.Join(s.root, filepath.FromSlash(webPath))

	info, err := os.Stat(full)
	if err != nil {
		return false
	}
	if info.IsDir() {
		webPath = path.Join(webPath, staticIndexFile)
		full = filepath.Join(s.root, filepath.FromSlash(webPath))
		if info, err = os.Stat(full); err != nil || info.IsDir() {
			return false
		}
	}

	f, err := os.Open(full)
	if err != nil {
		return false
	}
	defer f.Close() // nolint: errcheck

	c.Header("Cache-Control", s.cfg.CacheControlFor("/"+webPath))
	c.Header("ETag", fileETag(info))
	// ServeContent adds Last-Modified, answers If-None-Match with 304, and
	// handles Range requests without buffering the file in memory.
	http.ServeContent(c.Writer, c.Request, webPath, info.ModTime(), f)
	return true
}

// cleanWebPath makes an absolute URL path safe to join with the web root: any
// ".." is resolved against the root, so the result can never escape it.
func cleanWebPath(urlPath string) string {
	return strings.TrimPrefix(path.Clean("/"+urlPath), "/")
}

// isReservedPath reports whether the server owns the path. Unknown paths below
// these prefixes must not fall back to index.html.
func isReservedPath(urlPath string) bool {
	for _, prefix := range []string{"/api/", "/v2/", "/swagger/", "/__debug/"} {
		if strings.HasPrefix(urlPath, prefix) {
			return true
		}
	}
	switch urlPath {
	case "/api", "/v2", "/swagger", "/metrics", "/healthz", "/readyz", "/livez":
		return true
	}
	return false
}

// fileETag derives an ETag from file metadata. It is only stable for a given
// build, which is exactly what the web UI needs: Vite rewrites asset names on
// content changes and rewrites index.html on every release.
func fileETag(info os.FileInfo) string {
	return fmt.Sprintf(`"%x-%x"`, info.Size(), info.ModTime().UnixNano())
}
