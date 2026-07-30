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

package namespaces

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/dal/models"
	reponamespace "github.com/go-sigma/sigma/pkg/dal/repository/namespace"
	reporegistry "github.com/go-sigma/sigma/pkg/dal/repository/registry"
)

func TestListNamespaces(t *testing.T) {
	ctrl := gomock.NewController(t)
	namespaceRepository := reponamespace.NewMockNamespaceRepository(ctrl)
	items := []*models.Namespace{{ID: "namespace-1"}}
	namespaceRepository.EXPECT().
		ListNamespaceWithAuth(gomock.Any(), "user-1", nil, gomock.Any(), api.Sortable{}).
		Return(items, int64(1), nil)

	got, total, err := (&service{RepoNs: namespaceRepository}).
		ListNamespaces(t.Context(), "user-1", nil, api.Pagination{}, api.Sortable{})
	require.NoError(t, err)
	require.Equal(t, items, got)
	require.Equal(t, int64(1), total)
}

func TestGetNamespace(t *testing.T) {
	ctrl := gomock.NewController(t)
	namespaceRepository := reponamespace.NewMockNamespaceRepository(ctrl)
	repositoryRepository := reporegistry.NewMockRepositoryRepository(ctrl)
	tagRepository := reporegistry.NewMockTagRepository(ctrl)
	service := &service{
		RepoNs:       namespaceRepository,
		RepoRegistry: repositoryRepository,
		RepoTag:      tagRepository,
	}
	namespaceRepository.EXPECT().
		Get(gomock.Any(), "namespace-1").
		Return(&models.Namespace{ID: "namespace-1"}, nil)
	repositoryRepository.EXPECT().
		CountByNamespace(gomock.Any(), []string{"namespace-1"}).
		Return(map[string]int64{"namespace-1": 2}, nil)
	tagRepository.EXPECT().
		CountByNamespace(gomock.Any(), []string{"namespace-1"}).
		Return(map[string]int64{"namespace-1": 3}, nil)

	namespace, repositoryCount, tagCount, err := service.GetNamespace(t.Context(), "namespace-1")
	require.NoError(t, err)
	require.Equal(t, "namespace-1", namespace.ID)
	require.Equal(t, int64(2), repositoryCount)
	require.Equal(t, int64(3), tagCount)
}

func TestNamespaceMemberQueries(t *testing.T) {
	ctrl := gomock.NewController(t)
	memberRepository := reponamespace.NewMockNamespaceMemberRepository(ctrl)
	service := &service{RepoNsMember: memberRepository}
	members := []*models.NamespaceMember{{ID: "member-1"}}

	memberRepository.EXPECT().
		ListNamespaceMembers(gomock.Any(), "namespace-1", nil, api.Pagination{}, api.Sortable{}).
		Return(members, int64(1), nil)
	got, total, err := service.ListNamespaceMembers(
		t.Context(), "namespace-1", nil, api.Pagination{}, api.Sortable{},
	)
	require.NoError(t, err)
	require.Equal(t, members, got)
	require.Equal(t, int64(1), total)

	memberRepository.EXPECT().
		GetNamespaceMember(gomock.Any(), "namespace-1", "user-1").
		Return(members[0], nil)
	member, err := service.GetNamespaceMember(t.Context(), "namespace-1", "user-1")
	require.NoError(t, err)
	require.Equal(t, "member-1", member.ID)
}
