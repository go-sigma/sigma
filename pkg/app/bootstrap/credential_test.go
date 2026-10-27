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

package bootstrap

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/consts"
	reposetting "github.com/go-sigma/sigma/pkg/dal/repository/setting"
	"github.com/go-sigma/sigma/pkg/service/credential"
	"github.com/go-sigma/sigma/pkg/testkit"
)

func TestCredentialBootstrapGeneratesKey(t *testing.T) {
	cfg := &config.Configuration{
		Locker: config.ConfigurationLocker{
			Type:   enums.LockerTypeRedis,
			Prefix: "sigma-locker",
			Redis:  config.ConfigurationLockerRedis{},
		},
		Redis: config.ConfigurationRedis{
			URL: "redis://:sigma@localhost:6379/0",
		},
	}

	digCon := dig.New()
	require.NoError(t, digCon.Provide(func() *config.Configuration { return cfg }))
	require.NoError(t, testkit.InitializeIntegration(t, digCon))

	require.NoError(t, credentialBootstrap(digCon))

	settingRepository := reposetting.NewSettingRepository()
	settingObj, err := settingRepository.Get(t.Context(), consts.SettingCredentialMasterKey)
	require.NoError(t, err)
	require.Len(t, settingObj.Val, credential.MinMasterKeyLength)

	// Running the task again must keep the existing key, otherwise credentials written by
	// an earlier start could no longer be decrypted.
	require.NoError(t, credentialBootstrap(digCon))
	again, err := settingRepository.Get(t.Context(), consts.SettingCredentialMasterKey)
	require.NoError(t, err)
	require.Equal(t, settingObj.Val, again.Val)
}

func TestCredentialBootstrapWithConfiguredKey(t *testing.T) {
	cfg := &config.Configuration{
		Auth: config.ConfigurationAuth{CredentialKey: strings.Repeat("k", credential.MinMasterKeyLength)},
		Locker: config.ConfigurationLocker{
			Type:   enums.LockerTypeRedis,
			Prefix: "sigma-locker",
			Redis:  config.ConfigurationLockerRedis{},
		},
		Redis: config.ConfigurationRedis{
			URL: "redis://:sigma@localhost:6379/0",
		},
	}

	digCon := dig.New()
	require.NoError(t, digCon.Provide(func() *config.Configuration { return cfg }))
	require.NoError(t, testkit.InitializeIntegration(t, digCon))

	require.NoError(t, credentialBootstrap(digCon))

	// A configured key is authoritative: nothing is generated into the settings table.
	_, err := reposetting.NewSettingRepository().Get(t.Context(), consts.SettingCredentialMasterKey)
	require.Error(t, err)
}

func TestCredentialBootstrapRejectsShortKey(t *testing.T) {
	cfg := &config.Configuration{
		Auth: config.ConfigurationAuth{CredentialKey: "too-short"},
		Locker: config.ConfigurationLocker{
			Type:   enums.LockerTypeRedis,
			Prefix: "sigma-locker",
			Redis:  config.ConfigurationLockerRedis{},
		},
		Redis: config.ConfigurationRedis{
			URL: "redis://:sigma@localhost:6379/0",
		},
	}

	digCon := dig.New()
	require.NoError(t, digCon.Provide(func() *config.Configuration { return cfg }))
	require.NoError(t, testkit.InitializeIntegration(t, digCon))

	require.Error(t, credentialBootstrap(digCon))
}
