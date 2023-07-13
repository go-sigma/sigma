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

	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

//go:generate mockgen -destination=mocks/tag.go -package=mocks github.com/go-sigma/sigma/pkg/dal/dao TagService
//go:generate mockgen -destination=mocks/tag_factory.go -package=mocks github.com/go-sigma/sigma/pkg/dal/dao TagServiceFactory

// TagService is the interface that provides the tag service methods.
type TagService interface {
	// Create save a new tag if conflict do nothing.
	Create(ctx context.Context, tag *models.Tag) error
	// Get gets the tag with the specified tag ID.
	GetByID(ctx context.Context, tagID int64) (*models.Tag, error)
	// GetByName gets the tag with the specified tag name.
	GetByName(ctx context.Context, repositoryID int64, tag string) (*models.Tag, error)
	// DeleteByName deletes the tag with the specified tag name.
	DeleteByName(ctx context.Context, repositoryID int64, tag string) error
	// DeleteByArtifactID deletes the tag with the specified artifact ID.
	DeleteByArtifactID(ctx context.Context, artifactID int64) error
	// Incr increases the pull times of the artifact.
	Incr(ctx context.Context, id int64) error
	// ListByDtPagination lists the tags by the specified repository and pagination.
	ListByDtPagination(ctx context.Context, repository string, limit int, lastID ...int64) ([]*models.Tag, error)
	// ListTag lists the tags by the specified request.
	ListTag(ctx context.Context, req types.ListTagRequest) ([]*models.Tag, error)
	// CountArtifact counts the artifacts by the specified request.
	CountTag(ctx context.Context, req types.ListTagRequest) (int64, error)
	// CountByNamespace counts the tags by the specified namespace.
	CountByNamespace(ctx context.Context, namespaceIDs []int64) (map[int64]int64, error)
	// CountByRepository counts the tags by the specified repository.
	CountByRepository(ctx context.Context, repositoryIDs []int64) (map[int64]int64, error)
	// DeleteByID deletes the tag with the specified tag ID.
	DeleteByID(ctx context.Context, id int64) error
	// CountByArtifact counts the tags by the specified artifact.
	CountByArtifact(ctx context.Context, artifactIDs []int64) (map[int64]int64, error)
}

type tagService struct {
	tx *query.Query
}

// TagServiceFactory is the interface that provides the tag service factory methods.
type TagServiceFactory interface {
	New(txs ...*query.Query) TagService
}

type tagServiceFactory struct{}

// NewTagServiceFactory creates a new tag service factory.
func NewTagServiceFactory() TagServiceFactory {
	return &tagServiceFactory{}
}

func (f *tagServiceFactory) New(txs ...*query.Query) TagService {
	tx := query.Q
	if len(txs) > 0 {
		tx = txs[0]
	}
	return &tagService{
		tx: tx,
	}
}

// Create save a new tag if conflict do nothing.
func (s *tagService) Create(ctx context.Context, tag *models.Tag) error {
	return s.tx.Tag.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(tag)
}

// Get gets the tag with the specified tag ID.
func (s *tagService) GetByID(ctx context.Context, tagID int64) (*models.Tag, error) {
	return s.tx.Tag.WithContext(ctx).Where(s.tx.Tag.ID.Eq(tagID)).First()
}

// GetByName gets the tag with the specified tag name.
func (s *tagService) GetByName(ctx context.Context, repositoryID int64, tag string) (*models.Tag, error) {
	return s.tx.Tag.WithContext(ctx).
		Where(s.tx.Tag.RepositoryID.Eq(repositoryID), s.tx.Tag.Name.Eq(tag)).
		Preload(s.tx.Tag.Artifact).
		First()
}

// DeleteByName deletes the tag with the specified tag name.
func (s *tagService) DeleteByName(ctx context.Context, repositoryID int64, tag string) error {
	_, err := s.tx.Tag.WithContext(ctx).Where(s.tx.Tag.RepositoryID.Eq(repositoryID), s.tx.Tag.Name.Eq(tag)).Delete()
	return err
}

// DeleteByArtifactID deletes the tag with the specified artifact ID.
func (s *tagService) DeleteByArtifactID(ctx context.Context, artifactID int64) error {
	// sql: update tags set deleted_at = now() where artifact_id = ?
	_, err := s.tx.Tag.WithContext(ctx).Where(s.tx.Tag.ArtifactID.Eq(artifactID)).Delete()
	return err
}

// Incr increases the pull times of the artifact.
func (s *tagService) Incr(ctx context.Context, id int64) error {
	_, err := s.tx.Tag.WithContext(ctx).Where(s.tx.Tag.ID.Eq(id)).
		UpdateColumns(map[string]interface{}{
			"pull_times": gorm.Expr("pull_times + ?", 1),
			"last_pull":  time.Now(),
		})
	return err
}

