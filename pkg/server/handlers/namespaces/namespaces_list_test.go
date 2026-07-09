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

package namespaces

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"go.uber.org/mock/gomock"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/server/errcode"
	svcnamespace "github.com/go-sigma/sigma/pkg/service/namespaces"
	"github.com/go-sigma/sigma/pkg/validators"
)

func TestListNamespace(t *testing.T) {
	require.NoError(t, validators.Initialize())

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	serviceObj := svcnamespace.NewMockNamespaceService(ctrl)
	serviceObj.EXPECT().
		ListNamespaces(gomock.Any(), "10", gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, userID string, name *string, pagination api.Pagination, sort api.Sortable) ([]*models.Namespace, int64, error) {
			require.Equal(t, "10", userID)
			require.NotNil(t, name)
			require.Equal(t, "test", *name)
			require.NotNil(t, pagination.Page)
			require.Equal(t, 1, *pagination.Page)
			require.NotNil(t, pagination.Limit)
			require.Equal(t, 100, *pagination.Limit)
			require.Nil(t, sort.Sort)
			require.Nil(t, sort.Method)
			return []*models.Namespace{
				{
					ID:              "1",
					Name:            "test",
					Description:     nil,
					Visibility:      enums.VisibilityPrivate,
					RepositoryCount: 2,
					TagCount:        3,
					CreatedAt:       1,
					UpdatedAt:       2,
				},
			}, 1, nil
		})

	router := newNamespacesTestRouter(handler{NsSvc: serviceObj}, &models.User{ID: "10"})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, consts.APIV1+"/namespaces/?page=1&limit=100&name=test", nil)

	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, int64(1), gjson.Get(rec.Body.String(), "total").Int())
	require.Equal(t, "test", gjson.Get(rec.Body.String(), "items.0.name").String())
	require.Equal(t, int64(2), gjson.Get(rec.Body.String(), "items.0.repository_count").Int())
	require.Equal(t, int64(3), gjson.Get(rec.Body.String(), "items.0.tag_count").Int())
}

func TestListNamespaceReturnsInternalError(t *testing.T) {
	require.NoError(t, validators.Initialize())

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	serviceObj := svcnamespace.NewMockNamespaceService(ctrl)
	serviceObj.EXPECT().
		ListNamespaces(gomock.Any(), "", gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil, int64(0), errors.New("list failed"))

	router := newNamespacesTestRouter(handler{NsSvc: serviceObj}, nil)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, consts.APIV1+"/namespaces/?page=1&limit=100", nil)

	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
	require.Equal(t, errcode.HTTPErrCodeInternalError.Code, gjson.Get(rec.Body.String(), "code").String())
}
