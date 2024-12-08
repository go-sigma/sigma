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

package dao_test

import (
	"context"
	"testing"

	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"

	"github.com/go-sigma/sigma/pkg/dal/dao"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/logger"
)

func TestWebhookServiceFactory(t *testing.T) {
	f := dao.NewWebhookServiceFactory()
	require.NotNil(t, f.New())
	require.NotNil(t, f.New(query.Q))
}

func TestWebhookService(t *testing.T) {
	logger.SetLevel("debug")

	digCon := initDal(t)
	require.NotNil(t, digCon)

	ctx := log.Logger.WithContext(context.Background())

	webhookService := dao.NewWebhookServiceFactory().New()
	nsSvc := dao.NewNamespaceServiceFactory().New()

	nsObj := &models.Namespace{Name: "test"}
	err := nsSvc.Create(ctx, nsObj)
	require.NoError(t, err)

	webhookObj := &models.Webhook{NamespaceID: &nsObj.ID, URL: "http://test.com", SslVerify: false}
	err = webhookService.Create(ctx, webhookObj)
	require.NoError(t, err)

	{
		result, err := webhookService.GetByFilter(ctx, map[string]any{
			query.Webhook.ID.ColumnName().String(): webhookObj.ID,
		})
		require.NoError(t, err)
		require.Equal(t, 1, len(result))
		require.Equal(t, webhookObj.ID, result[0].ID)
	}

	{
		result, err := webhookService.GetByFilter(ctx, map[string]any{
			query.Webhook.ID.ColumnName().String(): 9999,
		})
		require.NoError(t, err)
		require.Equal(t, 0, len(result))
	}
}