// ListByDtPagination lists the tags by the specified repository and pagination.
func (s *tagService) ListByDtPagination(ctx context.Context, repository string, limit int, lastID ...int64) ([]*models.Tag, error) {
	do := s.tx.Tag.WithContext(ctx).
		LeftJoin(s.tx.Repository, s.tx.Tag.RepositoryID.EqCol(s.tx.Repository.ID)).
		Where(s.tx.Repository.Name.Eq(repository))
	if len(lastID) > 0 {
		do = do.Where(s.tx.Tag.ID.Gt(lastID[0]))
	}
	tags, err := do.Order(s.tx.Tag.ID).Limit(limit).Find()
	return tags, err
}

// ListTag lists the tags by the specified request.
func (s *tagService) ListTag(ctx context.Context, req types.ListTagRequest) ([]*models.Tag, error) {
	return s.tx.Tag.WithContext(ctx).
		LeftJoin(s.tx.Repository, s.tx.Tag.RepositoryID.EqCol(s.tx.Repository.ID)).
		Where(s.tx.Repository.Name.Eq(req.Repository)).
		Where(s.tx.Tag.ID.Gt(int64(ptr.To(req.Page)))).Limit(ptr.To(req.Limit)).Find()
}

// CountArtifact counts the artifacts by the specified request.
func (s *tagService) CountTag(ctx context.Context, req types.ListTagRequest) (int64, error) {
	return s.tx.Tag.WithContext(ctx).
		LeftJoin(s.tx.Repository, s.tx.Tag.RepositoryID.EqCol(s.tx.Repository.ID)).
		Where(s.tx.Repository.Name.Eq(req.Repository)).
		Count()
}

// DeleteByID deletes the tag with the specified tag ID.
func (s *tagService) DeleteByID(ctx context.Context, id int64) error {
	matched, err := s.tx.Tag.WithContext(ctx).Where(s.tx.Tag.ID.Eq(id)).Delete()
	if err != nil {
		return err
	}
	if matched.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// CountByArtifact counts the tags by the specified artifact.
func (s *tagService) CountByArtifact(ctx context.Context, artifactIDs []int64) (map[int64]int64, error) {
	tagCount := make(map[int64]int64)
	var count []struct {
		ArtifactID int64 `gorm:"column:artifact_id"`
		Count      int64 `gorm:"column:count"`
	}
	err := s.tx.Tag.WithContext(ctx).
		LeftJoin(s.tx.Artifact, s.tx.Tag.ArtifactID.EqCol(s.tx.Artifact.ID)).
		Where(s.tx.Artifact.ID.In(artifactIDs...)).
		Group(s.tx.Artifact.ID).
		Select(s.tx.Artifact.ID.As("artifact_id"), s.tx.Tag.ID.Count().As("count")).
		Scan(&count)
	if err != nil {
		return nil, err
	}
	for _, c := range count {
		tagCount[c.ArtifactID] = c.Count
	}
	return tagCount, nil
}

// CountByNamespace counts the tags by the specified namespace.
func (s *tagService) CountByNamespace(ctx context.Context, namespaceIDs []int64) (map[int64]int64, error) {
	tagCount := make(map[int64]int64)
	var count []struct {
		NamespaceID int64 `gorm:"column:namespace_id"`
		Count       int64 `gorm:"column:count"`
	}
	err := s.tx.Tag.WithContext(ctx).
		LeftJoin(s.tx.Repository, s.tx.Repository.ID.EqCol(s.tx.Tag.RepositoryID)).
		Where(s.tx.Repository.NamespaceID.In(namespaceIDs...)).
		Group(s.tx.Repository.NamespaceID).
		Select(s.tx.Repository.NamespaceID, s.tx.Tag.ID.Count().As("count")).
		Scan(&count)
	if err != nil {
		return nil, err
	}
	for _, c := range count {
		tagCount[c.NamespaceID] = c.Count
	}
	return tagCount, nil
}

// CountByRepository counts the tags by the specified repository.
func (s *tagService) CountByRepository(ctx context.Context, repositoryIDs []int64) (map[int64]int64, error) {
	tagCount := make(map[int64]int64)
	var count []struct {
		RepositoryID int64 `gorm:"column:repository_id"`
		Count        int64 `gorm:"column:count"`
	}
	err := s.tx.Tag.WithContext(ctx).
		Where(s.tx.Tag.RepositoryID.In(repositoryIDs...)).
		Group(s.tx.Tag.RepositoryID).
		Select(s.tx.Tag.RepositoryID, s.tx.Tag.ID.Count().As("count")).
		Scan(&count)
	if err != nil {
		return nil, err
	}
	for _, c := range count {
		tagCount[c.RepositoryID] = c.Count
	}
	return tagCount, nil
}
