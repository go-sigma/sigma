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

package analytics

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/authz"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	svcanalytics "github.com/go-sigma/sigma/pkg/service/analytics"
)

func TestGetUserPushHeatmap(t *testing.T) {
	tests := []struct {
		name       string
		user       *models.User
		userID     string
		days       string
		expectMock bool
		wantStatus int
	}{
		{name: "success", user: &models.User{ID: "user-1"}, userID: "user-1", days: "30", expectMock: true, wantStatus: http.StatusOK},
		{name: "invalid days", user: &models.User{ID: "user-1"}, userID: "user-1", days: "401", wantStatus: http.StatusBadRequest},
		{name: "unauthorized user", user: &models.User{ID: "user-2"}, userID: "user-1", days: "30", wantStatus: http.StatusUnauthorized},
		{name: "missing user", userID: "user-1", days: "30", wantStatus: http.StatusUnauthorized},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			service := svcanalytics.NewMockService(ctrl)
			if tt.expectMock {
				service.EXPECT().GetUserPushHeatmap(gomock.Any(), tt.userID, 30).Return([]api.DailyCount{}, nil)
			}
			recorder := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(recorder)
			c.Request = httptest.NewRequest(http.MethodGet, "/?days="+tt.days, nil)
			c.Params = gin.Params{{Key: "user_id", Value: tt.userID}}
			if tt.user != nil {
				c.Set(consts.ContextUser, tt.user)
			}

			(&handler{AnalyticsSvc: service}).GetUserPushHeatmap(c)
			c.Writer.WriteHeaderNow()

			require.Equal(t, tt.wantStatus, recorder.Code)
		})
	}
}

func TestGetNamespaceTrends(t *testing.T) {
	tests := []struct {
		name       string
		authorized bool
		expectMock bool
		wantStatus int
	}{
		{name: "success", authorized: true, expectMock: true, wantStatus: http.StatusOK},
		{name: "unauthorized", authorized: false, wantStatus: http.StatusUnauthorized},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			service := svcanalytics.NewMockService(ctrl)
			if tt.expectMock {
				service.EXPECT().GetNamespaceTrends(gomock.Any(), "namespace-1", 7).Return([]api.NamespaceHourlyMetric{}, nil)
			}
			recorder := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(recorder)
			c.Request = httptest.NewRequest(http.MethodGet, "/?days=7", nil)
			c.Params = gin.Params{{Key: "namespace_id", Value: "namespace-1"}}
			c.Set(consts.ContextUser, &models.User{ID: "user-1"})

			(&handler{
				AnalyticsSvc: service,
				Authorizer: fakeAuthorizer{namespace: func(context.Context, models.User, string, enums.Auth) (bool, error) {
					return tt.authorized, nil
				}},
			}).GetNamespaceTrends(c)
			c.Writer.WriteHeaderNow()

			require.Equal(t, tt.wantStatus, recorder.Code)
		})
	}
}

type fakeAuthorizer struct {
	namespace func(context.Context, models.User, string, enums.Auth) (bool, error)
}

var _ authz.Authorizer = fakeAuthorizer{}

func (f fakeAuthorizer) Authorize(context.Context, string, bool, string, string) (bool, error) {
	return false, nil
}

func (f fakeAuthorizer) Namespace(ctx context.Context, user models.User, namespaceID string, permission enums.Auth) (bool, error) {
	return f.namespace(ctx, user, namespaceID, permission)
}

func (fakeAuthorizer) NamespaceRole(context.Context, models.User, string) (*enums.NamespaceRole, error) {
	return nil, nil
}

func (fakeAuthorizer) NamespacesRole(context.Context, models.User, []string) (map[string]*enums.NamespaceRole, error) {
	return nil, nil
}

func (fakeAuthorizer) Repository(context.Context, models.User, string, enums.Auth) (bool, error) {
	return false, nil
}

func (fakeAuthorizer) Tag(context.Context, models.User, string, enums.Auth) (bool, error) {
	return false, nil
}

func (fakeAuthorizer) Artifact(context.Context, models.User, string, enums.Auth) (bool, error) {
	return false, nil
}
