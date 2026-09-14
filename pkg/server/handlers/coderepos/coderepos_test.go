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

package coderepos

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/server/errcode"
	svccoderepo "github.com/go-sigma/sigma/pkg/service/coderepos"
)

func TestList(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svccoderepo.NewMockCodeRepositoryService(ctrl)
	serviceObj.EXPECT().
		ListCodeRepositories(gomock.Any(), "10", enums.ProviderGithub, nil, nil, gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, userID string, provider enums.Provider, owner, name *string, pagination api.Pagination, sort api.Sortable) ([]*models.CodeRepository, []*models.CodeRepositoryOwner, int64, error) {
			require.Equal(t, "10", userID)
			return []*models.CodeRepository{{
				ID:           "1",
				RepositoryID: "2",
				OwnerID:      "3",
				Owner:        "octocat",
				IsOrg:        true,
				Name:         "hello-world",
				SshUrl:       "git@github.com:octocat/hello-world.git",
				CloneUrl:     "https://github.com/octocat/hello-world.git",
				OciRepoCount: 1,
				CreatedAt:    1,
				UpdatedAt:    1,
			}}, []*models.CodeRepositoryOwner{{
				ID:        "3",
				OwnerID:   "owner-1",
				Owner:     "octocat",
				IsOrg:     true,
				CreatedAt: 1,
				UpdatedAt: 1,
			}}, 1, nil
		})

	recorder, c := newCodereposContext(t)
	(&handler{CodeRepoSvc: serviceObj}).List(c, &api.ListCodeRepositoryRequest{Provider: enums.ProviderGithub})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Contains(t, recorder.Body.String(), `"total":1`)
	require.Contains(t, recorder.Body.String(), `"name":"hello-world"`)
	require.Contains(t, recorder.Body.String(), `"owner_id":"3"`)
}

