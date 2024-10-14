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
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/rs/xid"
	"github.com/stretchr/testify/require"
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/dal"
	"github.com/go-sigma/sigma/pkg/dal/badger"
	"github.com/go-sigma/sigma/pkg/modules/locker"
	"github.com/go-sigma/sigma/pkg/modules/locker/definition"
	"github.com/go-sigma/sigma/pkg/types/enums"
)

func TestInitialize(t *testing.T) {
	tests := []struct {
		name      string
		newDigCon func(*testing.T) *dig.Container
		wantErr   bool
	}{
		{
			name: "normal",
			newDigCon: func(t *testing.T) *dig.Container {
				digCon := dig.New()
				err := digCon.Provide(func() configs.Configuration {
					return configs.Configuration{}
				})
				require.NoError(t, err)
				return digCon
			},
			wantErr: true,
		},
		{
			name: "sqlite3",
			newDigCon: func(t *testing.T) *dig.Container {
				badgerDir, err := os.MkdirTemp("", "badger")
				require.NoError(t, err)

				digCon := dig.New()
				err = digCon.Provide(func() configs.Configuration {
					return configs.Configuration{
						Database: configs.ConfigurationDatabase{
							Type: enums.DatabaseSqlite3,
							Sqlite3: configs.ConfigurationDatabaseSqlite3{
								Path: fmt.Sprintf("%s.db", uuid.Must(uuid.NewV7()).String()),
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

				err = digCon.Provide(badger.New)
				require.NoError(t, err)

				err = digCon.Provide(func() (definition.Locker, error) {
					return locker.Initialize(digCon)
				})
				require.NoError(t, err)

				return digCon
			},
			wantErr: false,
		},
		{
			name: "mysql",
			newDigCon: func(t *testing.T) *dig.Container {
				badgerDir, err := os.MkdirTemp("", "badger")
				require.NoError(t, err)

				db, err := sql.Open("mysql", "root:sigma@tcp(127.0.0.1:3306)/")
				require.NoError(t, err)
				dbname := xid.New().String()
				_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", dbname))
				require.NoError(t, err)

				digCon := dig.New()
				err = digCon.Provide(func() configs.Configuration {
					return configs.Configuration{
						Database: configs.ConfigurationDatabase{
							Type: enums.DatabaseMysql,
							Mysql: configs.ConfigurationDatabaseMysql{
								Host:     "127.0.0.1",
								Port:     3306,
								Username: "root",
								Password: "sigma",
								Database: dbname,
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

				err = digCon.Provide(badger.New)
				require.NoError(t, err)

				err = digCon.Provide(func() (definition.Locker, error) {
					return locker.Initialize(digCon)
				})
				require.NoError(t, err)

				return digCon
			},
			wantErr: false,
		},
		{
			name: "postgresql",
			newDigCon: func(t *testing.T) *dig.Container {
				badgerDir, err := os.MkdirTemp("", "badger")
				require.NoError(t, err)

				ctx := context.Background()
				conn, err := pgx.Connect(ctx, "postgres://sigma:sigma@localhost:5432/?sslmode=disable")
				require.NoError(t, err)

				dbname := xid.New().String()
				_, err = conn.Exec(ctx, fmt.Sprintf("CREATE DATABASE \"%s\"", dbname))
				require.NoError(t, err)
				require.NoError(t, conn.Close(ctx))

				digCon := dig.New()
				err = digCon.Provide(func() configs.Configuration {
					return configs.Configuration{
						Database: configs.ConfigurationDatabase{
							Type: enums.DatabasePostgresql,
							Postgresql: configs.ConfigurationDatabasePostgresql{
								Host:     "localhost",
								Port:     5432,
								Username: "sigma",
								Password: "sigma",
								Database: dbname,
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

				err = digCon.Provide(badger.New)
				require.NoError(t, err)

				err = digCon.Provide(func() (definition.Locker, error) {
					return locker.Initialize(digCon)
				})
				require.NoError(t, err)

				return digCon
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := dal.Initialize(tt.newDigCon(t))
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
