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

package registry

import (
	"context"
	"errors"
	"time"

	"github.com/spf13/cast"
	"gorm.io/gen"
	"gorm.io/gen/field"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

//go:generate mockgen -destination=artifact_mocks.go -package=registry github.com/go-sigma/sigma/pkg/dal/repository/registry ArtifactRepository

// ArtifactRepository defines artifact repository operations
type ArtifactRepository interface {
	// Create creates a new artifact and ignores conflicts
	Create(ctx context.Context, artifact *models.Artifact) error
	// FindWithLastPull finds artifacts older than the given pull or update time with cursor pagination
	FindWithLastPull(ctx context.Context, repositoryID string, before int64, limit int64, last string) ([]*models.Artifact, error)
	// FindDeletableWithCursor finds artifacts that are safe to delete with cursor pagination
	FindDeletableWithCursor(ctx context.Context, repositoryID string, before int64, last string, limit int64) ([]*models.Artifact, error)
	// FindReferrersBySubjects finds referrers for the given subject digests
	FindReferrersBySubjects(ctx context.Context, repositoryID string, digests []string) ([]*models.Artifact, error)
	// Get gets the artifact with the specified artifact ID
	Get(ctx context.Context, id string) (*models.Artifact, error)
	// GetByDigest gets the artifact with the specified repository and digest
	GetByDigest(ctx context.Context, repositoryID string, digest string) (*models.Artifact, error)
	// GetByDigests gets artifacts in a repository with the specified digests
	GetByDigests(ctx context.Context, repository string, digests []string) ([]*models.Artifact, error)
	// DeleteByDigest deletes the artifact with the specified repository and digest
	DeleteByDigest(ctx context.Context, repository, digest string) error
	// AssociateBlobs associates blobs with the artifact
	AssociateBlobs(ctx context.Context, artifact *models.Artifact, blobs []*models.Blob) error
	// AssociateArtifact associates child artifacts with the artifact
	AssociateArtifact(ctx context.Context, artifact *models.Artifact, artifacts []*models.Artifact) error
	// CountByNamespace counts artifacts grouped by namespace ID
	CountByNamespace(ctx context.Context, namespaceIDs []string) (map[string]int64, error)
	// CountByRepository counts artifacts grouped by repository ID
	CountByRepository(ctx context.Context, repositoryIDs []string) (map[string]int64, error)
	// Incr increases the artifact pull count
	Incr(ctx context.Context, id string) error
	// ListArtifact lists artifacts by the specified request
	ListArtifact(ctx context.Context, req api.ListArtifactRequest) ([]*models.Artifact, error)
	// CountArtifact counts artifacts by the specified request
	CountArtifact(ctx context.Context, req api.ListArtifactRequest) (int64, error)
	// DeleteByID deletes the artifact with the specified artifact ID
	DeleteByID(ctx context.Context, id string) error
	// DeleteByIDs deletes artifacts with the specified artifact IDs
	DeleteByIDs(ctx context.Context, ids []string) error
	// CreateSbom creates an artifact sbom if one does not exist
	CreateSbom(ctx context.Context, sbom *models.ArtifactSbom) error
	// GetSbom gets the next artifact sbom task
	GetSbom(ctx context.Context) (*models.ArtifactSbom, error)
	// CreateVulnerability creates an artifact vulnerability if one does not exist
	CreateVulnerability(ctx context.Context, vulnerability *models.ArtifactVulnerability) error
	// ClaimVulnerability claims the next pending or stale artifact vulnerability task
	ClaimVulnerability(ctx context.Context, staleBefore int64) (*models.ArtifactVulnerability, error)
	// UpdateSbom updates the artifact sbom with the specified artifact ID
	UpdateSbom(ctx context.Context, artifactID string, updates map[string]any) error
	// UpdateVulnerability updates the artifact vulnerability with the specified artifact ID
	UpdateVulnerability(ctx context.Context, artifactID string, updates map[string]any) error
	// GetNamespaceSize gets the total artifact blob size in the specified namespace
	GetNamespaceSize(ctx context.Context, namespaceID string) (int64, error)
	// GetRepositorySize gets the total artifact blob size in the specified repository
	GetRepositorySize(ctx context.Context, repositoryID string) (int64, error)
	// GetReferrers gets artifacts that refer to the specified subject digest
	GetReferrers(ctx context.Context, repositoryID string, digest string, artifactTypes []string) ([]*models.Artifact, error)
	// IsArtifactAssociatedWithArtifact checks whether the artifact is associated with another artifact
	IsArtifactAssociatedWithArtifact(ctx context.Context, artifactID string) error
}

type artifactRepository struct {
	tx *query.Query
}

// NewArtifactRepository creates a new artifact repository with the optional query transaction
func NewArtifactRepository(txs ...*query.Query) ArtifactRepository {
	tx := query.Q
	if len(txs) > 0 {
		tx = txs[0]
	}
	return &artifactRepository{
		tx: tx,
	}
}

// Create creates a new artifact and ignores conflicts
func (s *artifactRepository) Create(ctx context.Context, artifact *models.Artifact) error {
	return s.tx.Artifact.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(artifact)
}

// FindWithLastPull finds artifacts older than the given pull or update time with cursor pagination
func (s *artifactRepository) FindWithLastPull(ctx context.Context, repositoryID string, before int64, limit int64, last string) ([]*models.Artifact, error) {
	return s.tx.Artifact.WithContext(ctx).
		Where(s.tx.Artifact.ID.Gt(last), s.tx.Artifact.RepositoryID.Eq(repositoryID)).
		Where(s.tx.Artifact.LastPull.Lt(before)).
		Or(s.tx.Artifact.LastPull.IsNull(), s.tx.Artifact.UpdatedAt.Lt(before)).
		Limit(int(limit)).Order(s.tx.Artifact.ID).Find()
}

// FindDeletableWithCursor finds artifacts that are safe to delete with cursor pagination
func (s *artifactRepository) FindDeletableWithCursor(ctx context.Context, repositoryID string, before int64, last string, limit int64) ([]*models.Artifact, error) {
	parent := s.tx.Artifact.As("parent")
	artifactArtifacts := artifactArtifactsJoinTable{}
	artifactID := field.NewString(artifactArtifacts.TableName(), "artifact_id")
	artifactSubID := field.NewString(artifactArtifacts.TableName(), "artifact_sub_id")

	tagSubQuery := s.tx.Tag.WithContext(ctx).
		Select(s.tx.Tag.ID).
		Where(s.tx.Tag.ArtifactID.EqCol(s.tx.Artifact.ID), s.tx.Tag.DeletedAt.Eq(0))
	parentSubQuery := parent.WithContext(ctx).
		Select(parent.ID).
		Join(artifactArtifacts, parent.ID.EqCol(artifactID)).
		Where(parent.DeletedAt.Eq(0), artifactSubID.EqCol(s.tx.Artifact.ID))

	return s.tx.Artifact.WithContext(ctx).
		Where(s.tx.Artifact.ID.Gt(last), s.tx.Artifact.RepositoryID.Eq(repositoryID), s.tx.Artifact.DeletedAt.Eq(0)).
		Where(s.tx.Artifact.LastPull.Lt(before)).
		Or(s.tx.Artifact.LastPull.IsNull(), s.tx.Artifact.UpdatedAt.Lt(before)).
		Where(s.tx.Artifact.ReferrerID.IsNull()).
		Not(gen.Exists(tagSubQuery), gen.Exists(parentSubQuery)).
		Limit(int(limit)).
		Order(s.tx.Artifact.ID).
		Find()
}

// FindReferrersBySubjects finds referrers for the given subject digests
func (s *artifactRepository) FindReferrersBySubjects(ctx context.Context, repositoryID string, digests []string) ([]*models.Artifact, error) {
	if len(digests) == 0 {
		return nil, nil
	}
	subject := s.tx.Artifact.As("subject")
	return s.tx.Artifact.WithContext(ctx).
		Join(subject, subject.ID.EqCol(s.tx.Artifact.ReferrerID)).
		Where(subject.RepositoryID.Eq(repositoryID), subject.DeletedAt.Eq(0), subject.Digest.In(digests...)).
		Where(s.tx.Artifact.RepositoryID.Eq(repositoryID), s.tx.Artifact.DeletedAt.Eq(0)).
		Find()
}

// Get gets the artifact with the specified artifact ID
func (s *artifactRepository) Get(ctx context.Context, id string) (*models.Artifact, error) {
	// SELECT * FROM `repositories` WHERE `repositories`.`id` = 1 AND `repositories`.`deleted_at` = 0
	// SELECT * FROM `artifacts` WHERE `artifacts`.`id` = 1 AND `artifacts`.`deleted_at` = 0 ORDER BY `artifacts`.`id` LIMIT 1
	return s.tx.Artifact.WithContext(ctx).
		Preload(s.tx.Artifact.Repository).
		Where(s.tx.Artifact.ID.Eq(id)).First()
}

// GetByDigest gets the artifact with the specified repository and digest
func (s *artifactRepository) GetByDigest(ctx context.Context, repositoryID string, digest string) (*models.Artifact, error) {
	return s.tx.Artifact.WithContext(ctx).
		Where(s.tx.Artifact.RepositoryID.Eq(repositoryID)).
		Where(s.tx.Artifact.Digest.Eq(digest)).
		First()
}

// GetByDigests gets artifacts in a repository with the specified digests
func (s *artifactRepository) GetByDigests(ctx context.Context, repository string, digests []string) ([]*models.Artifact, error) {
	return s.tx.Artifact.WithContext(ctx).
		LeftJoin(s.tx.Repository, s.tx.Repository.ID.EqCol(s.tx.Artifact.RepositoryID)).
		Where(s.tx.Repository.Name.Eq(repository)).
		Where(s.tx.Artifact.Digest.In(digests...)).
		Preload(s.tx.Artifact.Tags.Order(s.tx.Tag.UpdatedAt.Desc()).Limit(10)).
		Find()
}

// DeleteByDigest deletes the artifact with the specified repository and digest
func (s *artifactRepository) DeleteByDigest(ctx context.Context, repository, digest string) error {
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

// AssociateBlobs associates blobs with the artifact
func (s *artifactRepository) AssociateBlobs(ctx context.Context, artifact *models.Artifact, blobs []*models.Blob) error {
	return s.tx.Artifact.Blobs.WithContext(ctx).Model(artifact).Append(blobs...)
}

// AssociateArtifact associates child artifacts with the artifact
func (s *artifactRepository) AssociateArtifact(ctx context.Context, artifact *models.Artifact, artifacts []*models.Artifact) error {
	return s.tx.Artifact.ArtifactSubs.WithContext(ctx).Model(artifact).Append(artifacts...)
}

// Incr increases the artifact pull count
func (s *artifactRepository) Incr(ctx context.Context, id string) error {
	_, err := s.tx.Artifact.WithContext(ctx).Where(s.tx.Artifact.ID.Eq(id)).
		UpdateColumns(map[string]any{
			"pull_times": gorm.Expr("pull_times + ?", 1),
			"last_pull":  time.Now().UnixMilli(),
		})
	return err
}

// CountByNamespace counts artifacts grouped by namespace ID
func (s *artifactRepository) CountByNamespace(ctx context.Context, namespaceIDs []string) (map[string]int64, error) {
	artifactCount := make(map[string]int64)
	if len(namespaceIDs) == 0 {
		return artifactCount, nil
	}
	var count []struct {
		NamespaceID string `gorm:"column:namespace_id"`
		Count       int64  `gorm:"column:count"`
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

// CountByRepository counts artifacts grouped by repository ID
func (s *artifactRepository) CountByRepository(ctx context.Context, repositoryIDs []string) (map[string]int64, error) {
	artifactCount := make(map[string]int64)
	if len(repositoryIDs) == 0 {
		return artifactCount, nil
	}
	var count []struct {
		RepositoryID string `gorm:"column:repository_id"`
		Count        int64  `gorm:"column:count"`
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

// ListArtifact lists artifacts by the specified request
func (s *artifactRepository) ListArtifact(ctx context.Context, req api.ListArtifactRequest) ([]*models.Artifact, error) {
	return s.tx.Artifact.WithContext(ctx).
		LeftJoin(s.tx.Repository, s.tx.Repository.ID.EqCol(s.tx.Artifact.RepositoryID), s.tx.Repository.Name.Eq(req.Repository)).
		LeftJoin(s.tx.Namespace, s.tx.Namespace.Name.EqCol(s.tx.Repository.Name), s.tx.Namespace.Name.Eq(req.Namespace)).
		Preload(s.tx.Artifact.Tags.Order(s.tx.Tag.UpdatedAt.Desc()).Limit(10)).
		Where(s.tx.Artifact.ID.Gt("")).
		Limit(ptr.To(req.Limit)).Find()
}

// CountArtifact counts artifacts by the specified request
func (s *artifactRepository) CountArtifact(ctx context.Context, req api.ListArtifactRequest) (int64, error) {
	return s.tx.Artifact.WithContext(ctx).
		LeftJoin(s.tx.Repository, s.tx.Repository.ID.EqCol(s.tx.Artifact.RepositoryID), s.tx.Repository.Name.Eq(req.Repository)).
		LeftJoin(s.tx.Namespace, s.tx.Namespace.Name.EqCol(s.tx.Repository.Name), s.tx.Namespace.Name.Eq(req.Namespace)).
		Count()
}

// DeleteByID deletes the artifact with the specified artifact ID
func (s *artifactRepository) DeleteByID(ctx context.Context, id string) error {
	matched, err := s.tx.Artifact.WithContext(ctx).Where(s.tx.Artifact.ID.Eq(id)).Delete()
	if err != nil {
		return err
	}
	if matched.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// DeleteByIDs deletes artifacts with the specified artifact IDs
func (s *artifactRepository) DeleteByIDs(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	_, err := s.tx.Artifact.WithContext(ctx).Where(s.tx.Artifact.ID.In(ids...)).Delete()
	return err
}

// CreateSbom creates an artifact sbom if one does not exist
func (s *artifactRepository) CreateSbom(ctx context.Context, sbom *models.ArtifactSbom) error {
	_, err := s.tx.ArtifactSbom.WithContext(ctx).Where(s.tx.ArtifactSbom.ArtifactID.Eq(sbom.ArtifactID)).First()
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		return s.tx.ArtifactSbom.WithContext(ctx).Create(sbom)
	}
	return nil
}

// CreateVulnerability creates an artifact vulnerability if one does not exist
func (s *artifactRepository) CreateVulnerability(ctx context.Context, vulnerability *models.ArtifactVulnerability) error {
	_, err := s.tx.ArtifactVulnerability.WithContext(ctx).Where(s.tx.ArtifactVulnerability.ArtifactID.Eq(vulnerability.ArtifactID)).First()
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		return s.tx.ArtifactVulnerability.WithContext(ctx).Create(vulnerability)
	}
	return nil
}

// GetSbom gets the next artifact sbom task
func (s *artifactRepository) GetSbom(ctx context.Context) (*models.ArtifactSbom, error) {
	res, err := s.tx.ArtifactSbom.WithContext(ctx).
		Where(s.tx.ArtifactSbom.Status.Eq(enums.TaskCommonStatusPending)).
		Order(s.tx.ArtifactSbom.ID.Desc()).First()
	if err == gorm.ErrRecordNotFound {
		return s.tx.ArtifactSbom.WithContext(ctx).
			Where(s.tx.ArtifactSbom.Status.Eq(enums.TaskCommonStatusDoing)).
			Where(s.tx.ArtifactSbom.UpdatedAt.Lt(time.Now().Add(time.Hour * 6 * -1).UnixMilli())).
			Order(s.tx.ArtifactSbom.ID.Desc()).First()
	}
	return res, err
}

// ClaimVulnerability claims the next pending or stale artifact vulnerability task.
func (s *artifactRepository) ClaimVulnerability(ctx context.Context, staleBefore int64) (*models.ArtifactVulnerability, error) {
	var vulnerabilityID string
	err := s.tx.Transaction(func(tx *query.Query) error {
		now := time.Now().UnixMilli()
		vulnerabilityObj, claimStatus, err := claimPendingVulnerability(ctx, tx)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			vulnerabilityObj, claimStatus, err = claimTimedOutVulnerability(ctx, tx, staleBefore)
		}
		if err != nil {
			return err
		}

		taskQuery := tx.ArtifactVulnerability.WithContext(ctx).
			Where(tx.ArtifactVulnerability.ID.Eq(vulnerabilityObj.ID)).
			Where(tx.ArtifactVulnerability.Status.Eq(claimStatus))
		if claimStatus == enums.TaskCommonStatusDoing {
			taskQuery = taskQuery.Where(tx.ArtifactVulnerability.UpdatedAt.Lt(staleBefore))
		}
		matched, err := taskQuery.UpdateColumns(map[string]any{
			query.ArtifactVulnerability.Status.ColumnName().String():    enums.TaskCommonStatusDoing,
			query.ArtifactVulnerability.UpdatedAt.ColumnName().String(): now,
		})
		if err != nil {
			return err
		}
		if matched.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		vulnerabilityID = vulnerabilityObj.ID
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.tx.ArtifactVulnerability.WithContext(ctx).
		Preload(s.tx.ArtifactVulnerability.Artifact.Repository).
		Where(s.tx.ArtifactVulnerability.ID.Eq(vulnerabilityID)).First()
}

func claimPendingVulnerability(ctx context.Context, tx *query.Query) (*models.ArtifactVulnerability, enums.TaskCommonStatus, error) {
	vulnerabilityObj, err := tx.ArtifactVulnerability.WithContext(ctx).
		Where(tx.ArtifactVulnerability.Status.Eq(enums.TaskCommonStatusPending)).
		Order(tx.ArtifactVulnerability.ID.Desc()).First()
	return vulnerabilityObj, enums.TaskCommonStatusPending, err
}

func claimTimedOutVulnerability(ctx context.Context, tx *query.Query, timeoutBefore int64) (*models.ArtifactVulnerability, enums.TaskCommonStatus, error) {
	vulnerabilityObj, err := tx.ArtifactVulnerability.WithContext(ctx).
		Where(tx.ArtifactVulnerability.Status.Eq(enums.TaskCommonStatusDoing)).
		Where(tx.ArtifactVulnerability.UpdatedAt.Lt(timeoutBefore)).
		Order(tx.ArtifactVulnerability.ID.Desc()).First()
	return vulnerabilityObj, enums.TaskCommonStatusDoing, err
}

// UpdateSbom updates the artifact sbom with the specified artifact ID
func (s *artifactRepository) UpdateSbom(ctx context.Context, artifactID string, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}
	_, err := s.tx.ArtifactSbom.WithContext(ctx).Where(s.tx.ArtifactSbom.ArtifactID.Eq(artifactID)).UpdateColumns(updates)
	return err
}

// UpdateVulnerability updates the artifact vulnerability with the specified artifact ID
func (s *artifactRepository) UpdateVulnerability(ctx context.Context, artifactID string, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}
	_, err := s.tx.ArtifactVulnerability.WithContext(ctx).Where(s.tx.ArtifactVulnerability.ArtifactID.Eq(artifactID)).UpdateColumns(updates)
	return err
}

// GetNamespaceSize gets the total artifact blob size in the specified namespace
func (s *artifactRepository) GetNamespaceSize(ctx context.Context, namespaceID string) (int64, error) {
	var res struct {
		BlobsSize int64 `gorm:"column:blobs_size"`
	}
	err := s.tx.Artifact.WithContext(ctx).Select(s.tx.Artifact.BlobsSize.Sum().As("blobs_size")).
		Where(s.tx.Artifact.NamespaceID.Eq(namespaceID)).Scan(&res)
	if err != nil {
		return 0, err
	}
	return res.BlobsSize, nil
}

// GetRepositorySize gets the total artifact blob size in the specified repository
func (s *artifactRepository) GetRepositorySize(ctx context.Context, repositoryID string) (int64, error) {
	var res struct {
		BlobsSize int64 `gorm:"column:blobs_size"`
	}
	err := s.tx.Artifact.WithContext(ctx).Select(s.tx.Artifact.BlobsSize.Sum().As("blobs_size")).
		Where(s.tx.Artifact.RepositoryID.Eq(repositoryID)).Scan(&res)
	if err != nil {
		return 0, err
	}
	return res.BlobsSize, nil
}

// GetReferrers gets artifacts that refer to the specified subject digest
func (s *artifactRepository) GetReferrers(ctx context.Context, repositoryID string, digest string, artifactTypes []string) ([]*models.Artifact, error) {
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

// IsArtifactAssociatedWithArtifact checks whether the artifact is associated with another artifact
func (s *artifactRepository) IsArtifactAssociatedWithArtifact(ctx context.Context, artifactID string) error {
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
