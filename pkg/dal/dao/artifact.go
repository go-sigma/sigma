// Copyright 2023 XImager
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

package dao

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/ximager/ximager/pkg/dal/models"
	"github.com/ximager/ximager/pkg/dal/query"
	"github.com/ximager/ximager/pkg/types"
	"github.com/ximager/ximager/pkg/types/enums"
)

//go:generate mockgen -destination=mocks/artifact.go -package=mocks github.com/ximager/ximager/pkg/dal/dao ArtifactService
//go:generate mockgen -destination=mocks/artifact_factory.go -package=mocks github.com/ximager/ximager/pkg/dal/dao ArtifactServiceFactory

// ArtifactService is the interface that provides the artifact service methods.
type ArtifactService interface {
	// Create create a new artifact if conflict do nothing.
	Create(ctx context.Context, artifact *models.Artifact) error
	// Get gets the artifact with the specified artifact ID.
	Get(ctx context.Context, id int64) (*models.Artifact, error)
	// GetByDigest gets the artifact with the specified digest.
	GetByDigest(ctx context.Context, repositoryID int64, digest string) (*models.Artifact, error)
	// GetByDigests gets the artifacts with the specified digests.
	GetByDigests(ctx context.Context, repository string, digests []string) ([]*models.Artifact, error)
	// DeleteByDigest deletes the artifact with the specified digest.
	DeleteByDigest(ctx context.Context, repository, digest string) error
	// AssociateBlobs associates the blobs with the artifact.
	AssociateBlobs(ctx context.Context, artifact *models.Artifact, blobs []*models.Blob) error
	// AssociateArtifact associates the artifacts with the artifact.
	AssociateArtifact(ctx context.Context, artifact *models.Artifact, artifacts []*models.Artifact) error
	// CountByNamespace counts the artifacts by the specified namespace.
	CountByNamespace(ctx context.Context, namespaceIDs []int64) (map[int64]int64, error)
	// CountByRepository counts the artifacts by the specified repository.
	CountByRepository(ctx context.Context, repositoryIDs []int64) (map[int64]int64, error)
	// Incr increases the pull times of the artifact.
	Incr(ctx context.Context, id int64) error
	// ListArtifact lists the artifacts by the specified request.
	ListArtifact(ctx context.Context, req types.ListArtifactRequest) ([]*models.Artifact, error)
	// CountArtifact counts the artifacts by the specified request.
	CountArtifact(ctx context.Context, req types.ListArtifactRequest) (int64, error)
	// DeleteByID deletes the artifact with the specified artifact ID.
	DeleteByID(ctx context.Context, id int64) error
	// SaveSbom save a new artifact sbom if conflict update.
	SaveSbom(ctx context.Context, sbom *models.ArtifactSbom) error
	// SaveVulnerability save a new artifact vulnerability if conflict update.
	SaveVulnerability(ctx context.Context, vulnerability *models.ArtifactVulnerability) error
	// UpdateSbomStatus update the artifact sbom status.
	UpdateSbomStatus(ctx context.Context, artifactID int64, status enums.TaskCommonStatus) error
	// UpdateVulnerabilityStatus update the artifact vulnerability status.
	UpdateVulnerabilityStatus(ctx context.Context, artifactID int64, status enums.TaskCommonStatus) error
}

type artifactService struct {
	tx *query.Query
}

// ArtifactServiceFactory is the interface that provides the artifact service factory methods.
type ArtifactServiceFactory interface {
	New(txs ...*query.Query) ArtifactService
}

type artifactServiceFactory struct{}

// NewArtifactServiceFactory creates a new artifact service factory.
func NewArtifactServiceFactory() ArtifactServiceFactory {
	return &artifactServiceFactory{}
}

func (f *artifactServiceFactory) New(txs ...*query.Query) ArtifactService {
	tx := query.Q
	if len(txs) > 0 {
		tx = txs[0]
	}
	return &artifactService{
		tx: tx,
	}
}

// Create create a new artifact if conflict do nothing.
func (s *artifactService) Create(ctx context.Context, artifact *models.Artifact) error {
	return s.tx.Artifact.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(artifact)
}

