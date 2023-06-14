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

	"github.com/ximager/ximager/pkg/dal"
	"github.com/ximager/ximager/pkg/dal/models"
	"github.com/ximager/ximager/pkg/dal/query"
	"github.com/ximager/ximager/pkg/logger"
	"github.com/ximager/ximager/pkg/tests"
)

func TestReferenceServiceFactory(t *testing.T) {
	f := NewReferenceServiceFactory()
	referenceService := f.New()
	assert.NotNil(t, referenceService)
	referenceService = f.New(query.Q)
	assert.NotNil(t, referenceService)
}

func TestReferenceService(t *testing.T) {
	viper.SetDefault("log.level", "debug")
	logger.SetLevel("debug")
	err := tests.Initialize()
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

	tagServiceFactory := NewTagServiceFactory()
	namespaceServiceFactory := NewNamespaceServiceFactory()
	repositoryServiceFactory := NewRepositoryServiceFactory()
	referenceServiceFactory := NewReferenceServiceFactory()
	userServiceFactory := NewUserServiceFactory()

	err = query.Q.Transaction(func(tx *query.Query) error {
		userService := userServiceFactory.New(tx)
		userObj := &models.User{Username: "reference-service", Password: "test", Email: "test@gmail.com", Role: "admin"}
		err = userService.Create(ctx, userObj)
		assert.NoError(t, err)

		namespaceService := namespaceServiceFactory.New(tx)
		namespaceObj := &models.Namespace{Name: "test", UserID: userObj.ID}
		err = namespaceService.Create(ctx, namespaceObj)
		assert.NoError(t, err)

		repositoryService := repositoryServiceFactory.New(tx)
		repositoryObj := &models.Repository{Name: "test/busybox", NamespaceID: namespaceObj.ID}
		err = repositoryService.Create(ctx, repositoryObj)
		assert.NoError(t, err)

		tagService := tagServiceFactory.New(tx)
		tagObj := &models.Tag{
			RepositoryID: repositoryObj.ID,
			Name:         "latest",
			Artifact: &models.Artifact{
				RepositoryID: repositoryObj.ID,
				Digest:       "sha256:f7d81d5be30e617068bf53a9b136400b13d91c0f54d097a72bf91127f43d0157",
				Size:         123,
				ContentType:  "test",
				Raw:          "test",
			},
		}
		tagObj1, err := tagService.Save(ctx, tagObj)
		assert.NoError(t, err)
		assert.Equal(t, tagObj1.Name, tagObj.Name)

		referenceService := referenceServiceFactory.New(tx)
		referenceObj1, err := referenceService.Get(ctx, "test/busybox", "latest")
		assert.NoError(t, err)
		assert.Equal(t, referenceObj1.Artifact.Digest, "sha256:f7d81d5be30e617068bf53a9b136400b13d91c0f54d097a72bf91127f43d0157")
		assert.Equal(t, referenceObj1.Artifact.Size, uint64(123))

		referenceObj2, err := referenceService.Get(ctx, "test/busybox", "sha256:f7d81d5be30e617068bf53a9b136400b13d91c0f54d097a72bf91127f43d0157")
		assert.NoError(t, err)
		assert.Equal(t, referenceObj2.Artifact.Digest, "sha256:f7d81d5be30e617068bf53a9b136400b13d91c0f54d097a72bf91127f43d0157")
		assert.Equal(t, referenceObj2.Artifact.Size, uint64(123))

		return nil
	})
}
