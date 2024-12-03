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

	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/types/enums"
)

//go:generate mockgen -destination=mocks/audit.go -package=mocks github.com/go-sigma/sigma/pkg/dal/dao AuditService
//go:generate mockgen -destination=mocks/audit_factory.go -package=mocks github.com/go-sigma/sigma/pkg/dal/dao AuditServiceFactory

// AuditService is the interface that provides methods to operate on audit model
type AuditService interface {
	// Create creates a new Audit record in the database
	Create(ctx context.Context, audit *models.Audit) error
	// HotNamespace get top n hot namespace by user id
	HotNamespace(ctx context.Context, userID int64, top int) ([]*models.Namespace, error)
}

type auditService struct {
	tx *query.Query
}

// AuditServiceFactory is the interface that provides the audit service factory methods
type AuditServiceFactory interface {
	New(txs ...*query.Query) AuditService
}

type auditServiceFactory struct{}

// NewAuditServiceFactory creates a new audit service factory
func NewAuditServiceFactory() AuditServiceFactory {
	return &auditServiceFactory{}
}

// New creates a new audit service
func (s *auditServiceFactory) New(txs ...*query.Query) AuditService {
	tx := query.Q
	if len(txs) > 0 {
		tx = txs[0]
	}
	return &auditService{
		tx: tx,
	}
}

// Create create a new artifact if conflict do nothing
func (s *auditService) Create(ctx context.Context, audit *models.Audit) error {
	return s.tx.Audit.WithContext(ctx).Create(audit)
}

// HotNamespace get top n hot namespace by user id
func (s *auditService) HotNamespace(ctx context.Context, userID int64, top int) ([]*models.Namespace, error) {
	type result struct {
		NamespaceID int64
		CreatedAt   string
		Count       int64
	}
	var rs []result
	err := s.tx.Audit.WithContext(ctx).
		Where(s.tx.Audit.Action.Neq(enums.AuditActionDelete), s.tx.Audit.UserID.Eq(userID)).
		Group(s.tx.Audit.NamespaceID).
		Select(s.tx.Audit.NamespaceID, s.tx.Audit.CreatedAt.Max().As(s.tx.Audit.CreatedAt.ColumnName().String()), s.tx.Audit.ID.Count().As("count")).
		Limit(top).
		UnderlyingDB().
		Order("count desc, created_at desc").Find(&rs).Error
	if err != nil {
		return nil, err
	}
	if len(rs) == 0 {
		return nil, nil
	}
	var namespaceIDs = make([]int64, 0, len(rs))
	for _, audit := range rs {
		namespaceIDs = append(namespaceIDs, audit.NamespaceID)
	}
	return s.tx.Namespace.WithContext(ctx).Where(s.tx.Namespace.ID.In(namespaceIDs...)).Find()
}
