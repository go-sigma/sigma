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
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal"
	"github.com/go-sigma/sigma/pkg/dal/dao"
	"github.com/go-sigma/sigma/pkg/logger"
	"github.com/go-sigma/sigma/pkg/tests"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/password"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

func TestInitUser(t *testing.T) {
	logger.SetLevel("debug")

	tests := []struct {
		name       string
		genDigCon  func(*testing.T) *dig.Container
		afterCheck func(*testing.T, *dig.Container)
		afterEach  func(*testing.T, *dig.Container)
		wantErr    error
	}{
		{
			name: "normal",
			genDigCon: func(t *testing.T) *dig.Container {
				digCon := dig.New()

				err := digCon.Provide(func() configs.Configuration {
					return configs.Configuration{
						Auth: configs.ConfigurationAuth{
							Admin: configs.ConfigurationAuthAdmin{
								Username: "sigma",
								Password: "sigma",
								Email:    "sigma@gmail.com",
							},
						},
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

				err = digCon.Provide(func() password.Service {
					return password.New()
				})
				require.NoError(t, err)

				testInstance, err := tests.Initialize(t, digCon)
				require.NoError(t, err)

				err = digCon.Provide(func() *tests.Instance {
					return testInstance
				})
				require.NoError(t, err)

				return digCon
			},
			afterCheck: func(t *testing.T, c *dig.Container) {
				userService := utils.MustGetObjFromDigCon[dao.UserServiceFactory](c).New()
				passwordService := utils.MustGetObjFromDigCon[password.Service](c)

				ctx := log.Logger.WithContext(context.Background())
				count, err := userService.Count(ctx)
				require.NoError(t, err)
				require.Equal(t, count, int64(3))

				user, err := userService.GetByUsername(ctx, "sigma")
				require.NoError(t, err)
				require.NotNil(t, user)
				require.True(t, passwordService.Verify("sigma", ptr.To(user.Password)))

				user, err = userService.GetByUsername(ctx, consts.UserInternal)
				require.NoError(t, err)
				require.NotNil(t, user)
			},
			afterEach: func(t *testing.T, c *dig.Container) {
				require.NoError(t, dal.DeInitialize())
				require.NoError(t, utils.MustGetObjFromDigCon[*tests.Instance](c).DeInitialize())
			},
			wantErr: nil,
		},
		{
			name: "admin_no_password",
			genDigCon: func(t *testing.T) *dig.Container {
				digCon := dig.New()

				err := digCon.Provide(func() configs.Configuration {
					return configs.Configuration{
						Auth: configs.ConfigurationAuth{
							Admin: configs.ConfigurationAuthAdmin{
								Username: "sigma",
								Email:    "sigma@gmail.com",
							},
						},
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

				err = digCon.Provide(func() password.Service {
					return password.New()
				})
				require.NoError(t, err)

				testInstance, err := tests.Initialize(t, digCon)
				require.NoError(t, err)

				err = digCon.Provide(func() *tests.Instance {
					return testInstance
				})
				require.NoError(t, err)

				return digCon
			},
			afterCheck: nil,
			afterEach: func(t *testing.T, c *dig.Container) {
				require.NoError(t, dal.DeInitialize())
				require.NoError(t, utils.MustGetObjFromDigCon[*tests.Instance](c).DeInitialize())
			},
			wantErr: ErrAdminPassword,
		},
		{
			name: "admin_no_username",
			genDigCon: func(t *testing.T) *dig.Container {
				digCon := dig.New()

				err := digCon.Provide(func() configs.Configuration {
					return configs.Configuration{
						Auth: configs.ConfigurationAuth{
							Admin: configs.ConfigurationAuthAdmin{
								Password: "sigma",
								Email:    "sigma@gmail.com",
							},
						},
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

				err = digCon.Provide(func() password.Service {
					return password.New()
				})
				require.NoError(t, err)

				testInstance, err := tests.Initialize(t, digCon)
				require.NoError(t, err)

				err = digCon.Provide(func() *tests.Instance {
					return testInstance
				})
				require.NoError(t, err)

				return digCon
			},
			afterCheck: nil,
			afterEach: func(t *testing.T, c *dig.Container) {
				require.NoError(t, dal.DeInitialize())
				require.NoError(t, utils.MustGetObjFromDigCon[*tests.Instance](c).DeInitialize())
			},
			wantErr: ErrAdminUsername,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			digCon := tt.genDigCon(t)
			defer tt.afterEach(t, digCon)
			err := initUser(digCon)
			if tt.wantErr == nil {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tt.wantErr.Error())
			}
			if tt.afterCheck != nil {
				tt.afterCheck(t, digCon)
			}
		})
	}
}
