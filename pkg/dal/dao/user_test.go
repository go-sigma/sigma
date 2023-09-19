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

package dao_test

import (
	"context"
	"testing"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"github.com/go-sigma/sigma/pkg/dal"
	"github.com/go-sigma/sigma/pkg/dal/dao"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/logger"
	"github.com/go-sigma/sigma/pkg/tests"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

func TestUserServiceFactory(t *testing.T) {
	f := dao.NewUserServiceFactory()
	userService := f.New()
	assert.NotNil(t, userService)
	userService = f.New(query.Q)
	assert.NotNil(t, userService)
}

func TestUserGetByUsername(t *testing.T) {
	viper.SetDefault("log.level", "debug")
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

	f := dao.NewUserServiceFactory()

	ctx := log.Logger.WithContext(context.Background())

	err = query.Q.Transaction(func(tx *query.Query) error {
		userService := f.New(tx)
		assert.NotNil(t, userService)
		err := userService.Create(ctx, &models.User{Username: "test-case", Password: ptr.Of("test-case"), Email: ptr.Of("email")})
		assert.NoError(t, err)
		testUser, err := userService.GetByUsername(ctx, "test-case")
		assert.NoError(t, err)
		assert.Equal(t, ptr.To(testUser.Password), "test-case")
		total, err := userService.Count(ctx)
		assert.NoError(t, err)
		assert.Equal(t, total, int64(1))
		return nil
	})
	assert.NoError(t, err)
}
