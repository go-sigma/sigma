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
	"unicode"

	"github.com/stretchr/testify/require"
	pwdvalidate "github.com/wagslane/go-password-validator"
	"go.uber.org/dig"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	repouser "github.com/go-sigma/sigma/pkg/dal/repository/user"
	repousermocks "github.com/go-sigma/sigma/pkg/dal/repository/user/mocks"
	"github.com/go-sigma/sigma/pkg/logger"
	"github.com/go-sigma/sigma/pkg/service/password"
	"github.com/go-sigma/sigma/pkg/testkit"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

func TestGenerateAdminPassword(t *testing.T) {
	for range 20 {
		pwd, err := generateAdminPassword()
		require.NoError(t, err)
		require.Len(t, pwd, adminPasswordLength)

		var upper, lower, digit int
		for _, r := range pwd {
			switch {
			case unicode.IsUpper(r):
				upper++
			case unicode.IsLower(r):
				lower++
			case unicode.IsDigit(r):
				digit++
			default:
				require.Failf(t, "unexpected character", "password %q contains %q", pwd, r)
			}
		}
		require.NotZero(t, upper, pwd)
		require.NotZero(t, lower, pwd)
		require.Equal(t, adminPasswordDigits, digit, pwd)
		require.NoError(t, pwdvalidate.Validate(pwd, consts.PwdStrength), pwd)
	}
}

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

				ctx := t.Context()
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
			name: "admin_generated_password",
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
			afterCheck: func(t *testing.T, c *dig.Container) {
				var userRepository repouser.UserRepository
				require.NoError(t, c.Invoke(func(f repouser.UserRepository) { userRepository = f }))

				ctx := t.Context()
				count, err := userRepository.Count(ctx)
				require.NoError(t, err)
				require.Equal(t, int64(3), count)

				user, err := userRepository.GetByUsername(ctx, "sigma")
				require.NoError(t, err)
				require.NotNil(t, user.Password)
				require.NotEmpty(t, *user.Password)
			},
			wantErr: nil,
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

func TestEnsureUserSkipsExistingUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	mock := repousermocks.NewMockUserRepository(ctrl)
	user := &models.User{ID: "user-1", Username: "sigma"}
	mock.EXPECT().GetByUsername(gomock.Any(), "sigma").Return(user, nil)

	created, err := ensureUser(t.Context(), mock, user)
	require.NoError(t, err)
	require.False(t, created)
}

func TestEnsureUserCreatesMissingUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	mock := repousermocks.NewMockUserRepository(ctrl)
	user := &models.User{ID: "user-1", Username: "sigma"}
	gomock.InOrder(
		mock.EXPECT().GetByUsername(gomock.Any(), "sigma").Return(nil, gorm.ErrRecordNotFound),
		mock.EXPECT().Create(gomock.Any(), user).Return(nil),
	)

	created, err := ensureUser(t.Context(), mock, user)
	require.NoError(t, err)
	require.True(t, created)
}

func TestEnsureUserToleratesConcurrentCreate(t *testing.T) {
	ctrl := gomock.NewController(t)
	mock := repousermocks.NewMockUserRepository(ctrl)
	user := &models.User{ID: "user-1", Username: "sigma"}
	gomock.InOrder(
		mock.EXPECT().GetByUsername(gomock.Any(), "sigma").Return(nil, gorm.ErrRecordNotFound),
		mock.EXPECT().Create(gomock.Any(), user).
			Return(errors.New("duplicate key value violates unique constraint")),
		mock.EXPECT().GetByUsername(gomock.Any(), "sigma").Return(user, nil),
	)

	created, err := ensureUser(t.Context(), mock, user)
	require.NoError(t, err)
	require.False(t, created)
}

func TestEnsureUserReturnsCreateError(t *testing.T) {
	ctrl := gomock.NewController(t)
	mock := repousermocks.NewMockUserRepository(ctrl)
	user := &models.User{ID: "user-1", Username: "sigma"}
	createErr := errors.New("connection reset by peer")
	gomock.InOrder(
		mock.EXPECT().GetByUsername(gomock.Any(), "sigma").Return(nil, gorm.ErrRecordNotFound),
		mock.EXPECT().Create(gomock.Any(), user).Return(createErr),
		mock.EXPECT().GetByUsername(gomock.Any(), "sigma").Return(nil, gorm.ErrRecordNotFound),
	)

	created, err := ensureUser(t.Context(), mock, user)
	require.ErrorIs(t, err, createErr)
	require.False(t, created)
}

func TestEnsureUserReturnsLookupError(t *testing.T) {
	ctrl := gomock.NewController(t)
	mock := repousermocks.NewMockUserRepository(ctrl)
	user := &models.User{ID: "user-1", Username: "sigma"}
	lookupErr := errors.New("database unavailable")
	mock.EXPECT().GetByUsername(gomock.Any(), "sigma").Return(nil, lookupErr)

	created, err := ensureUser(t.Context(), mock, user)
	require.ErrorIs(t, err, lookupErr)
	require.False(t, created)
}
