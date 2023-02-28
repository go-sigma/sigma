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
	"fmt"

	"gorm.io/gorm"

	"github.com/ximager/ximager/pkg/dal/models"
	"github.com/ximager/ximager/pkg/dal/query"
	"github.com/ximager/ximager/pkg/types"
	"github.com/ximager/ximager/pkg/utils/ptr"
)

// NamespaceService is the interface that provides the namespace service methods.
type NamespaceService interface {
	// Create creates a new namespace.
	Create(ctx context.Context, namespace *models.Namespace) (*models.Namespace, error)
	// Get gets the namespace with the specified namespace ID.
	Get(ctx context.Context, id uint64) (*models.Namespace, error)
	// GetByName gets the namespace with the specified namespace name.
	GetByName(ctx context.Context, name string) (*models.Namespace, error)
	// ListNamespace lists all namespaces.
	ListNamespace(ctx context.Context, req types.ListNamespaceRequest) ([]*models.Namespace, error)
	// CountNamespace counts all namespaces.
	CountNamespace(ctx context.Context, req types.ListNamespaceRequest) (int64, error)
	// DeleteByID deletes the namespace with the specified namespace ID.
	DeleteByID(ctx context.Context, id uint64) error
	// UpdateByID updates the namespace with the specified namespace ID.
	UpdateByID(ctx context.Context, id uint64, req types.PutNamespaceRequest) error
}

type namespaceService struct {
	tx *query.Query
}

// NewNamespaceService creates a new namespace service.
func NewNamespaceService(txs ...*query.Query) NamespaceService {
	tx := query.Q
	if len(txs) > 0 {
		tx = txs[0]
	}
	return &namespaceService{
		tx: tx,
	}
}

// Create creates a new namespace.
func (s *namespaceService) Create(ctx context.Context, namespace *models.Namespace) (*models.Namespace, error) {
	err := s.tx.Namespace.WithContext(ctx).Create(namespace)
	if err != nil {
		return nil, err
	}
	return namespace, nil
}

// Get gets the namespace with the specified namespace ID.
func (s *namespaceService) Get(ctx context.Context, id uint64) (*models.Namespace, error) {
	ns, err := s.tx.Namespace.WithContext(ctx).Where(s.tx.Namespace.ID.Eq(id)).First()
	if err != nil {
		return nil, err
	}
	return ns, nil
}

// GetByName gets the namespace with the specified namespace name.
func (s *namespaceService) GetByName(ctx context.Context, name string) (*models.Namespace, error) {
	ns, err := s.tx.Namespace.WithContext(ctx).Where(s.tx.Namespace.Name.Eq(name)).First()
	if err != nil {
		return nil, err
	}
	return ns, nil
}

// ListNamespace lists all namespaces.
func (s *namespaceService) ListNamespace(ctx context.Context, req types.ListNamespaceRequest) ([]*models.Namespace, error) {
	query := s.tx.Namespace.WithContext(ctx).Offset(req.PageSize * (req.PageNum - 1)).Limit(req.PageSize)
	if req.Name != nil {
		query = query.Where(s.tx.Namespace.Name.Like(fmt.Sprintf("%%%s%%", ptr.To(req.Name))))
	}
	return query.Find()
}

// CountNamespace counts all namespaces.
func (s *namespaceService) CountNamespace(ctx context.Context, req types.ListNamespaceRequest) (int64, error) {
	query := s.tx.Namespace.WithContext(ctx)
	if req.Name != nil {
		query = query.Where(s.tx.Namespace.Name.Like(fmt.Sprintf("%%%s%%", ptr.To(req.Name))))
	}
	return query.Count()
}

// DeleteByID deletes the namespace with the specified namespace ID.
func (s *namespaceService) DeleteByID(ctx context.Context, id uint64) error {
	matched, err := s.tx.Namespace.WithContext(ctx).Where(s.tx.Namespace.ID.Eq(id)).Delete()
	if err != nil {
		return err
	}
	if matched.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// UpdateByID updates the namespace with the specified namespace ID.
func (s *namespaceService) UpdateByID(ctx context.Context, id uint64, req types.PutNamespaceRequest) error {
	query := s.tx.Namespace.WithContext(ctx).Where(s.tx.Namespace.ID.Eq(id))

	var update = make(map[string]interface{})
	if req.Description != nil {
		update[string(s.tx.Namespace.Description.ColumnName())] = ptr.To(req.Description)
	}
	matched, err := query.Updates(update)
	if err != nil {
		return err
	}
	if matched.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
