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
