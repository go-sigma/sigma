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
	"database/sql/driver"
	"errors"
	"fmt"
	"time"

	"github.com/jinzhu/copier"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

//go:generate go tool mockgen -destination=tag_mocks.go -package=registry github.com/go-sigma/sigma/pkg/dal/repository/registry TagRepository

// TagRepository defines tag repository operations
type TagRepository interface {
	// Create creates a tag or updates the existing tag artifact on name conflict
	Create(ctx context.Context, tag *models.Tag) error
	// FindWithQuantityCursor finds tags outside the newest retained quantity with cursor pagination
	FindWithQuantityCursor(ctx context.Context, repositoryID string, quantity, limit int, last string) ([]*models.Tag, error)
	// FindWithDayCursor finds tags older than the given number of days with cursor pagination
	FindWithDayCursor(ctx context.Context, repositoryID string, day, limit int, last string) ([]*models.Tag, error)
	// GetByID gets the tag with the specified tag ID
	GetByID(ctx context.Context, tagID string) (*models.Tag, error)
	// GetByName gets the tag with the specified repository and tag name
	GetByName(ctx context.Context, repositoryID string, tag string) (*models.Tag, error)
	// GetByArtifactID gets the tag with the specified repository and artifact ID
	GetByArtifactID(ctx context.Context, repositoryID, artifactID string) (*models.Tag, error)
	// DeleteByName deletes the tag with the specified repository and tag name
	DeleteByName(ctx context.Context, repositoryID string, tag string) error
	// DeleteByArtifactID deletes tags with the specified artifact ID
	DeleteByArtifactID(ctx context.Context, artifactID string) error
	// Incr increases the tag pull count
	Incr(ctx context.Context, id string) error
	// ListByDtPagination lists tags in the specified repository with cursor pagination
	ListByDtPagination(ctx context.Context, repository string, limit int, lastID ...string) ([]*models.Tag, error)
	// ListTag lists tags in a repository with filtering, pagination, and sorting
	ListTag(ctx context.Context, repositoryID string, name *string, types []enums.ArtifactType, pagination api.Pagination, sort api.Sortable) ([]*models.Tag, int64, error)
	// CountByNamespace counts tags grouped by namespace ID
	CountByNamespace(ctx context.Context, namespaceIDs []string) (map[string]int64, error)
	// CountByRepositories counts tags grouped by repository ID
	CountByRepositories(ctx context.Context, repositoryIDs []string) (map[string]int64, error)
	// CountByRepository counts tags in the specified repository
	CountByRepository(ctx context.Context, repositoryID string) (int64, error)
	// DeleteByID deletes the tag with the specified tag ID
	DeleteByID(ctx context.Context, id string) error
	// CountByArtifact counts tags grouped by artifact ID
	CountByArtifact(ctx context.Context, artifactIDs []string) (map[string]int64, error)
}

type tagRepository struct {
	tx *query.Query
}

// NewTagRepository creates a new tag repository with the optional query transaction
func NewTagRepository(txs ...*query.Query) TagRepository {
	tx := query.Q
	if len(txs) > 0 {
		tx = txs[0]
	}
	return &tagRepository{
		tx: tx,
	}
}

// Create creates a tag or updates the existing tag artifact on name conflict
func (s *tagRepository) Create(ctx context.Context, tag *models.Tag) error {
	findTagObj, err := s.tx.Tag.WithContext(ctx).Where(
		s.tx.Tag.RepositoryID.Eq(tag.RepositoryID),
		s.tx.Tag.Name.Eq(tag.Name)).First()
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		err = s.tx.Tag.WithContext(ctx).Create(tag)
		if err != nil {
			return err
		}
		return nil
	}
	_, err = s.tx.Tag.WithContext(ctx).Where(
		s.tx.Tag.RepositoryID.Eq(tag.RepositoryID),
		s.tx.Tag.Name.Eq(tag.Name)).Updates(map[string]any{
		query.Tag.ArtifactID.ColumnName().String(): tag.ArtifactID,
	})
	if err != nil {
		return err
	}
	findTagObj.ArtifactID = tag.ArtifactID
	return copier.Copy(findTagObj, tag)
}

