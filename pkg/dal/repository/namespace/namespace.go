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

package namespace

import (
	"context"
	"fmt"
	"log/slog"

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

//go:generate mockgen -destination=namespace_mocks.go -package=namespace github.com/go-sigma/sigma/pkg/dal/repository/namespace NamespaceRepository

// NamespaceRepository defines namespace repository operations
type NamespaceRepository interface {
	// Create creates a new namespace
	Create(ctx context.Context, namespace *models.Namespace) error
	// FindAll finds all namespaces
	FindAll(ctx context.Context) ([]*models.Namespace, error)
	// FindWithCursor finds namespaces with cursor pagination
	FindWithCursor(ctx context.Context, limit int64, last string) ([]*models.Namespace, error)
	// FindRecentlyUpdated finds recently updated namespaces for cache prewarming.
	FindRecentlyUpdated(ctx context.Context, limit int) ([]*models.Namespace, error)
	// UpdateQuota updates the namespace quota
	UpdateQuota(ctx context.Context, namespaceID string, limit int64) error
	// Get gets the namespace with the specified namespace ID
	Get(ctx context.Context, id string) (*models.Namespace, error)
	// GetByName gets the namespace with the specified namespace name
	GetByName(ctx context.Context, name string) (*models.Namespace, error)
	// ListNamespace lists namespaces with filtering, pagination, and sorting
	ListNamespace(ctx context.Context, name *string, pagination api.Pagination, sort api.Sortable) ([]*models.Namespace, int64, error)
	// ListNamespaceWithAuth lists namespaces visible to a user with pagination and sorting
	ListNamespaceWithAuth(ctx context.Context, userID string, name *string, pagination api.Pagination, sort api.Sortable) ([]*models.Namespace, int64, error)
	// CountNamespace counts namespaces by optional name filter
	CountNamespace(ctx context.Context, name *string) (int64, error)
	// DeleteByID deletes the namespace with the specified namespace ID
	DeleteByID(ctx context.Context, id string) error
	// UpdateByID updates the namespace with the specified namespace ID
	UpdateByID(ctx context.Context, id string, updates map[string]any) error
	// IncrementSize atomically increments namespace size by delta.
	// Returns quota-exceed error if SizeLimit > 0 and size+delta > SizeLimit.
	IncrementSize(ctx context.Context, namespaceID string, delta int64) error
	// DecrementSize atomically decrements namespace size by delta (floored at 0).
	DecrementSize(ctx context.Context, namespaceID string, delta int64) error
	// FindWithCursorDirty finds namespaces with size_dirty = true using cursor pagination
	FindWithCursorDirty(ctx context.Context, limit int64, last string) ([]*models.Namespace, error)
	// UpdateSizeAndClearDirty updates size and clears the dirty flag
	UpdateSizeAndClearDirty(ctx context.Context, namespaceID string, size int64) error
	// ClearSizeDirty clears only the size_dirty flag
	ClearSizeDirty(ctx context.Context, namespaceID string) error
}

type namespaceRepository struct {
	tx *query.Query
}

// NewNamespaceRepository creates a new namespace repository with the optional query transaction
func NewNamespaceRepository(txs ...*query.Query) NamespaceRepository {
	tx := query.Q
	if len(txs) > 0 {
		tx = txs[0]
	}
	return &namespaceRepository{
		tx: tx,
	}
}

// Create creates a new namespace
func (s *namespaceRepository) Create(ctx context.Context, namespaceObj *models.Namespace) error {
	return s.tx.Namespace.WithContext(ctx).Create(namespaceObj)
}

// FindAll finds all namespaces
func (s *namespaceRepository) FindAll(ctx context.Context) ([]*models.Namespace, error) {
	return s.tx.Namespace.WithContext(ctx).Find()
}

// FindWithCursor finds namespaces with cursor pagination
func (s *namespaceRepository) FindWithCursor(ctx context.Context, limit int64, last string) ([]*models.Namespace, error) {
	return s.tx.Namespace.WithContext(ctx).Where(s.tx.Namespace.ID.Gt(last)).Limit(int(limit)).Order(s.tx.Namespace.ID).Find()
}

// FindRecentlyUpdated finds recently updated namespaces for cache prewarming.
func (s *namespaceRepository) FindRecentlyUpdated(ctx context.Context, limit int) ([]*models.Namespace, error) {
	return s.tx.Namespace.WithContext(ctx).Order(s.tx.Namespace.UpdatedAt.Desc()).Limit(limit).Find()
}

// UpdateQuota updates the namespace quota
func (s *namespaceRepository) UpdateQuota(ctx context.Context, namespaceID string, limit int64) error {
	result, err := s.tx.Namespace.WithContext(ctx).Where(s.tx.Namespace.ID.Eq(namespaceID)).Update(s.tx.Namespace.SizeLimit, limit)
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return err
}

// Get gets the namespace with the specified namespace ID
func (s *namespaceRepository) Get(ctx context.Context, id string) (*models.Namespace, error) {
	return s.tx.Namespace.WithContext(ctx).Where(s.tx.Namespace.ID.Eq(id)).First()
}

// GetByName gets the namespace with the specified namespace name
func (s *namespaceRepository) GetByName(ctx context.Context, name string) (*models.Namespace, error) {
	return s.tx.Namespace.WithContext(ctx).Where(s.tx.Namespace.Name.Eq(name)).First()
}

// ListNamespace lists namespaces with filtering, pagination, and sorting
func (s *namespaceRepository) ListNamespace(ctx context.Context, name *string, pagination api.Pagination, sort api.Sortable) ([]*models.Namespace, int64, error) {
	pagination = utils.NormalizePagination(pagination)
	q := s.tx.Namespace.WithContext(ctx)
	if name != nil {
		q = q.Where(s.tx.Namespace.Name.Like(fmt.Sprintf("%%%s%%", ptr.To(name))))
	}
	field, ok := s.tx.Namespace.GetFieldByName(ptr.To(sort.Sort))
	if ok {
		switch ptr.To(sort.Method) {
		case enums.SortMethodDesc:
			q = q.Order(field.Desc())
		case enums.SortMethodAsc:
			q = q.Order(field)
		default:
			q = q.Order(s.tx.Namespace.UpdatedAt.Desc())
		}
	} else {
		q = q.Order(s.tx.Namespace.UpdatedAt.Desc())
	}
	return q.FindByPage(ptr.To(pagination.Limit)*(ptr.To(pagination.Page)-1), ptr.To(pagination.Limit))
}

// ListNamespaceWithAuth lists namespaces visible to a user with pagination and sorting
// userID 0 means anonymous
func (s *namespaceRepository) ListNamespaceWithAuth(ctx context.Context, userID string, name *string, pagination api.Pagination, sort api.Sortable) ([]*models.Namespace, int64, error) {
	pagination = utils.NormalizePagination(pagination)
	q := s.tx.Namespace.WithContext(ctx)
	if name != nil {
		q = q.Where(s.tx.Namespace.Name.Like(fmt.Sprintf("%s%%", ptr.To(name))))
	}
	if userID == "" { // find the public namespace
		q = q.Where(s.tx.Namespace.Visibility.Eq(enums.VisibilityPublic))
	} else { // find user id authenticated namespace
		userObj, err := s.tx.User.WithContext(ctx).Where(s.tx.User.ID.Eq(userID)).First()
		if err != nil {
			return nil, 0, err
		}
		if !(userObj.Role == enums.UserRoleAdmin || userObj.Role == enums.UserRoleRoot) { // nolint: staticcheck
			q = q.LeftJoin(s.tx.NamespaceMember, s.tx.Namespace.ID.EqCol(s.tx.NamespaceMember.NamespaceID), s.tx.NamespaceMember.UserID.Eq(userID)).
				Where(field.Or(s.tx.NamespaceMember.ID.IsNotNull(), s.tx.Namespace.Visibility.Eq(enums.VisibilityPublic)))
		}
	}
	field, ok := s.tx.Namespace.GetFieldByName(ptr.To(sort.Sort))
	if ok {
		switch ptr.To(sort.Method) {
		case enums.SortMethodDesc:
			q = q.Order(field.Desc())
		case enums.SortMethodAsc:
			q = q.Order(field)
		default:
			q = q.Order(s.tx.Namespace.UpdatedAt.Desc())
		}
	} else {
		q = q.Order(s.tx.Namespace.UpdatedAt.Desc())
	}
	return q.FindByPage(ptr.To(pagination.Limit)*(ptr.To(pagination.Page)-1), ptr.To(pagination.Limit))
}

// CountNamespace counts namespaces by optional name filter
func (s *namespaceRepository) CountNamespace(ctx context.Context, name *string) (int64, error) {
	q := s.tx.Namespace.WithContext(ctx)
	if name != nil {
		q = q.Where(s.tx.Namespace.Name.Like(fmt.Sprintf("%%%s%%", ptr.To(name))))
	}
	return q.Count()
}

// DeleteByID deletes the namespace with the specified namespace ID
func (s *namespaceRepository) DeleteByID(ctx context.Context, id string) error {
	matched, err := s.tx.Namespace.WithContext(ctx).Where(s.tx.Namespace.ID.Eq(id)).Delete()
	if err != nil {
		return err
	}
	if matched.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// UpdateByID updates the namespace with the specified namespace ID
func (s *namespaceRepository) UpdateByID(ctx context.Context, id string, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}
	_, err := s.tx.Namespace.WithContext(ctx).Where(s.tx.Namespace.ID.Eq(id)).Updates(updates)
	return err
}

// IncrementSize atomically increments namespace size by delta.
// If SizeLimit > 0 and size+delta exceeds the limit, returns a quota-exceed error.
// Also sets size_dirty = true to mark for reconciliation.
func (s *namespaceRepository) IncrementSize(ctx context.Context, namespaceID string, delta int64) error {
	result, err := s.tx.Namespace.WithContext(ctx).
		Where(s.tx.Namespace.ID.Eq(namespaceID)).
		Where(field.Or(s.tx.Namespace.SizeLimit.Eq(0), s.tx.Namespace.Size.Add(delta).LteCol(s.tx.Namespace.SizeLimit))).
		UpdateColumns(map[string]any{
			"size":       gorm.Expr("size + ?", delta),
			"size_dirty": true,
		})
	if err != nil {
		return err
	}
	if result.RowsAffected == 0 {
		ns, err := s.tx.Namespace.WithContext(ctx).Where(s.tx.Namespace.ID.Eq(namespaceID)).First()
		if err != nil {
			return err
		}
		if ns.SizeLimit > 0 {
			return errcode.GenDSErrCodeResourceSizeQuotaExceedNamespace(ns.Name, ns.Size, ns.SizeLimit, delta)
		}
		return gorm.ErrRecordNotFound
	}
	return nil
}

// DecrementSize atomically decrements namespace size by delta (floored at 0).
// Also sets size_dirty = true to mark for reconciliation.
func (s *namespaceRepository) DecrementSize(ctx context.Context, namespaceID string, delta int64) error {
	if delta <= 0 {
		return nil
	}
	result, err := s.tx.Namespace.WithContext(ctx).
		Where(s.tx.Namespace.ID.Eq(namespaceID)).
		Where(s.tx.Namespace.Size.Gte(delta)).
		UpdateColumns(map[string]any{
			"size":       gorm.Expr("size - ?", delta),
			"size_dirty": true,
		})
	if err != nil {
		return err
	}
	if result.RowsAffected == 0 {
		// Either namespace not found, or size < delta (already near zero).
		// Floor the size at 0 to avoid negative drift.
		ns, err := s.tx.Namespace.WithContext(ctx).Where(s.tx.Namespace.ID.Eq(namespaceID)).First()
		if err != nil {
			return err
		}
		if ns.Size > 0 {
			_, err = s.tx.Namespace.WithContext(ctx).
				Where(s.tx.Namespace.ID.Eq(namespaceID)).
				UpdateColumns(map[string]any{
					"size":       0,
					"size_dirty": true,
				})
			if err != nil {
				slog.Error("floor namespace size to zero failed", "err", err, "namespace_id", namespaceID)
			}
		}
	}
	return nil
}

// FindWithCursorDirty finds namespaces with size_dirty = true using cursor pagination.
func (s *namespaceRepository) FindWithCursorDirty(ctx context.Context, limit int64, last string) ([]*models.Namespace, error) {
	q := s.tx.Namespace.WithContext(ctx).Where(s.tx.Namespace.SizeDirty.Eq(true))
	if last != "" {
		q = q.Where(s.tx.Namespace.ID.Gt(last))
	}
	return q.Order(s.tx.Namespace.ID.Asc()).Limit(int(limit)).Find()
}

// UpdateSizeAndClearDirty updates size and clears the dirty flag.
func (s *namespaceRepository) UpdateSizeAndClearDirty(ctx context.Context, namespaceID string, size int64) error {
	_, err := s.tx.Namespace.WithContext(ctx).
		Where(s.tx.Namespace.ID.Eq(namespaceID)).
		UpdateColumns(map[string]any{
			"size":       size,
			"size_dirty": false,
		})
	return err
}

// ClearSizeDirty clears only the size_dirty flag.
func (s *namespaceRepository) ClearSizeDirty(ctx context.Context, namespaceID string) error {
	_, err := s.tx.Namespace.WithContext(ctx).
		Where(s.tx.Namespace.ID.Eq(namespaceID)).
		UpdateColumn(s.tx.Namespace.SizeDirty, false)
	return err
}
