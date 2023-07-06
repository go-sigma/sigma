// Copyright 2023 XImager
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

package dao

import (
	"context"
	"testing"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"github.com/ximager/ximager/pkg/dal"
	"github.com/ximager/ximager/pkg/dal/models"
	"github.com/ximager/ximager/pkg/dal/query"
	"github.com/ximager/ximager/pkg/logger"
	"github.com/ximager/ximager/pkg/tests"
	"github.com/ximager/ximager/pkg/types"
	"github.com/ximager/ximager/pkg/types/enums"
	"github.com/ximager/ximager/pkg/utils/ptr"
)

func TestRepositoryServiceFactory(t *testing.T) {
	f := NewRepositoryServiceFactory()
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
		err = conn.Close()
		assert.NoError(t, err)
		err = tests.DB.DeInit()
		assert.NoError(t, err)
	}()

	ctx := log.Logger.WithContext(context.Background())

	namespaceServiceFactory := NewNamespaceServiceFactory()
	repositoryServiceFactory := NewRepositoryServiceFactory()
	userServiceFactory := NewUserServiceFactory()

	err = query.Q.Transaction(func(tx *query.Query) error {
		userService := userServiceFactory.New(tx)
		userObj := &models.User{Provider: enums.ProviderLocal, Username: "repository-service", Password: ptr.Of("test"), Email: ptr.Of("test@gmail.com")}
		err = userService.Create(ctx, userObj)
		assert.NoError(t, err)

		namespaceService := namespaceServiceFactory.New(tx)
		namespaceObj := &models.Namespace{Name: "test", UserID: userObj.ID, Visibility: enums.VisibilityPrivate}
		err = namespaceService.Create(ctx, namespaceObj)
		assert.NoError(t, err)

		repositoryService := repositoryServiceFactory.New(tx)
		repositoryObj := &models.Repository{Name: "test/busybox", NamespaceID: namespaceObj.ID, Visibility: enums.VisibilityPrivate}
		err = repositoryService.Create(ctx, repositoryObj)
		assert.NoError(t, err)

		namespaceObj1 := &models.Namespace{Name: "test1", UserID: userObj.ID, Visibility: enums.VisibilityPrivate}
		err = namespaceService.Create(ctx, namespaceObj1)
		assert.NoError(t, err)
		err = repositoryService.Create(ctx, &models.Repository{Name: "test1/busybox", Visibility: enums.VisibilityPrivate})
		assert.NoError(t, err)

		count1, err := repositoryService.CountRepository(ctx, types.ListRepositoryRequest{
			Pagination: types.Pagination{
				Limit: ptr.Of(int(100)),
				Last:  ptr.Of(int64(0)),
			},
			Namespace: "test",
		})
		assert.NoError(t, err)
		assert.Equal(t, count1, int64(1))

		repository1, err := repositoryService.Get(ctx, repositoryObj.ID)
		assert.NoError(t, err)
		assert.Equal(t, repositoryObj.ID, repository1.ID)

		repository2, err := repositoryService.GetByName(ctx, "test/busybox")
		assert.NoError(t, err)
		assert.Equal(t, repositoryObj.ID, repository2.ID)

		repositories1, err := repositoryService.ListRepository(ctx, types.ListRepositoryRequest{
			Pagination: types.Pagination{
				Limit: ptr.Of(int(100)),
				Last:  ptr.Of(int64(0)),
			},
			Namespace: "test",
		})
		assert.NoError(t, err)
		assert.Equal(t, len(repositories1), int(1))

		repositories2, err := repositoryService.ListByDtPagination(ctx, 100, 1)
		assert.NoError(t, err)
		assert.Equal(t, len(repositories2), int(1))

		err = repositoryService.UpdateRepository(ctx, repository1.ID, models.Repository{Description: ptr.Of("test"), Overview: []byte("test")})
		assert.NoError(t, err)

		err = repositoryService.UpdateRepository(ctx, 10, models.Repository{Description: ptr.Of("test"), Overview: []byte("test")})
		assert.ErrorIs(t, err, gorm.ErrRecordNotFound)

		err = repositoryService.DeleteByID(ctx, repositoryObj.ID)
		assert.NoError(t, err)

		err = repositoryService.DeleteByID(ctx, repositoryObj.ID)
		assert.ErrorIs(t, err, gorm.ErrRecordNotFound)

		return nil
	})
	assert.NoError(t, err)
}
