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

package registry_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	reporegistry "github.com/go-sigma/sigma/pkg/dal/repository/registry"
	"github.com/go-sigma/sigma/pkg/testkit"
	"github.com/go-sigma/sigma/pkg/utils/uuid"
)

func TestNewRepositoryRepository(t *testing.T) {
	require.NotNil(t, reporegistry.NewRepositoryRepository())
	require.NotNil(t, reporegistry.NewRepositoryRepository(query.Q))
}

func TestRepositoryFindDeletableWithCursor(t *testing.T) {
	digCon := testkit.InitRepository(t)
	require.NotNil(t, digCon)

	ctx := t.Context()
	namespaceObj := &models.Namespace{ID: uuid.NewV7String(), Name: "gc-repository"}
	require.NoError(t, query.Q.Namespace.WithContext(ctx).Create(namespaceObj))

	oldUpdatedAt := time.Now().Add(-48 * time.Hour).UnixMilli()
	emptyRepository := &models.Repository{ID: uuid.NewV7String(), Name: "gc-repository/empty", NamespaceID: namespaceObj.ID, UpdatedAt: oldUpdatedAt}
	taggedRepository := &models.Repository{ID: uuid.NewV7String(), Name: "gc-repository/tagged", NamespaceID: namespaceObj.ID, UpdatedAt: oldUpdatedAt}
	require.NoError(t, query.Q.Repository.WithContext(ctx).Create(emptyRepository))
	require.NoError(t, query.Q.Repository.WithContext(ctx).Create(taggedRepository))

	artifactObj := &models.Artifact{
		ID:           uuid.NewV7String(),
		NamespaceID:  namespaceObj.ID,
		RepositoryID: taggedRepository.ID,
		Digest:       "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	}
	require.NoError(t, query.Q.Artifact.WithContext(ctx).Create(artifactObj))
	require.NoError(t, query.Q.Tag.WithContext(ctx).Create(&models.Tag{
		ID:           uuid.NewV7String(),
		RepositoryID: taggedRepository.ID,
		ArtifactID:   artifactObj.ID,
		Name:         "latest",
	}))

	repositories, err := reporegistry.NewRepositoryRepository().
		FindDeletableWithCursor(ctx, namespaceObj.ID, time.Now().UnixMilli(), "", 10)
	require.NoError(t, err)
	require.Len(t, repositories, 1)
	require.Equal(t, emptyRepository.ID, repositories[0].ID)

	repositories, err = reporegistry.NewRepositoryRepository().
		FindDeletableWithCursor(ctx, namespaceObj.ID, time.Now().UnixMilli(), emptyRepository.ID, 10)
	require.NoError(t, err)
	require.Empty(t, repositories)
}

func TestRepositoryRepository(t *testing.T) {
	fixture := newRegistryFixture(t)
	ctx := t.Context()
	repositoryRepository := fixture.RepositoryRepository
	description := "updated"
	nameFilter := fixture.Repository.Name

	got, err := repositoryRepository.Get(ctx, fixture.Repository.ID)
	require.NoError(t, err)
	require.Equal(t, fixture.Repository.Name, got.Name)

	got, err = repositoryRepository.GetByName(ctx, fixture.Repository.Name)
	require.NoError(t, err)
	require.Equal(t, fixture.Repository.ID, got.ID)

	count, err := repositoryRepository.CountRepository(ctx, fixture.Namespace.ID, nil)
	require.NoError(t, err)
	require.Equal(t, int64(1), count)

	repositories, total, err := repositoryRepository.ListRepository(ctx, fixture.Namespace.ID, &nameFilter, registryPagination(), registrySort("name", enums.SortMethodAsc))
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, repositories, 1)
	require.Equal(t, fixture.Repository.ID, repositories[0].ID)

	repositories, err = repositoryRepository.ListByDtPagination(ctx, 10)
	require.NoError(t, err)
	require.NotEmpty(t, repositories)

	require.NoError(t, repositoryRepository.UpdateRepository(ctx, fixture.Repository.ID, map[string]any{
		query.Repository.Description.ColumnName().String(): description,
		query.Repository.Overview.ColumnName().String():    []byte("overview"),
	}))
	got, err = repositoryRepository.Get(ctx, fixture.Repository.ID)
	require.NoError(t, err)
	require.Equal(t, description, *got.Description)
	require.Equal(t, []byte("overview"), got.Overview)

	countByNamespace, err := repositoryRepository.CountByNamespace(ctx, []string{fixture.Namespace.ID})
	require.NoError(t, err)
	require.Equal(t, int64(1), countByNamespace[fixture.Namespace.ID])

	emptyRepository := &models.Repository{
		ID:          uuid.NewV7String(),
		NamespaceID: fixture.Namespace.ID,
		Name:        fixture.Namespace.Name + "/empty",
	}
	require.NoError(t, repositoryRepository.Create(ctx, emptyRepository))
	deletedNames, err := repositoryRepository.DeleteEmpty(ctx, &fixture.Namespace.ID)
	require.NoError(t, err)
	require.Contains(t, deletedNames, emptyRepository.Name)

	require.NoError(t, repositoryRepository.DeleteByID(ctx, fixture.Repository.ID))
	_, err = repositoryRepository.Get(ctx, fixture.Repository.ID)
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestRepositoryRepositorySizeDirty(t *testing.T) {
	fixture := newRegistryFixture(t)
	ctx := t.Context()
	repositoryRepository := fixture.RepositoryRepository

	require.NoError(t, repositoryRepository.UpdateRepository(ctx, fixture.Repository.ID, map[string]any{
		query.Repository.SizeLimit.ColumnName().String(): int64(100),
		query.Repository.Size.ColumnName().String():      int64(10),
	}))

	require.NoError(t, repositoryRepository.IncrementSize(ctx, fixture.Repository.ID, 20))
	got, err := repositoryRepository.Get(ctx, fixture.Repository.ID)
	require.NoError(t, err)
	require.Equal(t, int64(30), got.Size)
	require.True(t, got.SizeDirty)

	err = repositoryRepository.IncrementSize(ctx, fixture.Repository.ID, 100)
	require.Error(t, err)

	dirty, err := repositoryRepository.FindWithCursorDirty(ctx, 10, "")
	require.NoError(t, err)
	require.Len(t, dirty, 1)
	require.Equal(t, fixture.Repository.ID, dirty[0].ID)

	require.NoError(t, repositoryRepository.UpdateSizeAndClearDirty(ctx, fixture.Repository.ID, 50))
	got, err = repositoryRepository.Get(ctx, fixture.Repository.ID)
	require.NoError(t, err)
	require.Equal(t, int64(50), got.Size)
	require.False(t, got.SizeDirty)

	require.NoError(t, repositoryRepository.DecrementSize(ctx, fixture.Repository.ID, 60))
	got, err = repositoryRepository.Get(ctx, fixture.Repository.ID)
	require.NoError(t, err)
	require.Zero(t, got.Size)
	require.True(t, got.SizeDirty)

	require.NoError(t, repositoryRepository.ClearSizeDirty(ctx, fixture.Repository.ID))
	got, err = repositoryRepository.Get(ctx, fixture.Repository.ID)
	require.NoError(t, err)
	require.False(t, got.SizeDirty)
}

func TestRepositoryFindDeletableWithCursorSQLiteSQL(t *testing.T) {
	testkit.RequireSQLiteDialect(t)

	digCon := testkit.InitRepository(t)
	require.NotNil(t, digCon)

	sqlLogger := &captureSQLLogger{Interface: logger.Discard}
	dryRunDB := query.Q.UnderlyingDB().Session(&gorm.Session{
		DryRun: true,
		Logger: sqlLogger,
	})
	dryRunQuery := query.Use(dryRunDB)

	const before = int64(1234567890)
	_, err := reporegistry.NewRepositoryRepository(dryRunQuery).
		FindDeletableWithCursor(t.Context(), "1", before, "0", 10)
	require.NoError(t, err)

	require.Equal(t, "SELECT * FROM `repositories` WHERE `repositories`.`id` > \"0\" AND `repositories`.`namespace_id` = \"1\" AND `repositories`.`updated_at` < 1234567890 AND NOT EXISTS (SELECT `tags`.`id` FROM `tags` WHERE `tags`.`repository_id` = `repositories`.`id` AND `tags`.`deleted_at` = 0) AND `repositories`.`deleted_at` = 0 ORDER BY `repositories`.`id` LIMIT 10", sqlLogger.sql)
}

type captureSQLLogger struct {
	logger.Interface
	sql string
}

func (l *captureSQLLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	l.sql, _ = fc()
}
