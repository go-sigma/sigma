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

package validators

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/testkit"
)

func TestFactory(t *testing.T) {
	digCon := dig.New()
	require.NoError(t, digCon.Provide(testkit.NewGin))
	require.NoError(t, factory{}.Initialize(digCon))
	require.NoError(t, digCon.Invoke(func(engine *gin.Engine) {
		routes := engine.Routes()
		require.Len(t, routes, 5)
		require.True(t, hasRoute(routes, http.MethodGet, "/api/v1/validators/reference"))
		require.True(t, hasRoute(routes, http.MethodGet, "/api/v1/validators/tag"))
		require.True(t, hasRoute(routes, http.MethodPost, "/api/v1/validators/password"))
		require.True(t, hasRoute(routes, http.MethodPost, "/api/v1/validators/cron"))
		require.True(t, hasRoute(routes, http.MethodPost, "/api/v1/validators/regexp"))
	}))
}

func TestValidators(t *testing.T) {
	tests := []struct {
		name       string
		invoke     func(*handler, *gin.Context)
		wantStatus int
	}{
		{
			name: "valid reference",
			invoke: func(h *handler, c *gin.Context) {
				h.GetReference(c, &api.GetValidatorReferenceRequest{Reference: "library/alpine"})
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name: "invalid reference",
			invoke: func(h *handler, c *gin.Context) {
				h.GetReference(c, &api.GetValidatorReferenceRequest{Reference: "alpine"})
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "valid tag template",
			invoke: func(h *handler, c *gin.Context) {
				h.GetTag(c, &api.GetValidatorTagRequest{Tag: "{{ .ScmBranch }}-{{ .ScmTag }}"})
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name: "invalid tag template",
			invoke: func(h *handler, c *gin.Context) {
				h.GetTag(c, &api.GetValidatorTagRequest{Tag: "{{"})
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "valid password",
			invoke: func(h *handler, c *gin.Context) {
				h.GetPassword(c, &api.ValidatePasswordRequest{Password: "Admin@123"})
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name: "weak password",
			invoke: func(h *handler, c *gin.Context) {
				h.GetPassword(c, &api.ValidatePasswordRequest{Password: "123"})
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "valid cron",
			invoke: func(h *handler, c *gin.Context) {
				h.ValidateCron(c, &api.ValidateCronRequest{Cron: "0 0 * * 6"})
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name: "invalid cron",
			invoke: func(h *handler, c *gin.Context) {
				h.ValidateCron(c, &api.ValidateCronRequest{Cron: "invalid"})
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "valid regexp",
			invoke: func(h *handler, c *gin.Context) {
				h.ValidateRegexp(c, &api.ValidateRegexpRequest{Regexp: "^v.*$"})
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name: "invalid regexp",
			invoke: func(h *handler, c *gin.Context) {
				h.ValidateRegexp(c, &api.ValidateRegexpRequest{Regexp: "["})
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			context, _ := gin.CreateTestContext(recorder)
			tt.invoke(&handler{}, context)
			context.Writer.WriteHeaderNow()
			require.Equal(t, tt.wantStatus, recorder.Code)
		})
	}
}

func hasRoute(routes gin.RoutesInfo, method, path string) bool {
	for _, route := range routes {
		if route.Method == method && route.Path == path {
			return true
		}
	}
	return false
}