// FindWithQuantityCursor finds tags outside the newest retained quantity with cursor pagination
func (s *tagRepository) FindWithQuantityCursor(ctx context.Context, repositoryID string, quantity, limit int, last string) ([]*models.Tag, error) {
	keptSubQuery := s.tx.Tag.WithContext(ctx).
		Select(s.tx.Tag.ID).
		Where(s.tx.Tag.RepositoryID.Eq(repositoryID)).
		Order(s.tx.Tag.UpdatedAt.Desc(), s.tx.Tag.ID.Desc()).
		Limit(quantity)
	return s.tx.Tag.WithContext(ctx).
		Where(s.tx.Tag.RepositoryID.Eq(repositoryID)).
		Where(s.tx.Tag.ID.Gt(last)).
		Where(s.tx.Tag.Columns(s.tx.Tag.ID).NotIn(keptSubQuery)).
		Order(s.tx.Tag.ID).
		Limit(limit).
		Find()
}

// FindWithDayCursor finds tags older than the given number of days with cursor pagination
func (s *tagRepository) FindWithDayCursor(ctx context.Context, repositoryID string, day, limit int, last string) ([]*models.Tag, error) {
	before := time.Now().Add(-24 * time.Hour * time.Duration(day)).UTC().UnixMilli()
	return s.tx.Tag.WithContext(ctx).
		Where(s.tx.Tag.RepositoryID.Eq(repositoryID)).
		Where(s.tx.Tag.ID.Gt(last)).
		Where(s.tx.Tag.UpdatedAt.Lt(before)).
		Limit(limit).Order(s.tx.Tag.ID).Find()
}

// GetByID gets the tag with the specified tag ID
func (s *tagRepository) GetByID(ctx context.Context, tagID string) (*models.Tag, error) {
	q := s.tx.Tag.WithContext(ctx).Where(s.tx.Tag.ID.Eq(tagID))
	q.UnderlyingDB().Preload("Artifact.ArtifactSubs.Vulnerability")
	q.UnderlyingDB().Preload("Artifact.ArtifactSubs.Sbom")
	q.Preload(s.tx.Tag.Artifact.ArtifactSubs)
	q.Preload(s.tx.Tag.Artifact.Vulnerability)
	q.Preload(s.tx.Tag.Artifact.Sbom)
	return q.First()
}

// GetByName gets the tag with the specified repository and tag name
func (s *tagRepository) GetByName(ctx context.Context, repositoryID string, tag string) (*models.Tag, error) {
	return s.tx.Tag.WithContext(ctx).
		Where(s.tx.Tag.RepositoryID.Eq(repositoryID), s.tx.Tag.Name.Eq(tag)).
		Preload(s.tx.Tag.Artifact).
		First()
}

// GetByArtifactID gets the tag with the specified repository and artifact ID
func (s *tagRepository) GetByArtifactID(ctx context.Context, repositoryID, artifactID string) (*models.Tag, error) {
	return s.tx.Tag.WithContext(ctx).Where(s.tx.Tag.RepositoryID.Eq(repositoryID), s.tx.Tag.ArtifactID.Eq(artifactID)).First()
}

