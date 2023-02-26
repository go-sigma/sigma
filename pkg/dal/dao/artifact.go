// The MIT License (MIT)
//
// Copyright © 2023 Tosone <i@tosone.cn>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package dao

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/ximager/ximager/pkg/dal/models"
	"github.com/ximager/ximager/pkg/dal/query"
	"github.com/ximager/ximager/pkg/types"
)

// ArtifactService is the interface that provides the artifact service methods.
type ArtifactService interface {
	// Save save a new artifact if conflict update.
	Save(ctx context.Context, artifact *models.Artifact) (*models.Artifact, error)
	// Get gets the artifact with the specified artifact ID.
	Get(ctx context.Context, id uint64) (*models.Artifact, error)
	// GetByDigest gets the artifact with the specified digest.
	GetByDigest(ctx context.Context, repository, digest string) (*models.Artifact, error)
	// DeleteByDigest deletes the artifact with the specified digest.
	DeleteByDigest(ctx context.Context, repository, digest string) error
	// AssociateBlobs associates the blobs with the artifact.
	AssociateBlobs(ctx context.Context, artifact *models.Artifact, blobs []*models.Blob) error
	// CountByNamespace counts the artifacts by the specified namespace.
	CountByNamespace(ctx context.Context, namespaceIDs []uint64) (map[uint64]int64, error)
	// CountByRepository counts the artifacts by the specified repository.
	CountByRepository(ctx context.Context, repositoryIDs []uint64) (map[uint64]int64, error)
	// Incr increases the pull times of the artifact.
	Incr(ctx context.Context, id uint64) error
	// ListArtifact lists the artifacts by the specified request.
	ListArtifact(ctx context.Context, req types.ListArtifactRequest) ([]*models.Artifact, error)
	// CountArtifact counts the artifacts by the specified request.
	CountArtifact(ctx context.Context, req types.ListArtifactRequest) (int64, error)
	// DeleteByID deletes the artifact with the specified artifact ID.
	DeleteByID(ctx context.Context, id uint64) error
}

type artifactService struct {
	tx *query.Query
}

// NewArtifactService creates a new artifact service.
func NewArtifactService(txs ...*query.Query) ArtifactService {
	tx := query.Q
	if len(txs) > 0 {
		tx = txs[0]
	}
	return &artifactService{
		tx: tx,
	}
}

// Save save a new artifact if conflict update.
func (s *artifactService) Save(ctx context.Context, artifact *models.Artifact) (*models.Artifact, error) {
	err := s.tx.Artifact.WithContext(ctx).Save(artifact)
	if err != nil {
		return nil, err
	}
	return artifact, nil
}

// Get gets the artifact with the specified artifact ID.
func (s *artifactService) Get(ctx context.Context, id uint64) (*models.Artifact, error) {
	artifact, err := s.tx.Artifact.WithContext(ctx).Where(s.tx.Artifact.ID.Eq(id)).First()
	if err != nil {
		return nil, err
	}
	return artifact, nil
}

// GetByDigest gets the artifact with the specified digest.
func (s *artifactService) GetByDigest(ctx context.Context, repository, digest string) (*models.Artifact, error) {
	artifact, err := s.tx.Artifact.WithContext(ctx).Where(s.tx.Artifact.Digest.Eq(digest)).First()
	if err != nil {
		return nil, err
	}
	return artifact, nil
}

// DeleteByDigest deletes the artifact with the specified digest.
func (s *artifactService) DeleteByDigest(ctx context.Context, repository, digest string) error {
	err := s.tx.Transaction(func(tx *query.Query) error {
		artifact, err := tx.Artifact.WithContext(ctx).Where(tx.Artifact.Digest.Eq(digest)).First()
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

func (s *artifactService) AssociateBlobs(ctx context.Context, artifact *models.Artifact, blobs []*models.Blob) error {
	return s.tx.Artifact.Blobs.Model(artifact).Append(blobs...)
}

// Incr increases the pull times of the artifact.
func (s *artifactService) Incr(ctx context.Context, id uint64) error {
	_, err := s.tx.Artifact.WithContext(ctx).Where(s.tx.Tag.ID.Eq(id)).
		UpdateColumns(map[string]interface{}{
			"pull_times": gorm.Expr("pull_times + ?", 1),
			"last_pull":  time.Now(),
		})
	if err != nil {
		return err
	}
	return nil
}

// CountByNamespace counts the artifacts by the specified namespace.
func (s *artifactService) CountByNamespace(ctx context.Context, namespaceIDs []uint64) (map[uint64]int64, error) {
	artifactCount := make(map[uint64]int64)
	if len(namespaceIDs) == 0 {
		return artifactCount, nil
	}
	var count []struct {
		NamespaceID uint64 `gorm:"column:namespace_id"`
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

// CountByRepository counts the artifacts by the specified repository.
func (s *artifactService) CountByRepository(ctx context.Context, repositoryIDs []uint64) (map[uint64]int64, error) {
	artifactCount := make(map[uint64]int64)
	if len(repositoryIDs) == 0 {
		return artifactCount, nil
	}
	var count []struct {
		RepositoryID uint64 `gorm:"column:repository_id"`
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

// ListArtifact lists the artifacts by the specified request.
func (s *artifactService) ListArtifact(ctx context.Context, req types.ListArtifactRequest) ([]*models.Artifact, error) {
	query := s.tx.Artifact.WithContext(ctx).
		LeftJoin(s.tx.Repository, s.tx.Repository.ID.EqCol(s.tx.Artifact.RepositoryID), s.tx.Repository.Name.Eq(req.Repository)).
		LeftJoin(s.tx.Namespace, s.tx.Namespace.Name.EqCol(s.tx.Repository.Name), s.tx.Namespace.Name.Eq(req.Namespace)).
		Preload(s.tx.Artifact.Tags.Order(s.tx.Tag.UpdatedAt.Desc()).Limit(10)).
		Offset(req.PageSize * (req.PageNum - 1)).Limit(req.PageSize)
	return query.Find()
}

// CountArtifact counts the artifacts by the specified request.
func (s *artifactService) CountArtifact(ctx context.Context, req types.ListArtifactRequest) (int64, error) {
	return s.tx.Artifact.WithContext(ctx).
		LeftJoin(s.tx.Repository, s.tx.Repository.ID.EqCol(s.tx.Artifact.RepositoryID), s.tx.Repository.Name.Eq(req.Repository)).
		LeftJoin(s.tx.Namespace, s.tx.Namespace.Name.EqCol(s.tx.Repository.Name), s.tx.Namespace.Name.Eq(req.Namespace)).
		Count()
}

// DeleteByID deletes the artifact with the specified ID.
func (s *artifactService) DeleteByID(ctx context.Context, id uint64) error {
	matched, err := s.tx.Artifact.WithContext(ctx).Where(s.tx.Artifact.ID.Eq(id)).Delete()
	if err != nil {
		return err
	}
	if matched.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
