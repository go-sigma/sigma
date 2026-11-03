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

package overview_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	repooverview "github.com/go-sigma/sigma/pkg/dal/repository/overview"
	"github.com/go-sigma/sigma/pkg/testkit"
	"github.com/go-sigma/sigma/pkg/utils/uuid"
)

func TestNewOverviewRepository(t *testing.T) {
	require.NotNil(t, repooverview.NewOverviewRepository())
	require.NotNil(t, repooverview.NewOverviewRepository(query.Q))
}

func TestOverviewRepository(t *testing.T) {
	testkit.InitRepository(t)

	ctx := t.Context()
	overviewRepository := repooverview.NewOverviewRepository()

	userID := uuid.NewV7String()
	otherUserID := uuid.NewV7String()

	managed := &models.Namespace{ID: uuid.NewV7String(), Name: "managed", Visibility: enums.VisibilityPrivate, Size: 300}
	readOnly := &models.Namespace{ID: uuid.NewV7String(), Name: "readonly", Visibility: enums.VisibilityPrivate, Size: 100}
	unrelated := &models.Namespace{ID: uuid.NewV7String(), Name: "unrelated", Visibility: enums.VisibilityPrivate, Size: 200}
	for _, namespaceObj := range []*models.Namespace{managed, readOnly, unrelated} {
		require.NoError(t, query.Q.Namespace.WithContext(ctx).Create(namespaceObj))
	}

	// Repositories are counted from the live table, not from the cached counters.
	repositories := []*models.Repository{
		{ID: uuid.NewV7String(), NamespaceID: managed.ID, Name: "api"},
		{ID: uuid.NewV7String(), NamespaceID: managed.ID, Name: "web"},
		{ID: uuid.NewV7String(), NamespaceID: unrelated.ID, Name: "worker"},
	}
	for _, repositoryObj := range repositories {
		require.NoError(t, query.Q.Repository.WithContext(ctx).Create(repositoryObj))
	}

	require.NoError(t, query.Q.NamespaceMember.WithContext(ctx).Create(&models.NamespaceMember{
		ID: uuid.NewV7String(), UserID: userID, NamespaceID: managed.ID, Role: enums.NamespaceRoleManager,
	}))
	require.NoError(t, query.Q.NamespaceMember.WithContext(ctx).Create(&models.NamespaceMember{
		ID: uuid.NewV7String(), UserID: userID, NamespaceID: readOnly.ID, Role: enums.NamespaceRoleReader,
	}))
	require.NoError(t, query.Q.NamespaceMember.WithContext(ctx).Create(&models.NamespaceMember{
		ID: uuid.NewV7String(), UserID: otherUserID, NamespaceID: unrelated.ID, Role: enums.NamespaceRoleAdmin,
	}))

	ids, err := overviewRepository.ListManagedNamespaceIDs(ctx, userID)
	require.NoError(t, err)
	require.Equal(t, []string{managed.ID}, ids)

	all, err := overviewRepository.AggregateNamespaces(ctx, nil)
	require.NoError(t, err)
	require.Equal(t, int64(3), all.Namespaces)
	require.Equal(t, int64(3), all.Repositories)
	require.Zero(t, all.Tags)
	require.Equal(t, int64(600), all.Size)

	managedUsage, err := overviewRepository.AggregateNamespaces(ctx, ids)
	require.NoError(t, err)
	require.Equal(t, int64(1), managedUsage.Namespaces)
	require.Equal(t, int64(2), managedUsage.Repositories)
	require.Equal(t, int64(300), managedUsage.Size)

	emptyUsage, err := overviewRepository.AggregateNamespaces(ctx, []string{})
	require.NoError(t, err)
	require.Zero(t, emptyUsage.Namespaces)
	require.Zero(t, emptyUsage.Size)

	top, err := overviewRepository.TopNamespaces(ctx, nil, 2)
	require.NoError(t, err)
	require.Len(t, top, 2)
	require.Equal(t, managed.ID, top[0].ID)
	require.Equal(t, unrelated.ID, top[1].ID)
	require.Equal(t, int64(2), top[0].RepositoryCount)
	require.Equal(t, int64(1), top[1].RepositoryCount)

	emptyTop, err := overviewRepository.TopNamespaces(ctx, []string{}, 2)
	require.NoError(t, err)
	require.Empty(t, emptyTop)

	// Builds: one triggered by the user, one by cron that falls back to the builder creator,
	// and one triggered by somebody else.
	now := time.Now().UTC().Truncate(time.Hour)
	builderObj := &models.Builder{
		ID:           uuid.NewV7String(),
		RepositoryID: uuid.NewV7String(),
		Source:       enums.BuilderSourceDockerfile,
		CreatorID:    &userID,
	}
	require.NoError(t, query.Q.Builder.WithContext(ctx).Create(builderObj))

	endedAt := now.Add(2 * time.Minute).UnixMilli()
	startedAt := now.UnixMilli()
	runners := []*models.BuilderRunner{
		{ID: uuid.NewV7String(), BuilderID: builderObj.ID, Status: enums.BuildStatusSuccess, StartedAt: &startedAt, EndedAt: &endedAt, TriggeredBy: &userID},
		{ID: uuid.NewV7String(), BuilderID: builderObj.ID, Status: enums.BuildStatusFailed, StartedAt: &startedAt},
		{ID: uuid.NewV7String(), BuilderID: builderObj.ID, Status: enums.BuildStatusSuccess, StartedAt: &startedAt, TriggeredBy: &otherUserID},
	}
	for _, runnerObj := range runners {
		require.NoError(t, query.Q.BuilderRunner.WithContext(ctx).Create(runnerObj))
	}

	since := now.Add(-time.Hour)
	allRecords, err := overviewRepository.ListBuildRecords(ctx, since, "")
	require.NoError(t, err)
	require.Len(t, allRecords, 3)

	mineRecords, err := overviewRepository.ListBuildRecords(ctx, since, userID)
	require.NoError(t, err)
	require.Len(t, mineRecords, 2)

	noneRecords, err := overviewRepository.ListBuildRecords(ctx, since, uuid.NewV7String())
	require.NoError(t, err)
	require.Empty(t, noneRecords)
}

