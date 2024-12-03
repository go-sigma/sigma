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
	"github.com/go-sigma/sigma/pkg/types/enums"
)

func TestWorkQueueServiceFactory(t *testing.T) {
	f := dao.NewWorkQueueServiceFactory()
	require.NotNil(t, f.New())
	require.NotNil(t, f.New(query.Q))
}

func TestWorkQueueService(t *testing.T) {
	logger.SetLevel("debug")

	digCon := initDal(t)
	require.NotNil(t, digCon)

	ctx := log.Logger.WithContext(context.Background())

	wqSvc := dao.NewWorkQueueServiceFactory().New()

	wqObj := &models.WorkQueue{
		Topic:   enums.DaemonGc,
		Payload: []byte("payload"),
		Version: "version",
	}
	require.NoError(t, wqSvc.Create(ctx, wqObj))

	require.NoError(t, wqSvc.UpdateStatus(ctx, wqObj.ID, "version", "newVersion", 1, enums.TaskCommonStatusPending))

	wqNewObj, err := wqSvc.Get(ctx, enums.DaemonGc)
	require.NoError(t, err)
	require.Equal(t, wqObj.ID, wqNewObj.ID)
	require.Equal(t, wqObj.Topic, wqNewObj.Topic)
	require.Equal(t, wqObj.Payload, wqNewObj.Payload)
	require.Equal(t, []byte("payload"), wqNewObj.Payload)
	require.Equal(t, 1, wqNewObj.Times)
	require.Equal(t, enums.TaskCommonStatusPending, wqNewObj.Status)
}
