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

package distribution

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	svcrepository "github.com/go-sigma/sigma/pkg/service/repositories"
)

func TestListRepositories(t *testing.T) {
	ctrl := gomock.NewController(t)
	svc := svcrepository.NewMockRepositoryService(ctrl)
	svc.EXPECT().ListRepositories(gomock.Any(), "user-1", "", nil, gomock.Any(), gomock.Any()).
		Return([]*models.Repository{{Name: "library/alpine"}, {Name: "library/busybox"}}, nil, int64(2), nil)
	recorder, c := newDistributionContext(http.MethodGet, "/v2/_catalog?n=200")
	c.Request.Host = "registry.example.com"
	c.Set(consts.ContextUser, &models.User{ID: "user-1"})

	(&handler{RepoSvc: svc}).ListRepositories(c)
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusOK, recorder.Code)
	require.JSONEq(t, `{"repositories":["library/alpine","library/busybox"]}`, recorder.Body.String())
	require.Equal(t, `<http://registry.example.com/v2/_catalog?last=library%2Fbusybox&n=200>; rel="next"`, recorder.Header().Get("Link"))
}

func TestListRepositoriesInvalidLimit(t *testing.T) {
	recorder, c := newDistributionContext(http.MethodGet, "/v2/_catalog?n=invalid")
	c.Set(consts.ContextUser, &models.User{ID: "user-1"})

	(&handler{}).ListRepositories(c)
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestListRepositoriesServiceError(t *testing.T) {
	ctrl := gomock.NewController(t)
	svc := svcrepository.NewMockRepositoryService(ctrl)
	svc.EXPECT().ListRepositories(gomock.Any(), "user-1", "", nil, gomock.Any(), gomock.Any()).Return(nil, nil, int64(0), errors.New("list failed"))
	recorder, c := newDistributionContext(http.MethodGet, "/v2/_catalog")
	c.Set(consts.ContextUser, &models.User{ID: "user-1"})

	(&handler{RepoSvc: svc}).ListRepositories(c)
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func newDistributionContext(method, target string) (*httptest.ResponseRecorder, *gin.Context) {
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(method, target, nil)
	return recorder, c
}
