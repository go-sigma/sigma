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

package handlers

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/dal"
	"github.com/go-sigma/sigma/pkg/dal/badger"
	"github.com/go-sigma/sigma/pkg/inits"
	"github.com/go-sigma/sigma/pkg/logger"
	"github.com/go-sigma/sigma/pkg/server/validators"
	"github.com/go-sigma/sigma/pkg/tests"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils/password"
)

const (
	privateKeyString = "LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlDWFFJQkFBS0JnUUN2bmwyeU1hRmR0NTJFOFhIN2tFdkVIbnBtelpWbFBTOWFrZTJ5TmQrNm13VXBlaVQ5CnVqVkZwTmJ2RkFna002TUd3dll5N1hkV1FwNTBaOXVVS0d1UlJEZSt4QXQvbklObVZCcVJwU3VnYzhPOVdMNzQKU294UldJSjFVcWJ3NnYvaFU3K1dSMFlORU1ubVlodzJDNXZPQ3c3UlIrQnJET2h5aEtuKzJ3MWRDUUlEQVFBQgpBb0dBSGtjY2VsTnFNY0V0YkRWQVpKSE5Ma1BlOEloelFHQWJJTzlWM3NyQkJ1Z2hMTFI5V2kxWGIrbHFrUStRCkU4Vy9UclFnUkVtQ3NLR050aDROMG01aGxRR3dBS0tsYUhLOWxzYUtPVDBpV0lwYk1HSm1rMWJQZEV5RTRlL1QKcjN2bUMwU0NaZGJOZElkL1FuMzlkY2hZY2I3MGtBaW5kNFlHQXYvNU45UXdSZ0VDUVFEa2JlcnU4bTRRdXhOagpmTysyTUJmL1NoaUtUbHdYZlNXYURvcW9tTE14MG9BeHpwVkU2RzdZMStJd0xYSXd6VEswUXdIUTdDWEl4ZmkvCi9pRyt6T3BCQWtFQXhOQ3ZhSHJhZklpWjVmZVFESlR6T0kzS3B4WDNSWFlaTytDTHlLeHlic0tZQklTSm9Db0YKVkw4K0diRGZJMU9adm5lTXZEcEE3WFhEQkt3TXFHMXd5UUpCQU9BMGRzUWpWUjY4ejdIMW5iNmZnOTVCbHNhaApWTWlGUUJQdXMrLzVPT0RzOElCeWVKWlM0UUdiRzFvWU1SMXZPcFl0c3FtaUx3L2FLR1loaEhPbTQwRUNRRWhLCmZxTlp2TGJSVmZYcUlMYitYdmYrM05qU2NLaks0Q25tS0hIbEpZTVpaczBDQWFzYXhDcUV0RUtyZk1wMUFwdTcKUGE1RmwyT2hSYWlKcVh5VDlrRUNRUUNYdXlrdWR3eXdudEhHL3d2SmVoeWFSYkxGczd5UG1SbUVEL0FHcEY0QgpKcFZrZFJNQVJpa1g1OE84OWF6WXQyT3pkTGNlTWQ3WWlJRGd4UVhBSEcyagotLS0tLUVORCBSU0EgUFJJVkFURSBLRVktLS0tLQo="
)

func TestInitializeSkipAuth(t *testing.T) {
	logger.SetLevel("debug")

	digCon := dig.New()
	require.NoError(t, digCon.Provide(tests.NewEcho))
	require.NoError(t, validators.Initialize(digCon))

	badgerDir, err := os.MkdirTemp("", "badger")
	require.NoError(t, err)

	err = digCon.Provide(func() configs.Configuration {
		return configs.Configuration{
			Auth: configs.ConfigurationAuth{
				Admin: configs.ConfigurationAuthAdmin{
					Username: "sigma",
					Password: "sigma",
					Email:    "sigma@gmail.com",
				},
				Jwt: configs.ConfigurationAuthJwt{
					PrivateKey: privateKeyString,
				},
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
			Locker: configs.ConfigurationLocker{
				Type:   enums.LockerTypeBadger,
				Badger: configs.ConfigurationLockerBadger{},
				Prefix: "sigma-locker",
			},
			Badger: configs.ConfigurationBadger{
				Enabled: true,
				Path:    badgerDir,
			},
		}
	})
	require.NoError(t, err)

	err = digCon.Provide(func() password.Service {
		return password.New()
	})
	require.NoError(t, err)

	require.NoError(t, digCon.Provide(badger.New))

	tests, err := tests.Initialize(t, digCon)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, dal.DeInitialize())
		require.NoError(t, tests.DeInitialize())
	}()

	require.NoError(t, inits.Initialize(digCon))

	require.NoError(t, Initialize(digCon))
}

type factoryOk struct{}

func (f *factoryOk) Initialize(*dig.Container) error {
	return nil
}

func TestInitializeOK(t *testing.T) {
	routerFactories = make(map[string]Factory)
	require.NoError(t, RegisterRouterFactory("ok", &factoryOk{}))
	digCon := dig.New()
	require.NoError(t, digCon.Provide(tests.NewEcho))
	require.NoError(t, Initialize(digCon))
}

type factoryErr struct{}

func (f *factoryErr) Initialize(*dig.Container) error {
	return errors.New("error")
}

func TestInitializeErr(t *testing.T) {
	routerFactories = make(map[string]Factory)
	require.NoError(t, RegisterRouterFactory("err", &factoryErr{}))
	digCon := dig.New()
	require.NoError(t, digCon.Provide(tests.NewEcho))
	require.Error(t, Initialize(digCon))
}

func TestInitializeDup(t *testing.T) {
	routerFactories = make(map[string]Factory)
	require.NoError(t, RegisterRouterFactory("err", &factoryErr{}))
	require.Error(t, RegisterRouterFactory("err", &factoryErr{}))
}
