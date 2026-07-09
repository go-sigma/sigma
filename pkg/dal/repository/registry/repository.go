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
	"fmt"
	"log/slog"

	"gorm.io/gen"
	"gorm.io/gen/field"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/server/errcode"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

//go:generate mockgen -destination=repository_mocks.go -package=registry github.com/go-sigma/sigma/pkg/dal/repository/registry RepositoryRepository

// RepositoryRepository defines repository operations
type RepositoryRepository interface {
	// Create creates a repository
	Create(ctx context.Context, repositoryObj *models.Repository) error
	// FindAll finds repositories in a namespace with cursor pagination
	FindAll(ctx context.Context, namespaceID string, limit int64, last string) ([]*models.Repository, error)
	// FindDeletableWithCursor finds empty repositories that are safe to delete with cursor pagination
	FindDeletableWithCursor(ctx context.Context, namespaceID string, before int64, last string, limit int64) ([]*models.Repository, error)
	// FindRecentlyUpdated finds recently updated repositories for cache prewarming.
	FindRecentlyUpdated(ctx context.Context, limit int) ([]*models.Repository, error)
	// Get gets the repository with the specified repository ID
	Get(ctx context.Context, repositoryID string) (*models.Repository, error)
	// GetByName gets the repository with the specified repository name
	GetByName(context.Context, string) (*models.Repository, error)
	// ListByDtPagination lists repositories with cursor pagination
	ListByDtPagination(ctx context.Context, limit int, lastID ...string) ([]*models.Repository, error)
	// ListWithScrollable lists repositories visible to a user with cursor pagination
	ListWithScrollable(ctx context.Context, namespaceID, userID string, name *string, limit int, lastID string) ([]*models.Repository, error)
	// ListRepository lists repositories in a namespace with pagination and sorting
	ListRepository(ctx context.Context, namespaceID string, name *string, pagination api.Pagination, sort api.Sortable) ([]*models.Repository, int64, error)
	// ListRepositoryWithAuth lists repositories visible to a user with pagination and sorting
	ListRepositoryWithAuth(ctx context.Context, namespaceID, userID string, name *string, pagination api.Pagination, sort api.Sortable) ([]*models.Repository, int64, error)
	// CountRepository counts repositories in a namespace
	CountRepository(ctx context.Context, namespaceID string, name *string) (int64, error)
	// UpdateRepository updates the repository with the specified ID
	UpdateRepository(ctx context.Context, id string, updates map[string]any) error
	// CountByNamespace counts repositories grouped by namespace ID
	CountByNamespace(ctx context.Context, namespaceIDs []string) (map[string]int64, error)
	// DeleteByID deletes the repository with the specified repository ID
	DeleteByID(ctx context.Context, id string) error
	// DeleteEmpty deletes empty repositories and returns their names
	DeleteEmpty(ctx context.Context, namespaceID *string) ([]string, error)
	// IncrementSize atomically increments repository size by delta.
	// Returns quota-exceed error if SizeLimit > 0 and size+delta > SizeLimit.
	IncrementSize(ctx context.Context, repositoryID string, delta int64) error
	// DecrementSize atomically decrements repository size by delta (floored at 0).
	DecrementSize(ctx context.Context, repositoryID string, delta int64) error
	// FindWithCursorDirty finds repositories with size_dirty = true using cursor pagination
	FindWithCursorDirty(ctx context.Context, limit int64, last string) ([]*models.Repository, error)
	// UpdateSizeAndClearDirty updates size and clears the dirty flag
	UpdateSizeAndClearDirty(ctx context.Context, repositoryID string, size int64) error
	// ClearSizeDirty clears only the size_dirty flag
	ClearSizeDirty(ctx context.Context, repositoryID string) error
}

type repositoryRepository struct {
	tx *query.Query
}

// NewRepositoryRepository creates a new repository repository with the optional query transaction
func NewRepositoryRepository(txs ...*query.Query) RepositoryRepository {
	tx := query.Q
	if len(txs) > 0 {
		tx = txs[0]
	}
	return &repositoryRepository{
		tx: tx,
	}
}

// Create creates a new repository
func (s *repositoryRepository) Create(ctx context.Context, repositoryObj *models.Repository) error {
	return s.tx.Repository.WithContext(ctx).Create(repositoryObj)
}

// FindAll finds repositories in a namespace with cursor pagination
func (s *repositoryRepository) FindAll(ctx context.Context, namespaceID string, limit int64, last string) ([]*models.Repository, error) {
	return s.tx.Repository.WithContext(ctx).
		Where(s.tx.Repository.ID.Gt(last), s.tx.Repository.NamespaceID.Eq(namespaceID)).
		Limit(int(limit)).Order(s.tx.Repository.ID).Find()
}

