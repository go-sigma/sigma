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
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"go.uber.org/mock/gomock"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/server/errcode"
	svcnamespace "github.com/go-sigma/sigma/pkg/service/namespaces"
	"github.com/go-sigma/sigma/pkg/validators"
)

func TestPostNamespace(t *testing.T) {
	require.NoError(t, validators.Initialize())

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	serviceObj := svcnamespace.NewMockNamespaceService(ctrl)
	serviceObj.EXPECT().
		CreateNamespace(gomock.Any(), "10", api.PostNamespaceRequest{Name: "test"}).
		DoAndReturn(func(_ context.Context, userID string, req api.PostNamespaceRequest) (*models.Namespace, error) {
			require.Equal(t, "10", userID)
			require.Equal(t, "test", req.Name)
			return &models.Namespace{ID: "100", Name: req.Name}, nil
		})

	router := newNamespacesTestRouter(handler{NsSvc: serviceObj}, &models.User{ID: "10"})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, consts.APIV1+"/namespaces/", strings.NewReader(`{"name":"test"}`))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code)
	require.Equal(t, "100", gjson.Get(rec.Body.String(), "id").String())
}

func TestPostNamespaceReturnsUnauthorizedWithoutUser(t *testing.T) {
	require.NoError(t, validators.Initialize())

	router := newNamespacesTestRouter(handler{}, nil)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, consts.APIV1+"/namespaces/", strings.NewReader(`{"name":"test"}`))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
	require.Equal(t, errcode.HTTPErrCodeUnauthorized.Code, gjson.Get(rec.Body.String(), "code").String())
}

func TestPostNamespaceReturnsBadRequestForInvalidName(t *testing.T) {
	require.NoError(t, validators.Initialize())

	router := newNamespacesTestRouter(handler{}, &models.User{ID: "10"})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, consts.APIV1+"/namespaces/", strings.NewReader(`{"name":"t"}`))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.Equal(t, errcode.HTTPErrCodeBadRequest.Code, gjson.Get(rec.Body.String(), "code").String())
}

func TestPostNamespaceReturnsServiceError(t *testing.T) {
	require.NoError(t, validators.Initialize())

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	serviceObj := svcnamespace.NewMockNamespaceService(ctrl)
	serviceObj.EXPECT().
		CreateNamespace(gomock.Any(), "10", gomock.Any()).
		Return(nil, errors.New("create failed"))

	router := newNamespacesTestRouter(handler{NsSvc: serviceObj}, &models.User{ID: "10"})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, consts.APIV1+"/namespaces/", strings.NewReader(`{"name":"test"}`))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
	require.Equal(t, errcode.HTTPErrCodeInternalError.Code, gjson.Get(rec.Body.String(), "code").String())
}
