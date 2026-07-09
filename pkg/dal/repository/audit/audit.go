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

package audit

import (
	"context"

	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
)

//go:generate mockgen -destination=audit_mocks.go -package=audit github.com/go-sigma/sigma/pkg/dal/repository/audit AuditRepository

// AuditRepository defines audit repository operations
type AuditRepository interface {
	// Create creates a new audit record
	Create(ctx context.Context, audit *models.Audit) error
	// SoftDeleteBefore soft deletes audit records created before the cutoff
	SoftDeleteBefore(ctx context.Context, createdBefore int64, limit int) (int64, error)
	// HardDeleteBefore hard deletes audit records created before the cutoff
	HardDeleteBefore(ctx context.Context, createdBefore int64, limit int) (int64, error)
}

type auditRepository struct {
	tx *query.Query
}

// NewAuditRepository creates a new audit repository with the optional query transaction
func NewAuditRepository(txs ...*query.Query) AuditRepository {
	tx := query.Q
	if len(txs) > 0 {
		tx = txs[0]
	}
	return &auditRepository{
		tx: tx,
	}
}

// Create creates a new audit record
func (s *auditRepository) Create(ctx context.Context, audit *models.Audit) error {
	return s.tx.Audit.WithContext(ctx).Create(audit)
}

// SoftDeleteBefore soft deletes audit records created before the cutoff
func (s *auditRepository) SoftDeleteBefore(ctx context.Context, createdBefore int64, limit int) (int64, error) {
	if limit <= 0 {
		return 0, nil
	}
	ids, err := s.findIDsForSoftDelete(ctx, createdBefore, limit)
	if err != nil {
		return 0, err
	}
	if len(ids) == 0 {
		return 0, nil
	}
	matched, err := s.tx.Audit.WithContext(ctx).Where(s.tx.Audit.ID.In(ids...)).Delete()
	if err != nil {
		return 0, err
	}
	return matched.RowsAffected, nil
}

// HardDeleteBefore hard deletes audit records created before the cutoff
func (s *auditRepository) HardDeleteBefore(ctx context.Context, createdBefore int64, limit int) (int64, error) {
	if limit <= 0 {
		return 0, nil
	}
	ids, err := s.findIDsForHardDelete(ctx, createdBefore, limit)
	if err != nil {
		return 0, err
	}
	if len(ids) == 0 {
		return 0, nil
	}
	matched, err := s.tx.Audit.WithContext(ctx).Unscoped().Where(s.tx.Audit.ID.In(ids...)).Delete()
	if err != nil {
		return 0, err
	}
	return matched.RowsAffected, nil
}

func (s *auditRepository) findIDsForSoftDelete(ctx context.Context, createdBefore int64, limit int) ([]string, error) {
	auditObjs, err := s.tx.Audit.WithContext(ctx).
		Select(s.tx.Audit.ID).
		Where(s.tx.Audit.CreatedAt.Lt(createdBefore), s.tx.Audit.DeletedAt.Eq(0)).
		Order(s.tx.Audit.CreatedAt).
		Limit(limit).
		Find()
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(auditObjs))
	for _, auditObj := range auditObjs {
		ids = append(ids, auditObj.ID)
	}
	return ids, nil
}

func (s *auditRepository) findIDsForHardDelete(ctx context.Context, createdBefore int64, limit int) ([]string, error) {
	auditObjs, err := s.tx.Audit.WithContext(ctx).
		Unscoped().
		Select(s.tx.Audit.ID).
		Where(s.tx.Audit.CreatedAt.Lt(createdBefore)).
		Order(s.tx.Audit.CreatedAt).
		Limit(limit).
		Find()
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(auditObjs))
	for _, auditObj := range auditObjs {
		ids = append(ids, auditObj.ID)
	}
	return ids, nil
}
