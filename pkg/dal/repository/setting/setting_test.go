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

package setting_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/dal/query"
	reposetting "github.com/go-sigma/sigma/pkg/dal/repository/setting"
	"github.com/go-sigma/sigma/pkg/logger"
	"github.com/go-sigma/sigma/pkg/testkit"
)

func TestNewSettingRepository(t *testing.T) {
	require.NotNil(t, reposetting.NewSettingRepository())
	require.NotNil(t, reposetting.NewSettingRepository(query.Q))
}

func TestSettingRepository(t *testing.T) {
	logger.SetLevel("debug")

	digCon := testkit.InitRepository(t)
	require.NotNil(t, digCon)

	ctx := context.Background()

	settingSvc := reposetting.NewSettingRepository()

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

	_, err = settingSvc.Get(ctx, "key")
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)

	err = settingSvc.Delete(ctx, "missing")
	require.ErrorIs(t, err, gorm.ErrRecordNotFound)
}