func TestListReturnsInternalError(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svccoderepo.NewMockCodeRepositoryService(ctrl)
	serviceObj.EXPECT().
		ListCodeRepositories(gomock.Any(), "10", enums.ProviderGithub, nil, nil, gomock.Any(), gomock.Any()).
		Return(nil, nil, int64(0), errors.New("list failed"))

	recorder, c := newCodereposContext(t)
	(&handler{CodeRepoSvc: serviceObj}).List(c, &api.ListCodeRepositoryRequest{Provider: enums.ProviderGithub})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func TestListReturnsErrCode(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svccoderepo.NewMockCodeRepositoryService(ctrl)
	serviceObj.EXPECT().
		ListCodeRepositories(gomock.Any(), "10", enums.ProviderGithub, nil, nil, gomock.Any(), gomock.Any()).
		Return(nil, nil, int64(0), errcode.HTTPErrCodeBadRequest.Detail("not a valid Provider"))

	recorder, c := newCodereposContext(t)
	(&handler{CodeRepoSvc: serviceObj}).List(c, &api.ListCodeRepositoryRequest{Provider: enums.ProviderGithub})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestGet(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svccoderepo.NewMockCodeRepositoryService(ctrl)
	serviceObj.EXPECT().
		GetCodeRepository(gomock.Any(), "10", enums.ProviderGithub, "1").
		DoAndReturn(func(_ context.Context, userID string, provider enums.Provider, id string) (*models.CodeRepository, []*models.CodeRepositoryOwner, error) {
			require.Equal(t, "10", userID)
			require.Equal(t, "1", id)
			return &models.CodeRepository{
				ID:           "1",
				RepositoryID: "2",
				OwnerID:      "3",
				Owner:        "octocat",
				IsOrg:        true,
				Name:         "hello-world",
				SshUrl:       "git@github.com:octocat/hello-world.git",
				CloneUrl:     "https://github.com/octocat/hello-world.git",
				OciRepoCount: 1,
				CreatedAt:    1,
				UpdatedAt:    1,
			}, nil, nil
		})

	recorder, c := newCodereposContext(t)
	(&handler{CodeRepoSvc: serviceObj}).Get(c, &api.GetCodeRepositoryRequest{Provider: enums.ProviderGithub, ID: "1"})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Contains(t, recorder.Body.String(), `"id":"1"`)
	require.Contains(t, recorder.Body.String(), `"name":"hello-world"`)
}

func TestGetReturnsInternalError(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svccoderepo.NewMockCodeRepositoryService(ctrl)
	serviceObj.EXPECT().
		GetCodeRepository(gomock.Any(), "10", enums.ProviderGithub, "1").
		Return(nil, nil, errors.New("get failed"))

	recorder, c := newCodereposContext(t)
	(&handler{CodeRepoSvc: serviceObj}).Get(c, &api.GetCodeRepositoryRequest{Provider: enums.ProviderGithub, ID: "1"})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func TestListOwners(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svccoderepo.NewMockCodeRepositoryService(ctrl)
	serviceObj.EXPECT().
		ListCodeRepositoryOwners(gomock.Any(), "10", enums.ProviderGithub, nil).
		DoAndReturn(func(_ context.Context, userID string, provider enums.Provider, name *string) ([]*models.CodeRepositoryOwner, int64, error) {
			require.Equal(t, "10", userID)
			return []*models.CodeRepositoryOwner{{
				ID:        "3",
				OwnerID:   "owner-1",
				Owner:     "octocat",
				IsOrg:     true,
				CreatedAt: 1,
				UpdatedAt: 1,
			}}, 1, nil
		})

	recorder, c := newCodereposContext(t)
	(&handler{CodeRepoSvc: serviceObj}).ListOwners(c, &api.ListCodeRepositoryOwnerRequest{Provider: enums.ProviderGithub})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Contains(t, recorder.Body.String(), `"total":1`)
	require.Contains(t, recorder.Body.String(), `"owner":"octocat"`)
}

func TestListBranches(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svccoderepo.NewMockCodeRepositoryService(ctrl)
	serviceObj.EXPECT().
		ListCodeRepositoryBranches(gomock.Any(), "1").
		DoAndReturn(func(_ context.Context, codeRepositoryID string) ([]*models.CodeRepositoryBranch, int64, error) {
			require.Equal(t, "1", codeRepositoryID)
			return []*models.CodeRepositoryBranch{{
				ID:        "2",
				Name:      "main",
				CreatedAt: 1,
				UpdatedAt: 1,
			}}, 1, nil
		})

	recorder, c := newCodereposContext(t)
	(&handler{CodeRepoSvc: serviceObj}).ListBranches(c, &api.ListCodeRepositoryBranchesRequest{ID: "1"})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Contains(t, recorder.Body.String(), `"total":1`)
	require.Contains(t, recorder.Body.String(), `"name":"main"`)
}

func TestListBranchesReturnsInternalError(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svccoderepo.NewMockCodeRepositoryService(ctrl)
	serviceObj.EXPECT().
		ListCodeRepositoryBranches(gomock.Any(), "1").
		Return(nil, int64(0), errors.New("list failed"))

	recorder, c := newCodereposContext(t)
	(&handler{CodeRepoSvc: serviceObj}).ListBranches(c, &api.ListCodeRepositoryBranchesRequest{ID: "1"})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func TestGetBranch(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svccoderepo.NewMockCodeRepositoryService(ctrl)
	serviceObj.EXPECT().
		GetCodeRepositoryBranch(gomock.Any(), "1", "main").
		Return(&models.CodeRepositoryBranch{
			ID:        "2",
			Name:      "main",
			CreatedAt: 1,
			UpdatedAt: 1,
		}, nil)

	recorder, c := newCodereposContext(t)
	(&handler{CodeRepoSvc: serviceObj}).GetBranch(c, &api.GetCodeRepositoryBranchRequest{ID: "1", Name: "main"})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Contains(t, recorder.Body.String(), `"name":"main"`)
}

func TestResync(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svccoderepo.NewMockCodeRepositoryService(ctrl)
	serviceObj.EXPECT().
		ResyncCodeRepositories(gomock.Any(), "10", enums.ProviderGithub).
		Return(nil)

	recorder, c := newCodereposContext(t)
	(&handler{CodeRepoSvc: serviceObj}).Resync(c, &api.GetCodeRepositoryResyncRequest{Provider: enums.ProviderGithub})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusAccepted, recorder.Code)
}

func TestResyncReturnsInternalError(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svccoderepo.NewMockCodeRepositoryService(ctrl)
	serviceObj.EXPECT().
		ResyncCodeRepositories(gomock.Any(), "10", enums.ProviderGithub).
		Return(errors.New("resync failed"))

	recorder, c := newCodereposContext(t)
	(&handler{CodeRepoSvc: serviceObj}).Resync(c, &api.GetCodeRepositoryResyncRequest{Provider: enums.ProviderGithub})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func TestProviders(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svccoderepo.NewMockCodeRepositoryService(ctrl)
	serviceObj.EXPECT().
		ListCodeRepositoryProviders(gomock.Any(), "10").
		DoAndReturn(func(_ context.Context, userID string) ([]*models.User3rdParty, error) {
			require.Equal(t, "10", userID)
			return []*models.User3rdParty{{
				ID:        "1",
				UserID:    "10",
				Provider:  enums.ProviderGithub,
				CreatedAt: 1,
				UpdatedAt: 1,
			}}, nil
		})

	recorder, c := newCodereposContext(t)
	(&handler{CodeRepoSvc: serviceObj}).Providers(c)
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Contains(t, recorder.Body.String(), `"provider":"github"`)
}

func TestUser3rdParty(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svccoderepo.NewMockCodeRepositoryService(ctrl)
	serviceObj.EXPECT().
		GetCodeRepositoryUser3rdParty(gomock.Any(), "10", enums.ProviderGithub).
		Return(&models.User3rdParty{
			ID:                    "1",
			UserID:                "10",
			Provider:              enums.ProviderGithub,
			AccountID:             new("account-1"),
			CrLastUpdateTimestamp: 1,
			CrLastUpdateStatus:    enums.TaskCommonStatusSuccess,
			CrLastUpdateMessage:   nil,
			CreatedAt:             1,
			UpdatedAt:             1,
		}, nil)

	recorder, c := newCodereposContext(t)
	(&handler{CodeRepoSvc: serviceObj}).User3rdParty(c, &api.GetCodeRepositoryUser3rdPartyRequest{Provider: enums.ProviderGithub})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Contains(t, recorder.Body.String(), `"id":"1"`)
	require.Contains(t, recorder.Body.String(), `"account_id":"account-1"`)
}

func TestUser3rdPartyReturnsNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	serviceObj := svccoderepo.NewMockCodeRepositoryService(ctrl)
	serviceObj.EXPECT().
		GetCodeRepositoryUser3rdParty(gomock.Any(), "10", enums.ProviderGithub).
		Return(nil, errors.New("not found"))

	recorder, c := newCodereposContext(t)
	(&handler{CodeRepoSvc: serviceObj}).User3rdParty(c, &api.GetCodeRepositoryUser3rdPartyRequest{Provider: enums.ProviderGithub})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusNotFound, recorder.Code)
}

func newCodereposContext(t *testing.T) (*httptest.ResponseRecorder, *gin.Context) {
	t.Helper()
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Set(consts.ContextUser, &models.User{ID: "10", Role: enums.UserRoleAdmin})
	return recorder, c
}
