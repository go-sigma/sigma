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

	"gorm.io/gen/field"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

//go:generate mockgen -destination=mocks/webhook.go -package=mocks github.com/go-sigma/sigma/pkg/dal/dao WebhookService
//go:generate mockgen -destination=mocks/webhook_factory.go -package=mocks github.com/go-sigma/sigma/pkg/dal/dao WebhookServiceFactory

// WebhookService is the interface that provides methods to operate on webhook model
type WebhookService interface {
	// Create a new webhook
	Create(ctx context.Context, webhook *models.Webhook) error
	// List all webhook with pagination
	List(ctx context.Context, namespaceID *int64, pagination types.Pagination, sort types.Sortable) ([]*models.Webhook, int64, error)
	// Get gets the webhook with the specified webhook ID.
	Get(ctx context.Context, id int64) (*models.Webhook, error)
	// GetByFilter gets the webhook with the specified filter.
	GetByFilter(ctx context.Context, filter map[string]any) ([]*models.Webhook, error)
	// DeleteByID deletes the webhook with the specified webhook ID.
	DeleteByID(ctx context.Context, id int64) error
	// UpdateByID updates the webhook with the specified webhook ID.
	UpdateByID(ctx context.Context, id int64, updates map[string]interface{}) error
	// CreateLog create a new webhook log
	CreateLog(ctx context.Context, webhookLog *models.WebhookLog) error
	// ListLogs all webhook logs with pagination
	ListLogs(ctx context.Context, webhookID int64, pagination types.Pagination, sort types.Sortable) ([]*models.WebhookLog, int64, error)
	// GetLog get webhook log with the specified webhook ID
	GetLog(ctx context.Context, webhookLogID int64) (*models.WebhookLog, error)
	// DeleteLogByID delete webhook log by id
	DeleteLogByID(ctx context.Context, webhookLogID int64) error
}

type webhookService struct {
	tx *query.Query
}

// WebhookServiceFactory is the interface that provides the webhook service factory methods.
type WebhookServiceFactory interface {
	New(txs ...*query.Query) WebhookService
}

type webhookServiceFactory struct{}

// NewWebhookServiceFactory creates a new webhook service factory.
func NewWebhookServiceFactory() WebhookServiceFactory {
	return &webhookServiceFactory{}
}

// New ...
func (s *webhookServiceFactory) New(txs ...*query.Query) WebhookService {
	tx := query.Q
	if len(txs) > 0 {
		tx = txs[0]
	}
	return &webhookService{
		tx: tx,
	}
}

// Create a new webhook
func (s *webhookService) Create(ctx context.Context, webhook *models.Webhook) error {
	return s.tx.Webhook.WithContext(ctx).Create(webhook)
}

// List all webhook with pagination
func (s *webhookService) List(ctx context.Context, namespaceID *int64, pagination types.Pagination, sort types.Sortable) ([]*models.Webhook, int64, error) {
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

// Get gets the webhook with the specified webhook ID.
func (s *webhookService) Get(ctx context.Context, id int64) (*models.Webhook, error) {
	return s.tx.Webhook.WithContext(ctx).Where(s.tx.Webhook.ID.Eq(id)).First()
}

// GetByFilter gets the webhook with the specified filter.
func (s *webhookService) GetByFilter(ctx context.Context, filter map[string]any) ([]*models.Webhook, error) {
	return s.tx.Webhook.WithContext(ctx).Where(field.Attrs(filter)).Find()
}

// ListLogs all webhook logs with pagination
func (s *webhookService) ListLogs(ctx context.Context, webhookID int64, pagination types.Pagination, sort types.Sortable) ([]*models.WebhookLog, int64, error) {
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

// DeleteByID deletes the webhook with the specified webhook ID.
func (s *webhookService) DeleteByID(ctx context.Context, id int64) error {
	matched, err := s.tx.Webhook.WithContext(ctx).Where(s.tx.Webhook.ID.Eq(id)).Delete()
	if err != nil {
		return err
	}
	if matched.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// UpdateByID updates the webhook with the specified webhook ID.
func (s *webhookService) UpdateByID(ctx context.Context, id int64, updates map[string]interface{}) error {
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

// CreateLog create a new webhook log
func (s *webhookService) CreateLog(ctx context.Context, webhookLog *models.WebhookLog) error {
	return s.tx.WebhookLog.WithContext(ctx).Create(webhookLog)
}

// GetLog get webhook log with the specified webhook ID
func (s *webhookService) GetLog(ctx context.Context, webhookLogID int64) (*models.WebhookLog, error) {
	return s.tx.WebhookLog.WithContext(ctx).Where(s.tx.WebhookLog.ID.Eq(webhookLogID)).Preload(s.tx.WebhookLog.Webhook).First()
}

// DeleteLogByID delete webhook log by id
func (s *webhookService) DeleteLogByID(ctx context.Context, webhookLogID int64) error {
	_, err := s.tx.WebhookLog.WithContext(ctx).Where(s.tx.WebhookLog.ID.Eq(webhookLogID)).Delete()
	return err
}
