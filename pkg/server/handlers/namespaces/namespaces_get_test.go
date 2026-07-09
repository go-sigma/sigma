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

	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/server/errcode"
	svcnamespace "github.com/go-sigma/sigma/pkg/service/namespaces"
	"github.com/go-sigma/sigma/pkg/validators"
)

func TestGetNamespace(t *testing.T) {
	require.NoError(t, validators.Initialize())

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	serviceObj := svcnamespace.NewMockNamespaceService(ctrl)
	serviceObj.EXPECT().
		GetNamespace(gomock.Any(), "1").
		DoAndReturn(func(_ context.Context, id string) (*models.Namespace, int64, int64, error) {
			require.Equal(t, "1", id)
			return &models.Namespace{
				ID:              id,
				Name:            "test",
				Overview:        []byte("overview"),
				Visibility:      enums.VisibilityPrivate,
				RepositoryLimit: 10,
				TagLimit:        20,
				CreatedAt:       1,
				UpdatedAt:       2,
			}, 2, 3, nil
		})

	router := newNamespacesTestRouter(handler{NsSvc: serviceObj}, &models.User{ID: "10"})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, consts.APIV1+"/namespaces/1", nil)

	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "1", gjson.Get(rec.Body.String(), "id").String())
	require.Equal(t, "test", gjson.Get(rec.Body.String(), "name").String())
	require.Equal(t, int64(2), gjson.Get(rec.Body.String(), "repository_count").Int())
	require.Equal(t, int64(3), gjson.Get(rec.Body.String(), "tag_count").Int())
}

func TestGetNamespaceReturnsInternalError(t *testing.T) {
	require.NoError(t, validators.Initialize())

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	serviceObj := svcnamespace.NewMockNamespaceService(ctrl)
	serviceObj.EXPECT().
		GetNamespace(gomock.Any(), "1").
		Return(nil, int64(0), int64(0), errors.New("get failed"))

	router := newNamespacesTestRouter(handler{NsSvc: serviceObj}, &models.User{ID: "10"})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, consts.APIV1+"/namespaces/1", nil)

	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
	require.Equal(t, errcode.HTTPErrCodeInternalError.Code, gjson.Get(rec.Body.String(), "code").String())
}
