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

package tags

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/dal/models"
	reponamespace "github.com/go-sigma/sigma/pkg/dal/repository/namespace"
	reporegistry "github.com/go-sigma/sigma/pkg/dal/repository/registry"
)

func TestListTags(t *testing.T) {
	ctrl := gomock.NewController(t)
	namespaceRepository := reponamespace.NewMockNamespaceRepository(ctrl)
	repositoryRepository := reporegistry.NewMockRepositoryRepository(ctrl)
	tagRepository := reporegistry.NewMockTagRepository(ctrl)
	service := &tagService{
		namespaceRepository:  namespaceRepository,
		repositoryRepository: repositoryRepository,
		tagRepository:        tagRepository,
	}
	namespaceRepository.EXPECT().
		Get(gomock.Any(), "namespace-1").
		Return(&models.Namespace{ID: "namespace-1"}, nil)
	repositoryRepository.EXPECT().
		Get(gomock.Any(), "repository-1").
		Return(&models.Repository{ID: "repository-1", NamespaceID: "namespace-1"}, nil)
	tagRepository.EXPECT().
		ListTag(gomock.Any(), "repository-1", nil, nil, api.Pagination{}, api.Sortable{}).
		Return([]*models.Tag{{ID: "tag-1"}}, int64(1), nil)

	items, total, err := service.ListTags(
		t.Context(), "namespace-1", "repository-1", nil, nil, api.Pagination{}, api.Sortable{},
	)
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, int64(1), total)
}

func TestGetTagAndInvalidArtifactDigest(t *testing.T) {
	ctrl := gomock.NewController(t)
	tagRepository := reporegistry.NewMockTagRepository(ctrl)
	service := &tagService{tagRepository: tagRepository}
	tagRepository.EXPECT().
		GetByID(gomock.Any(), "tag-1").
		Return(&models.Tag{ID: "tag-1"}, nil)

	tag, err := service.GetTag(t.Context(), "tag-1")
	require.NoError(t, err)
	require.Equal(t, "tag-1", tag.ID)

	_, err = service.GetArtifactRaw(t.Context(), "invalid")
	require.Error(t, err)
}