// FindDeletableWithCursor finds empty repositories that are safe to delete with cursor pagination
func (s *repositoryRepository) FindDeletableWithCursor(ctx context.Context, namespaceID string, before int64, last string, limit int64) ([]*models.Repository, error) {
	tagSubQuery := s.tx.Tag.WithContext(ctx).
		Select(s.tx.Tag.ID).
		Where(s.tx.Tag.RepositoryID.EqCol(s.tx.Repository.ID))
	return s.tx.Repository.WithContext(ctx).
		Where(s.tx.Repository.ID.Gt(last), s.tx.Repository.NamespaceID.Eq(namespaceID)).
		Where(s.tx.Repository.UpdatedAt.Lt(before)).
		Not(gen.Exists(tagSubQuery)).
		Limit(int(limit)).
		Order(s.tx.Repository.ID).
		Find()
}

// FindRecentlyUpdated finds recently updated repositories for cache prewarming.
func (s *repositoryRepository) FindRecentlyUpdated(ctx context.Context, limit int) ([]*models.Repository, error) {
	return s.tx.Repository.WithContext(ctx).
		Order(s.tx.Repository.UpdatedAt.Desc()).
		Limit(limit).
		Find()
}

// Get gets the repository with the specified repository ID
func (s *repositoryRepository) Get(ctx context.Context, repositoryID string) (*models.Repository, error) {
	return s.tx.Repository.WithContext(ctx).
		Where(s.tx.Repository.ID.Eq(repositoryID)).
		Preload(s.tx.Repository.Builder.CodeRepository).
		Preload(s.tx.Repository.Builder.CodeRepository.User3rdParty).
		Preload(s.tx.Repository.Builder).First()
}

// GetByName gets the repository with the specified repository name
func (s *repositoryRepository) GetByName(ctx context.Context, name string) (*models.Repository, error) {
	return s.tx.Repository.WithContext(ctx).Where(s.tx.Repository.Name.Eq(name)).First()
}

// ListByDtPagination lists repositories with cursor pagination
func (s *repositoryRepository) ListByDtPagination(ctx context.Context, limit int, lastID ...string) ([]*models.Repository, error) {
	do := s.tx.Repository.WithContext(ctx)
	if len(lastID) > 0 {
		do = do.Where(s.tx.Repository.ID.Gt(lastID[0]))
	}
	return do.Order(s.tx.Repository.ID).Limit(limit).Find()
}

// ListWithScrollable lists repositories visible to a user with cursor pagination
func (s *repositoryRepository) ListWithScrollable(ctx context.Context, namespaceID, userID string, name *string, limit int, lastID string) ([]*models.Repository, error) {
	q := s.tx.Repository.WithContext(ctx)
	if namespaceID != "" {
		q = q.Where(s.tx.Repository.NamespaceID.Eq(namespaceID))
	}
	userObj, err := s.tx.User.WithContext(ctx).Where(s.tx.User.ID.Eq(userID)).First()
	if err != nil {
		return nil, err
	}
	if !(userObj.Role == enums.UserRoleAdmin || userObj.Role == enums.UserRoleRoot) { // nolint: staticcheck
		q = q.LeftJoin(s.tx.NamespaceMember, s.tx.Repository.NamespaceID.EqCol(s.tx.NamespaceMember.NamespaceID), s.tx.NamespaceMember.UserID.Eq(userID)).
			Where(s.tx.NamespaceMember.ID.IsNotNull())
	}
	if name != nil {
		q = q.Where(s.tx.Repository.Name.Like(fmt.Sprintf("%s%%", ptr.To(name))))
	}
	if lastID != "" {
		q = q.Where(s.tx.Repository.ID.Gt(lastID))
	}
	return q.Order(s.tx.Repository.ID).Limit(limit).Find()
}

