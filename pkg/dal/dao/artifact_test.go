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
	"golang.org/x/exp/slices"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/dal"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/logger"
	"github.com/go-sigma/sigma/pkg/tests"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

func TestArtifactServiceFactory(t *testing.T) {
	f := NewArtifactServiceFactory()
	artifactService := f.New()
	assert.NotNil(t, artifactService)
	artifactService = f.New(query.Q)
	assert.NotNil(t, artifactService)
}

func TestArtifactServiceAssociateArtifact(t *testing.T) {
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

	userServiceFactory := NewUserServiceFactory()
	namespaceServiceFactory := NewNamespaceServiceFactory()
	repositoryServiceFactory := NewRepositoryServiceFactory()

	var repositoryObj *models.Repository
	err = query.Q.Transaction(func(tx *query.Query) error {
		userService := userServiceFactory.New(tx)
		userObj := &models.User{Provider: enums.ProviderLocal, Username: "artifact-service", Password: ptr.Of("test"), Email: ptr.Of("test@gmail.com")}
		err = userService.Create(ctx, userObj)
		assert.NoError(t, err)

		namespaceService := namespaceServiceFactory.New(tx)
		namespaceObj := &models.Namespace{Name: "test", Visibility: enums.VisibilityPrivate}
		err = namespaceService.Create(ctx, namespaceObj)
		assert.NoError(t, err)

		repositoryService := repositoryServiceFactory.New(tx)
		repositoryObj = &models.Repository{Name: "test/busybox", NamespaceID: namespaceObj.ID, Visibility: enums.VisibilityPrivate}
		err = repositoryService.Create(ctx, repositoryObj, AutoCreateNamespace{UserID: userObj.ID})
		assert.NoError(t, err)
		return nil
	})
	assert.NoError(t, err)

	artifactServiceFactory := NewArtifactServiceFactory()
	artifactService := artifactServiceFactory.New()
	artifactObj1 := &models.Artifact{
		RepositoryID: repositoryObj.ID,
		Digest:       "sha256:xxxx",
		Size:         123,
		ContentType:  "test",
		Raw:          []byte("test"),
	}
	err = artifactService.Create(ctx, artifactObj1)
	assert.NoError(t, err)

	artifactObj2 := &models.Artifact{
		RepositoryID: repositoryObj.ID,
		Digest:       "sha256:xxxxx",
		Size:         1234,
		ContentType:  "test",
		Raw:          []byte("test"),
	}
	err = artifactService.Create(ctx, artifactObj2)
	assert.NoError(t, err)
	err = artifactService.AssociateArtifact(ctx, artifactObj1, []*models.Artifact{artifactObj2})
	assert.NoError(t, err)
}

