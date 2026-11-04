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

package authz

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/authz"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
)

// stubAuthorizer only answers the middleware call; the embedded interface panics if another
// method is used, which would mean the test exercised an unexpected path.
type stubAuthorizer struct {
	authz.Authorizer
	allowed bool
	calls   int
}

func (s *stubAuthorizer) Authorize(_ context.Context, _ string, _ bool, _ string, _ string) (bool, error) {
	s.calls++
	return s.allowed, nil
}

func TestAuthzSelfScopedPaths(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name             string
		role             enums.UserRole
		method           string
		path             string
		authorizerAllows bool
		wantStatus       int
		wantCalls        int
	}{
		{name: "self profile", role: enums.UserRoleUser, method: http.MethodGet, path: "/api/v1/users/self", wantStatus: http.StatusOK},
		{name: "self update", role: enums.UserRoleUser, method: http.MethodPut, path: "/api/v1/users/self", wantStatus: http.StatusOK},
		{name: "self reset password", role: enums.UserRoleUser, method: http.MethodPut, path: "/api/v1/users/self/reset-password", wantStatus: http.StatusOK},
		{name: "logout", role: enums.UserRoleUser, method: http.MethodPost, path: "/api/v1/users/logout", wantStatus: http.StatusOK},
		{name: "password validator", role: enums.UserRoleUser, method: http.MethodPost, path: "/api/v1/validators/password", wantStatus: http.StatusOK},
		{name: "system config", role: enums.UserRoleUser, method: http.MethodGet, path: "/api/v1/systems/config", wantStatus: http.StatusOK},
		{name: "overview", role: enums.UserRoleUser, method: http.MethodGet, path: "/api/v1/overview", wantStatus: http.StatusOK},
		{name: "daily activity", role: enums.UserRoleUser, method: http.MethodGet, path: "/api/v1/activity/daily", wantStatus: http.StatusOK},
		{name: "daily builds", role: enums.UserRoleUser, method: http.MethodGet, path: "/api/v1/builds/daily", wantStatus: http.StatusOK},
		{name: "activity heatmap", role: enums.UserRoleUser, method: http.MethodGet, path: "/api/v1/users/01a1/activity/heatmap", wantStatus: http.StatusOK},
		{name: "user list stays denied", role: enums.UserRoleUser, method: http.MethodGet, path: "/api/v1/users/", wantStatus: http.StatusUnauthorized},
		{name: "namespace list with pagination", role: enums.UserRoleUser, method: http.MethodGet, path: "/api/v1/namespaces/?limit=20", wantStatus: http.StatusOK},
		{name: "namespace allowed by authorizer", role: enums.UserRoleUser, method: http.MethodGet, path: "/api/v1/namespaces/01a1", authorizerAllows: true, wantStatus: http.StatusOK, wantCalls: 1},
		{name: "namespace denied by authorizer", role: enums.UserRoleUser, method: http.MethodGet, path: "/api/v1/namespaces/01a1", wantStatus: http.StatusUnauthorized, wantCalls: 1},
		{name: "admin bypasses the authorizer", role: enums.UserRoleAdmin, method: http.MethodGet, path: "/api/v1/users/", wantStatus: http.StatusOK},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			authorizer := &stubAuthorizer{allowed: test.authorizerAllows}
			engine := gin.New()
			engine.Use(func(c *gin.Context) {
				c.Set(consts.ContextUser, &models.User{ID: "01a1", Role: test.role})
				c.Next()
			})
			engine.Use(AuthzWithConfig(Config{Authorizer: authorizer}))
			engine.Any("/*path", func(c *gin.Context) { c.Status(http.StatusOK) })

			request := httptest.NewRequest(test.method, test.path, nil)
			recorder := httptest.NewRecorder()
			engine.ServeHTTP(recorder, request)

			require.Equal(t, test.wantStatus, recorder.Code)
			require.Equal(t, test.wantCalls, authorizer.calls)
		})
	}
}
