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

package webhooks

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"go.uber.org/dig"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	repowebhook "github.com/go-sigma/sigma/pkg/dal/repository/webhook"
	"github.com/go-sigma/sigma/pkg/infra/workq"
	"github.com/go-sigma/sigma/pkg/server/errcode"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
	"github.com/go-sigma/sigma/pkg/utils/uuid"
)

//go:generate mockgen -mock_names Service=MockWebhookService -destination=webhooks_mocks.go -package=webhooks github.com/go-sigma/sigma/pkg/service/webhooks Service

// Service encapsulates webhook-related business logic.
type Service interface {
	// CreateWebhook creates a webhook (includes: validate url + check quota + create).
	CreateWebhook(ctx context.Context, userID string, req api.PostWebhookRequest) error
	// ListWebhooks lists webhooks with pagination.
	ListWebhooks(ctx context.Context, namespaceID *string, pagination api.Pagination, sort api.Sortable) ([]*models.Webhook, int64, error)
	// GetWebhook gets a webhook by ID.
	GetWebhook(ctx context.Context, id string) (*models.Webhook, error)
	// UpdateWebhook updates a webhook.
	UpdateWebhook(ctx context.Context, userID string, id string, req api.PutWebhookRequest) error
	// DeleteWebhook deletes a webhook.
	DeleteWebhook(ctx context.Context, userID string, id string) error
	// PingWebhook sends a ping event.
	PingWebhook(ctx context.Context, userID string, webhookID string) error
	// GetWebhookLog gets a webhook log by ID.
	GetWebhookLog(ctx context.Context, webhookLogID string) (*models.WebhookLog, error)
	// DeleteWebhookLog deletes a webhook log.
	DeleteWebhookLog(ctx context.Context, userID string, webhookID string, webhookLogID string) error
	// ListWebhookLogs lists webhook logs with pagination.
	ListWebhookLogs(ctx context.Context, webhookID string, pagination api.Pagination, sort api.Sortable) ([]*models.WebhookLog, int64, error)
	// ResendWebhookLog resends a webhook log.
	ResendWebhookLog(ctx context.Context, userID string, webhookID string, webhookLogID string) error
}

type service struct {
	dig.In

	RepoWebhook repowebhook.WebhookRepository
	Producer    workq.Producer
}

func NewService(params service) Service {
	return &params
}

func (s *service) CreateWebhook(ctx context.Context, userID string, req api.PostWebhookRequest) error {
	if !(strings.HasPrefix(req.URL, "http://") || strings.HasPrefix(req.URL, "https://")) { // nolint: staticcheck
		slog.Error("uRL is invalid", "URL", req.URL)
		return errcode.HTTPErrCodeBadRequest.Detail("URL is invalid, should start with 'http://' or 'https://'")
	}

	_, total, err := s.RepoWebhook.List(ctx, req.NamespaceID, api.Pagination{}, api.Sortable{})
	if err != nil {
		slog.Error("get webhook count failed", "err", err)
		return errcode.HTTPErrCodeInternalError.Detail(err.Error())
	}
	if total >= consts.MaxWebhooks {
		slog.Error("reached the maximum webhooks", "total", total)
		return errcode.HTTPErrCodeBadRequest.Detail("Reached the maximum webhooks")
	}

	return query.Q.Transaction(func(tx *query.Query) error {
		repoWebhook := repowebhook.NewWebhookRepository(tx)
		namespaceID := req.NamespaceID
		if ptr.To(req.NamespaceID) == "" {
			namespaceID = nil
		}
		webhookObj := &models.Webhook{
			ID:                uuid.NewV7String(),
			NamespaceID:       namespaceID,
			URL:               req.URL,
			Secret:            req.Secret,
			SslVerify:         req.SslVerify,
			RetryTimes:        req.RetryTimes,
			RetryDuration:     req.RetryDuration,
			Enable:            req.Enable,
			EventNamespace:    req.EventNamespace,
			EventRepository:   req.EventRepository,
			EventTag:          req.EventTag,
			EventArtifact:     req.EventArtifact,
			EventMember:       req.EventMember,
			EventDaemonTaskGc: req.EventDaemonTaskGc,
		}
		err = repoWebhook.Create(ctx, webhookObj)
		if err != nil {
			slog.Error("create webhook failed", "err", err)
			return errcode.HTTPErrCodeInternalError.Detail("Create webhook failed")
		}
		return nil
	})
}