func TestArtifactService(t *testing.T) {
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

	tagServiceFactory := NewTagServiceFactory()
	namespaceServiceFactory := NewNamespaceServiceFactory()
	repositoryServiceFactory := NewRepositoryServiceFactory()
	artifactServiceFactory := NewArtifactServiceFactory()
	userServiceFactory := NewUserServiceFactory()

	var artifactObj *models.Artifact
	var tagObj1 *models.Tag
	err = query.Q.Transaction(func(tx *query.Query) error {
		userService := userServiceFactory.New(tx)
		userObj := &models.User{Provider: enums.ProviderLocal, Username: "artifact-service", Password: ptr.Of("test"), Email: ptr.Of("test@gmail.com")}
		err = userService.Create(ctx, userObj)
		assert.NoError(t, err)

		namespaceService := namespaceServiceFactory.New(tx)
		namespaceObj := &models.Namespace{Name: "test", Visibility: enums.VisibilityPrivate}
		err = namespaceService.Create(ctx, namespaceObj)
		assert.NoError(t, err)

		repositoryService := repositoryServiceFactory.New(tx)
		repositoryObj := &models.Repository{Name: "test/busybox", NamespaceID: namespaceObj.ID, Visibility: enums.VisibilityPrivate}
		err = repositoryService.Create(ctx, repositoryObj, AutoCreateNamespace{UserID: userObj.ID})
		assert.NoError(t, err)

		artifactService := artifactServiceFactory.New(tx)
		artifactObj = &models.Artifact{
			RepositoryID: repositoryObj.ID,
			Digest:       "sha256:xxxx",
			Size:         123,
			ContentType:  "test",
			Raw:          []byte("test"),
		}
		err = artifactService.Create(ctx, artifactObj)
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
		tagObj1 = tagObj
		assert.Equal(t, tagObj1.Name, tagObj.Name)

		artifact1, err := artifactService.Get(ctx, artifactObj.ID)
		assert.NoError(t, err)
		assert.Equal(t, artifact1.ID, artifactObj.ID)

		artifacts1, err := artifactService.GetByDigests(ctx, "test/busybox", []string{"sha256:xxxx"})
		assert.NoError(t, err)
		assert.Equal(t, len(artifacts1), int(1))

		err = artifactService.Incr(ctx, artifactObj.ID)
		assert.NoError(t, err)
		artifact1, err = artifactService.Get(ctx, artifactObj.ID)
		assert.NoError(t, err)
		assert.Equal(t, artifact1.ID, artifactObj.ID)
		assert.Equal(t, artifact1.PullTimes, int64(1))
		assert.True(t, artifact1.LastPull.Valid)

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

		return nil
	})
	assert.NoError(t, err)

	err = query.Q.Transaction(func(tx *query.Query) error {
		artifactService := artifactServiceFactory.New(tx)
		err = artifactService.AssociateBlobs(ctx, artifactObj,
			[]*models.Blob{{
				Digest:      "sha256:123",
				Size:        123,
				ContentType: "test",
			}})
		assert.NoError(t, err)
		return nil
	})
	assert.NoError(t, err)

	artifactService := artifactServiceFactory.New()
	err = artifactService.DeleteByDigest(ctx, "test/busybox", artifactObj.Digest)
	assert.NoError(t, err)
	err = artifactService.DeleteByID(ctx, 10)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	err = artifactService.DeleteByID(ctx, tagObj1.ArtifactID)
	assert.NoError(t, err)

	err = query.Q.Transaction(func(tx *query.Query) error {
		userService := userServiceFactory.New(tx)
		userObj := &models.User{Provider: enums.ProviderLocal, Username: "artifact-service1", Password: ptr.Of("test"), Email: ptr.Of("test1@gmail.com")}
		err = userService.Create(ctx, userObj)
		assert.NoError(t, err)

		namespaceService := namespaceServiceFactory.New(tx)
		namespaceObj := &models.Namespace{Name: "test1", Visibility: enums.VisibilityPrivate}
		err = namespaceService.Create(ctx, namespaceObj)
		assert.NoError(t, err)

		repositoryService := repositoryServiceFactory.New(tx)
		repositoryObj := &models.Repository{Name: "test1/busybox", NamespaceID: namespaceObj.ID, Visibility: enums.VisibilityPrivate}
		err = repositoryService.Create(ctx, repositoryObj, AutoCreateNamespace{UserID: userObj.ID})
		assert.NoError(t, err)

		artifactObj = &models.Artifact{
			RepositoryID: repositoryObj.ID,
			Digest:       "sha256:xxxx",
			Size:         123,
			ContentType:  "test",
			Raw:          []byte("test"),
		}
		artifactService := artifactServiceFactory.New(tx)
		err = artifactService.Create(ctx, artifactObj)
		assert.NoError(t, err)

		sbomObj := &models.ArtifactSbom{ArtifactID: artifactObj.ID, Raw: []byte("test"), Status: enums.TaskCommonStatusPending}
		err = artifactService.CreateSbom(ctx, sbomObj)
		assert.NoError(t, err)
		err = artifactService.UpdateSbom(ctx, artifactObj.ID, map[string]any{
			query.ArtifactSbom.Status.ColumnName().String(): enums.TaskCommonStatusSuccess,
		})
		assert.NoError(t, err)

		vulnObj := &models.ArtifactVulnerability{ArtifactID: artifactObj.ID, Raw: []byte("test"), Status: enums.TaskCommonStatusPending}
		err = artifactService.CreateVulnerability(ctx, vulnObj)
		assert.NoError(t, err)
		err = artifactService.UpdateVulnerability(ctx, artifactObj.ID, map[string]any{
			query.ArtifactVulnerability.Status.ColumnName().String(): enums.TaskCommonStatusSuccess,
		})
		assert.NoError(t, err)
		return nil
	})
	assert.NoError(t, err)
}
