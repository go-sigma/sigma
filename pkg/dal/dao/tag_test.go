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

func TestTagServiceFactory(t *testing.T) {
	f := NewTagServiceFactory()
	tagService := f.New()
	assert.NotNil(t, tagService)
	tagService = f.New(query.Q)
	assert.NotNil(t, tagService)
}

func TestTagService(t *testing.T) {
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
	userServiceFactory := NewUserServiceFactory()

	err = query.Q.Transaction(func(tx *query.Query) error {
		userService := userServiceFactory.New(tx)
		userObj := &models.User{Username: "tag-service", Password: "test", Email: "test@gmail.com", Role: "admin"}
		err = userService.Create(ctx, userObj)
		assert.NoError(t, err)

		namespaceService := namespaceServiceFactory.New(tx)
		namespaceObj := &models.Namespace{Name: "test", UserID: userObj.ID, Visibility: ptr.Of(enums.VisibilityPrivate)}
		err = namespaceService.Create(ctx, namespaceObj)
		assert.NoError(t, err)

		repositoryService := repositoryServiceFactory.New(tx)
		repositoryObj := &models.Repository{Name: "test/busybox", NamespaceID: namespaceObj.ID, Visibility: ptr.Of(enums.VisibilityPrivate)}
		err = repositoryService.Create(ctx, repositoryObj)
		assert.NoError(t, err)

		tagService := tagServiceFactory.New(tx)
		tagObj := &models.Tag{
			RepositoryID: repositoryObj.ID,
			Name:         "latest",
			Artifact: &models.Artifact{
				RepositoryID: repositoryObj.ID,
				Digest:       "sha256:xxx",
				Size:         123,
				ContentType:  "test",
				Raw:          []byte("test"),
			},
		}
		err = tagService.Create(ctx, tagObj)
		assert.NoError(t, err)

		tag1, err := tagService.GetByID(ctx, tagObj.ID)
		assert.NoError(t, err)
		assert.Equal(t, tag1.ID, tagObj.ID)

		tag2, err := tagService.GetByName(ctx, repositoryObj.ID, "latest")
		assert.NoError(t, err)
		assert.Equal(t, tag2.ID, tagObj.ID)

		err = tagService.Incr(ctx, tagObj.ID)
		assert.NoError(t, err)
		tag3, err := tagService.GetByID(ctx, tagObj.ID)
		assert.NoError(t, err)
		assert.Equal(t, tag3.PullTimes, int64(1))
		assert.True(t, tag3.LastPull.Valid)

		tags1, err := tagService.ListTag(ctx, types.ListTagRequest{
			Pagination: types.Pagination{
				PageSize: 100,
				PageNum:  1,
			},
			Repository: "test/busybox",
		})
		assert.NoError(t, err)
		assert.Equal(t, len(tags1), int(1))

		count1, err := tagService.CountTag(ctx, types.ListTagRequest{
			Pagination: types.Pagination{
				PageSize: 100,
				PageNum:  1,
			},
			Repository: "test/busybox",
		})
		assert.NoError(t, err)
		assert.Equal(t, count1, int64(1))

		err = tagService.DeleteByName(ctx, repositoryObj.ID, "latest")
		assert.NoError(t, err)

		artifactObj := &models.Artifact{
			RepositoryID: repositoryObj.ID,
			Digest:       "sha256:xxxxx",
			Size:         123,
			ContentType:  "test",
			Raw:          []byte("test"),
		}
		err = tx.Artifact.WithContext(ctx).Create(artifactObj)
		assert.NoError(t, err)

		tagObj1 := &models.Tag{
			RepositoryID: repositoryObj.ID,
			Name:         "latest1",
			Artifact:     artifactObj,
		}
		err = tagService.Create(ctx, tagObj1)
		assert.NoError(t, err)

		err = tagService.DeleteByID(ctx, tagObj1.ID)
		assert.NoError(t, err)

		err = tagService.DeleteByID(ctx, 10)
		assert.ErrorIs(t, err, gorm.ErrRecordNotFound)

		artifactObj2 := &models.Artifact{
			RepositoryID: repositoryObj.ID,
			Digest:       "sha256:xxxxxxxx",
			Size:         123,
			ContentType:  "test",
			Raw:          []byte("test"),
		}
		err = tx.Artifact.WithContext(ctx).Create(artifactObj2)
		assert.NoError(t, err)
		tagObj2 := &models.Tag{
			RepositoryID: repositoryObj.ID,
			Name:         "latest1",
			Artifact:     artifactObj2,
		}
		err = tagService.Create(ctx, tagObj2)
		assert.NoError(t, err)

		tags2, err := tagService.ListByDtPagination(ctx, "test/busybox", 10, 1)
		assert.NoError(t, err)
		assert.Equal(t, len(tags2), int(1))

		tagCount1, err := tagService.CountByArtifact(ctx, []int64{tagObj2.ArtifactID})
		assert.NoError(t, err)
		assert.Equal(t, len(tagCount1), int(1))
		assert.Equal(t, tagCount1[tagObj2.ArtifactID], int64(1))

		return nil
	})
	assert.NoError(t, err)
}
