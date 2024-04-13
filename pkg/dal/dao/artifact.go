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

package dao

import (
	"context"
	"errors"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/spf13/cast"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

//go:generate mockgen -destination=mocks/artifact.go -package=mocks github.com/go-sigma/sigma/pkg/dal/dao ArtifactService
//go:generate mockgen -destination=mocks/artifact_factory.go -package=mocks github.com/go-sigma/sigma/pkg/dal/dao ArtifactServiceFactory

// ArtifactService is the interface that provides the artifact service methods.
type ArtifactService interface {
	// Create create a new artifact if conflict do nothing.
	Create(ctx context.Context, artifact *models.Artifact) error
	// FindWithLastPull ...
	FindWithLastPull(ctx context.Context, repositoryID int64, before int64, limit, last int64) ([]*models.Artifact, error)
	// FindAssociateWithTag ...
	FindAssociateWithTag(ctx context.Context, ids []int64) ([]int64, error)
	// FindAssociateWithArtifact ...
	FindAssociateWithArtifact(ctx context.Context, ids []int64) ([]int64, error)
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
	// DeleteByIDs deletes the artifact with the specified artifact ID.
	DeleteByIDs(ctx context.Context, ids []int64) error
	// CreateSbom create a new artifact sbom.
	CreateSbom(ctx context.Context, sbom *models.ArtifactSbom) error
	// CreateVulnerability save a new artifact vulnerability.
	CreateVulnerability(ctx context.Context, vulnerability *models.ArtifactVulnerability) error
	// UpdateSbom update the artifact sbom.
	UpdateSbom(ctx context.Context, artifactID int64, updates map[string]any) error
	// UpdateVulnerability update the artifact vulnerability.
	UpdateVulnerability(ctx context.Context, artifactID int64, updates map[string]any) error
	// GetNamespaceSize get the specific namespace size
	GetNamespaceSize(ctx context.Context, namespaceID int64) (int64, error)
	// GetRepositorySize get the specific repository size
	GetRepositorySize(ctx context.Context, repositoryID int64) (int64, error)
	// GetReferrers ...
	GetReferrers(ctx context.Context, repositoryID int64, digest string, artifactTypes []string) ([]*models.Artifact, error)
	// IsArtifactAssociatedWithArtifact ...
	IsArtifactAssociatedWithArtifact(ctx context.Context, artifactID int64) error
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

func (s *artifactServiceFactory) New(txs ...*query.Query) ArtifactService {
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

// FindWithLastPull ...
func (s *artifactService) FindWithLastPull(ctx context.Context, repositoryID int64, before int64, limit, last int64) ([]*models.Artifact, error) {
	return s.tx.Artifact.WithContext(ctx).
		Where(s.tx.Artifact.ID.Gt(last), s.tx.Artifact.RepositoryID.Eq(repositoryID)).
		Where(s.tx.Artifact.LastPull.Lt(before)).
		Or(s.tx.Artifact.LastPull.IsNull(), s.tx.Artifact.UpdatedAt.Lt(before)).
		Limit(int(limit)).Order(s.tx.Artifact.ID).Find()
}

// FindAssociateWithTag ...
func (s *artifactService) FindAssociateWithTag(ctx context.Context, ids []int64) ([]int64, error) {
	var result []int64
	err := s.tx.Blob.WithContext(ctx).UnderlyingDB().Raw("SELECT artifact_id FROM tags WHERE artifact_id in (?)", ids).Scan(&result).Error
	return result, err
}

// FindAssociateWithArtifact ...
func (s *artifactService) FindAssociateWithArtifact(ctx context.Context, ids []int64) ([]int64, error) {
	var artifacts []int64
	err := s.tx.Blob.WithContext(ctx).UnderlyingDB().Raw("SELECT artifact_id FROM artifact_artifacts WHERE artifact_id in (?)", ids).Scan(&artifacts).Error
	if err != nil {
		return nil, err
	}
	var artifactSubs []int64
	err = s.tx.Blob.WithContext(ctx).UnderlyingDB().Raw("SELECT artifact_sub_id FROM artifact_artifacts WHERE artifact_sub_id in (?)", ids).Scan(&artifactSubs).Error
	if err != nil {
		return nil, err
	}
	resultSet := mapset.NewSet(artifacts...)
	resultSet.Append(artifactSubs...)
	return resultSet.ToSlice(), err
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
	return s.tx.Artifact.ArtifactSubs.WithContext(ctx).Model(artifact).Append(artifacts...)
}

// Incr increases the pull times of the artifact.
func (s *artifactService) Incr(ctx context.Context, id int64) error {
	_, err := s.tx.Artifact.WithContext(ctx).Where(s.tx.Artifact.ID.Eq(id)).
		UpdateColumns(map[string]interface{}{
			"pull_times": gorm.Expr("pull_times + ?", 1),
			"last_pull":  time.Now().UnixMilli(),
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
		Where(s.tx.Artifact.ID.Gt(int64(ptr.To(req.Page)))).
		Limit(ptr.To(req.Limit)).Find()
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

// DeleteByIDs deletes the artifact with the specified ID.
func (s *artifactService) DeleteByIDs(ctx context.Context, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	_, err := s.tx.Artifact.WithContext(ctx).Where(s.tx.Artifact.ID.In(ids...)).Delete()
	return err
}

// CreateSbom save a new artifact sbom.
func (s *artifactService) CreateSbom(ctx context.Context, sbom *models.ArtifactSbom) error {
	_, err := s.tx.ArtifactSbom.WithContext(ctx).Where(s.tx.ArtifactSbom.ArtifactID.Eq(sbom.ArtifactID)).First()
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		return s.tx.ArtifactSbom.WithContext(ctx).Create(sbom)
	}
	return nil
}

// CreateVulnerability save a new artifact vulnerability.
func (s *artifactService) CreateVulnerability(ctx context.Context, vulnerability *models.ArtifactVulnerability) error {
	_, err := s.tx.ArtifactVulnerability.WithContext(ctx).Where(s.tx.ArtifactVulnerability.ArtifactID.Eq(vulnerability.ArtifactID)).First()
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		return s.tx.ArtifactVulnerability.WithContext(ctx).Create(vulnerability)
	}
	return nil
}

// UpdateSbom update the artifact sbom.
func (s *artifactService) UpdateSbom(ctx context.Context, artifactID int64, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}
	_, err := s.tx.ArtifactSbom.WithContext(ctx).Where(s.tx.ArtifactSbom.ArtifactID.Eq(artifactID)).UpdateColumns(updates)
	return err
}

// UpdateVulnerability update the artifact vulnerability.
func (s *artifactService) UpdateVulnerability(ctx context.Context, artifactID int64, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}
	_, err := s.tx.ArtifactVulnerability.WithContext(ctx).Where(s.tx.ArtifactVulnerability.ArtifactID.Eq(artifactID)).UpdateColumns(updates)
	return err
}

// GetNamespaceSize get the specific namespace size
func (s *artifactService) GetNamespaceSize(ctx context.Context, namespaceID int64) (int64, error) {
	q := s.tx.Artifact.WithContext(ctx).Select(s.tx.Artifact.BlobsSize.Sum().As("blobs_size")).
		Where(s.tx.Artifact.NamespaceID.Eq(namespaceID)).
		Clauses(clause.Locking{Strength: clause.LockingStrengthUpdate,
			Table: clause.Table{Name: clause.CurrentTable}})
	res, err := q.First()
	if err != nil {
		return 0, err
	}
	return res.BlobsSize, nil
}

// GetRepositorySize get the specific repository size
func (s *artifactService) GetRepositorySize(ctx context.Context, repositoryID int64) (int64, error) {
	q := s.tx.Artifact.WithContext(ctx).Select(s.tx.Artifact.BlobsSize.Sum().As("blobs_size")).
		Where(s.tx.Artifact.RepositoryID.Eq(repositoryID)).
		Clauses(clause.Locking{Strength: clause.LockingStrengthUpdate,
			Table: clause.Table{Name: clause.CurrentTable}})
	res, err := q.First()
	if err != nil {
		return 0, err
	}
	return res.BlobsSize, nil
}

// GetReferrers ...
func (s *artifactService) GetReferrers(ctx context.Context, repositoryID int64, digest string, artifactTypes []string) ([]*models.Artifact, error) {
	artifactObj, err := s.tx.Artifact.WithContext(ctx).Where(s.tx.Artifact.RepositoryID.Eq(repositoryID)).
		Where(s.tx.Artifact.Digest.Eq(digest)).First()
	if err != nil {
		return nil, err
	}
	q := s.tx.Artifact.WithContext(ctx).Where(s.tx.Artifact.RepositoryID.Eq(repositoryID))
	if len(artifactTypes) > 0 {
		q = q.Where(s.tx.Artifact.ConfigMediaType.In(artifactTypes...))
	}
	return q.Where(s.tx.Artifact.ReferrerID.Eq(artifactObj.ID)).Find()
}

// IsArtifactAssociatedWithArtifact ...
func (s *artifactService) IsArtifactAssociatedWithArtifact(ctx context.Context, artifactID int64) error {
	result, err := s.tx.Artifact.WithContext(ctx).ArtifactAssociated(artifactID)
	if err != nil {
		return err
	}
	r := cast.ToStringMapInt64(result)
	if r["count"] == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
