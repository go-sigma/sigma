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
	"golang.org/x/exp/slices"
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

func TestArtifactServiceFactory(t *testing.T) {
	f := dao.NewArtifactServiceFactory()
	assert.NotNil(t, f.New())
	assert.NotNil(t, f.New(query.Q))
}

func TestArtifactServiceAssociateArtifact(t *testing.T) {
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

	userServiceFactory := dao.NewUserServiceFactory()
	namespaceServiceFactory := dao.NewNamespaceServiceFactory()
	repositoryServiceFactory := dao.NewRepositoryServiceFactory()

	userService := userServiceFactory.New()
	userObj := &models.User{Username: "artifact-service", Password: ptr.Of("test"), Email: ptr.Of("test@gmail.com")}
	assert.NoError(t, userService.Create(ctx, userObj))

	namespaceService := namespaceServiceFactory.New()
	namespaceObj := &models.Namespace{Name: "test", Visibility: enums.VisibilityPrivate}
	assert.NoError(t, namespaceService.Create(ctx, namespaceObj))

	repositoryService := repositoryServiceFactory.New()
	repositoryObj := &models.Repository{Name: "test/busybox", NamespaceID: namespaceObj.ID}
	assert.NoError(t, repositoryService.Create(ctx, repositoryObj, dao.AutoCreateNamespace{UserID: userObj.ID}))

	artifactServiceFactory := dao.NewArtifactServiceFactory()
	artifactService := artifactServiceFactory.New()
	artifactObj1 := &models.Artifact{
		NamespaceID:  namespaceObj.ID,
		RepositoryID: repositoryObj.ID,
		Digest:       "sha256:xxxx",
		Size:         123,
		ContentType:  "test",
		Raw:          []byte("test"),
	}
	assert.NoError(t, artifactService.Create(ctx, artifactObj1))

	artifactObj2 := &models.Artifact{
		NamespaceID:  namespaceObj.ID,
		RepositoryID: repositoryObj.ID,
		Digest:       "sha256:xxxxx",
		Size:         1234,
		ContentType:  "test",
		Raw:          []byte("test"),
	}
	assert.NoError(t, artifactService.Create(ctx, artifactObj2))
	assert.NoError(t, artifactService.AssociateArtifact(ctx, artifactObj1, []*models.Artifact{artifactObj2}))
}

