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

package artifacts

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/dal/models"
	reporegistry "github.com/go-sigma/sigma/pkg/dal/repository/registry"
)

func TestListArtifacts(t *testing.T) {
	ctrl := gomock.NewController(t)
	artifactRepository := reporegistry.NewMockArtifactRepository(ctrl)
	request := api.ListArtifactRequest{Repository: "sigma/demo"}
	items := []*models.Artifact{{ID: "artifact-1"}}

	artifactRepository.EXPECT().ListArtifact(gomock.Any(), request).Return(items, nil)
	artifactRepository.EXPECT().CountArtifact(gomock.Any(), request).Return(int64(1), nil)

	got, total, err := (&service{RepoArtifact: artifactRepository}).
		ListArtifacts(t.Context(), request)
	require.NoError(t, err)
	require.Equal(t, items, got)
	require.Equal(t, int64(1), total)
}

func TestGetArtifact(t *testing.T) {
	ctrl := gomock.NewController(t)
	repositoryRepository := reporegistry.NewMockRepositoryRepository(ctrl)
	artifactRepository := reporegistry.NewMockArtifactRepository(ctrl)
	service := &service{
		RepoRegistry: repositoryRepository,
		RepoArtifact: artifactRepository,
	}
	repositoryRepository.EXPECT().
		GetByName(gomock.Any(), "sigma/demo").
		Return(&models.Repository{ID: "repository-1"}, nil)
	artifactRepository.EXPECT().
		GetByDigest(gomock.Any(), "repository-1", "sha256:digest").
		Return(&models.Artifact{ID: "artifact-1"}, nil)

	got, err := service.GetArtifact(t.Context(), "sigma/demo", "sha256:digest")
	require.NoError(t, err)
	require.Equal(t, "artifact-1", got.ID)

	repositoryRepository.EXPECT().
		GetByName(gomock.Any(), "missing").
		Return(nil, gorm.ErrRecordNotFound)
	_, err = service.GetArtifact(t.Context(), "missing", "sha256:digest")
	require.Error(t, err)
}

func TestDeleteArtifactError(t *testing.T) {
	ctrl := gomock.NewController(t)
	artifactRepository := reporegistry.NewMockArtifactRepository(ctrl)
	artifactRepository.EXPECT().
		DeleteByDigest(gomock.Any(), "sigma/demo", "sha256:digest").
		Return(errors.New("delete failed"))

	err := (&service{RepoArtifact: artifactRepository}).
		DeleteArtifact(t.Context(), "sigma/demo", "sha256:digest")
	require.Error(t, err)
}
