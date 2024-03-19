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
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

func TestRepositoryServiceFactory(t *testing.T) {
	f := dao.NewRepositoryServiceFactory()
	repositoryService := f.New()
	assert.NotNil(t, repositoryService)
	repositoryService = f.New(query.Q)
	assert.NotNil(t, repositoryService)
}

func TestRepositoryService(t *testing.T) {
	viper.SetDefault("log.level", "debug")
	logger.SetLevel("debug")
	err := tests.Initialize(t)
	assert.NoError(t, err)
	err = tests.DB.Init()
	assert.NoError(t, err)
	defer func() {
		conn, err := dal.DB.DB()
		assert.NoError(t, err)
		assert.NoError(t, conn.Close())
		assert.NoError(t, tests.DB.DeInit())
	}()

	ctx := log.Logger.WithContext(context.Background())

	namespaceServiceFactory := dao.NewNamespaceServiceFactory()
	repositoryServiceFactory := dao.NewRepositoryServiceFactory()
	userServiceFactory := dao.NewUserServiceFactory()

	err = query.Q.Transaction(func(tx *query.Query) error {
		userService := userServiceFactory.New(tx)
		userObj := &models.User{Username: "repository-service", Password: ptr.Of("test"), Email: ptr.Of("test@gmail.com")}
		err = userService.Create(ctx, userObj)
		assert.NoError(t, err)

		namespaceService := namespaceServiceFactory.New(tx)
		namespaceObj := &models.Namespace{Name: "test", Visibility: enums.VisibilityPrivate}
		err = namespaceService.Create(ctx, namespaceObj)
		assert.NoError(t, err)

		repositoryService := repositoryServiceFactory.New(tx)
		repositoryObj := &models.Repository{Name: "test/busybox", NamespaceID: namespaceObj.ID}
		err = repositoryService.Create(ctx, repositoryObj, dao.AutoCreateNamespace{UserID: userObj.ID})
		assert.NoError(t, err)

		namespaceObj1 := &models.Namespace{Name: "test1", Visibility: enums.VisibilityPrivate}
		err = namespaceService.Create(ctx, namespaceObj1)
		assert.NoError(t, err)
		err = repositoryService.Create(ctx, &models.Repository{Name: "test1/busybox"}, dao.AutoCreateNamespace{UserID: userObj.ID})
		assert.NoError(t, err)

		count1, err := repositoryService.CountRepository(ctx, namespaceObj.ID, nil)
		assert.NoError(t, err)
		assert.Equal(t, count1, int64(1))

		repository1, err := repositoryService.Get(ctx, repositoryObj.ID)
		assert.NoError(t, err)
		assert.Equal(t, repositoryObj.ID, repository1.ID)

		repository2, err := repositoryService.GetByName(ctx, "test/busybox")
		assert.NoError(t, err)
		assert.Equal(t, repositoryObj.ID, repository2.ID)

		repositories1, count, err := repositoryService.ListRepository(ctx, namespaceObj.ID, nil, types.Pagination{
			Limit: ptr.Of(int(100)),
			Page:  ptr.Of(int(1)),
		}, types.Sortable{Sort: ptr.Of("created_at"), Method: ptr.Of(enums.SortMethodAsc)})
		assert.NoError(t, err)
		assert.Equal(t, int64(len(repositories1)), count)

		repositories2, err := repositoryService.ListByDtPagination(ctx, 100, 1)
		assert.NoError(t, err)
		assert.Equal(t, len(repositories2), int(1))

		err = repositoryService.UpdateRepository(ctx, repository1.ID, map[string]any{"description": ptr.Of("test"), "overview": []byte("test")})
		assert.NoError(t, err)

		err = repositoryService.DeleteByID(ctx, repositoryObj.ID)
		assert.NoError(t, err)

		return nil
	})
	assert.NoError(t, err)
}
