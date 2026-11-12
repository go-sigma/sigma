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

package bootstrap

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/dig"
	"go.uber.org/mock/gomock"

	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	reposettingmocks "github.com/go-sigma/sigma/pkg/dal/repository/setting/mocks"
	lockmocks "github.com/go-sigma/sigma/pkg/infra/lock/mocks"
	"github.com/go-sigma/sigma/pkg/logger"
	"github.com/go-sigma/sigma/pkg/testkit"
)

func TestSigning(t *testing.T) {
	logger.SetLevel("debug")

	digCon := dig.New()

	err := digCon.Provide(func() *config.Configuration {
		return &config.Configuration{
			Locker: config.ConfigurationLocker{
				Type:   enums.LockerTypeRedis,
				Prefix: "sigma-locker",
				Redis:  config.ConfigurationLockerRedis{},
			},
			Redis: config.ConfigurationRedis{
				URL: "redis://:sigma@localhost:6379/0",
			},
		}
	})
	require.NoError(t, err)

	require.NoError(t, testkit.InitializeIntegration(t, digCon))

	require.NoError(t, signing(digCon))
}

func TestSigningSkipsExistingKey(t *testing.T) {
	ctrl := gomock.NewController(t)
	settingRepository := reposettingmocks.NewMockSettingRepository(ctrl)
	locker := lockmocks.NewMockLocker(ctrl)
	lockObj := lockmocks.NewMockLock(ctrl)

	locker.EXPECT().
		Acquire(gomock.Any(), signingKeyLockKey, signingKeyLockTTL, signingKeyLockWait).
		Return(lockObj, nil)
	lockObj.EXPECT().Unlock(gomock.Any()).Return(nil)
	settingRepository.EXPECT().Get(gomock.Any(), consts.SettingSignPrivateKey).
		Return(&models.Setting{Key: consts.SettingSignPrivateKey}, nil)

	require.NoError(t, signingWithParams(signingParams{SettingRepository: settingRepository, Locker: locker}))
}

func TestSigningReturnsLockError(t *testing.T) {
	ctrl := gomock.NewController(t)
	settingRepository := reposettingmocks.NewMockSettingRepository(ctrl)
	locker := lockmocks.NewMockLocker(ctrl)

	lockErr := errors.New("lock timeout")
	locker.EXPECT().
		Acquire(gomock.Any(), signingKeyLockKey, signingKeyLockTTL, signingKeyLockWait).
		Return(nil, lockErr)

	err := signingWithParams(signingParams{SettingRepository: settingRepository, Locker: locker})
	require.ErrorIs(t, err, lockErr)
}
