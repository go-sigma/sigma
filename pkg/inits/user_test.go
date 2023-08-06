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
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"github.com/go-sigma/sigma/pkg/dal"
	"github.com/go-sigma/sigma/pkg/dal/dao"
	"github.com/go-sigma/sigma/pkg/logger"
	"github.com/go-sigma/sigma/pkg/tests"
	"github.com/go-sigma/sigma/pkg/utils/password"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

func TestInitInternalUser(t *testing.T) {
	logger.SetLevel("debug")
	err := tests.Initialize(t)
	assert.NoError(t, err)
	err = tests.DB.Init()
	assert.NoError(t, err)
	defer func() {
		conn, err := dal.DB.DB()
		assert.NoError(t, err)
		err = conn.Close()
		assert.NoError(t, err)
		err = tests.DB.DeInit()
		assert.NoError(t, err)
	}()

	err = initUser()
	assert.Error(t, err)

	viper.SetDefault("auth.internalUser.password", "internal-sigma")
	err = initUser()
	assert.Error(t, err)

	viper.SetDefault("auth.internalUser.username", "internal-sigma")
	err = initUser()
	assert.Error(t, err)
}

func TestInitAdminUser1(t *testing.T) {
	logger.SetLevel("debug")
	err := tests.Initialize(t)
	assert.NoError(t, err)
	err = tests.DB.Init()
	assert.NoError(t, err)
	defer func() {
		conn, err := dal.DB.DB()
		assert.NoError(t, err)
		err = conn.Close()
		assert.NoError(t, err)
		err = tests.DB.DeInit()
		assert.NoError(t, err)
	}()

	viper.SetDefault("auth.internalUser.password", "internal-sigma")
	viper.SetDefault("auth.internalUser.username", "internal-sigma")

	viper.SetDefault("auth.admin.password", "sigma")
	err = initUser()
	assert.Error(t, err)
}

func TestInitAdminUser2(t *testing.T) {
	logger.SetLevel("debug")
	err := tests.Initialize(t)
	assert.NoError(t, err)
	err = tests.DB.Init()
	assert.NoError(t, err)
	defer func() {
		conn, err := dal.DB.DB()
		assert.NoError(t, err)
		err = conn.Close()
		assert.NoError(t, err)
		err = tests.DB.DeInit()
		assert.NoError(t, err)
	}()

	viper.SetDefault("auth.internalUser.password", "internal-sigma")
	viper.SetDefault("auth.internalUser.username", "internal-sigma")
	viper.SetDefault("auth.admin.password", "sigma")
	viper.SetDefault("auth.admin.username", "sigma")
	viper.SetDefault("auth.admin.email", "sigma@gmail.com")
	err = initUser()
	assert.NoError(t, err)

	userServiceFactory := dao.NewUserServiceFactory()
	userService := userServiceFactory.New()
	passwordService := password.New()

	count, err := userService.Count(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, count, int64(2))

	ctx := context.Background()
	user, err := userService.GetByUsername(ctx, "sigma")
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.True(t, passwordService.Verify("sigma", ptr.To(user.Password)))

	user, err = userService.GetByUsername(ctx, "internal-sigma")
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.True(t, passwordService.Verify("internal-sigma", ptr.To(user.Password)))
}