func TestArtifactService(t *testing.T) {
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

	userService := dao.NewUserServiceFactory().New()
	namespaceService := dao.NewNamespaceServiceFactory().New()
	repositoryService := dao.NewRepositoryServiceFactory().New()
	tagService := dao.NewTagServiceFactory().New()
	artifactService := dao.NewArtifactServiceFactory().New()

	userObj := &models.User{Username: "artifact-service", Password: ptr.Of("test"), Email: ptr.Of("test@gmail.com")}
	assert.NoError(t, userService.Create(ctx, userObj))

	namespaceObj := &models.Namespace{Name: "test", Visibility: enums.VisibilityPrivate}
	assert.NoError(t, namespaceService.Create(ctx, namespaceObj))

	repositoryObj := &models.Repository{Name: "test/busybox", NamespaceID: namespaceObj.ID}
	assert.NoError(t, repositoryService.Create(ctx, repositoryObj, dao.AutoCreateNamespace{UserID: userObj.ID}))

	artifactObj := &models.Artifact{
		NamespaceID:  repositoryObj.ID,
		RepositoryID: repositoryObj.ID,
		Digest:       "sha256:xxxx",
		Size:         123,
		ContentType:  "test",
		Raw:          []byte("test"),
	}
	assert.NoError(t, artifactService.Create(ctx, artifactObj))

	tagObj := &models.Tag{
		RepositoryID: repositoryObj.ID,
		Name:         "latest",
		Artifact: &models.Artifact{
			NamespaceID:  repositoryObj.ID,
			RepositoryID: repositoryObj.ID,
			Digest:       "sha256:xxx",
			Size:         123,
			ContentType:  "test",
			Raw:          []byte("test"),
		},
	}
	assert.NoError(t, tagService.Create(ctx, tagObj))
	assert.Equal(t, tagObj.Name, tagObj.Name)

	artifact1, err := artifactService.Get(ctx, artifactObj.ID)
	assert.NoError(t, err)
	assert.Equal(t, artifact1.ID, artifactObj.ID)

	artifacts1, err := artifactService.GetByDigests(ctx, "test/busybox", []string{"sha256:xxxx"})
	assert.NoError(t, err)
	assert.Equal(t, len(artifacts1), int(1))

	assert.NoError(t, artifactService.Incr(ctx, artifactObj.ID))
	artifact1, err = artifactService.Get(ctx, artifactObj.ID)
	assert.NoError(t, err)
	assert.Equal(t, artifact1.ID, artifactObj.ID)
	assert.Equal(t, artifact1.PullTimes, int64(1))

	nsCount1, err := artifactService.CountByNamespace(ctx, []int64{namespaceObj.ID})
	assert.NoError(t, err)
	assert.Equal(t, len(nsCount1), 1)
	assert.Equal(t, nsCount1[namespaceObj.ID], int64(2))

	nsCount2, err := artifactService.CountByNamespace(ctx, []int64{})
	assert.NoError(t, err)
	assert.Equal(t, len(nsCount2), 0)

	repoCount1, err := artifactService.CountByRepository(ctx, []int64{repositoryObj.ID})
	assert.NoError(t, err)
	assert.Equal(t, len(repoCount1), 1)
	assert.Equal(t, repoCount1[repositoryObj.ID], int64(2))

	repoCount2, err := artifactService.CountByRepository(ctx, []int64{})
	assert.NoError(t, err)
	assert.Equal(t, len(repoCount2), 0)

	artifacts2, err := artifactService.ListArtifact(ctx, types.ListArtifactRequest{
		Pagination: types.Pagination{
			Limit: ptr.Of(int(100)),
			Page:  ptr.Of(int(0)),
		},
		Namespace:  namespaceObj.Name,
		Repository: repositoryObj.Name,
	})
	assert.NoError(t, err)
	assert.Equal(t, len(artifacts2), 2)
	assert.True(t, slices.Contains([]string{artifacts2[0].Digest, artifacts2[1].Digest}, "sha256:xxxx"))
	assert.True(t, slices.Contains([]string{artifacts2[0].Digest, artifacts2[1].Digest}, "sha256:xxx"))

	artifactCount1, err := artifactService.CountArtifact(ctx, types.ListArtifactRequest{
		Pagination: types.Pagination{
			Limit: ptr.Of(int(100)),
			Page:  ptr.Of(int(0)),
		},
		Namespace:  namespaceObj.Name,
		Repository: repositoryObj.Name,
	})
	assert.NoError(t, err)
	assert.Equal(t, artifactCount1, int64(2))

	assert.NoError(t, artifactService.AssociateBlobs(ctx, artifactObj,
		[]*models.Blob{{
			Digest:      "sha256:123",
			Size:        123,
			ContentType: "test",
		}}))

	assert.NoError(t, artifactService.DeleteByDigest(ctx, "test/busybox", artifactObj.Digest))
	assert.ErrorIs(t, artifactService.DeleteByID(ctx, 10), gorm.ErrRecordNotFound)
	assert.NoError(t, artifactService.DeleteByID(ctx, tagObj.ArtifactID))

	assert.NoError(t, userService.Create(ctx,
		&models.User{Username: "artifact-service1", Password: ptr.Of("test"), Email: ptr.Of("test1@gmail.com")}))

	assert.NoError(t, namespaceService.Create(ctx, &models.Namespace{Name: "test1", Visibility: enums.VisibilityPrivate}))

	assert.NoError(t, repositoryService.Create(ctx,
		&models.Repository{Name: "test1/busybox", NamespaceID: namespaceObj.ID},
		dao.AutoCreateNamespace{UserID: userObj.ID}))

	assert.NoError(t, artifactService.Create(ctx, &models.Artifact{
		NamespaceID:  repositoryObj.ID,
		RepositoryID: repositoryObj.ID,
		Digest:       "sha256:xxxx",
		Size:         123,
		ContentType:  "test",
		Raw:          []byte("test"),
	}))

	assert.NoError(t, artifactService.CreateSbom(ctx,
		&models.ArtifactSbom{ArtifactID: artifactObj.ID, Raw: []byte("test"), Status: enums.TaskCommonStatusPending}))
	assert.NoError(t, artifactService.UpdateSbom(ctx, artifactObj.ID, map[string]any{
		query.ArtifactSbom.Status.ColumnName().String(): enums.TaskCommonStatusSuccess,
	}))

	assert.NoError(t, artifactService.CreateVulnerability(ctx,
		&models.ArtifactVulnerability{ArtifactID: artifactObj.ID, Raw: []byte("test"), Status: enums.TaskCommonStatusPending}))
	assert.NoError(t, artifactService.UpdateVulnerability(ctx, artifactObj.ID, map[string]any{
		query.ArtifactVulnerability.Status.ColumnName().String(): enums.TaskCommonStatusSuccess,
	}))
}

func TestArtifactServiceGetNamespaceSize(t *testing.T) {
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

	userService := dao.NewUserServiceFactory().New()
	namespaceService := dao.NewNamespaceServiceFactory().New()
	repositoryService := dao.NewRepositoryServiceFactory().New()
	artifactService := dao.NewArtifactServiceFactory().New()

	userObj := &models.User{Username: "artifact-service", Password: ptr.Of("test"), Email: ptr.Of("test@gmail.com")}
	assert.NoError(t, userService.Create(ctx, userObj))

	namespaceObj := &models.Namespace{Name: "test", Visibility: enums.VisibilityPrivate}
	assert.NoError(t, namespaceService.Create(ctx, namespaceObj))

	repositoryObj := &models.Repository{Name: "test/busybox", NamespaceID: namespaceObj.ID}
	assert.NoError(t, repositoryService.Create(ctx, repositoryObj, dao.AutoCreateNamespace{UserID: userObj.ID}))

	artifactObj := &models.Artifact{
		NamespaceID:  namespaceObj.ID,
		RepositoryID: repositoryObj.ID,
		Digest:       "sha256:xxxx",
		Size:         123,
		BlobsSize:    123,
		ContentType:  "test",
		Raw:          []byte("test"),
	}
	assert.NoError(t, artifactService.Create(ctx, artifactObj))

	assert.NoError(t, query.Q.Transaction(func(tx *query.Query) error {
		artifactService := dao.NewArtifactServiceFactory().New(tx)
		size, err := artifactService.GetNamespaceSize(ctx, namespaceObj.ID)
		if err != nil {
			return err
		}
		assert.Equal(t, int64(123), size)
		return nil
	}))
}
