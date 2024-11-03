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

package inits

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/dal"
	"github.com/go-sigma/sigma/pkg/logger"
	"github.com/go-sigma/sigma/pkg/tests"
	"github.com/go-sigma/sigma/pkg/types/enums"
)

func TestSigning(t *testing.T) {
	logger.SetLevel("debug")

	digCon := dig.New()

	err := digCon.Provide(func() configs.Configuration {
		return configs.Configuration{
			Locker: configs.ConfigurationLocker{
				Type:   enums.LockerTypeRedis,
				Prefix: "sigma-locker",
				Redis:  configs.ConfigurationLockerRedis{},
			},
			Redis: configs.ConfigurationRedis{
				Type: enums.RedisTypeExternal,
				URL:  "redis://:sigma@localhost:6379/0",
			},
			Database: configs.ConfigurationDatabase{
				Type: enums.DatabaseSqlite3,
				Sqlite3: configs.ConfigurationDatabaseSqlite3{
					Path: fmt.Sprintf("%s.db", strings.ReplaceAll(uuid.Must(uuid.NewV7()).String(), "-", "")),
				},
				Mysql: configs.ConfigurationDatabaseMysql{
					Host:     "127.0.0.1",
					Port:     3306,
					Username: "root",
					Password: "sigma",
					Database: strings.ReplaceAll(uuid.Must(uuid.NewV7()).String(), "-", ""),
				},
				Postgresql: configs.ConfigurationDatabasePostgresql{
					Host:     "127.0.0.1",
					Port:     5432,
					Username: "sigma",
					Password: "sigma",
					Database: strings.ReplaceAll(uuid.Must(uuid.NewV7()).String(), "-", ""),
					SslMode:  "disable",
				},
			},
		}
	})
	require.NoError(t, err)

	tests, err := tests.Initialize(t, digCon)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, dal.DeInitialize())
		require.NoError(t, tests.DeInitialize())
	}()

	require.NoError(t, signing(digCon))
}
