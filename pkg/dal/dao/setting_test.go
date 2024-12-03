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
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/go-sigma/sigma/pkg/dal/dao"
	"github.com/go-sigma/sigma/pkg/dal/query"
)

func TestSettingServiceFactory(t *testing.T) {
	f := dao.NewWorkQueueServiceFactory()
	assert.NotNil(t, f.New())
	assert.NotNil(t, f.New(query.Q))
}

// func TestSettingService(t *testing.T) {
// 	logger.SetLevel("debug")
// 	assert.NoError(t, tests.Initialize(t))
// 	assert.NoError(t, tests.DB.Init())
// 	defer func() {
// 		conn, err := dal.DB.DB()
// 		assert.NoError(t, err)
// 		assert.NoError(t, conn.Close())
// 		assert.NoError(t, tests.DB.DeInit())
// 	}()

// 	ctx := log.Logger.WithContext(context.Background())

// 	settingServiceFactory := dao.NewSettingServiceFactory()
// 	settingService := settingServiceFactory.New()
// 	assert.NotNil(t, settingService)

// 	assert.NoError(t, settingService.Create(ctx, "key", []byte("val")))

// 	settingObj, err := settingService.Get(ctx, "key")
// 	assert.NoError(t, err)
// 	assert.NotNil(t, settingObj)
// 	assert.Equal(t, "key", settingObj.Key)
// 	assert.Equal(t, []byte("val"), settingObj.Val)

// 	assert.NoError(t, settingService.Update(ctx, "key", []byte("new")))

// 	settingObj, err = settingService.Get(ctx, "key")
// 	assert.NoError(t, err)
// 	assert.NotNil(t, settingObj)
// 	assert.Equal(t, "key", settingObj.Key)
// 	assert.Equal(t, []byte("new"), settingObj.Val)

// 	assert.NoError(t, settingService.Delete(ctx, "key"))
// }