func (s *service) ListWebhooks(ctx context.Context, namespaceID *string, pagination api.Pagination, sort api.Sortable) ([]*models.Webhook, int64, error) {
	return s.RepoWebhook.List(ctx, namespaceID, pagination, sort)
}

func (s *service) GetWebhook(ctx context.Context, id string) (*models.Webhook, error) {
	webhookObj, err := s.RepoWebhook.Get(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Error("webhook not found", "err", err, "id", id)
			return nil, errcode.HTTPErrCodeNotFound.Detail("Webhook not found")
		}
		slog.Error("get webhook failed", "err", err)
		return nil, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Get webhook(%s) failed", id))
	}
	return webhookObj, nil
}

func (s *service) UpdateWebhook(ctx context.Context, userID string, id string, req api.PutWebhookRequest) error {
	_, err := s.RepoWebhook.Get(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Error("webhook not found", "err", err, "id", id)
			return errcode.HTTPErrCodeNotFound.Detail("Webhook not found")
		}
		slog.Error("get webhook failed", "err", err)
		return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Get webhook(%s) failed", id))
	}

	updates := make(map[string]any, 15)
	if req.Url != nil {
		updates[query.Webhook.URL.ColumnName().String()] = ptr.To(req.Url)
	}
	if req.Secret != nil {
		updates[query.Webhook.Secret.ColumnName().String()] = ptr.To(req.Secret)
	}
	if req.SslVerify != nil {
		updates[query.Webhook.SslVerify.ColumnName().String()] = ptr.To(req.SslVerify)
	}
	if req.RetryTimes != nil {
		updates[query.Webhook.RetryTimes.ColumnName().String()] = ptr.To(req.RetryTimes)
	}
	if req.RetryDuration != nil {
		updates[query.Webhook.RetryDuration.ColumnName().String()] = ptr.To(req.RetryDuration)
	}
	if req.Enable != nil {
		updates[query.Webhook.Enable.ColumnName().String()] = ptr.To(req.Enable)
	}
	if req.EventNamespace != nil {
		updates[query.Webhook.EventNamespace.ColumnName().String()] = ptr.To(req.EventNamespace)
	}
	if req.EventRepository != nil {
		updates[query.Webhook.EventRepository.ColumnName().String()] = ptr.To(req.EventRepository)
	}
	if req.EventTag != nil {
		updates[query.Webhook.EventTag.ColumnName().String()] = ptr.To(req.EventTag)
	}
	if req.EventArtifact != nil {
		updates[query.Webhook.EventArtifact.ColumnName().String()] = ptr.To(req.EventArtifact)
	}
	if req.EventMember != nil {
		updates[query.Webhook.EventMember.ColumnName().String()] = ptr.To(req.EventMember)
	}
	if req.EventDaemonTaskGc != nil {
		updates[query.Webhook.EventDaemonTaskGc.ColumnName().String()] = ptr.To(req.EventDaemonTaskGc)
	}

	return query.Q.Transaction(func(tx *query.Query) error {
		repoWebhook := repowebhook.NewWebhookRepository(tx)
		err = repoWebhook.UpdateByID(ctx, id, updates)
		if err != nil {
			slog.Error("update webhook failed", "err", err)
			return errcode.HTTPErrCodeInternalError.Detail("Update webhook failed")
		}
		return nil
	})
}

func (s *service) DeleteWebhook(ctx context.Context, userID string, id string) error {
	_, err := s.RepoWebhook.Get(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Error("webhook not found", "err", err, "id", id)
			return errcode.HTTPErrCodeNotFound.Detail("Webhook not found")
		}
		slog.Error("get webhook failed", "err", err)
		return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Get webhook(%s) failed", id))
	}

	return query.Q.Transaction(func(tx *query.Query) error {
		repoWebhook := repowebhook.NewWebhookRepository(tx)
		err = repoWebhook.DeleteByID(ctx, id)
		if err != nil {
			slog.Error("delete webhook failed", "err", err)
			return errcode.HTTPErrCodeInternalError.Detail("Delete webhook failed")
		}
		return nil
	})
}

