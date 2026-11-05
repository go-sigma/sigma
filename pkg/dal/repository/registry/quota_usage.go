// Copyright 2026 sigma
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

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/server/errcode"
)

const (
	// quotaUsageTable is the table name used by raw GORM queries.
	quotaUsageTable = "quota_usages"
	// ScopeTypeGlobal is the global physical usage scope.
	ScopeTypeGlobal = "global"
)

//go:generate go tool mockgen -destination=mocks/quota_usage_mocks.go -package=mocks github.com/go-sigma/sigma/pkg/dal/repository/registry QuotaUsageRepository

// QuotaUsageRepository maintains the aggregated usage counters. Only the global
// physical usage row is used today; namespace/repository sizes stay on their
// own tables.
type QuotaUsageRepository interface {
	// ReserveGlobal atomically reserves delta bytes for the global physical
	// capacity. capacity == 0 means unlimited. Returns a 507 errcode when the
	// reservation would exceed the capacity.
	ReserveGlobal(ctx context.Context, delta, capacity int64) error
	// ReleaseGlobal releases delta bytes from the global physical usage.
	ReleaseGlobal(ctx context.Context, delta int64) error
	// GetGlobal returns the current global physical usage.
	GetGlobal(ctx context.Context) (int64, error)
	// SyncGlobal recomputes the global usage from live blobs and stores it.
	SyncGlobal(ctx context.Context) (int64, error)
}

type quotaUsageRepository struct {
	tx *query.Query
}

// NewQuotaUsageRepository creates a new quota usage repository with the
// optional query transaction.
func NewQuotaUsageRepository(txs ...*query.Query) QuotaUsageRepository {
	tx := query.Q
	if len(txs) > 0 {
		tx = txs[0]
	}
	return &quotaUsageRepository{tx: tx}
}

// ReserveGlobal atomically reserves delta bytes for the global capacity.
func (s *quotaUsageRepository) ReserveGlobal(ctx context.Context, delta, capacity int64) error {
	if delta <= 0 {
		return nil
	}
	if capacity > 0 && delta > capacity {
		current, err := s.GetGlobal(ctx)
		if err != nil {
			return err
		}
		return errcode.GenDSErrCodeStorageCapacityExceed(current, capacity, delta)
	}

	db := s.tx.UnderlyingDB().WithContext(ctx).Table(quotaUsageTable)
	stmt := db.Where("scope_type = ? AND scope_id = ?", ScopeTypeGlobal, "")
	if capacity > 0 {
		stmt = stmt.Where("size + ? <= ?", delta, capacity)
	}
	result := stmt.Updates(map[string]any{
		"size":       gorm.Expr("size + ?", delta),
		"updated_at": time.Now().UnixMilli(),
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		current, err := s.GetGlobal(ctx)
		if err != nil {
			return err
		}
		return errcode.GenDSErrCodeStorageCapacityExceed(current, capacity, delta)
	}
	return nil
}

// ReleaseGlobal releases delta bytes from the global physical usage.
func (s *quotaUsageRepository) ReleaseGlobal(ctx context.Context, delta int64) error {
	if delta <= 0 {
		return nil
	}
	db := s.tx.UnderlyingDB().WithContext(ctx).Table(quotaUsageTable)
	result := db.Where("scope_type = ? AND scope_id = ? AND size >= ?", ScopeTypeGlobal, "", delta).
		Updates(map[string]any{
			"size":       gorm.Expr("size - ?", delta),
			"updated_at": time.Now().UnixMilli(),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		// Floor at zero when the recorded size is already below delta.
		return db.Where("scope_type = ? AND scope_id = ?", ScopeTypeGlobal, "").
			Updates(map[string]any{
				"size":       0,
				"updated_at": time.Now().UnixMilli(),
			}).Error
	}
	return nil
}

// GetGlobal returns the current global physical usage.
func (s *quotaUsageRepository) GetGlobal(ctx context.Context) (int64, error) {
	var usage models.QuotaUsage
	err := s.tx.UnderlyingDB().WithContext(ctx).
		Table(quotaUsageTable).
		Where("scope_type = ? AND scope_id = ?", ScopeTypeGlobal, "").
		First(&usage).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, nil
		}
		return 0, err
	}
	return usage.Size, nil
}

// SyncGlobal recomputes the global usage from live blobs and stores it.
//
// The blob sum is scanned before taking the quota row lock to avoid blocking
// uploads for the duration of the scan. The transaction then re-reads the
// counter under the lock and skips the write when it changed, which means a
// reservation or release happened concurrently and the scan is stale.
func (s *quotaUsageRepository) SyncGlobal(ctx context.Context) (int64, error) {
	current, err := s.GetGlobal(ctx)
	if err != nil {
		return 0, err
	}

	var row struct {
		Total int64
	}
	if err := s.tx.UnderlyingDB().WithContext(ctx).
		Table("blobs").
		Select("COALESCE(SUM(size), 0) AS total").
		Where("deleted_at = 0").
		Scan(&row).Error; err != nil {
		return 0, err
	}

	err = s.tx.Transaction(func(tx *query.Query) error {
		// Ensure the singleton row exists, then lock it before comparing the
		// counter with the value observed before the scan.
		if err := tx.QuotaUsage.WithContext(ctx).
			Clauses(clause.OnConflict{DoNothing: true}).
			Create(&models.QuotaUsage{ScopeType: ScopeTypeGlobal, ScopeID: ""}); err != nil {
			return err
		}
		now := time.Now().UnixMilli()
		if _, err := tx.QuotaUsage.WithContext(ctx).
			Where(tx.QuotaUsage.ScopeType.Eq(ScopeTypeGlobal), tx.QuotaUsage.ScopeID.Eq("")).
			UpdateColumn(tx.QuotaUsage.UpdatedAt, now); err != nil {
			return err
		}
		latest, err := tx.QuotaUsage.WithContext(ctx).
			Where(tx.QuotaUsage.ScopeType.Eq(ScopeTypeGlobal), tx.QuotaUsage.ScopeID.Eq("")).
			First()
		if err != nil {
			return err
		}
		if latest.Size != current {
			// A concurrent reservation/release changed the counter while the
			// blob sum was scanned; leave the write to the next round.
			return nil
		}
		_, err = tx.QuotaUsage.WithContext(ctx).
			Where(tx.QuotaUsage.ScopeType.Eq(ScopeTypeGlobal), tx.QuotaUsage.ScopeID.Eq("")).
			UpdateColumns(map[string]any{
				"size":       row.Total,
				"updated_at": now,
			})
		return err
	})
	if err != nil {
		return 0, err
	}
	return row.Total, nil
}
