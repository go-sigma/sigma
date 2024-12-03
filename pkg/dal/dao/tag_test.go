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

func TestTagServiceFactory(t *testing.T) {
	f := dao.NewTagServiceFactory()
	assert.NotNil(t, f.New())
	assert.NotNil(t, f.New(query.Q))
}

// func TestTagService(t *testing.T) {
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
// 	namespaceService := dao.NewNamespaceServiceFactory().New()
// 	repositoryService := dao.NewRepositoryServiceFactory().New()
// 	artifactService := dao.NewArtifactServiceFactory().New()

// 	userObj := &models.User{Username: "tag-service", Password: ptr.Of("test"), Email: ptr.Of("test@gmail.com")}
// 	assert.NoError(t, userService.Create(ctx, userObj))

// 	namespaceObj := &models.Namespace{Name: "test", Visibility: enums.VisibilityPrivate}
// 	assert.NoError(t, namespaceService.Create(ctx, namespaceObj))

// 	repositoryObj := &models.Repository{Name: "test/busybox", NamespaceID: namespaceObj.ID}
// 	assert.NoError(t, repositoryService.Create(ctx, repositoryObj, dao.AutoCreateNamespace{UserID: userObj.ID}))

// 	tagService := dao.NewTagServiceFactory().New()
// 	tagObj := &models.Tag{
// 		RepositoryID: repositoryObj.ID,
// 		Name:         "latest",
// 		Artifact: &models.Artifact{
// 			NamespaceID:  namespaceObj.ID,
// 			RepositoryID: repositoryObj.ID,
// 			Digest:       "sha256:xxx",
// 			Size:         123,
// 			ContentType:  "test",
// 			Raw:          []byte("test"),
// 			Type:         enums.ArtifactTypeImage,
// 		},
// 	}
// 	assert.NoError(t, tagService.Create(ctx, tagObj))

// 	tag1, err := tagService.GetByID(ctx, tagObj.ID)
// 	assert.NoError(t, err)
// 	assert.Equal(t, tag1.ID, tagObj.ID)

// 	tag2, err := tagService.GetByName(ctx, repositoryObj.ID, "latest")
// 	assert.NoError(t, err)
// 	assert.Equal(t, tag2.ID, tagObj.ID)

// 	err = tagService.Incr(ctx, tagObj.ID)
// 	assert.NoError(t, err)
// 	tag3, err := tagService.GetByID(ctx, tagObj.ID)
// 	assert.NoError(t, err)
// 	assert.Equal(t, tag3.PullTimes, int64(1))

// 	tags1, _, err := tagService.ListTag(ctx, repositoryObj.ID, nil, nil, types.Pagination{
// 		Limit: ptr.Of(int(100)),
// 		Page:  ptr.Of(int(0)),
// 	}, types.Sortable{})
// 	assert.NoError(t, err)
// 	assert.Equal(t, len(tags1), int(1))

// 	// count1, err := tagService.CountTag(ctx, types.ListTagRequest{
// 	// 	Pagination: types.Pagination{
// 	// 		Limit: ptr.Of(int(100)),
// 	// 		Page:  ptr.Of(int(0)),
// 	// 	},
// 	// 	Repository: "test/busybox",
// 	// })
// 	// assert.NoError(t, err)
// 	// assert.Equal(t, count1, int64(1))

// 	err = tagService.DeleteByName(ctx, repositoryObj.ID, "latest")
// 	assert.NoError(t, err)

// 	artifactObj := &models.Artifact{
// 		NamespaceID:  namespaceObj.ID,
// 		RepositoryID: repositoryObj.ID,
// 		Digest:       "sha256:xxxxx",
// 		Size:         123,
// 		ContentType:  "test",
// 		Raw:          []byte("test"),
// 	}
// 	assert.NoError(t, artifactService.Create(ctx, artifactObj))

// 	tagObj1 := &models.Tag{
// 		RepositoryID: repositoryObj.ID,
// 		Name:         "latest1",
// 		Artifact:     artifactObj,
// 	}
// 	err = tagService.Create(ctx, tagObj1)
// 	assert.NoError(t, err)

// 	err = tagService.DeleteByID(ctx, tagObj1.ID)
// 	assert.NoError(t, err)

// 	err = tagService.DeleteByID(ctx, 10)
// 	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)

// 	artifactObj2 := &models.Artifact{
// 		NamespaceID:  namespaceObj.ID,
// 		RepositoryID: repositoryObj.ID,
// 		Digest:       "sha256:xxxxxxxx",
// 		Size:         123,
// 		ContentType:  "test",
// 		Raw:          []byte("test"),
// 	}
// 	assert.NoError(t, artifactService.Create(ctx, artifactObj2))

// 	tagObj2 := &models.Tag{
// 		RepositoryID: repositoryObj.ID,
// 		Name:         "latest1",
// 		Artifact:     artifactObj2,
// 	}
// 	assert.NoError(t, tagService.Create(ctx, tagObj2))

// 	tags2, err := tagService.ListByDtPagination(ctx, "test/busybox", 10, 1)
// 	assert.NoError(t, err)
// 	assert.Equal(t, len(tags2), int(1))

// 	tagCount1, err := tagService.CountByArtifact(ctx, []int64{tagObj2.ArtifactID})
// 	assert.NoError(t, err)
// 	assert.Equal(t, len(tagCount1), int(1))
// 	assert.Equal(t, tagCount1[tagObj2.ArtifactID], int64(1))
// }
