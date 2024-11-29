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
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/go-sigma/sigma/pkg/dal/dao"
	"github.com/go-sigma/sigma/pkg/dal/query"
)

func TestRepositoryServiceFactory(t *testing.T) {
	f := dao.NewRepositoryServiceFactory()
	assert.NotNil(t, f.New())
	assert.NotNil(t, f.New(query.Q))
}

// func TestRepositoryService(t *testing.T) {
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

// 	namespaceServiceFactory := dao.NewNamespaceServiceFactory()
// 	repositoryServiceFactory := dao.NewRepositoryServiceFactory()
// 	userServiceFactory := dao.NewUserServiceFactory()

// 	userService := userServiceFactory.New()
// 	userObj := &models.User{Username: "repository-service", Password: ptr.Of("test"), Email: ptr.Of("test@gmail.com")}
// 	assert.NoError(t, userService.Create(ctx, userObj))

// 	namespaceService := namespaceServiceFactory.New()
// 	namespaceObj := &models.Namespace{Name: "test", Visibility: enums.VisibilityPrivate}
// 	assert.NoError(t, namespaceService.Create(ctx, namespaceObj))

// 	repositoryService := repositoryServiceFactory.New()
// 	repositoryObj := &models.Repository{Name: "test/busybox", NamespaceID: namespaceObj.ID}
// 	assert.NoError(t, repositoryService.Create(ctx, repositoryObj, dao.AutoCreateNamespace{UserID: userObj.ID}))

// 	namespaceObj1 := &models.Namespace{Name: "test1", Visibility: enums.VisibilityPrivate}
// 	assert.NoError(t, namespaceService.Create(ctx, namespaceObj1))
// 	assert.NoError(t, repositoryService.Create(ctx, &models.Repository{Name: "test1/busybox"}, dao.AutoCreateNamespace{UserID: userObj.ID}))

// 	count1, err := repositoryService.CountRepository(ctx, namespaceObj.ID, nil)
// 	assert.NoError(t, err)
// 	assert.Equal(t, count1, int64(1))

// 	repository1, err := repositoryService.Get(ctx, repositoryObj.ID)
// 	assert.NoError(t, err)
// 	assert.Equal(t, repositoryObj.ID, repository1.ID)

// 	repository2, err := repositoryService.GetByName(ctx, "test/busybox")
// 	assert.NoError(t, err)
// 	assert.Equal(t, repositoryObj.ID, repository2.ID)

// 	repositories1, count, err := repositoryService.ListRepository(ctx, namespaceObj.ID, nil, types.Pagination{
// 		Limit: ptr.Of(int(100)),
// 		Page:  ptr.Of(int(1)),
// 	}, types.Sortable{Sort: ptr.Of("created_at"), Method: ptr.Of(enums.SortMethodAsc)})
// 	assert.NoError(t, err)
// 	assert.Equal(t, int64(len(repositories1)), count)

// 	repositories2, err := repositoryService.ListByDtPagination(ctx, 100, 1)
// 	assert.NoError(t, err)
// 	assert.Equal(t, len(repositories2), int(1))

// 	assert.NoError(t, repositoryService.UpdateRepository(ctx, repository1.ID, map[string]any{"description": ptr.Of("test"), "overview": []byte("test")}))

// 	assert.NoError(t, repositoryService.DeleteByID(ctx, repositoryObj.ID))
// }
