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

package webhooks

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	repowebhook "github.com/go-sigma/sigma/pkg/dal/repository/webhook"
	"github.com/go-sigma/sigma/pkg/testkit"
	"github.com/go-sigma/sigma/pkg/utils/uuid"
)

func TestWebhookQueries(t *testing.T) {
	repository := newTestWebhookRepository(t)
	webhook := createTestWebhook(t, repository)
	service := &webhookService{webhookRepository: repository}

	items, total, err := service.ListWebhooks(t.Context(), nil, api.Pagination{}, api.Sortable{})
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, int64(1), total)

	got, err := service.GetWebhook(t.Context(), webhook.ID)
	require.NoError(t, err)
	require.Equal(t, webhook.ID, got.ID)
}

func TestCreateWebhookRejectsLimitBoundary(t *testing.T) {
	repository := newTestWebhookRepository(t)
	for range consts.MaxWebhooks {
		createTestWebhook(t, repository)
	}
	service := &webhookService{webhookRepository: repository}

	err := service.CreateWebhook(t.Context(), "user-1", api.PostWebhookRequest{
		URL: "https://example.test/hook",
	})
	require.Error(t, err)

	_, total, listErr := repository.List(t.Context(), nil, api.Pagination{}, api.Sortable{})
	require.NoError(t, listErr)
	require.Equal(t, int64(consts.MaxWebhooks), total)
}

func TestUpdateWebhookUpdatesDaemonGCEvent(t *testing.T) {
	repository := newTestWebhookRepository(t)
	webhook := createTestWebhook(t, repository)
	service := &webhookService{webhookRepository: repository}
	enabled := true

	err := service.UpdateWebhook(t.Context(), "user-1", webhook.ID, api.PutWebhookRequest{
		EventDaemonTaskGc: &enabled,
	})
	require.NoError(t, err)

	got, err := repository.Get(t.Context(), webhook.ID)
	require.NoError(t, err)
	require.True(t, got.EventDaemonTaskGc)
}

func TestDeleteWebhookLogDeletesLog(t *testing.T) {
	repository := newTestWebhookRepository(t)
	webhook := createTestWebhook(t, repository)
	log := &models.WebhookLog{
		ID:           uuid.NewV7String(),
		WebhookID:    webhook.ID,
		TraceContext: []byte("{}"),
		ReqHeader:    []byte("{}"),
		ReqBody:      []byte("{}"),
		RespHeader:   []byte("{}"),
		RespBody:     []byte("{}"),
	}
	require.NoError(t, repository.CreateLog(t.Context(), log))
	service := &webhookService{webhookRepository: repository}

	require.NoError(t, service.DeleteWebhookLog(
		t.Context(), "user-1", webhook.ID, log.ID,
	))

	_, err := repository.GetLog(t.Context(), log.ID)
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
	_, err = repository.Get(t.Context(), webhook.ID)
	require.NoError(t, err)
}

func newTestWebhookRepository(t *testing.T) repowebhook.WebhookRepository {
	t.Helper()
	digCon := testkit.InitRepository(t)
	var repository repowebhook.WebhookRepository
	require.NoError(t, digCon.Invoke(func(repo repowebhook.WebhookRepository) {
		repository = repo
	}))
	return repository
}

func createTestWebhook(t *testing.T, repository repowebhook.WebhookRepository) *models.Webhook {
	t.Helper()
	webhook := &models.Webhook{
		ID:  uuid.NewV7String(),
		URL: "https://example.test/hook",
	}
	require.NoError(t, repository.Create(t.Context(), webhook))
	return webhook
}