// Get gets the artifact with the specified artifact ID.
func (s *artifactService) Get(ctx context.Context, id int64) (*models.Artifact, error) {
	// SELECT * FROM `repositories` WHERE `repositories`.`id` = 1 AND `repositories`.`deleted_at` = 0
	// SELECT * FROM `artifacts` WHERE `artifacts`.`id` = 1 AND `artifacts`.`deleted_at` = 0 ORDER BY `artifacts`.`id` LIMIT 1
	return s.tx.Artifact.WithContext(ctx).
		Preload(s.tx.Artifact.Repository).
		Where(s.tx.Artifact.ID.Eq(id)).First()
}

// GetByDigest gets the artifact with the specified digest.
func (s *artifactService) GetByDigest(ctx context.Context, repositoryID int64, digest string) (*models.Artifact, error) {
	return s.tx.Artifact.WithContext(ctx).
		Where(s.tx.Artifact.RepositoryID.Eq(repositoryID)).
		Where(s.tx.Artifact.Digest.Eq(digest)).
		First()
}

// GetByDigests gets the artifacts with the specified digests.
func (s *artifactService) GetByDigests(ctx context.Context, repository string, digests []string) ([]*models.Artifact, error) {
	return s.tx.Artifact.WithContext(ctx).
		LeftJoin(s.tx.Repository, s.tx.Repository.ID.EqCol(s.tx.Artifact.RepositoryID)).
		Where(s.tx.Repository.Name.Eq(repository)).
		Where(s.tx.Artifact.Digest.In(digests...)).
		Preload(s.tx.Artifact.Tags.Order(s.tx.Tag.UpdatedAt.Desc()).Limit(10)).
		Find()
}

// DeleteByDigest deletes the artifact with the specified digest.
func (s *artifactService) DeleteByDigest(ctx context.Context, repository, digest string) error {
	artifact, err := s.tx.Artifact.WithContext(ctx).Where(s.tx.Artifact.Digest.Eq(digest)).Preload(s.tx.Artifact.Blobs).First()
	if err != nil {
		return err
	}
	err = s.tx.Transaction(func(tx *query.Query) error {
		err = tx.Artifact.Blobs.Model(artifact).Delete(artifact.Blobs...)
		if err != nil {
			return err
		}
		_, err = tx.Artifact.WithContext(ctx).Where(tx.Artifact.Digest.Eq(digest)).Delete()
		if err != nil {
			return err
		}
		_, err = tx.Tag.WithContext(ctx).Where(tx.Tag.ArtifactID.Eq(artifact.ID)).Delete()
		if err != nil {
			return err
		}
		return nil
	})
	return err
}

// AssociateBlobs ...
func (s *artifactService) AssociateBlobs(ctx context.Context, artifact *models.Artifact, blobs []*models.Blob) error {
	return s.tx.Artifact.Blobs.WithContext(ctx).Model(artifact).Append(blobs...)
}

// AssociateArtifact ...
func (s *artifactService) AssociateArtifact(ctx context.Context, artifact *models.Artifact, artifacts []*models.Artifact) error {
	return s.tx.Artifact.ArtifactIndexes.WithContext(ctx).Model(artifact).Append(artifacts...)
}

// Incr increases the pull times of the artifact.
func (s *artifactService) Incr(ctx context.Context, id int64) error {
	_, err := s.tx.Artifact.WithContext(ctx).Where(s.tx.Artifact.ID.Eq(id)).
		UpdateColumns(map[string]interface{}{
			"pull_times": gorm.Expr("pull_times + ?", 1),
			"last_pull":  time.Now(),
		})
	return err
}

// CountByNamespace counts the artifacts by the specified namespace.
func (s *artifactService) CountByNamespace(ctx context.Context, namespaceIDs []int64) (map[int64]int64, error) {
	artifactCount := make(map[int64]int64)
	if len(namespaceIDs) == 0 {
		return artifactCount, nil
	}
	var count []struct {
		NamespaceID int64 `gorm:"column:namespace_id"`
		Count       int64 `gorm:"column:count"`
	}
	err := s.tx.Artifact.WithContext(ctx).
		LeftJoin(s.tx.Repository, s.tx.Repository.ID.EqCol(s.tx.Artifact.RepositoryID)).
		Where(s.tx.Repository.NamespaceID.In(namespaceIDs...)).
		Group(s.tx.Repository.NamespaceID).
		Select(s.tx.Repository.NamespaceID, s.tx.Artifact.ID.Count().As("count")).
		Scan(&count)
	if err != nil {
		return nil, err
	}
	for _, c := range count {
		artifactCount[c.NamespaceID] = c.Count
	}
	return artifactCount, nil
}

