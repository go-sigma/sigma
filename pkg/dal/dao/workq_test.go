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
	"github.com/stretchr/testify/assert"

	"github.com/go-sigma/sigma/pkg/dal"
	"github.com/go-sigma/sigma/pkg/dal/dao"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/logger"
	"github.com/go-sigma/sigma/pkg/tests"
	"github.com/go-sigma/sigma/pkg/types/enums"
)

func TestWorkQueueServiceFactory(t *testing.T) {
	f := dao.NewWorkQueueServiceFactory()
	artifactService := f.New()
	assert.NotNil(t, artifactService)
	artifactService = f.New(query.Q)
	assert.NotNil(t, artifactService)
}

func TestWorkQueueService(t *testing.T) {
	logger.SetLevel("debug")
	assert.NoError(t, tests.Initialize(t))
	assert.NoError(t, tests.DB.Init())
	defer func() {
		conn, err := dal.DB.DB()
		assert.NoError(t, err)
		assert.NoError(t, conn.Close())
		assert.NoError(t, tests.DB.DeInit())
	}()

	ctx := log.Logger.WithContext(context.Background())

	workqServiceFactory := dao.NewWorkQueueServiceFactory()
	workqService := workqServiceFactory.New()
	assert.NotNil(t, workqService)

	workqObj := &models.WorkQueue{
		Topic:   enums.DaemonGc,
		Payload: []byte("payload"),
		Version: "version",
	}
	assert.NoError(t, workqService.Create(ctx, workqObj))

	assert.NoError(t, workqService.UpdateStatus(ctx, workqObj.ID, "version", "newVersion", 1, enums.TaskCommonStatusPending))

	workqNewObj, err := workqService.Get(ctx, enums.DaemonGc)
	assert.NoError(t, err)
	assert.Equal(t, workqObj.ID, workqNewObj.ID)
	assert.Equal(t, workqObj.Topic, workqNewObj.Topic)
	assert.Equal(t, workqObj.Payload, workqNewObj.Payload)
	assert.Equal(t, []byte("payload"), workqNewObj.Payload)
	assert.Equal(t, 1, workqNewObj.Times)
	assert.Equal(t, enums.TaskCommonStatusPending, workqNewObj.Status)
}
