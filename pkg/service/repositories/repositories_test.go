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
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/dal/models"
	reporegistry "github.com/go-sigma/sigma/pkg/dal/repository/registry"
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
