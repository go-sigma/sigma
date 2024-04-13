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
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

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

func TestNamespaceServiceFactory(t *testing.T) {
	f := dao.NewNamespaceServiceFactory()
	assert.NotNil(t, f.New())
	assert.NotNil(t, f.New(query.Q))
}

func TestNamespaceService(t *testing.T) {
	logger.SetLevel("debug")
	assert.NoError(t, tests.Initialize(t))
	assert.NoError(t, tests.DB.Init())
	defer func() {
		conn, err := dal.DB.DB()
		assert.NoError(t, err)
		assert.NoError(t, conn.Close())
		assert.NoError(t, tests.DB.DeInit())
	}()

	ctx := log.Logger.WithContext(context.Background())

	namespaceServiceFactory := dao.NewNamespaceServiceFactory()
	namespaceService := namespaceServiceFactory.New()

	userServiceFactory := dao.NewUserServiceFactory()
	userService := userServiceFactory.New()
	userObj := &models.User{Username: "namespace-service", Password: ptr.Of("test"), Email: ptr.Of("test@gmail.com")}
	assert.NoError(t, userService.Create(ctx, userObj))

	namespaceObj := &models.Namespace{
		Name:       "test",
		Visibility: enums.VisibilityPrivate,
	}
	assert.NoError(t, namespaceService.Create(ctx, namespaceObj))

	ns, err := namespaceService.Get(ctx, namespaceObj.ID)
	assert.NoError(t, err)
	assert.Equal(t, ns.ID, namespaceObj.ID)
	assert.Equal(t, ns.Name, namespaceObj.Name)

	ns1, err := namespaceService.GetByName(ctx, "test")
	assert.NoError(t, err)
	assert.Equal(t, ns1.ID, namespaceObj.ID)
	assert.Equal(t, ns1.Name, namespaceObj.Name)

	namespaceList, _, err := namespaceService.ListNamespace(ctx, ptr.Of("t"), types.Pagination{
		Limit: ptr.Of(int(100)),
		Page:  ptr.Of(int(0)),
	}, types.Sortable{})
	assert.NoError(t, err)
	assert.Equal(t, len(namespaceList), int(1))

	namespaceList, _, err = namespaceService.ListNamespace(ctx, ptr.Of("t"), types.Pagination{
		Limit: ptr.Of(int(100)),
		Page:  ptr.Of(int(0)),
	}, types.Sortable{
		Sort:   ptr.Of("created_at"),
		Method: ptr.Of(enums.SortMethodDesc),
	})
	assert.NoError(t, err)
	assert.Equal(t, len(namespaceList), int(1))

	count, err := namespaceService.CountNamespace(ctx, ptr.Of("t"))
	assert.NoError(t, err)
	assert.Equal(t, count, int64(1))

	assert.NoError(t, namespaceService.UpdateByID(ctx, namespaceObj.ID, map[string]interface{}{query.Namespace.Description.ColumnName().String(): "test"}))

	assert.NoError(t, namespaceService.DeleteByID(ctx, namespaceObj.ID))

	assert.ErrorIs(t, namespaceService.DeleteByID(ctx, 10), gorm.ErrRecordNotFound)
}

func TestNamespaceServiceQuota(t *testing.T) {
	logger.SetLevel("debug")
	assert.NoError(t, tests.Initialize(t))
	assert.NoError(t, tests.DB.Init())
	defer func() {
		conn, err := dal.DB.DB()
		assert.NoError(t, err)
		assert.NoError(t, conn.Close())
		assert.NoError(t, tests.DB.DeInit())
	}()

	ctx := log.Logger.WithContext(context.Background())

	namespaceServiceFactory := dao.NewNamespaceServiceFactory()
	userServiceFactory := dao.NewUserServiceFactory()

	userService := userServiceFactory.New()
	userObj := &models.User{Username: "artifact-service", Password: ptr.Of("test"), Email: ptr.Of("test@gmail.com")}
	assert.NoError(t, userService.Create(ctx, userObj))

	namespaceService := namespaceServiceFactory.New()

	namespaceObj := &models.Namespace{
		Name: "test",
	}
	assert.NoError(t, namespaceService.Create(ctx, namespaceObj))

	assert.NoError(t, namespaceService.UpdateQuota(ctx, namespaceObj.ID, 100))

	assert.ErrorIs(t, namespaceService.UpdateQuota(ctx, 10, 100), gorm.ErrRecordNotFound)
}
