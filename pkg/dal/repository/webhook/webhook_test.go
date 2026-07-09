// Copyright 2024 sigma
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

package webhook_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	reponamespace "github.com/go-sigma/sigma/pkg/dal/repository/namespace"
	repowebhook "github.com/go-sigma/sigma/pkg/dal/repository/webhook"
	"github.com/go-sigma/sigma/pkg/testkit"
	"github.com/go-sigma/sigma/pkg/utils/uuid"
)

func TestNewWebhookRepository(t *testing.T) {
	require.NotNil(t, repowebhook.NewWebhookRepository())
	require.NotNil(t, repowebhook.NewWebhookRepository(query.Q))
}

func TestWebhookRepository(t *testing.T) {
	testkit.InitRepository(t)

	ctx := t.Context()
	webhookRepository := repowebhook.NewWebhookRepository()
	nsSvc := reponamespace.NewNamespaceRepository()

	nsObj := &models.Namespace{ID: uuid.NewV7String(), Name: "test"}
	err := nsSvc.Create(ctx, nsObj)
	require.NoError(t, err)

	webhookObj := &models.Webhook{ID: uuid.NewV7String(), NamespaceID: &nsObj.ID, URL: "http://test.com", SslVerify: false}
	err = webhookRepository.Create(ctx, webhookObj)
	require.NoError(t, err)

	{
		var result []*models.Webhook
		result, err = webhookRepository.GetByFilter(ctx, map[string]any{
			query.Webhook.ID.ColumnName().String(): webhookObj.ID,
		})
		require.NoError(t, err)
		require.Equal(t, 1, len(result))
		require.Equal(t, webhookObj.ID, result[0].ID)
	}

	{
		var result []*models.Webhook
		result, err = webhookRepository.GetByFilter(ctx, map[string]any{
			query.Webhook.ID.ColumnName().String(): uuid.NewV7String(),
		})
		require.NoError(t, err)
		require.Equal(t, 0, len(result))
	}

	got, err := webhookRepository.Get(ctx, webhookObj.ID)
	require.NoError(t, err)
	require.Equal(t, webhookObj.URL, got.URL)

	method := enums.SortMethodAsc
	sortBy := "url"
	page := 1
	limit := 10
	webhooks, total, err := webhookRepository.List(ctx, &nsObj.ID, api.Pagination{
		Page:  &page,
		Limit: &limit,
	}, api.Sortable{Sort: &sortBy, Method: &method})
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, webhooks, 1)
	require.Equal(t, webhookObj.ID, webhooks[0].ID)

	require.NoError(t, webhookRepository.UpdateByID(ctx, webhookObj.ID, map[string]any{
		query.Webhook.URL.ColumnName().String(): "http://updated.test",
	}))
	require.NoError(t, webhookRepository.UpdateByID(ctx, webhookObj.ID, map[string]any{}))
	got, err = webhookRepository.Get(ctx, webhookObj.ID)
	require.NoError(t, err)
	require.Equal(t, "http://updated.test", got.URL)

	err = webhookRepository.UpdateByID(ctx, uuid.NewV7String(), map[string]any{
		query.Webhook.URL.ColumnName().String(): "http://missing.test",
	})
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)

	logObj := &models.WebhookLog{
		ID:           uuid.NewV7String(),
		WebhookID:    webhookObj.ID,
		ResourceType: enums.WebhookResourceTypeWebhook,
		Action:       enums.WebhookActionPing,
		StatusCode:   200,
		TraceContext: []byte("{}"),
		ReqHeader:    []byte("{}"),
		ReqBody:      []byte("{}"),
		RespHeader:   []byte("{}"),
		RespBody:     []byte("{}"),
	}
	require.NoError(t, webhookRepository.CreateLog(ctx, logObj))

	gotLog, err := webhookRepository.GetLog(ctx, logObj.ID)
	require.NoError(t, err)
	require.Equal(t, webhookObj.ID, gotLog.Webhook.ID)

	logs, total, err := webhookRepository.ListLogs(ctx, webhookObj.ID, api.Pagination{
		Page:  &page,
		Limit: &limit,
	}, api.Sortable{Sort: &sortBy, Method: &method})
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, logs, 1)
	require.Equal(t, logObj.ID, logs[0].ID)

	require.NoError(t, webhookRepository.DeleteLogByID(ctx, logObj.ID))
	_, err = webhookRepository.GetLog(ctx, logObj.ID)
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)

	require.NoError(t, webhookRepository.DeleteByID(ctx, webhookObj.ID))
	err = webhookRepository.DeleteByID(ctx, uuid.NewV7String())
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}
