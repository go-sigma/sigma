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

package analytics

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/utils/uuid"
)

//go:generate mockgen -destination=analytics_mocks.go -package=analytics github.com/go-sigma/sigma/pkg/dal/repository/analytics AnalyticsRepository

// AnalyticsRepository defines analytics rollup repository operations
type AnalyticsRepository interface {
	// IncrUserPush increments a user's hourly push count
	IncrUserPush(ctx context.Context, hour int64, userID string, delta int64) error
	// IncrNamespace increments namespace hourly activity counters
	IncrNamespace(ctx context.Context, hour int64, namespaceID string, pushDelta, pullDelta, sizeDelta, tagDelta int64) error
	// ListUserPush lists hourly user push rollups
	ListUserPush(ctx context.Context, userID string, startHour, endHour int64) ([]*models.UserActivityHourly, error)
	// ListNamespaceActivity lists hourly namespace activity rollups
	ListNamespaceActivity(ctx context.Context, namespaceID string, startHour, endHour int64) ([]*models.NamespaceActivityHourly, error)
	// DeleteBeforeHour deletes rollups older than the retention cutoff
	DeleteBeforeHour(ctx context.Context, hour int64) error
}

type analyticsRepository struct {
	tx *query.Query
}

// NewAnalyticsRepository creates a new analytics repository with the optional query transaction
func NewAnalyticsRepository(txs ...*query.Query) AnalyticsRepository {
	tx := query.Q
	if len(txs) > 0 {
		tx = txs[0]
	}
	return &analyticsRepository{tx: tx}
}

// IncrUserPush increments a user's hourly push count
func (s *analyticsRepository) IncrUserPush(ctx context.Context, hour int64, userID string, delta int64) error {
	now := time.Now().UTC().UnixMilli()
	return s.tx.UnderlyingDB().WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "hour"}, {Name: "user_id"}},
			DoUpdates: clause.Assignments(map[string]any{
				"push_count": gorm.Expr("user_activity_hourlies.push_count + ?", delta),
				"updated_at": now,
			}),
		}).
		Create(&models.UserActivityHourly{
			ID:        uuid.NewV7String(),
			Hour:      hour,
			UserID:    userID,
			PushCount: delta,
			CreatedAt: now,
			UpdatedAt: now,
		}).Error
}

// IncrNamespace increments namespace hourly activity counters
func (s *analyticsRepository) IncrNamespace(
	ctx context.Context,
	hour int64, namespaceID string, pushDelta, pullDelta, sizeDelta, tagDelta int64,
) error {

	now := time.Now().UTC().UnixMilli()
	return s.tx.UnderlyingDB().WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "hour"}, {Name: "namespace_id"}},
			DoUpdates: clause.Assignments(map[string]any{
				"push_count": gorm.Expr("namespace_activity_hourlies.push_count + ?", pushDelta),
				"pull_count": gorm.Expr("namespace_activity_hourlies.pull_count + ?", pullDelta),
				"size_delta": gorm.Expr("namespace_activity_hourlies.size_delta + ?", sizeDelta),
				"tag_delta":  gorm.Expr("namespace_activity_hourlies.tag_delta + ?", tagDelta),
				"updated_at": now,
			}),
		}).
		Create(&models.NamespaceActivityHourly{
			ID:          uuid.NewV7String(),
			Hour:        hour,
			NamespaceID: namespaceID,
			PushCount:   pushDelta,
			PullCount:   pullDelta,
			SizeDelta:   sizeDelta,
			TagDelta:    tagDelta,
			CreatedAt:   now,
			UpdatedAt:   now,
		}).Error
}

// ListUserPush lists hourly user push rollups
func (s *analyticsRepository) ListUserPush(
	ctx context.Context,
	userID string, startHour, endHour int64,
) ([]*models.UserActivityHourly, error) {

	return s.tx.UserActivityHourly.WithContext(ctx).
		Where(
			s.tx.UserActivityHourly.UserID.Eq(userID),
			s.tx.UserActivityHourly.Hour.Gte(startHour),
			s.tx.UserActivityHourly.Hour.Lte(endHour),
		).
		Order(s.tx.UserActivityHourly.Hour).
		Find()
}

// ListNamespaceActivity lists hourly namespace activity rollups
func (s *analyticsRepository) ListNamespaceActivity(
	ctx context.Context,
	namespaceID string, startHour, endHour int64,
) ([]*models.NamespaceActivityHourly, error) {

	return s.tx.NamespaceActivityHourly.WithContext(ctx).
		Where(
			s.tx.NamespaceActivityHourly.NamespaceID.Eq(namespaceID),
			s.tx.NamespaceActivityHourly.Hour.Gte(startHour),
			s.tx.NamespaceActivityHourly.Hour.Lte(endHour),
		).
		Order(s.tx.NamespaceActivityHourly.Hour).
		Find()
}

// DeleteBeforeHour deletes rollups older than the retention cutoff
func (s *analyticsRepository) DeleteBeforeHour(ctx context.Context, hour int64) error {
	if _, err := s.tx.UserActivityHourly.WithContext(ctx).
		Where(s.tx.UserActivityHourly.Hour.Lt(hour)).
		Delete(); err != nil {
		return err
	}
	_, err := s.tx.NamespaceActivityHourly.WithContext(ctx).
		Where(s.tx.NamespaceActivityHourly.Hour.Lt(hour)).
		Delete()
	return err
}
