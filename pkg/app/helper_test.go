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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestEtagSkipper(t *testing.T) {
	tests := []struct {
		method string
		target string
		skip   bool
	}{
		{http.MethodGet, "/api/v1/repositories", false},
		{http.MethodGet, "/api/v1", false},
		{http.MethodHead, "/api/v1/repositories", false},
		{http.MethodPost, "/api/v1/repositories", true},
		{http.MethodGet, "/v2/library/alpine/blobs/sha256:abc", true},
		{http.MethodGet, "/assets/index-abc123.js", true},
		{http.MethodGet, "/namespaces/demo/repository/tags", true},
		{http.MethodGet, "/metrics", true},
		{http.MethodGet, "/healthz", true},
	}

	gin.SetMode(gin.TestMode)
	for _, tt := range tests {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(tt.method, tt.target, nil)
		require.Equal(t, tt.skip, etagSkipper(c), "%s %s", tt.method, tt.target)
	}
}
