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

package namespace_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	reponamespace "github.com/go-sigma/sigma/pkg/dal/repository/namespace"
	"github.com/go-sigma/sigma/pkg/testkit"
	"github.com/go-sigma/sigma/pkg/utils/uuid"
)

func TestNewNamespaceRepository(t *testing.T) {
	require.NotNil(t, reponamespace.NewNamespaceRepository())
	require.NotNil(t, reponamespace.NewNamespaceRepository(query.Q))
}

func TestNamespaceRepository(t *testing.T) {
	testkit.InitRepository(t)

	ctx := t.Context()
	namespaceRepository := reponamespace.NewNamespaceRepository()
	description := "updated"
	nameFilter := "test"
	namespaceObj := &models.Namespace{
		ID:         uuid.NewV7String(),
		Name:       "test",
		Visibility: enums.VisibilityPrivate,
	}
	require.NoError(t, namespaceRepository.Create(ctx, namespaceObj))

	got, err := namespaceRepository.Get(ctx, namespaceObj.ID)
	require.NoError(t, err)
	require.Equal(t, namespaceObj.Name, got.Name)

	got, err = namespaceRepository.GetByName(ctx, namespaceObj.Name)
	require.NoError(t, err)
	require.Equal(t, namespaceObj.ID, got.ID)

	namespaces, total, err := namespaceRepository.ListNamespace(ctx, &nameFilter, testPagination(), testSort("name"))
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, namespaces, 1)
	require.Equal(t, namespaceObj.ID, namespaces[0].ID)

	count, err := namespaceRepository.CountNamespace(ctx, &nameFilter)
	require.NoError(t, err)
	require.Equal(t, int64(1), count)

	require.NoError(t, namespaceRepository.UpdateByID(ctx, namespaceObj.ID, map[string]any{
		query.Namespace.Description.ColumnName().String(): description,
	}))
	got, err = namespaceRepository.Get(ctx, namespaceObj.ID)
	require.NoError(t, err)
	require.Equal(t, description, *got.Description)

	require.NoError(t, namespaceRepository.DeleteByID(ctx, namespaceObj.ID))
	err = namespaceRepository.DeleteByID(ctx, uuid.NewV7String())
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestNamespaceRepositoryQuota(t *testing.T) {
	testkit.InitRepository(t)

	ctx := t.Context()
	namespaceRepository := reponamespace.NewNamespaceRepository()
	namespaceObj := &models.Namespace{
		ID:   uuid.NewV7String(),
		Name: "quota",
	}
	require.NoError(t, namespaceRepository.Create(ctx, namespaceObj))

	require.NoError(t, namespaceRepository.UpdateQuota(ctx, namespaceObj.ID, 100))
	got, err := namespaceRepository.Get(ctx, namespaceObj.ID)
	require.NoError(t, err)
	require.Equal(t, int64(100), got.SizeLimit)

	err = namespaceRepository.UpdateQuota(ctx, uuid.NewV7String(), 100)
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestNamespaceRepositorySizeDirty(t *testing.T) {
	testkit.InitRepository(t)

	ctx := t.Context()
	namespaceRepository := reponamespace.NewNamespaceRepository()
	namespaceObj := &models.Namespace{
		ID:        uuid.NewV7String(),
		Name:      "size",
		SizeLimit: 100,
		Size:      10,
	}
	require.NoError(t, namespaceRepository.Create(ctx, namespaceObj))

	require.NoError(t, namespaceRepository.IncrementSize(ctx, namespaceObj.ID, 20))
	got, err := namespaceRepository.Get(ctx, namespaceObj.ID)
	require.NoError(t, err)
	require.Equal(t, int64(30), got.Size)
	require.True(t, got.SizeDirty)

	err = namespaceRepository.IncrementSize(ctx, namespaceObj.ID, 100)
	require.Error(t, err)

	dirty, err := namespaceRepository.FindWithCursorDirty(ctx, 10, "")
	require.NoError(t, err)
	require.Len(t, dirty, 1)
	require.Equal(t, namespaceObj.ID, dirty[0].ID)

	require.NoError(t, namespaceRepository.UpdateSizeAndClearDirty(ctx, namespaceObj.ID, 50))
	got, err = namespaceRepository.Get(ctx, namespaceObj.ID)
	require.NoError(t, err)
	require.Equal(t, int64(50), got.Size)
	require.False(t, got.SizeDirty)

	require.NoError(t, namespaceRepository.DecrementSize(ctx, namespaceObj.ID, 60))
	got, err = namespaceRepository.Get(ctx, namespaceObj.ID)
	require.NoError(t, err)
	require.Zero(t, got.Size)
	require.True(t, got.SizeDirty)

	require.NoError(t, namespaceRepository.ClearSizeDirty(ctx, namespaceObj.ID))
	got, err = namespaceRepository.Get(ctx, namespaceObj.ID)
	require.NoError(t, err)
	require.False(t, got.SizeDirty)
}

func TestNamespaceRepositoryListWithAuth(t *testing.T) {
	testkit.InitRepository(t)

	ctx := t.Context()
	namespaceRepository := reponamespace.NewNamespaceRepository()
	memberRepository := reponamespace.NewNamespaceMemberRepository()
	normalUser := &models.User{ID: uuid.NewV7String(), Username: "normal", Role: enums.UserRoleUser}
	adminUser := &models.User{ID: uuid.NewV7String(), Username: "admin", Role: enums.UserRoleAdmin}
	require.NoError(t, query.Q.User.WithContext(ctx).Create(normalUser))
	require.NoError(t, query.Q.User.WithContext(ctx).Create(adminUser))

	publicNamespace := &models.Namespace{ID: uuid.NewV7String(), Name: "auth-public", Visibility: enums.VisibilityPublic}
	privateNamespace := &models.Namespace{ID: uuid.NewV7String(), Name: "auth-private", Visibility: enums.VisibilityPrivate}
	hiddenNamespace := &models.Namespace{ID: uuid.NewV7String(), Name: "auth-hidden", Visibility: enums.VisibilityPrivate}
	require.NoError(t, namespaceRepository.Create(ctx, publicNamespace))
	require.NoError(t, namespaceRepository.Create(ctx, privateNamespace))
	require.NoError(t, namespaceRepository.Create(ctx, hiddenNamespace))
	_, err := memberRepository.AddNamespaceMember(ctx, normalUser.ID, *privateNamespace, enums.NamespaceRoleReader)
	require.NoError(t, err)

	nameFilter := "auth-"
	namespaces, total, err := namespaceRepository.ListNamespaceWithAuth(ctx, "", &nameFilter, testPagination(), testSort("name"))
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Equal(t, []string{publicNamespace.Name}, namespaceNames(namespaces))

	namespaces, total, err = namespaceRepository.ListNamespaceWithAuth(ctx, normalUser.ID, &nameFilter, testPagination(), testSort("name"))
	require.NoError(t, err)
	require.Equal(t, int64(2), total)
	require.Equal(t, []string{privateNamespace.Name, publicNamespace.Name}, namespaceNames(namespaces))

	namespaces, total, err = namespaceRepository.ListNamespaceWithAuth(ctx, adminUser.ID, &nameFilter, testPagination(), testSort("name"))
	require.NoError(t, err)
	require.Equal(t, int64(3), total)
	require.Equal(t, []string{hiddenNamespace.Name, privateNamespace.Name, publicNamespace.Name}, namespaceNames(namespaces))
}

func namespaceNames(namespaces []*models.Namespace) []string {
	names := make([]string, 0, len(namespaces))
	for _, namespaceObj := range namespaces {
		names = append(names, namespaceObj.Name)
	}
	return names
}

func testPagination() api.Pagination {
	page := 1
	limit := 10
	return api.Pagination{Page: &page, Limit: &limit}
}

func testSort(column string) api.Sortable {
	method := enums.SortMethodAsc
	return api.Sortable{Sort: &column, Method: &method}
}
