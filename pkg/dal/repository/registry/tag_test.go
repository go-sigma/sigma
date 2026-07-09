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

func TestNewTagRepository(t *testing.T) {
	require.NotNil(t, reporegistry.NewTagRepository())
	require.NotNil(t, reporegistry.NewTagRepository(query.Q))
}

func TestTagFindWithQuantityCursorSQLiteSQL(t *testing.T) {
	testkit.RequireSQLiteDialect(t)

	digCon := testkit.InitRepository(t)
	require.NotNil(t, digCon)

	sqlLogger := &captureSQLLogger{Interface: logger.Discard}
	dryRunDB := query.Q.UnderlyingDB().Session(&gorm.Session{
		DryRun: true,
		Logger: sqlLogger,
	})
	dryRunQuery := query.Use(dryRunDB)

	_, err := reporegistry.NewTagRepository(dryRunQuery).
		FindWithQuantityCursor(t.Context(), "1", 3, 10, "5")
	require.NoError(t, err)

	require.Equal(t, "SELECT * FROM `tags` WHERE `tags`.`repository_id` = \"1\" AND `tags`.`id` > \"5\" AND NOT `tags`.`id` IN (SELECT `tags`.`id` FROM `tags` WHERE `tags`.`repository_id` = \"1\" AND `tags`.`deleted_at` = 0 ORDER BY `tags`.`updated_at` DESC,`tags`.`id` DESC LIMIT 3) AND `tags`.`deleted_at` = 0 ORDER BY `tags`.`id` LIMIT 10", sqlLogger.sql)
}

func TestTagRepository(t *testing.T) {
	fixture := newRegistryFixture(t)
	ctx := t.Context()
	tagRepository := fixture.TagRepository
	artifactRepository := fixture.ArtifactRepository
	artifactOne := newRegistryArtifact(fixture.Namespace.ID, fixture.Repository.ID, "sha256:tag-one")
	artifactTwo := newRegistryArtifact(fixture.Namespace.ID, fixture.Repository.ID, "sha256:tag-two")
	require.NoError(t, artifactRepository.Create(ctx, artifactOne))
	require.NoError(t, artifactRepository.Create(ctx, artifactTwo))

	tagObj := &models.Tag{
		ID:           uuid.NewV7String(),
		RepositoryID: fixture.Repository.ID,
		ArtifactID:   artifactOne.ID,
		Name:         "latest",
	}
	require.NoError(t, tagRepository.Create(ctx, tagObj))

	got, err := tagRepository.GetByID(ctx, tagObj.ID)
	require.NoError(t, err)
	require.Equal(t, tagObj.ID, got.ID)

	got, err = tagRepository.GetByName(ctx, fixture.Repository.ID, tagObj.Name)
	require.NoError(t, err)
	require.Equal(t, tagObj.ID, got.ID)
	require.Equal(t, artifactOne.ID, got.Artifact.ID)

	got, err = tagRepository.GetByArtifactID(ctx, fixture.Repository.ID, artifactOne.ID)
	require.NoError(t, err)
	require.Equal(t, tagObj.ID, got.ID)

	require.NoError(t, tagRepository.Incr(ctx, tagObj.ID))
	got, err = tagRepository.GetByID(ctx, tagObj.ID)
	require.NoError(t, err)
	require.Equal(t, int64(1), got.PullTimes)

	nameFilter := "late"
	tags, total, err := tagRepository.ListTag(ctx, fixture.Repository.ID, &nameFilter, []enums.ArtifactType{enums.ArtifactTypeImage}, registryPagination(), registrySort("name", enums.SortMethodAsc))
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, tags, 1)
	require.Equal(t, tagObj.ID, tags[0].ID)

	tags, err = tagRepository.ListByDtPagination(ctx, fixture.Repository.Name, 10)
	require.NoError(t, err)
	require.Len(t, tags, 1)

	countByArtifact, err := tagRepository.CountByArtifact(ctx, []string{artifactOne.ID})
	require.NoError(t, err)
	require.Equal(t, int64(1), countByArtifact[artifactOne.ID])

	countByNamespace, err := tagRepository.CountByNamespace(ctx, []string{fixture.Namespace.ID})
	require.NoError(t, err)
	require.Equal(t, int64(1), countByNamespace[fixture.Namespace.ID])

	countByRepositories, err := tagRepository.CountByRepositories(ctx, []string{fixture.Repository.ID})
	require.NoError(t, err)
	require.Equal(t, int64(1), countByRepositories[fixture.Repository.ID])

	countByRepository, err := tagRepository.CountByRepository(ctx, fixture.Repository.ID)
	require.NoError(t, err)
	require.Equal(t, int64(1), countByRepository)

	replacement := &models.Tag{
		ID:           uuid.NewV7String(),
		RepositoryID: fixture.Repository.ID,
		ArtifactID:   artifactTwo.ID,
		Name:         "latest",
	}
	require.NoError(t, tagRepository.Create(ctx, replacement))
	got, err = tagRepository.GetByName(ctx, fixture.Repository.ID, "latest")
	require.NoError(t, err)
	require.Equal(t, artifactTwo.ID, got.ArtifactID)

	require.NoError(t, tagRepository.DeleteByArtifactID(ctx, artifactTwo.ID))
	_, err = tagRepository.GetByName(ctx, fixture.Repository.ID, "latest")
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)

	tagByID := &models.Tag{
		ID:           uuid.NewV7String(),
		RepositoryID: fixture.Repository.ID,
		ArtifactID:   artifactOne.ID,
		Name:         "delete-by-id",
	}
	require.NoError(t, tagRepository.Create(ctx, tagByID))
	require.NoError(t, tagRepository.DeleteByID(ctx, tagByID.ID))
	err = tagRepository.DeleteByID(ctx, uuid.NewV7String())
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestTagRepositoryFindWithDayCursor(t *testing.T) {
	fixture := newRegistryFixture(t)
	ctx := t.Context()
	artifactObj := newRegistryArtifact(fixture.Namespace.ID, fixture.Repository.ID, "sha256:old-tag")
	require.NoError(t, fixture.ArtifactRepository.Create(ctx, artifactObj))
	oldTag := &models.Tag{
		ID:           uuid.NewV7String(),
		RepositoryID: fixture.Repository.ID,
		ArtifactID:   artifactObj.ID,
		Name:         "old",
		UpdatedAt:    time.Now().Add(-48 * time.Hour).UnixMilli(),
	}
	require.NoError(t, query.Q.Tag.WithContext(ctx).Create(oldTag))

	tags, err := fixture.TagRepository.FindWithDayCursor(ctx, fixture.Repository.ID, 1, 10, "")
	require.NoError(t, err)
	require.Len(t, tags, 1)
	require.Equal(t, oldTag.ID, tags[0].ID)
}
