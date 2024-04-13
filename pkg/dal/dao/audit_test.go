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

	"github.com/go-sigma/sigma/pkg/dal"
	"github.com/go-sigma/sigma/pkg/dal/dao"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/logger"
	"github.com/go-sigma/sigma/pkg/tests"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

func TestAuditServiceFactory(t *testing.T) {
	f := dao.NewAuditServiceFactory()
	assert.NotNil(t, f.New())
	assert.NotNil(t, f.New(query.Q))
}

func TestAuditService(t *testing.T) {
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

	auditServiceFactory := dao.NewAuditServiceFactory()
	auditService := auditServiceFactory.New()
	assert.NotNil(t, auditService)

	namespaceServiceFactory := dao.NewNamespaceServiceFactory()
	namespaceService := namespaceServiceFactory.New()
	assert.NotNil(t, namespaceService)

	namespaceObj1 := &models.Namespace{Name: "test"}
	err := namespaceService.Create(ctx, namespaceObj1)
	assert.NoError(t, err)

	userServiceFactory := dao.NewUserServiceFactory()
	userService := userServiceFactory.New()
	assert.NotNil(t, userService)

	userObj := &models.User{Username: "test-case", Password: ptr.Of("test-case"), Email: ptr.Of("email")}
	err = userService.Create(ctx, userObj)
	assert.NoError(t, err)

	err = auditService.Create(ctx, &models.Audit{
		UserID:       userObj.ID,
		NamespaceID:  ptr.Of(namespaceObj1.ID),
		Action:       enums.AuditActionCreate,
		ResourceType: enums.AuditResourceTypeNamespace,
		Resource:     namespaceObj1.Name,
	})
	assert.NoError(t, err)

	hotNamespaceObjs, err := auditService.HotNamespace(ctx, userObj.ID, 3)
	assert.NoError(t, err)
	assert.Equal(t, len(hotNamespaceObjs), 1)
	assert.Equal(t, hotNamespaceObjs[0].Name, namespaceObj1.Name)
}
