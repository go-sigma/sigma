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
	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/logger"
)

func TestSettingServiceFactory(t *testing.T) {
	f := dao.NewWorkQueueServiceFactory()
	require.NotNil(t, f.New())
	require.NotNil(t, f.New(query.Q))
}

func TestSettingService(t *testing.T) {
	logger.SetLevel("debug")

	digCon := initDal(t)
	require.NotNil(t, digCon)

	ctx := log.Logger.WithContext(context.Background())

	settingSvc := dao.NewSettingServiceFactory().New()

	require.NoError(t, settingSvc.Create(ctx, "key", []byte("val")))

	settingObj, err := settingSvc.Get(ctx, "key")
	require.NoError(t, err)
	require.NotNil(t, settingObj)
	require.Equal(t, "key", settingObj.Key)
	require.Equal(t, []byte("val"), settingObj.Val)

	require.NoError(t, settingSvc.Update(ctx, "key", []byte("new")))

	settingObj, err = settingSvc.Get(ctx, "key")
	require.NoError(t, err)
	require.NotNil(t, settingObj)
	require.Equal(t, "key", settingObj.Key)
	require.Equal(t, []byte("new"), settingObj.Val)

	require.NoError(t, settingSvc.Delete(ctx, "key"))
}