// DeleteByName deletes the tag with the specified repository and tag name
func (s *tagRepository) DeleteByName(ctx context.Context, repositoryID string, tag string) error {
	tagObj, err := s.tx.Tag.WithContext(ctx).Where(s.tx.Tag.RepositoryID.Eq(repositoryID), s.tx.Tag.Name.Eq(tag)).First()
	if err != nil {
		return err
	}
	delTagObj := &models.Tag{ID: tagObj.ID}
	matched, err := s.tx.Tag.WithContext(ctx).Delete(delTagObj)
	if err != nil {
		return err
	}
	if matched.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// DeleteByArtifactID deletes tags with the specified artifact ID
func (s *tagRepository) DeleteByArtifactID(ctx context.Context, artifactID string) error {
	// sql: update tags set deleted_at = now() where artifact_id = ?
	_, err := s.tx.Tag.WithContext(ctx).Where(s.tx.Tag.ArtifactID.Eq(artifactID)).Delete()
	return err
}

// Incr increases the tag pull count
func (s *tagRepository) Incr(ctx context.Context, id string) error {
	_, err := s.tx.Tag.WithContext(ctx).Where(s.tx.Tag.ID.Eq(id)).
		UpdateColumns(map[string]any{
			"pull_times": gorm.Expr("pull_times + ?", 1),
			"last_pull":  time.Now().UnixMilli(),
		})
	return err
}

// ListByDtPagination lists tags in the specified repository with cursor pagination
func (s *tagRepository) ListByDtPagination(ctx context.Context, repository string, limit int, lastID ...string) ([]*models.Tag, error) {
	do := s.tx.Tag.WithContext(ctx).
		LeftJoin(s.tx.Repository, s.tx.Tag.RepositoryID.EqCol(s.tx.Repository.ID)).
		Where(s.tx.Repository.Name.Eq(repository))
	if len(lastID) > 0 {
		do = do.Where(s.tx.Tag.ID.Gt(lastID[0]))
	}
	tags, err := do.Order(s.tx.Tag.ID).Limit(limit).Find()
	return tags, err
}

// ListTag lists tags in a repository with filtering, pagination, and sorting
func (s *tagRepository) ListTag(ctx context.Context, repositoryID string, name *string, types []enums.ArtifactType, pagination api.Pagination, sort api.Sortable) ([]*models.Tag, int64, error) {
	var mTypes []driver.Valuer

	if len(types) > 0 {
		for _, t := range types {
			mTypes = append(mTypes, t)
		}
	}
	pagination = utils.NormalizePagination(pagination)
	q := s.tx.Tag.WithContext(ctx).Where(s.tx.Tag.RepositoryID.Eq(repositoryID))
	if len(types) > 0 {
		q = q.RightJoin(s.tx.Artifact, s.tx.Tag.ArtifactID.EqCol(s.tx.Artifact.ID), s.tx.Artifact.Type.In(mTypes...))
	} else {
		q = q.RightJoin(s.tx.Artifact, s.tx.Tag.ArtifactID.EqCol(s.tx.Artifact.ID))
	}
	if name != nil {
		q = q.Where(s.tx.Tag.Name.Like(fmt.Sprintf("%%%s%%", ptr.To(name))))
	}
	field, ok := s.tx.Tag.GetFieldByName(ptr.To(sort.Sort))
	if ok {
		switch ptr.To(sort.Method) {
		case enums.SortMethodDesc:
			q = q.Order(field.Desc())
		case enums.SortMethodAsc:
			q = q.Order(field)
		default:
			q = q.Order(s.tx.Tag.UpdatedAt.Desc())
		}
	} else {
		q = q.Order(s.tx.Tag.UpdatedAt.Desc())
	}
	if len(types) > 0 {
		q = q.Preload(s.tx.Tag.Artifact.ArtifactSubs.On(s.tx.Artifact.Type.In(mTypes...)))
	} else {
		q = q.Preload(s.tx.Tag.Artifact.ArtifactSubs)
	}
	q = q.Preload(s.tx.Tag.Artifact.Vulnerability)
	q = q.Preload(s.tx.Tag.Artifact.Sbom)
	q.UnderlyingDB().Preload("Artifact.ArtifactSubs.Vulnerability")
	q.UnderlyingDB().Preload("Artifact.ArtifactSubs.Sbom")
	return q.FindByPage(ptr.To(pagination.Limit)*(ptr.To(pagination.Page)-1), ptr.To(pagination.Limit))
}

// DeleteByID deletes the tag with the specified tag ID
func (s *tagRepository) DeleteByID(ctx context.Context, id string) error {
	t := &models.Tag{ID: id}
	matched, err := s.tx.Tag.WithContext(ctx).Delete(t)
	if err != nil {
		return err
	}
	if matched.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// CountByArtifact counts tags grouped by artifact ID
func (s *tagRepository) CountByArtifact(ctx context.Context, artifactIDs []string) (map[string]int64, error) {
	tagCount := make(map[string]int64)
	var count []struct {
		ArtifactID string `gorm:"column:artifact_id"`
		Count      int64  `gorm:"column:count"`
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

// CountByNamespace counts tags grouped by namespace ID
func (s *tagRepository) CountByNamespace(ctx context.Context, namespaceIDs []string) (map[string]int64, error) {
	tagCount := make(map[string]int64)
	var count []struct {
		NamespaceID string `gorm:"column:namespace_id"`
		Count       int64  `gorm:"column:count"`
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

// CountByRepositories counts tags grouped by repository ID
func (s *tagRepository) CountByRepositories(ctx context.Context, repositoryIDs []string) (map[string]int64, error) {
	tagCount := make(map[string]int64)
	var count []struct {
		RepositoryID string `gorm:"column:repository_id"`
		Count        int64  `gorm:"column:count"`
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

// CountByRepository counts tags in the specified repository
func (s *tagRepository) CountByRepository(ctx context.Context, repositoryID string) (int64, error) {
	return s.tx.Tag.WithContext(ctx).Where(s.tx.Tag.RepositoryID.Eq(repositoryID)).Count()
}
