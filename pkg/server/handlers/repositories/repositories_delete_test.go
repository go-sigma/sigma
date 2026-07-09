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

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/server/errcode"
	svcrepository "github.com/go-sigma/sigma/pkg/service/repositories"
	"github.com/go-sigma/sigma/pkg/validators"
)

func TestDeleteRepository(t *testing.T) {
	require.NoError(t, validators.Initialize())

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	serviceObj := svcrepository.NewMockRepositoryService(ctrl)
	serviceObj.EXPECT().
		DeleteRepository(gomock.Any(), "123", "10", "20").
		DoAndReturn(func(_ context.Context, userID, namespaceID, repositoryID string) error {
			require.Equal(t, "123", userID)
			require.Equal(t, "10", namespaceID)
			require.Equal(t, "20", repositoryID)
			return nil
		})

	router := newRepositoriesTestRouter(handler{
		RepoSvc: serviceObj,
		Authorizer: fakeAuthorizer{
			repository: func(_ context.Context, user models.User, repositoryID string, auth enums.Auth) (bool, error) {
				require.Equal(t, "123", user.ID)
				require.Equal(t, "20", repositoryID)
				require.Equal(t, enums.AuthManage, auth)
				return true, nil
			},
		},
	}, &models.User{ID: "123"})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/namespaces/10/repositories/20", nil)

	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNoContent, rec.Code)
}

func TestDeleteRepositoryReturnsUnauthorizedWithoutUser(t *testing.T) {
	require.NoError(t, validators.Initialize())

	router := newRepositoriesTestRouter(handler{}, nil)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/namespaces/10/repositories/20", nil)

	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
	require.Equal(t, errcode.HTTPErrCodeUnauthorized.Code, gjson.Get(rec.Body.String(), "code").String())
}

func TestDeleteRepositoryReturnsUnauthorizedWhenAuthDenied(t *testing.T) {
	require.NoError(t, validators.Initialize())

	router := newRepositoriesTestRouter(handler{
		Authorizer: fakeAuthorizer{
			repository: func(context.Context, models.User, string, enums.Auth) (bool, error) {
				return false, nil
			},
		},
	}, &models.User{ID: "123"})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/namespaces/10/repositories/20", nil)

	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
	require.Equal(t, errcode.HTTPErrCodeUnauthorized.Code, gjson.Get(rec.Body.String(), "code").String())
}

func TestDeleteRepositoryReturnsNotFoundWhenAuthorizerCannotFindResource(t *testing.T) {
	require.NoError(t, validators.Initialize())

	router := newRepositoriesTestRouter(handler{
		Authorizer: fakeAuthorizer{
			repository: func(context.Context, models.User, string, enums.Auth) (bool, error) {
				return false, errors.Join(gorm.ErrRecordNotFound, errors.New("repository missing"))
			},
		},
	}, &models.User{ID: "123"})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/namespaces/10/repositories/20", nil)

	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
	require.Equal(t, errcode.HTTPErrCodeNotFound.Code, gjson.Get(rec.Body.String(), "code").String())
}

func TestDeleteRepositoryReturnsInternalError(t *testing.T) {
	require.NoError(t, validators.Initialize())

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	serviceObj := svcrepository.NewMockRepositoryService(ctrl)
	serviceObj.EXPECT().
		DeleteRepository(gomock.Any(), "123", "10", "20").
		Return(errors.New("delete failed"))

	router := newRepositoriesTestRouter(handler{
		RepoSvc: serviceObj,
		Authorizer: fakeAuthorizer{
			repository: func(context.Context, models.User, string, enums.Auth) (bool, error) {
				return true, nil
			},
		},
	}, &models.User{ID: "123"})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/namespaces/10/repositories/20", nil)

	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
	require.Equal(t, errcode.HTTPErrCodeInternalError.Code, gjson.Get(rec.Body.String(), "code").String())
}