// ListRepositoryWithAuth lists repositories visible to a user with pagination and sorting
func (s *repositoryRepository) ListRepositoryWithAuth(ctx context.Context, namespaceID, userID string, name *string, pagination api.Pagination, sort api.Sortable) ([]*models.Repository, int64, error) {
	pagination = utils.NormalizePagination(pagination)
	q := s.tx.Repository.WithContext(ctx)
	if namespaceID != "" {
		q = q.Where(s.tx.Repository.NamespaceID.Eq(namespaceID))
	}
	if name != nil {
		q = q.Where(s.tx.Repository.Name.Like(fmt.Sprintf("%s%%", ptr.To(name))))
	}
	userObj, err := s.tx.User.WithContext(ctx).Where(s.tx.User.ID.Eq(userID)).First()
	if err != nil {
		return nil, 0, err
	}
	if !(userObj.Role == enums.UserRoleAdmin || userObj.Role == enums.UserRoleRoot) { // nolint: staticcheck
		q = q.LeftJoin(s.tx.NamespaceMember, s.tx.Repository.NamespaceID.EqCol(s.tx.NamespaceMember.NamespaceID), s.tx.NamespaceMember.UserID.Eq(userID)).
			Where(s.tx.NamespaceMember.ID.IsNotNull())
	}
	field, ok := s.tx.Repository.GetFieldByName(ptr.To(sort.Sort))
	if ok {
		switch ptr.To(sort.Method) {
		case enums.SortMethodDesc:
			q = q.Order(field.Desc())
		case enums.SortMethodAsc:
			q = q.Order(field)
		default:
			q = q.Order(s.tx.Repository.UpdatedAt.Desc())
		}
	} else {
		q = q.Order(s.tx.Repository.UpdatedAt.Desc())
	}
	q = q.Preload(s.tx.Repository.Builder.CodeRepository).
		Preload(s.tx.Repository.Builder.CodeRepository.User3rdParty)
	return q.FindByPage(ptr.To(pagination.Limit)*(ptr.To(pagination.Page)-1), ptr.To(pagination.Limit))
}

// ListRepository lists repositories in a namespace with pagination and sorting
func (s *repositoryRepository) ListRepository(ctx context.Context, namespaceID string, name *string, pagination api.Pagination, sort api.Sortable) ([]*models.Repository, int64, error) {
	pagination = utils.NormalizePagination(pagination)
	q := s.tx.Repository.WithContext(ctx).Where(s.tx.Repository.NamespaceID.Eq(namespaceID))
	if name != nil {
		q = q.Where(s.tx.Repository.Name.Like(fmt.Sprintf("%s%%", ptr.To(name))))
	}
	field, ok := s.tx.Repository.GetFieldByName(ptr.To(sort.Sort))
	if ok {
		switch ptr.To(sort.Method) {
		case enums.SortMethodDesc:
			q = q.Order(field.Desc())
		case enums.SortMethodAsc:
			q = q.Order(field)
		default:
			q = q.Order(s.tx.Repository.UpdatedAt.Desc())
		}
	} else {
		q = q.Order(s.tx.Repository.UpdatedAt.Desc())
	}
	return q.FindByPage(ptr.To(pagination.Limit)*(ptr.To(pagination.Page)-1), ptr.To(pagination.Limit))
}

// UpdateRepository updates the repository with the specified ID
func (s *repositoryRepository) UpdateRepository(ctx context.Context, id string, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}
	_, err := s.tx.Repository.WithContext(ctx).Where(s.tx.Repository.ID.Eq(id)).UpdateColumns(updates)
	if err != nil {
		return err
	}
	return nil
}

// CountRepository counts repositories in a namespace
func (s *repositoryRepository) CountRepository(ctx context.Context, namespaceID string, name *string) (int64, error) {
	q := s.tx.Repository.WithContext(ctx).Where(s.tx.Repository.NamespaceID.Eq(namespaceID))
	if name != nil {
		q = q.Where(s.tx.Repository.Name.Like(fmt.Sprintf("%s%%", ptr.To(name))))
	}
	return q.Count()
}

// DeleteByID deletes the repository with the specified repository ID
func (s *repositoryRepository) DeleteByID(ctx context.Context, id string) error {
	_, err := s.tx.Repository.WithContext(ctx).Where(s.tx.Repository.ID.Eq(id)).Delete()
	if err != nil {
		return err
	}
	return nil
}

// CountByNamespace counts repositories grouped by namespace ID
func (s *repositoryRepository) CountByNamespace(ctx context.Context, namespaceIDs []string) (map[string]int64, error) {
	tagCount := make(map[string]int64)
	var count []struct {
		NamespaceID string `gorm:"column:namespace_id"`
		Count       int64  `gorm:"column:count"`
	}
	err := s.tx.Repository.WithContext(ctx).
		Where(s.tx.Repository.NamespaceID.In(namespaceIDs...)).
		Group(s.tx.Repository.NamespaceID).
		Select(s.tx.Repository.NamespaceID, s.tx.Repository.ID.Count().As("count")).
		Scan(&count)
	if err != nil {
		return nil, err
	}
	for _, c := range count {
		tagCount[c.NamespaceID] = c.Count
	}
	return tagCount, nil
}

