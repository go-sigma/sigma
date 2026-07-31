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

package repositories

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/dal/models"
	repobuilder "github.com/go-sigma/sigma/pkg/dal/repository/builder"
	reponamespace "github.com/go-sigma/sigma/pkg/dal/repository/namespace"
	reporegistry "github.com/go-sigma/sigma/pkg/dal/repository/registry"
	"github.com/go-sigma/sigma/pkg/server/errcode"
)

func TestGetRepository(t *testing.T) {
	ctrl := gomock.NewController(t)
	repositoryRepository := reporegistry.NewMockRepositoryRepository(ctrl)
	service := &service{RepoRegistry: repositoryRepository}

	repositoryRepository.EXPECT().
		Get(gomock.Any(), "repository-1").
		Return(&models.Repository{ID: "repository-1"}, nil)
	got, err := service.GetRepository(t.Context(), "repository-1")
	require.NoError(t, err)
	require.Equal(t, "repository-1", got.ID)

	repositoryRepository.EXPECT().
		Get(gomock.Any(), "missing").
		Return(nil, gorm.ErrRecordNotFound)
	_, err = service.GetRepository(t.Context(), "missing")
	require.Error(t, err)
}

func TestGetRepositoryByName(t *testing.T) {
	ctrl := gomock.NewController(t)
	repositoryRepository := reporegistry.NewMockRepositoryRepository(ctrl)
	service := &service{RepoRegistry: repositoryRepository}

	repositoryRepository.EXPECT().
		GetByName(gomock.Any(), "sigma/demo").
		Return(&models.Repository{ID: "repository-1", Name: "sigma/demo"}, nil)

	got, err := service.GetRepositoryByName(t.Context(), "sigma/demo")
	require.NoError(t, err)
	require.Equal(t, "sigma/demo", got.Name)
}

func TestListRepositories(t *testing.T) {
	ctrl := gomock.NewController(t)
	namespaceRepository := reponamespace.NewMockNamespaceRepository(ctrl)
	repositoryRepository := reporegistry.NewMockRepositoryRepository(ctrl)
	builderRepository := repobuilder.NewMockBuilderRepository(ctrl)
	service := &service{
		RepoNs:       namespaceRepository,
		RepoRegistry: repositoryRepository,
		RepoBuilder:  builderRepository,
	}
	items := []*models.Repository{{ID: "repository-1"}}
	builders := map[string]*models.Builder{"repository-1": {RepositoryID: "repository-1"}}

	namespaceRepository.EXPECT().Get(gomock.Any(), "namespace-1").Return(&models.Namespace{ID: "namespace-1"}, nil)
	repositoryRepository.EXPECT().
		ListRepositoryWithAuth(gomock.Any(), "namespace-1", "user-1", nil, api.Pagination{}, api.Sortable{}).
		Return(items, int64(1), nil)
	builderRepository.EXPECT().GetByRepositoryIDs(gomock.Any(), []string{"repository-1"}).Return(builders, nil)

	got, gotBuilders, total, err := service.ListRepositories(t.Context(), "user-1", "namespace-1", nil, api.Pagination{}, api.Sortable{})
	require.NoError(t, err)
	require.Equal(t, items, got)
	require.Equal(t, builders, gotBuilders)
	require.Equal(t, int64(1), total)
}

func TestListRepositoriesMapsNamespaceNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	namespaceRepository := reponamespace.NewMockNamespaceRepository(ctrl)
	namespaceRepository.EXPECT().Get(gomock.Any(), "missing").Return(nil, gorm.ErrRecordNotFound)

	_, _, _, err := (&service{RepoNs: namespaceRepository}).ListRepositories(t.Context(), "user-1", "missing", nil, api.Pagination{}, api.Sortable{})
	code, ok := errcode.AsType[errcode.ErrCode](err)
	require.True(t, ok)
	require.Equal(t, "NOT_FOUND", code.Code)
}

func TestUpdateRepositoryRejectsNamespaceMismatchAndSkipsEmptyUpdate(t *testing.T) {
	ctrl := gomock.NewController(t)
	namespaceRepository := reponamespace.NewMockNamespaceRepository(ctrl)
	repositoryRepository := reporegistry.NewMockRepositoryRepository(ctrl)
	service := &service{RepoNs: namespaceRepository, RepoRegistry: repositoryRepository}

	namespaceRepository.EXPECT().Get(gomock.Any(), "namespace-1").Return(&models.Namespace{ID: "namespace-1"}, nil)
	repositoryRepository.EXPECT().Get(gomock.Any(), "repository-1").Return(&models.Repository{ID: "repository-1", NamespaceID: "other"}, nil)
	err := service.UpdateRepository(t.Context(), "user-1", api.UpdateRepositoryRequest{ID: "repository-1", NamespaceID: "namespace-1"})
	code, ok := errcode.AsType[errcode.ErrCode](err)
	require.True(t, ok)
	require.Equal(t, "NOT_FOUND", code.Code)

	namespaceRepository.EXPECT().Get(gomock.Any(), "namespace-1").Return(&models.Namespace{ID: "namespace-1"}, nil)
	repositoryRepository.EXPECT().Get(gomock.Any(), "repository-1").Return(&models.Repository{ID: "repository-1", NamespaceID: "namespace-1"}, nil)
	err = service.UpdateRepository(t.Context(), "user-1", api.UpdateRepositoryRequest{ID: "repository-1", NamespaceID: "namespace-1"})
	require.NoError(t, err)
}

func TestGetRepositoryMapsUnexpectedError(t *testing.T) {
	ctrl := gomock.NewController(t)
	repositoryRepository := reporegistry.NewMockRepositoryRepository(ctrl)
	repositoryRepository.EXPECT().Get(gomock.Any(), "repository-1").Return(nil, errors.New("database unavailable"))

	_, err := (&service{RepoRegistry: repositoryRepository}).GetRepository(t.Context(), "repository-1")
	code, ok := errcode.AsType[errcode.ErrCode](err)
	require.True(t, ok)
	require.Equal(t, "INTERNAL_ERROR", code.Code)
}
