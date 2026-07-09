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

	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	reporegistry "github.com/go-sigma/sigma/pkg/dal/repository/registry"
	"github.com/go-sigma/sigma/pkg/testkit"
	"github.com/go-sigma/sigma/pkg/utils/uuid"
)

func TestNewBlobRepository(t *testing.T) {
	require.NotNil(t, reporegistry.NewBlobRepository())
	require.NotNil(t, reporegistry.NewBlobRepository(query.Q))
}

func TestBlobFindDeletableWithCursorSQLiteSQL(t *testing.T) {
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
	_, err := reporegistry.NewBlobRepository(dryRunQuery).
		FindDeletableWithCursor(t.Context(), before, "0", 10)
	require.NoError(t, err)

	require.Equal(t, "SELECT * FROM `blobs` WHERE `blobs`.`id` > \"0\" AND (`blobs`.`last_pull` < 1234567890 OR (`blobs`.`last_pull` IS NULL AND `blobs`.`updated_at` < 1234567890)) AND NOT EXISTS (SELECT `artifacts`.`id` FROM `artifacts` INNER JOIN `artifact_blobs` ON `artifacts`.`id` = `artifact_blobs`.`artifact_id` WHERE `artifact_blobs`.`blob_id` = `blobs`.`id` AND `artifacts`.`deleted_at` = 0) AND `blobs`.`deleted_at` = 0 ORDER BY `blobs`.`id` LIMIT 10", sqlLogger.sql)
}

func TestBlobFindAssociateWithArtifactSQLiteSQL(t *testing.T) {
	testkit.RequireSQLiteDialect(t)

	digCon := testkit.InitRepository(t)
	require.NotNil(t, digCon)

	sqlLogger := &captureSQLLogger{Interface: logger.Discard}
	dryRunDB := query.Q.UnderlyingDB().Session(&gorm.Session{
		DryRun: true,
		Logger: sqlLogger,
	})
	dryRunQuery := query.Use(dryRunDB)

	_, err := reporegistry.NewBlobRepository(dryRunQuery).
		FindAssociateWithArtifact(t.Context(), []string{"1", "2", "3"})
	require.NoError(t, err)

	require.Equal(t, "SELECT * FROM `blobs` WHERE `blobs`.`id` IN (\"1\",\"2\",\"3\") AND EXISTS (SELECT `artifacts`.`id` FROM `artifacts` INNER JOIN `artifact_blobs` ON `artifacts`.`id` = `artifact_blobs`.`artifact_id` WHERE `artifact_blobs`.`blob_id` = `blobs`.`id` AND `artifacts`.`deleted_at` = 0) AND `blobs`.`deleted_at` = 0", sqlLogger.sql)
}

func TestBlobRepository(t *testing.T) {
	fixture := newRegistryFixture(t)
	ctx := t.Context()
	blobRepository := fixture.BlobRepository
	artifactRepository := fixture.ArtifactRepository
	oldUpdatedAt := time.Now().Add(-2 * time.Hour).UnixMilli()
	blobOne := &models.Blob{
		ID:          uuid.NewV7String(),
		Digest:      "sha256:blob-one",
		Size:        123,
		ContentType: "application/octet-stream",
		UpdatedAt:   oldUpdatedAt,
	}
	blobTwo := &models.Blob{
		ID:          uuid.NewV7String(),
		Digest:      "sha256:blob-two",
		Size:        234,
		ContentType: "application/octet-stream",
		UpdatedAt:   oldUpdatedAt,
	}
	require.NoError(t, blobRepository.Create(ctx, blobOne))
	require.NoError(t, blobRepository.Create(ctx, blobTwo))

	got, err := blobRepository.FindByDigest(ctx, blobOne.Digest)
	require.NoError(t, err)
	require.Equal(t, blobOne.Size, got.Size)

	blobs, err := blobRepository.FindByDigests(ctx, []string{blobOne.Digest, blobTwo.Digest})
	require.NoError(t, err)
	require.Len(t, blobs, 2)

	exists, err := blobRepository.Exists(ctx, blobOne.Digest)
	require.NoError(t, err)
	require.True(t, exists)
	exists, err = blobRepository.Exists(ctx, "sha256:missing")
	require.NoError(t, err)
	require.False(t, exists)

	oldBlobs, err := blobRepository.FindWithLastPull(ctx, time.Now().UnixMilli(), "", 10)
	require.NoError(t, err)
	require.Len(t, oldBlobs, 2)

	artifactObj := newRegistryArtifact(fixture.Namespace.ID, fixture.Repository.ID, "sha256:artifact-with-blob")
	require.NoError(t, artifactRepository.Create(ctx, artifactObj))
	require.NoError(t, artifactRepository.AssociateBlobs(ctx, artifactObj, []*models.Blob{blobOne}))
	associatedIDs, err := blobRepository.FindAssociateWithArtifact(ctx, []string{blobOne.ID, blobTwo.ID})
	require.NoError(t, err)
	require.Equal(t, []string{blobOne.ID}, associatedIDs)

	require.NoError(t, blobRepository.Incr(ctx, blobOne.ID))
	got, err = blobRepository.FindByDigest(ctx, blobOne.Digest)
	require.NoError(t, err)
	require.Equal(t, uint(1), got.PullTimes)
	require.NotZero(t, got.LastPull)

	require.NoError(t, blobRepository.DeleteByID(ctx, blobTwo.ID))
	err = blobRepository.DeleteByID(ctx, uuid.NewV7String())
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestBlobRepositoryUpload(t *testing.T) {
	testkit.InitRepository(t)

	ctx := t.Context()
	blobRepository := reporegistry.NewBlobRepository()
	uploadID := uuid.NewV7String()
	partOne := &models.BlobUpload{
		ID:         uuid.NewV7String(),
		UploadID:   uploadID,
		PartNumber: 1,
		Etag:       "etag-1",
		Repository: "library/alpine",
		FileID:     "file-1",
		Size:       10,
	}
	partTwo := &models.BlobUpload{
		ID:         uuid.NewV7String(),
		UploadID:   uploadID,
		PartNumber: 2,
		Etag:       "etag-2",
		Repository: "library/alpine",
		FileID:     "file-2",
		Size:       20,
	}
	require.NoError(t, blobRepository.UploadCreate(ctx, partOne))
	require.NoError(t, blobRepository.UploadCreate(ctx, partTwo))

	lastPart, err := blobRepository.UploadGetLastPart(ctx, uploadID)
	require.NoError(t, err)
	require.Equal(t, partTwo.ID, lastPart.ID)

	parts, err := blobRepository.UploadFindAllByUploadID(ctx, uploadID)
	require.NoError(t, err)
	require.Equal(t, []int32{1, 2}, []int32{parts[0].PartNumber, parts[1].PartNumber})

	totalSize, err := blobRepository.UploadTotalSizeByUploadID(ctx, uploadID)
	require.NoError(t, err)
	require.Equal(t, int64(30), totalSize)

	etags, err := blobRepository.UploadTotalEtagsByUploadID(ctx, uploadID)
	require.NoError(t, err)
	require.Equal(t, []string{"etag-2"}, etags)

	require.NoError(t, blobRepository.UploadDeleteByUploadID(ctx, uploadID))
	parts, err = blobRepository.UploadFindAllByUploadID(ctx, uploadID)
	require.NoError(t, err)
	require.Empty(t, parts)
}
