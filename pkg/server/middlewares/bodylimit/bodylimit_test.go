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

package bodylimit

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/testkit"
)

var (
	zeroLimit    int64
	threeBytes   int64 = 3
	fiveBytes    int64 = 5
	hundredBytes int64 = 100
)

func TestBodyLimitRejectsAPIContentLengthAboveDefaultLimit(t *testing.T) {
	router := newBodyLimitRouter(config.ConfigurationHTTPBodyLimit{
		Default: &threeBytes,
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/items", strings.NewReader("test"))
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusRequestEntityTooLarge, rec.Code)
	require.Contains(t, rec.Body.String(), "PAYLOAD_TOO_LARGE")
}

func TestBodyLimitUsesManifestLimit(t *testing.T) {
	router := newBodyLimitRouter(config.ConfigurationHTTPBodyLimit{
		Default:  &hundredBytes,
		Manifest: &threeBytes,
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/v2/library/busybox/manifests/latest", strings.NewReader("test"))
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusRequestEntityTooLarge, rec.Code)
	require.Contains(t, rec.Body.String(), "DENIED")
}

func TestBodyLimitUsesBlobUploadLimit(t *testing.T) {
	router := newBodyLimitRouter(config.ConfigurationHTTPBodyLimit{
		Default: &threeBytes,
		Blob:    &fiveBytes,
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/v2/library/busybox/blobs/uploads/upload-id", strings.NewReader("test"))
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}

func TestBodyLimitZeroLimitDisablesRouteClass(t *testing.T) {
	router := newBodyLimitRouter(config.ConfigurationHTTPBodyLimit{
		Default: &threeBytes,
		Blob:    &zeroLimit,
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v2/library/busybox/blobs/uploads/", strings.NewReader("large-body"))
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}

func TestBodyLimitDisabled(t *testing.T) {
	enabled := false
	router := newBodyLimitRouter(config.ConfigurationHTTPBodyLimit{
		Enabled: &enabled,
		Default: &threeBytes,
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/items", strings.NewReader("test"))
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}

func TestBodyLimitWrapsUnknownLengthBody(t *testing.T) {
	router := testkit.NewGin()
	router.Use(BodyLimit(config.ConfigurationHTTPBodyLimit{
		Default: &threeBytes,
	}))
	router.POST("/api/v1/items", func(c *gin.Context) {
		_, err := io.ReadAll(c.Request.Body)
		var maxBytesErr *http.MaxBytesError
		require.ErrorAs(t, err, &maxBytesErr)
		c.Status(http.StatusRequestEntityTooLarge)
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/items", io.NopCloser(strings.NewReader("test")))
	req.ContentLength = -1
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusRequestEntityTooLarge, rec.Code)
}

func newBodyLimitRouter(cfg config.ConfigurationHTTPBodyLimit) *gin.Engine {
	router := testkit.NewGin()
	router.Use(BodyLimit(cfg))
	router.Any("/*path", func(c *gin.Context) {
		if c.Request.Body != nil {
			_, _ = io.Copy(io.Discard, c.Request.Body)
		}
		c.Status(http.StatusOK)
	})
	return router
}