func TestOverviewRepositorySumBlobSize(t *testing.T) {
	testkit.InitRepository(t)

	ctx := t.Context()
	overviewRepository := repooverview.NewOverviewRepository()

	userID := uuid.NewV7String()
	managed := &models.Namespace{ID: uuid.NewV7String(), Name: "managed-blobs", Visibility: enums.VisibilityPrivate}
	other := &models.Namespace{ID: uuid.NewV7String(), Name: "other-blobs", Visibility: enums.VisibilityPrivate}
	for _, namespaceObj := range []*models.Namespace{managed, other} {
		require.NoError(t, query.Q.Namespace.WithContext(ctx).Create(namespaceObj))
	}
	require.NoError(t, query.Q.NamespaceMember.WithContext(ctx).Create(&models.NamespaceMember{
		ID: uuid.NewV7String(), UserID: userID, NamespaceID: managed.ID, Role: enums.NamespaceRoleManager,
	}))

	managedRepo := &models.Repository{ID: uuid.NewV7String(), NamespaceID: managed.ID, Name: "api"}
	otherRepo := &models.Repository{ID: uuid.NewV7String(), NamespaceID: other.ID, Name: "web"}
	for _, repositoryObj := range []*models.Repository{managedRepo, otherRepo} {
		require.NoError(t, query.Q.Repository.WithContext(ctx).Create(repositoryObj))
	}

	// Two blobs, one of them shared by an artifact of each namespace.
	shared := &models.Blob{ID: uuid.NewV7String(), Digest: "sha256:shared", Size: 100, ContentType: "application/octet-stream"}
	unique := &models.Blob{ID: uuid.NewV7String(), Digest: "sha256:unique", Size: 50, ContentType: "application/octet-stream"}
	for _, blobObj := range []*models.Blob{shared, unique} {
		require.NoError(t, query.Q.Blob.WithContext(ctx).Create(blobObj))
	}

	managedArtifact := &models.Artifact{
		ID: uuid.NewV7String(), NamespaceID: managed.ID, RepositoryID: managedRepo.ID,
		Digest: "sha256:managed", ContentType: "application/vnd.oci.image.manifest.v1+json",
	}
	otherArtifact := &models.Artifact{
		ID: uuid.NewV7String(), NamespaceID: other.ID, RepositoryID: otherRepo.ID,
		Digest: "sha256:other", ContentType: "application/vnd.oci.image.manifest.v1+json",
	}
	for _, artifactObj := range []*models.Artifact{managedArtifact, otherArtifact} {
		require.NoError(t, query.Q.Artifact.WithContext(ctx).Create(artifactObj))
	}
	require.NoError(t, query.Q.UnderlyingDB().WithContext(ctx).Exec(
		"INSERT INTO artifact_blobs (artifact_id, blob_id) VALUES (?, ?), (?, ?), (?, ?)",
		managedArtifact.ID, shared.ID, managedArtifact.ID, unique.ID, otherArtifact.ID, shared.ID,
	).Error)

	global, err := overviewRepository.SumBlobSize(ctx, nil)
	require.NoError(t, err)
	require.Equal(t, int64(150), global)

	ids, err := overviewRepository.ListManagedNamespaceIDs(ctx, userID)
	require.NoError(t, err)
	scoped, err := overviewRepository.SumBlobSize(ctx, ids)
	require.NoError(t, err)
	require.Equal(t, int64(150), scoped)

	empty, err := overviewRepository.SumBlobSize(ctx, []string{})
	require.NoError(t, err)
	require.Zero(t, empty)
}
