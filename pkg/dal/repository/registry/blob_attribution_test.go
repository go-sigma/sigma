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

package registry_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/go-sigma/sigma/pkg/dal/models"
	reporegistry "github.com/go-sigma/sigma/pkg/dal/repository/registry"
	"github.com/go-sigma/sigma/pkg/utils/uuid"
)

func TestBlobAttributionRepository(t *testing.T) {
	fixture := newRegistryFixture(t)
	ctx := t.Context()
	attributionRepository := reporegistry.NewBlobAttributionRepository()

	blobOne := &models.Blob{
		ID:          uuid.NewV7String(),
		Digest:      "sha256:attribution-one",
		Size:        100,
		ContentType: "application/octet-stream",
	}
	blobTwo := &models.Blob{
		ID:          uuid.NewV7String(),
		Digest:      "sha256:attribution-two",
		Size:        50,
		ContentType: "application/octet-stream",
	}
	require.NoError(t, fixture.BlobRepository.Create(ctx, blobOne))
	require.NoError(t, fixture.BlobRepository.Create(ctx, blobTwo))

	artifactObj := newRegistryArtifact(fixture.Namespace.ID, fixture.Repository.ID, "sha256:attribution-artifact")
	require.NoError(t, fixture.ArtifactRepository.Create(ctx, artifactObj))
	require.NoError(t, fixture.ArtifactRepository.AssociateBlobs(ctx, artifactObj, []*models.Blob{blobOne, blobTwo}))

	blobs := []*models.Blob{blobOne, blobTwo}

	firstBatch := uuid.NewV7String()
	require.NoError(t, attributionRepository.UpsertBatch(ctx, reporegistry.ScopeTypeNamespace, fixture.Namespace.ID, blobs, firstBatch))
	delta, err := attributionRepository.SumBatch(ctx, reporegistry.ScopeTypeNamespace, fixture.Namespace.ID, firstBatch)
	require.NoError(t, err)
	require.Equal(t, int64(150), delta)

	// Re-inserting the same scope/blob pairs is idempotent: the old rows keep
	// their batch_id, so the new batch contributes nothing.
	secondBatch := uuid.NewV7String()
	require.NoError(t, attributionRepository.UpsertBatch(ctx, reporegistry.ScopeTypeNamespace, fixture.Namespace.ID, blobs, secondBatch))
	delta, err = attributionRepository.SumBatch(ctx, reporegistry.ScopeTypeNamespace, fixture.Namespace.ID, secondBatch)
	require.NoError(t, err)
	require.Zero(t, delta)

	repositoryBatch := uuid.NewV7String()
	require.NoError(t, attributionRepository.UpsertBatch(ctx, reporegistry.ScopeTypeRepository, fixture.Repository.ID, blobs, repositoryBatch))
	delta, err = attributionRepository.SumBatch(ctx, reporegistry.ScopeTypeRepository, fixture.Repository.ID, repositoryBatch)
	require.NoError(t, err)
	require.Equal(t, int64(150), delta)

	namespaceSize, err := attributionRepository.SumByScope(ctx, reporegistry.ScopeTypeNamespace, fixture.Namespace.ID)
	require.NoError(t, err)
	require.Equal(t, int64(150), namespaceSize)

	repositorySize, err := attributionRepository.SumByScope(ctx, reporegistry.ScopeTypeRepository, fixture.Repository.ID)
	require.NoError(t, err)
	require.Equal(t, int64(150), repositorySize)

	// The artifact is still reachable, so reconciliation keeps the size.
	namespaceSize, err = attributionRepository.SyncAndSum(ctx, reporegistry.ScopeTypeNamespace, fixture.Namespace.ID)
	require.NoError(t, err)
	require.Equal(t, int64(150), namespaceSize)

	// Soft-deleting the artifact makes the blobs unreachable; reconciliation
	// removes the attribution rows and the size drops to zero.
	require.NoError(t, fixture.ArtifactRepository.DeleteByID(ctx, artifactObj.ID))
	namespaceSize, err = attributionRepository.SyncAndSum(ctx, reporegistry.ScopeTypeNamespace, fixture.Namespace.ID)
	require.NoError(t, err)
	require.Zero(t, namespaceSize)

	repositorySize, err = attributionRepository.SyncAndSum(ctx, reporegistry.ScopeTypeRepository, fixture.Repository.ID)
	require.NoError(t, err)
	require.Zero(t, repositorySize)
}