func (s *service) PingWebhook(ctx context.Context, userID string, webhookID string) error {
	webhookObj, err := s.RepoWebhook.Get(ctx, webhookID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Error("webhook not found", "err", err, "WebhookID", webhookID)
			return errcode.HTTPErrCodeNotFound.Detail("Webhook not found")
		}
		slog.Error("get webhook failed", "err", err, "WebhookID", webhookID)
		return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Get webhook(%s) failed", webhookID))
	}

	return query.Q.Transaction(func(tx *query.Query) error {
		err := s.Producer.Produce(ctx, enums.DaemonWebhook, api.DaemonWebhookPayload{
			NamespaceID:  webhookObj.NamespaceID,
			WebhookID:    webhookObj.ID,
			Type:         enums.WebhookTypePing,
			Action:       enums.WebhookActionPing,
			ResourceType: enums.WebhookResourceTypeWebhook,
		})
		if err != nil {
			slog.Error("webhook event produce failed", "err", err)
			return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Webhook event produce failed: %v", err))
		}
		return nil
	})
}

func (s *service) GetWebhookLog(ctx context.Context, webhookLogID string) (*models.WebhookLog, error) {
	webhookLogObj, err := s.RepoWebhook.GetLog(ctx, webhookLogID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Error("webhook log not found", "err", err, "id", webhookLogID)
			return nil, errcode.HTTPErrCodeNotFound.Detail("Webhook log not found")
		}
		slog.Error("get webhook log failed", "err", err, "id", webhookLogID)
		return nil, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Get webhook log(%s) failed", webhookLogID))
	}
	return webhookLogObj, nil
}

func (s *service) DeleteWebhookLog(ctx context.Context, userID string, webhookID string, webhookLogID string) error {
	_, err := s.RepoWebhook.Get(ctx, webhookID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Error("webhook not found", "err", err, "WebhookID", webhookID, "WebhookLogID", webhookLogID)
			return errcode.HTTPErrCodeNotFound.Detail("Webhook not found")
		}
		slog.Error("get webhook failed", "err", err, "WebhookID", webhookID, "WebhookLogID", webhookLogID)
		return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Get webhook(%s) failed", webhookID))
	}

	return query.Q.Transaction(func(tx *query.Query) error {
		repoWebhook := repowebhook.NewWebhookRepository(tx)
		err = repoWebhook.DeleteLogByID(ctx, webhookLogID)
		if err != nil {
			slog.Error("delete webhook log failed", "err", err)
			return errcode.HTTPErrCodeInternalError.Detail("Delete webhook log failed")
		}
		return nil
	})
}

func (s *service) ListWebhookLogs(ctx context.Context, webhookID string, pagination api.Pagination, sort api.Sortable) ([]*models.WebhookLog, int64, error) {
	return s.RepoWebhook.ListLogs(ctx, webhookID, pagination, sort)
}

func (s *service) ResendWebhookLog(ctx context.Context, userID string, webhookID string, webhookLogID string) error {
	webhookLogObj, err := s.RepoWebhook.GetLog(ctx, webhookLogID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Error("webhook log not found", "err", err, "WebhookID", webhookID, "WebhookLogID", webhookLogID)
			return errcode.HTTPErrCodeNotFound.Detail("Webhook log not found")
		}
		slog.Error("get webhook failed", "err", err, "WebhookID", webhookID, "WebhookLogID", webhookLogID)
		return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Get webhook log(%s) failed", webhookLogID))
	}

	return query.Q.Transaction(func(tx *query.Query) error {
		err := s.Producer.Produce(ctx, enums.DaemonWebhook, api.DaemonWebhookPayload{
			NamespaceID:  webhookLogObj.Webhook.NamespaceID,
			WebhookID:    webhookLogObj.Webhook.ID,
			WebhookLogID: new(webhookLogID),
			Type:         enums.WebhookTypeResend,
		})
		if err != nil {
			slog.Error("webhook event produce failed", "err", err)
			return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Webhook event produce failed: %v", err))
		}
		return nil
	})
}
