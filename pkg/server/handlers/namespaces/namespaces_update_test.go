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
	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/server/errcode"
	svcnamespace "github.com/go-sigma/sigma/pkg/service/namespaces"
	"github.com/go-sigma/sigma/pkg/validators"
)

func TestPutNamespace(t *testing.T) {
	require.NoError(t, validators.Initialize())

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	serviceObj := svcnamespace.NewMockNamespaceService(ctrl)
	serviceObj.EXPECT().
		UpdateNamespace(gomock.Any(), "10", "1", gomock.Any()).
		DoAndReturn(func(_ context.Context, userID, id string, req api.UpdateNamespaceRequest) error {
			require.Equal(t, "10", userID)
			require.Equal(t, "1", id)
			require.NotNil(t, req.Description)
			require.Equal(t, "updated", *req.Description)
			return nil
		})

	authorizer := fakeAuthorizer{
		namespace: func(_ context.Context, user models.User, namespaceID string, auth enums.Auth) (bool, error) {
			require.Equal(t, "10", user.ID)
			require.Equal(t, "1", namespaceID)
			require.Equal(t, enums.AuthAdmin, auth)
			return true, nil
		},
	}
	router := newNamespacesTestRouter(handler{NsSvc: serviceObj, Authorizer: authorizer}, &models.User{ID: "10"})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, consts.APIV1+"/namespaces/1", strings.NewReader(`{"description":"updated"}`))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNoContent, rec.Code)
}

func TestPutNamespaceReturnsUnauthorizedWhenAuthFails(t *testing.T) {
	require.NoError(t, validators.Initialize())

	authorizer := fakeAuthorizer{
		namespace: func(context.Context, models.User, string, enums.Auth) (bool, error) {
			return false, nil
		},
	}
	router := newNamespacesTestRouter(handler{Authorizer: authorizer}, &models.User{ID: "10"})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, consts.APIV1+"/namespaces/1", strings.NewReader(`{"description":"updated"}`))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
	require.Equal(t, errcode.HTTPErrCodeUnauthorized.Code, gjson.Get(rec.Body.String(), "code").String())
}

func TestPutNamespaceReturnsServiceError(t *testing.T) {
	require.NoError(t, validators.Initialize())

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	serviceObj := svcnamespace.NewMockNamespaceService(ctrl)
	serviceObj.EXPECT().
		UpdateNamespace(gomock.Any(), "10", "1", gomock.Any()).
		Return(errors.New("update failed"))

	authorizer := fakeAuthorizer{
		namespace: func(context.Context, models.User, string, enums.Auth) (bool, error) {
			return true, nil
		},
	}
	router := newNamespacesTestRouter(handler{NsSvc: serviceObj, Authorizer: authorizer}, &models.User{ID: "10"})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, consts.APIV1+"/namespaces/1", strings.NewReader(`{"description":"updated"}`))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
	require.Equal(t, errcode.HTTPErrCodeInternalError.Code, gjson.Get(rec.Body.String(), "code").String())
}