// DeleteEmpty deletes empty repositories and returns their names
func (s *repositoryRepository) DeleteEmpty(ctx context.Context, namespaceID *string) ([]string, error) {
	q := s.tx.Repository.WithContext(ctx).
		LeftJoin(s.tx.Artifact, s.tx.Repository.ID.EqCol(s.tx.Artifact.RepositoryID)).
		LeftJoin(s.tx.Tag, s.tx.Repository.ID.EqCol(s.tx.Tag.RepositoryID)).
		Where(s.tx.Artifact.RepositoryID.IsNull(), s.tx.Tag.RepositoryID.IsNull())
	if namespaceID != nil {
		q = q.Where(s.tx.Repository.NamespaceID.Eq(ptr.To(namespaceID)))
	}
	repositoryObjs, err := q.Find()
	if err != nil {
		return nil, err
	}
	IDs := make([]string, 0, len(repositoryObjs))
	result := make([]string, 0, len(repositoryObjs))
	for _, r := range repositoryObjs {
		result = append(result, r.Name)
		IDs = append(IDs, r.ID)
	}
	_, err = s.tx.Repository.WithContext(ctx).Where(s.tx.Repository.ID.In(IDs...)).Delete()
	if err != nil {
		return nil, err
	}
	return result, nil
}

// IncrementSize atomically increments repository size by delta.
// If SizeLimit > 0 and size+delta exceeds the limit, returns a quota-exceed error.
// Also sets size_dirty = true to mark for reconciliation.
func (s *repositoryRepository) IncrementSize(ctx context.Context, repositoryID string, delta int64) error {
	result, err := s.tx.Repository.WithContext(ctx).
		Where(s.tx.Repository.ID.Eq(repositoryID)).
		Where(field.Or(s.tx.Repository.SizeLimit.Eq(0), s.tx.Repository.Size.Add(delta).LteCol(s.tx.Repository.SizeLimit))).
		UpdateColumns(map[string]any{
			"size":       gorm.Expr("size + ?", delta),
			"size_dirty": true,
		})
	if err != nil {
		return err
	}
	if result.RowsAffected == 0 {
		repo, err := s.tx.Repository.WithContext(ctx).Where(s.tx.Repository.ID.Eq(repositoryID)).First()
		if err != nil {
			return err
		}
		if repo.SizeLimit > 0 {
			return errcode.GenDSErrCodeResourceSizeQuotaExceedRepository(repo.Name, repo.Size, repo.SizeLimit, delta)
		}
		return gorm.ErrRecordNotFound
	}
	return nil
}

// DecrementSize atomically decrements repository size by delta (floored at 0).
// Also sets size_dirty = true to mark for reconciliation.
func (s *repositoryRepository) DecrementSize(ctx context.Context, repositoryID string, delta int64) error {
	if delta <= 0 {
		return nil
	}
	result, err := s.tx.Repository.WithContext(ctx).
		Where(s.tx.Repository.ID.Eq(repositoryID)).
		Where(s.tx.Repository.Size.Gte(delta)).
		UpdateColumns(map[string]any{
			"size":       gorm.Expr("size - ?", delta),
			"size_dirty": true,
		})
	if err != nil {
		return err
	}
	if result.RowsAffected == 0 {
		// Either repository not found, or size < delta (already near zero).
		// Floor the size at 0 to avoid negative drift.
		repo, err := s.tx.Repository.WithContext(ctx).Where(s.tx.Repository.ID.Eq(repositoryID)).First()
		if err != nil {
			return err
		}
		if repo.Size > 0 {
			_, err = s.tx.Repository.WithContext(ctx).
				Where(s.tx.Repository.ID.Eq(repositoryID)).
				UpdateColumns(map[string]any{
					"size":       0,
					"size_dirty": true,
				})
			if err != nil {
				slog.Error("floor repository size to zero failed", "err", err, "repository_id", repositoryID)
			}
		}
	}
	return nil
}

// FindWithCursorDirty finds repositories with size_dirty = true using cursor pagination.
func (s *repositoryRepository) FindWithCursorDirty(ctx context.Context, limit int64, last string) ([]*models.Repository, error) {
	q := s.tx.Repository.WithContext(ctx).Where(s.tx.Repository.SizeDirty.Eq(true))
	if last != "" {
		q = q.Where(s.tx.Repository.ID.Gt(last))
	}
	return q.Order(s.tx.Repository.ID.Asc()).Limit(int(limit)).Find()
}

// UpdateSizeAndClearDirty updates size and clears the dirty flag.
func (s *repositoryRepository) UpdateSizeAndClearDirty(ctx context.Context, repositoryID string, size int64) error {
	_, err := s.tx.Repository.WithContext(ctx).
		Where(s.tx.Repository.ID.Eq(repositoryID)).
		UpdateColumns(map[string]any{
			"size":       size,
			"size_dirty": false,
		})
	return err
}

// ClearSizeDirty clears only the size_dirty flag.
func (s *repositoryRepository) ClearSizeDirty(ctx context.Context, repositoryID string) error {
	_, err := s.tx.Repository.WithContext(ctx).
		Where(s.tx.Repository.ID.Eq(repositoryID)).
		UpdateColumn(s.tx.Repository.SizeDirty, false)
	return err
}