// CountByRepository counts the artifacts by the specified repository.
func (s *artifactService) CountByRepository(ctx context.Context, repositoryIDs []int64) (map[int64]int64, error) {
	artifactCount := make(map[int64]int64)
	if len(repositoryIDs) == 0 {
		return artifactCount, nil
	}
	var count []struct {
		RepositoryID int64 `gorm:"column:repository_id"`
		Count        int64 `gorm:"column:count"`
	}
	err := s.tx.Artifact.WithContext(ctx).Where(s.tx.Artifact.RepositoryID.In(repositoryIDs...)).
		Group(s.tx.Artifact.RepositoryID).
		Select(s.tx.Artifact.RepositoryID, s.tx.Artifact.ID.Count().As("count")).
		Scan(&count)
	if err != nil {
		return nil, err
	}
	for _, c := range count {
		artifactCount[c.RepositoryID] = c.Count
	}
	return artifactCount, nil
}

// ListArtifact lists the artifacts by the specified request.
func (s *artifactService) ListArtifact(ctx context.Context, req types.ListArtifactRequest) ([]*models.Artifact, error) {
	return s.tx.Artifact.WithContext(ctx).
		LeftJoin(s.tx.Repository, s.tx.Repository.ID.EqCol(s.tx.Artifact.RepositoryID), s.tx.Repository.Name.Eq(req.Repository)).
		LeftJoin(s.tx.Namespace, s.tx.Namespace.Name.EqCol(s.tx.Repository.Name), s.tx.Namespace.Name.Eq(req.Namespace)).
		Preload(s.tx.Artifact.Tags.Order(s.tx.Tag.UpdatedAt.Desc()).Limit(10)).
		Offset(req.PageSize * (req.PageNum - 1)).Limit(req.PageSize).Find()
}

// CountArtifact counts the artifacts by the specified request.
func (s *artifactService) CountArtifact(ctx context.Context, req types.ListArtifactRequest) (int64, error) {
	return s.tx.Artifact.WithContext(ctx).
		LeftJoin(s.tx.Repository, s.tx.Repository.ID.EqCol(s.tx.Artifact.RepositoryID), s.tx.Repository.Name.Eq(req.Repository)).
		LeftJoin(s.tx.Namespace, s.tx.Namespace.Name.EqCol(s.tx.Repository.Name), s.tx.Namespace.Name.Eq(req.Namespace)).
		Count()
}

// DeleteByID deletes the artifact with the specified ID.
func (s *artifactService) DeleteByID(ctx context.Context, id int64) error {
	matched, err := s.tx.Artifact.WithContext(ctx).Where(s.tx.Artifact.ID.Eq(id)).Delete()
	if err != nil {
		return err
	}
	if matched.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// SaveSbom save a new artifact sbom if conflict do nothing.
func (s *artifactService) SaveSbom(ctx context.Context, sbom *models.ArtifactSbom) error {
	return s.tx.ArtifactSbom.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(sbom)
}

// SaveVulnerability save a new artifact vulnerability if conflict do nothing.
func (s *artifactService) SaveVulnerability(ctx context.Context, vulnerability *models.ArtifactVulnerability) error {
	return s.tx.ArtifactVulnerability.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(vulnerability)
}

// UpdateSbomStatus update the artifact sbom status.
func (s *artifactService) UpdateSbomStatus(ctx context.Context, artifactID int64, status enums.TaskCommonStatus) error {
	_, err := s.tx.ArtifactSbom.WithContext(ctx).Where(s.tx.ArtifactSbom.ID.Eq(artifactID)).
		UpdateColumns(map[string]interface{}{
			"status": status,
		})
	return err
}

// UpdateVulnerabilityStatus update the artifact vulnerability status.
func (s *artifactService) UpdateVulnerabilityStatus(ctx context.Context, artifactID int64, status enums.TaskCommonStatus) error {
	_, err := s.tx.ArtifactVulnerability.WithContext(ctx).Where(s.tx.ArtifactVulnerability.ID.Eq(artifactID)).
		UpdateColumns(map[string]interface{}{
			"status": status,
		})
	return err
}
