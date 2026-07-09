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

package webhook

import (
	"context"

	"gorm.io/gen/field"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

//go:generate mockgen -destination=webhook_mocks.go -package=webhook github.com/go-sigma/sigma/pkg/dal/repository/webhook WebhookRepository

// WebhookRepository defines webhook repository operations
type WebhookRepository interface {
	// Create creates a new webhook
	Create(ctx context.Context, webhook *models.Webhook) error
	// List lists webhooks with filtering, pagination, and sorting
	List(ctx context.Context, namespaceID *string, pagination api.Pagination, sort api.Sortable) ([]*models.Webhook, int64, error)
	// Get gets the webhook with the specified webhook ID
	Get(ctx context.Context, id string) (*models.Webhook, error)
	// GetByFilter gets webhooks with the specified filter
	GetByFilter(ctx context.Context, filter map[string]any) ([]*models.Webhook, error)
	// DeleteByID deletes the webhook with the specified webhook ID
	DeleteByID(ctx context.Context, id string) error
	// UpdateByID updates the webhook with the specified webhook ID
	UpdateByID(ctx context.Context, id string, updates map[string]any) error
	// CreateLog creates a new webhook log
	CreateLog(ctx context.Context, webhookLog *models.WebhookLog) error
	// ListLogs lists webhook logs with pagination and sorting
	ListLogs(ctx context.Context, webhookID string, pagination api.Pagination, sort api.Sortable) ([]*models.WebhookLog, int64, error)
	// GetLog gets a webhook log with the specified log ID
	GetLog(ctx context.Context, webhookLogID string) (*models.WebhookLog, error)
	// DeleteLogByID deletes a webhook log by ID
	DeleteLogByID(ctx context.Context, webhookLogID string) error
}

type webhookRepository struct {
	tx *query.Query
}

// NewWebhookRepository creates a new webhook repository with the optional query transaction
func NewWebhookRepository(txs ...*query.Query) WebhookRepository {
	tx := query.Q
	if len(txs) > 0 {
		tx = txs[0]
	}
	return &webhookRepository{
		tx: tx,
	}
}

// Create creates a new webhook
func (s *webhookRepository) Create(ctx context.Context, webhook *models.Webhook) error {
	return s.tx.Webhook.WithContext(ctx).Create(webhook)
}

// List lists webhooks with filtering, pagination, and sorting
func (s *webhookRepository) List(ctx context.Context, namespaceID *string, pagination api.Pagination, sort api.Sortable) ([]*models.Webhook, int64, error) {
	pagination = utils.NormalizePagination(pagination)
	q := s.tx.Webhook.WithContext(ctx)
	if namespaceID != nil {
		q = q.Where(s.tx.Webhook.NamespaceID.Eq(ptr.To(namespaceID)))
	}
	f, ok := s.tx.Webhook.GetFieldByName(ptr.To(sort.Sort))
	if ok {
		switch ptr.To(sort.Method) {
		case enums.SortMethodDesc:
			q = q.Order(f.Desc())
		case enums.SortMethodAsc:
			q = q.Order(f)
		default:
			q = q.Order(s.tx.Webhook.UpdatedAt.Desc())
		}
	} else {
		q = q.Order(s.tx.Webhook.UpdatedAt.Desc())
	}
	return q.FindByPage(ptr.To(pagination.Limit)*(ptr.To(pagination.Page)-1), ptr.To(pagination.Limit))
}

// Get gets the webhook with the specified webhook ID
func (s *webhookRepository) Get(ctx context.Context, id string) (*models.Webhook, error) {
	return s.tx.Webhook.WithContext(ctx).Where(s.tx.Webhook.ID.Eq(id)).First()
}

// GetByFilter gets webhooks with the specified filter
func (s *webhookRepository) GetByFilter(ctx context.Context, filter map[string]any) ([]*models.Webhook, error) {
	return s.tx.Webhook.WithContext(ctx).Where(field.Attrs(filter)).Find()
}

// ListLogs lists webhook logs with pagination and sorting
func (s *webhookRepository) ListLogs(ctx context.Context, webhookID string, pagination api.Pagination, sort api.Sortable) ([]*models.WebhookLog, int64, error) {
	pagination = utils.NormalizePagination(pagination)
	q := s.tx.WebhookLog.WithContext(ctx).Where(s.tx.WebhookLog.WebhookID.Eq(webhookID))
	f, ok := s.tx.WebhookLog.GetFieldByName(ptr.To(sort.Sort))
	if ok {
		switch ptr.To(sort.Method) {
		case enums.SortMethodDesc:
			q = q.Order(f.Desc())
		case enums.SortMethodAsc:
			q = q.Order(f)
		default:
			q = q.Order(s.tx.WebhookLog.UpdatedAt.Desc())
		}
	} else {
		q = q.Order(s.tx.WebhookLog.UpdatedAt.Desc())
	}
	return q.FindByPage(ptr.To(pagination.Limit)*(ptr.To(pagination.Page)-1), ptr.To(pagination.Limit))
}

// DeleteByID deletes the webhook with the specified webhook ID
func (s *webhookRepository) DeleteByID(ctx context.Context, id string) error {
	matched, err := s.tx.Webhook.WithContext(ctx).Where(s.tx.Webhook.ID.Eq(id)).Delete()
	if err != nil {
		return err
	}
	if matched.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// UpdateByID updates the webhook with the specified webhook ID
func (s *webhookRepository) UpdateByID(ctx context.Context, id string, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}
	matched, err := s.tx.Webhook.WithContext(ctx).Where(s.tx.Webhook.ID.Eq(id)).Updates(updates)
	if err != nil {
		return err
	}
	if matched.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// CreateLog creates a new webhook log
func (s *webhookRepository) CreateLog(ctx context.Context, webhookLog *models.WebhookLog) error {
	return s.tx.WebhookLog.WithContext(ctx).Create(webhookLog)
}

// GetLog gets a webhook log with the specified log ID
func (s *webhookRepository) GetLog(ctx context.Context, webhookLogID string) (*models.WebhookLog, error) {
	return s.tx.WebhookLog.WithContext(ctx).Where(s.tx.WebhookLog.ID.Eq(webhookLogID)).Preload(s.tx.WebhookLog.Webhook).First()
}

// DeleteLogByID deletes a webhook log by ID
func (s *webhookRepository) DeleteLogByID(ctx context.Context, webhookLogID string) error {
	_, err := s.tx.WebhookLog.WithContext(ctx).Where(s.tx.WebhookLog.ID.Eq(webhookLogID)).Delete()
	return err
}
