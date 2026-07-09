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
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/consts"
	repouser "github.com/go-sigma/sigma/pkg/dal/repository/user"
	"github.com/go-sigma/sigma/pkg/logger"
	"github.com/go-sigma/sigma/pkg/service/password"
	"github.com/go-sigma/sigma/pkg/testkit"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

func TestInitUser(t *testing.T) {
	logger.SetLevel("debug")

	tests := []struct {
		name       string
		genDigCon  func(*testing.T) *dig.Container
		afterCheck func(*testing.T, *dig.Container)
		wantErr    error
	}{
		{
			name: "normal",
			genDigCon: func(t *testing.T) *dig.Container {
				digCon := dig.New()

				err := digCon.Provide(func() *config.Configuration {
					return &config.Configuration{
						Auth: config.ConfigurationAuth{
							Admin: config.ConfigurationAuthAdmin{
								Username: "sigma",
								Password: "sigma",
								Email:    "sigma@gmail.com",
							},
						},
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

				err = digCon.Provide(func() password.Service {
					return password.New()
				})
				require.NoError(t, err)

				require.NoError(t, testkit.InitializeIntegration(t, digCon))

				return digCon
			},
			afterCheck: func(t *testing.T, c *dig.Container) {
				var userRepository repouser.UserRepository
				require.NoError(t, c.Invoke(func(f repouser.UserRepository) { userRepository = f }))
				var passwordSvc password.Service
				require.NoError(t, c.Invoke(func(s password.Service) { passwordSvc = s }))

				ctx := context.Background()
				count, err := userRepository.Count(ctx)
				require.NoError(t, err)
				require.Equal(t, count, int64(3))

				user, err := userRepository.GetByUsername(ctx, "sigma")
				require.NoError(t, err)
				require.NotNil(t, user)
				require.True(t, passwordSvc.Verify("sigma", ptr.To(user.Password)))

				user, err = userRepository.GetByUsername(ctx, consts.UserInternal)
				require.NoError(t, err)
				require.NotNil(t, user)
			},
			wantErr: nil,
		},
		{
			name: "admin_no_password",
			genDigCon: func(t *testing.T) *dig.Container {
				digCon := dig.New()

				err := digCon.Provide(func() *config.Configuration {
					return &config.Configuration{
						Auth: config.ConfigurationAuth{
							Admin: config.ConfigurationAuthAdmin{
								Username: "sigma",
								Email:    "sigma@gmail.com",
							},
						},
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

				err = digCon.Provide(func() password.Service {
					return password.New()
				})
				require.NoError(t, err)

				require.NoError(t, testkit.InitializeIntegration(t, digCon))

				return digCon
			},
			afterCheck: nil,
			wantErr:    ErrAdminPassword,
		},
		{
			name: "admin_no_username",
			genDigCon: func(t *testing.T) *dig.Container {
				digCon := dig.New()

				err := digCon.Provide(func() *config.Configuration {
					return &config.Configuration{
						Auth: config.ConfigurationAuth{
							Admin: config.ConfigurationAuthAdmin{
								Password: "sigma",
								Email:    "sigma@gmail.com",
							},
						},
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

				err = digCon.Provide(func() password.Service {
					return password.New()
				})
				require.NoError(t, err)

				require.NoError(t, testkit.InitializeIntegration(t, digCon))

				return digCon
			},
			afterCheck: nil,
			wantErr:    ErrAdminUsername,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			digCon := tt.genDigCon(t)
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
