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
	"github.com/stretchr/testify/require"

	"github.com/go-sigma/sigma/pkg/dal/dao"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/logger"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

func TestUserServiceFactory(t *testing.T) {
	f := dao.NewUserServiceFactory()
	require.NotNil(t, f.New())
	require.NotNil(t, f.New(query.Q))
}

func TestUserService(t *testing.T) {
	logger.SetLevel("debug")

	digCon := initDal(t)
	require.NotNil(t, digCon)

	ctx := log.Logger.WithContext(context.Background())

	userSvc := dao.NewUserServiceFactory().New()

	require.NoError(t, userSvc.Create(ctx, &models.User{Username: "test-case", Password: ptr.Of("test-case"), Email: ptr.Of("email")}))

	testUser, err := userSvc.GetByUsername(ctx, "test-case")
	require.NoError(t, err)
	require.Equal(t, ptr.To(testUser.Password), "test-case")
	total, err := userSvc.Count(ctx)
	require.NoError(t, err)
	require.Equal(t, total, int64(1))
}

// func TestUserGetByUsername(t *testing.T) {
// 	logger.SetLevel("debug")
// 	assert.NoError(t, tests.Initialize(t))
// 	assert.NoError(t, tests.DB.Init())
// 	defer func() {
// 		conn, err := dal.DB.DB()
// 		assert.NoError(t, err)
// 		assert.NoError(t, conn.Close())
// 		assert.NoError(t, tests.DB.DeInit())
// 	}()

// 	ctx := log.Logger.WithContext(context.Background())

// 	userService := dao.NewUserServiceFactory().New()
// 	assert.NotNil(t, userService)
// 	assert.NoError(t, userService.Create(ctx, &models.User{Username: "test-case", Password: ptr.Of("test-case"), Email: ptr.Of("email")}))

// 	testUser, err := userService.GetByUsername(ctx, "test-case")
// 	assert.NoError(t, err)
// 	assert.Equal(t, ptr.To(testUser.Password), "test-case")
// 	total, err := userService.Count(ctx)
// 	assert.NoError(t, err)
// 	assert.Equal(t, total, int64(1))
// }
