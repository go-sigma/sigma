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
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	reporegistry "github.com/go-sigma/sigma/pkg/dal/repository/registry"
	"github.com/go-sigma/sigma/pkg/testkit"
	"github.com/go-sigma/sigma/pkg/utils/uuid"
)

func TestNewArtifactRepository(t *testing.T) {
	require.NotNil(t, reporegistry.NewArtifactRepository())
	require.NotNil(t, reporegistry.NewArtifactRepository(query.Q))
}

func TestArtifactRepositoryClaimVulnerability(t *testing.T) {
	digCon := testkit.InitRepository(t)
	require.NotNil(t, digCon)

	ctx := t.Context()
	staleBefore := time.Now().Add(-6 * time.Hour).UnixMilli()
	namespaceObj := &models.Namespace{ID: uuid.NewV7String(), Name: "claim-vulnerability"}
	require.NoError(t, query.Q.Namespace.WithContext(ctx).Create(namespaceObj))
	repositoryObj := &models.Repository{ID: uuid.NewV7String(), Name: "claim-vulnerability/repo", NamespaceID: namespaceObj.ID}
	require.NoError(t, query.Q.Repository.WithContext(ctx).Create(repositoryObj))

	pendingArtifact := &models.Artifact{
		ID:           uuid.NewV7String(),
		NamespaceID:  namespaceObj.ID,
		RepositoryID: repositoryObj.ID,
		Digest:       "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	}
	require.NoError(t, query.Q.Artifact.WithContext(ctx).Create(pendingArtifact))
	pendingVulnerability := &models.ArtifactVulnerability{
		ID:         uuid.NewV7String(),
		ArtifactID: pendingArtifact.ID,
		Status:     enums.TaskCommonStatusPending,
	}
	require.NoError(t, query.Q.ArtifactVulnerability.WithContext(ctx).Create(pendingVulnerability))

	artifactRepository := reporegistry.NewArtifactRepository()
	claimed, err := artifactRepository.ClaimVulnerability(ctx, staleBefore)
	require.NoError(t, err)
	require.Equal(t, pendingVulnerability.ID, claimed.ID)
	require.NotNil(t, claimed.Artifact)
	require.NotNil(t, claimed.Artifact.Repository)

	claimedObj, err := query.Q.ArtifactVulnerability.WithContext(ctx).Where(
		query.ArtifactVulnerability.ID.Eq(pendingVulnerability.ID),
	).First()
	require.NoError(t, err)
	require.Equal(t, enums.TaskCommonStatusDoing, claimedObj.Status)

	_, err = artifactRepository.ClaimVulnerability(ctx, staleBefore)
	require.True(t, errors.Is(err, gorm.ErrRecordNotFound))

	staleArtifact := &models.Artifact{
		ID:           uuid.NewV7String(),
		NamespaceID:  namespaceObj.ID,
		RepositoryID: repositoryObj.ID,
		Digest:       "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
	}
	require.NoError(t, query.Q.Artifact.WithContext(ctx).Create(staleArtifact))
	staleVulnerability := &models.ArtifactVulnerability{
		ID:         uuid.NewV7String(),
		ArtifactID: staleArtifact.ID,
		Status:     enums.TaskCommonStatusDoing,
	}
	require.NoError(t, query.Q.ArtifactVulnerability.WithContext(ctx).Create(staleVulnerability))
	oldUpdatedAt := time.Now().Add(-7 * time.Hour).UnixMilli()
	_, err = query.Q.ArtifactVulnerability.WithContext(ctx).
		Where(query.ArtifactVulnerability.ID.Eq(staleVulnerability.ID)).
		UpdateColumn(query.ArtifactVulnerability.UpdatedAt, oldUpdatedAt)
	require.NoError(t, err)

	claimed, err = artifactRepository.ClaimVulnerability(ctx, staleBefore)
	require.NoError(t, err)
	require.Equal(t, staleVulnerability.ID, claimed.ID)

	refreshed, err := query.Q.ArtifactVulnerability.WithContext(ctx).
		Where(query.Q.ArtifactVulnerability.ID.Eq(staleVulnerability.ID)).
		First()
	require.NoError(t, err)
	require.Greater(t, refreshed.UpdatedAt, staleBefore)

	_, err = artifactRepository.ClaimVulnerability(ctx, staleBefore)
	require.True(t, errors.Is(err, gorm.ErrRecordNotFound))
}

func TestArtifactRepositoryAssociateArtifact(t *testing.T) {
	fixture := newRegistryFixture(t)
	ctx := t.Context()
	parent := newRegistryArtifact(fixture.Namespace.ID, fixture.Repository.ID, "sha256:parent")
	child := newRegistryArtifact(fixture.Namespace.ID, fixture.Repository.ID, "sha256:child")
	referrer := newRegistryArtifact(fixture.Namespace.ID, fixture.Repository.ID, "sha256:referrer")
	referrer.ReferrerID = &parent.ID
	require.NoError(t, fixture.ArtifactRepository.Create(ctx, parent))
	require.NoError(t, fixture.ArtifactRepository.Create(ctx, child))
	require.NoError(t, fixture.ArtifactRepository.Create(ctx, referrer))

	require.NoError(t, fixture.ArtifactRepository.AssociateArtifact(ctx, parent, []*models.Artifact{child}))
	require.NoError(t, fixture.ArtifactRepository.IsArtifactAssociatedWithArtifact(ctx, child.ID))

	referrers, err := fixture.ArtifactRepository.FindReferrersBySubjects(ctx, fixture.Repository.ID, []string{parent.Digest})
	require.NoError(t, err)
	require.Len(t, referrers, 1)
	require.Equal(t, referrer.ID, referrers[0].ID)

	referrers, err = fixture.ArtifactRepository.GetReferrers(ctx, fixture.Repository.ID, parent.Digest, nil)
	require.NoError(t, err)
	require.Len(t, referrers, 1)
	require.Equal(t, referrer.ID, referrers[0].ID)
}

func TestArtifactRepository(t *testing.T) {
	fixture := newRegistryFixture(t)
	ctx := t.Context()
	artifactRepository := fixture.ArtifactRepository
	tagRepository := fixture.TagRepository
	artifactObj := newRegistryArtifact(fixture.Namespace.ID, fixture.Repository.ID, "sha256:artifact-main")
	require.NoError(t, artifactRepository.Create(ctx, artifactObj))

	tagObj := &models.Tag{
		ID:           uuid.NewV7String(),
		RepositoryID: fixture.Repository.ID,
		ArtifactID:   artifactObj.ID,
		Name:         "latest",
	}
	require.NoError(t, tagRepository.Create(ctx, tagObj))

	got, err := artifactRepository.Get(ctx, artifactObj.ID)
	require.NoError(t, err)
	require.Equal(t, fixture.Repository.ID, got.Repository.ID)

	got, err = artifactRepository.GetByDigest(ctx, fixture.Repository.ID, artifactObj.Digest)
	require.NoError(t, err)
	require.Equal(t, artifactObj.ID, got.ID)

	artifacts, err := artifactRepository.GetByDigests(ctx, fixture.Repository.Name, []string{artifactObj.Digest})
	require.NoError(t, err)
	require.Len(t, artifacts, 1)
	require.Equal(t, artifactObj.ID, artifacts[0].ID)

	require.NoError(t, artifactRepository.Incr(ctx, artifactObj.ID))
	got, err = artifactRepository.Get(ctx, artifactObj.ID)
	require.NoError(t, err)
	require.Equal(t, int64(1), got.PullTimes)

	countByNamespace, err := artifactRepository.CountByNamespace(ctx, []string{fixture.Namespace.ID})
	require.NoError(t, err)
	require.Equal(t, int64(1), countByNamespace[fixture.Namespace.ID])

	countByRepository, err := artifactRepository.CountByRepository(ctx, []string{fixture.Repository.ID})
	require.NoError(t, err)
	require.Equal(t, int64(1), countByRepository[fixture.Repository.ID])

	limit := 10
	listed, err := artifactRepository.ListArtifact(ctx, api.ListArtifactRequest{
		Pagination: api.Pagination{Limit: &limit},
		Namespace:  fixture.Namespace.Name,
		Repository: fixture.Repository.Name,
	})
	require.NoError(t, err)
	require.NotEmpty(t, listed)

	total, err := artifactRepository.CountArtifact(ctx, api.ListArtifactRequest{
		Namespace:  fixture.Namespace.Name,
		Repository: fixture.Repository.Name,
	})
	require.NoError(t, err)
	require.NotZero(t, total)

	blobObj := &models.Blob{
		ID:          uuid.NewV7String(),
		Digest:      "sha256:artifact-blob",
		Size:        123,
		ContentType: "application/octet-stream",
	}
	require.NoError(t, fixture.BlobRepository.Create(ctx, blobObj))
	require.NoError(t, artifactRepository.AssociateBlobs(ctx, artifactObj, []*models.Blob{blobObj}))

	sbom := &models.ArtifactSbom{
		ID:         uuid.NewV7String(),
		ArtifactID: artifactObj.ID,
		Raw:        []byte("sbom"),
		Status:     enums.TaskCommonStatusPending,
	}
	require.NoError(t, artifactRepository.CreateSbom(ctx, sbom))
	gotSbom, err := artifactRepository.GetSbom(ctx)
	require.NoError(t, err)
	require.Equal(t, sbom.ID, gotSbom.ID)
	require.NoError(t, artifactRepository.UpdateSbom(ctx, artifactObj.ID, map[string]any{
		query.ArtifactSbom.Status.ColumnName().String(): enums.TaskCommonStatusSuccess,
	}))

	vulnerability := &models.ArtifactVulnerability{
		ID:         uuid.NewV7String(),
		ArtifactID: artifactObj.ID,
		Raw:        []byte("vulnerability"),
		Status:     enums.TaskCommonStatusPending,
	}
	require.NoError(t, artifactRepository.CreateVulnerability(ctx, vulnerability))
	require.NoError(t, artifactRepository.UpdateVulnerability(ctx, artifactObj.ID, map[string]any{
		query.ArtifactVulnerability.Status.ColumnName().String(): enums.TaskCommonStatusSuccess,
	}))

	deleteByIDsArtifact := newRegistryArtifact(fixture.Namespace.ID, fixture.Repository.ID, "sha256:delete-by-ids")
	require.NoError(t, artifactRepository.Create(ctx, deleteByIDsArtifact))
	require.NoError(t, artifactRepository.DeleteByIDs(ctx, []string{deleteByIDsArtifact.ID}))

	require.NoError(t, artifactRepository.DeleteByDigest(ctx, fixture.Repository.Name, artifactObj.Digest))
	err = artifactRepository.DeleteByID(ctx, uuid.NewV7String())
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestArtifactRepositoryGetNamespaceSize(t *testing.T) {
	fixture := newRegistryFixture(t)
	ctx := t.Context()
	artifactObj := newRegistryArtifact(fixture.Namespace.ID, fixture.Repository.ID, "sha256:size")
	artifactObj.BlobsSize = 321
	require.NoError(t, fixture.ArtifactRepository.Create(ctx, artifactObj))

	namespaceSize, err := fixture.ArtifactRepository.GetNamespaceSize(ctx, fixture.Namespace.ID)
	require.NoError(t, err)
	require.Equal(t, int64(321), namespaceSize)

	repositorySize, err := fixture.ArtifactRepository.GetRepositorySize(ctx, fixture.Repository.ID)
	require.NoError(t, err)
	require.Equal(t, int64(321), repositorySize)
}
