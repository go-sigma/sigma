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

// import (
// 	"context"
// 	"testing"

// 	"github.com/rs/zerolog/log"
// 	"github.com/spf13/viper"
// 	"github.com/stretchr/testify/assert"

// 	"github.com/go-sigma/sigma/pkg/dal"
// 	"github.com/go-sigma/sigma/pkg/dal/dao"
// 	"github.com/go-sigma/sigma/pkg/dal/models"
// 	"github.com/go-sigma/sigma/pkg/logger"
// 	"github.com/go-sigma/sigma/pkg/tests"
// 	"github.com/go-sigma/sigma/pkg/types/enums"
// 	"github.com/go-sigma/sigma/pkg/utils/ptr"
// )

// func TestDaemonService(t *testing.T) {
// 	viper.SetDefault("log.level", "debug")
// 	logger.SetLevel("debug")
// 	err := tests.Initialize(t)
// 	assert.NoError(t, err)
// 	err = tests.DB.Init()
// 	assert.NoError(t, err)
// 	defer func() {
// 		conn, err := dal.DB.DB()
// 		assert.NoError(t, err)
// 		err = conn.Close()
// 		assert.NoError(t, err)
// 		err = tests.DB.DeInit()
// 		assert.NoError(t, err)
// 	}()

// 	ctx := log.Logger.WithContext(context.Background())

// 	namespaceService := dao.NewNamespaceServiceFactory().New()
// 	daemonService := dao.NewDaemonServiceFactory().New()

// 	namespaceObj := models.Namespace{
// 		Name: "test",
// 	}
// 	err = namespaceService.Create(ctx, &namespaceObj)
// 	assert.NoError(t, err)

// 	err = daemonService.CreateGcRepositoryRunner(ctx, &models.DaemonGcRepositoryRunner{
// 		Status: enums.TaskCommonStatusPending,
// 	})
// 	assert.NoError(t, err)
// 	err = daemonService.CreateGcRepositoryRunner(ctx, &models.DaemonGcRepositoryRunner{
// 		NamespaceID: ptr.Of(namespaceObj.ID),
// 		Status:      enums.TaskCommonStatusPending,
// 	})
// 	assert.NoError(t, err)

// 	_, err = daemonService.GetLastGcRepositoryRunner(ctx, nil)
// 	assert.NoError(t, err)

// 	_, err = daemonService.GetLastGcRepositoryRunner(ctx, &namespaceObj.ID)
// 	assert.NoError(t, err)
// }
