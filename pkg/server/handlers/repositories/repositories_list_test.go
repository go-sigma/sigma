// Copyright 2023 sigma
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

package repositories

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"go.uber.org/mock/gomock"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/authz"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/server"
	"github.com/go-sigma/sigma/pkg/server/errcode"
	svcrepository "github.com/go-sigma/sigma/pkg/service/repositories"
	"github.com/go-sigma/sigma/pkg/testkit"
	"github.com/go-sigma/sigma/pkg/validators"
)

func TestListRepositories(t *testing.T) {
	require.NoError(t, validators.Initialize())

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	serviceObj := svcrepository.NewMockRepositoryService(ctrl)
	serviceObj.EXPECT().
		ListRepositories(gomock.Any(), "123", "10", gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, userID string, namespaceID string, name *string, pagination api.Pagination, sort api.Sortable) ([]*models.Repository, map[string]*models.Builder, int64, error) {
			require.Equal(t, "123", userID)
			require.Equal(t, "10", namespaceID)
			require.NotNil(t, name)
			require.Equal(t, "busybox", *name)
			require.NotNil(t, pagination.Page)
			require.Equal(t, 1, *pagination.Page)
			require.NotNil(t, pagination.Limit)
			require.Equal(t, 100, *pagination.Limit)
			require.Nil(t, sort.Sort)
			require.Nil(t, sort.Method)
			return []*models.Repository{
				{
					ID:          "20",
					NamespaceID: "10",
					Name:        "library/busybox",
					Overview:    []byte("overview"),
					TagCount:    2,
					CreatedAt:   1,
					UpdatedAt:   2,
				},
			}, nil, 1, nil
		})

	router := newRepositoriesTestRouter(handler{RepoSvc: serviceObj}, &models.User{ID: "123"})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/namespaces/10/repositories/?page=1&limit=100&name=busybox", nil)

	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, int64(1), gjson.Get(rec.Body.String(), "total").Int())
	require.Equal(t, "library/busybox", gjson.Get(rec.Body.String(), "items.0.name").String())
	require.Equal(t, int64(2), gjson.Get(rec.Body.String(), "items.0.tag_count").Int())
}

func TestListRepositoriesReturnsInternalError(t *testing.T) {
	require.NoError(t, validators.Initialize())

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	serviceObj := svcrepository.NewMockRepositoryService(ctrl)
	serviceObj.EXPECT().
		ListRepositories(gomock.Any(), "", "10", gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil, nil, int64(0), errors.New("list failed"))

	router := newRepositoriesTestRouter(handler{RepoSvc: serviceObj}, nil)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/namespaces/10/repositories/?page=1&limit=100", nil)

	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
	require.Equal(t, errcode.HTTPErrCodeInternalError.Code, gjson.Get(rec.Body.String(), "code").String())
}

func newRepositoriesTestRouter(h handler, user *models.User) *gin.Engine {
	router := testkit.NewGin()
	router.Use(func(c *gin.Context) {
		if user != nil {
			c.Set(consts.ContextUser, user)
		}
	})
	group := router.Group(consts.APIV1 + "/namespaces/:namespace_id/repositories")
	group.GET("/", server.WrapRequest(h.ListRepositories))
	group.DELETE("/:repository_id", server.WrapRequest(h.DeleteRepository))
	return router
}

type fakeAuthorizer struct {
	repository func(ctx context.Context, user models.User, repositoryID string, auth enums.Auth) (bool, error)
}

var _ authz.Authorizer = fakeAuthorizer{}

func (f fakeAuthorizer) Authorize(context.Context, string, bool, string, string) (bool, error) {
	return false, nil
}

func (f fakeAuthorizer) Namespace(context.Context, models.User, string, enums.Auth) (bool, error) {
	return false, nil
}

func (f fakeAuthorizer) NamespaceRole(context.Context, models.User, string) (*enums.NamespaceRole, error) {
	return nil, nil
}

func (f fakeAuthorizer) NamespacesRole(context.Context, models.User, []string) (map[string]*enums.NamespaceRole, error) {
	return nil, nil
}

func (f fakeAuthorizer) Repository(ctx context.Context, user models.User, repositoryID string, auth enums.Auth) (bool, error) {
	if f.repository == nil {
		return false, nil
	}
	return f.repository(ctx, user, repositoryID, auth)
}

func (f fakeAuthorizer) Tag(context.Context, models.User, string, enums.Auth) (bool, error) {
	return false, nil
}

func (f fakeAuthorizer) Artifact(context.Context, models.User, string, enums.Auth) (bool, error) {
	return false, nil
}
