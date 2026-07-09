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

package dal_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/dal"
	"github.com/go-sigma/sigma/pkg/infra/lock"
	"github.com/go-sigma/sigma/pkg/testkit"
)

func TestInitialize(t *testing.T) {
	databaseType, err := testkit.DatabaseType(enums.DatabaseSqlite3)
	require.NoError(t, err)
	databaseConfig, err := testkit.PrepareDatabase(t, databaseType)
	require.NoError(t, err)

	digCon := dig.New()
	err = digCon.Provide(func() *config.Configuration {
		return &config.Configuration{
			Database: databaseConfig,
			Locker: config.ConfigurationLocker{
				Type:   enums.LockerTypeInmemory,
				Prefix: "sigma-locker",
			},
		}
	})
	require.NoError(t, err)

	err = digCon.Provide(lock.Initialize)
	require.NoError(t, err)

	require.NoError(t, dal.Initialize(digCon))
	t.Cleanup(func() {
		require.NoError(t, dal.DeInitialize(digCon))
	})
}

func TestInitializeUnknownDatabase(t *testing.T) {
	digCon := dig.New()
	require.NoError(t, digCon.Provide(func() *config.Configuration {
		return &config.Configuration{}
	}))
	require.Error(t, dal.Initialize(digCon))
}
