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

package dao_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/dal"
	"github.com/go-sigma/sigma/pkg/dal/badger"
	"github.com/go-sigma/sigma/pkg/dal/dao"
	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/logger"
	"github.com/go-sigma/sigma/pkg/modules/locker"
	"github.com/go-sigma/sigma/pkg/modules/locker/definition"
	"github.com/go-sigma/sigma/pkg/tests"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

func TestWebhookServiceFactory(t *testing.T) {
	f := dao.NewWebhookServiceFactory()
	require.NotNil(t, f.New())
	require.NotNil(t, f.New(query.Q))
}

func TestWebhookService(t *testing.T) {
	logger.SetLevel("debug")

	config, err := tests.GetConfig()
	require.NoError(t, err)

	digCon := dig.New()
	require.NoError(t, digCon.Provide(func() configs.Configuration { return ptr.To(config) }))
	require.NoError(t, digCon.Provide(func() (definition.Locker, error) { return locker.Initialize(digCon) }))
	require.NoError(t, digCon.Provide(badger.New))
	require.NoError(t, dal.Initialize(digCon))

	// webhookService := dao.NewWebhookServiceFactory().New()
}

// SIGMA_DATABASE_TYPE=mysql SIGMA_DATABASE_MYSQL_HOST=127.0.0.1 SIGMA_DATABASE_MYSQL_PORT=3306 SIGMA_DATABASE_MYSQL_USERNAME=root SIGMA_DATABASE_MYSQL_PASSWORD=sigma SIGMA_DATABASE_MYSQL_DATABASE=sigma go test -v -run TestWebhookService . -tags viper_bind_struct
// import (
// 	"context"
// 	"testing"

// 	"github.com/rs/zerolog/log"
// 	"github.com/stretchr/testify/assert"

// 	"github.com/go-sigma/sigma/pkg/dal"
// 	"github.com/go-sigma/sigma/pkg/dal/dao"
// 	"github.com/go-sigma/sigma/pkg/logger"
// 	"github.com/go-sigma/sigma/pkg/tests"
// )

// func TestWebhook(t *testing.T) {
// 	logger.SetLevel("debug")
// 	assert.NoError(t, tests.Initialize(t))
// 	assert.NoError(t, tests.DB.Init())
// 	defer func() {
// 		conn, err := dal.DB.DB()
// 		assert.NoError(t, err)
// 		assert.NoError(t, conn.Close())
// 		assert.NoError(t, tests.DB.DeInit())
// 	}()

// 	webhookService := dao.NewWebhookServiceFactory().New()

// 	ctx := log.Logger.WithContext(context.Background())

// 	webhookService.GetByFilter(ctx, map[string]any{"id": 1, "namespace_id": nil}) // nolint: errcheck
// }
