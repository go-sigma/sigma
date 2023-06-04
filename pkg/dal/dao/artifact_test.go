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

func TestArtifactServiceFactory(t *testing.T) {
	f := NewArtifactServiceFactory()
	artifactService := f.New()
	assert.NotNil(t, artifactService)
	artifactService = f.New(query.Q)
	assert.NotNil(t, artifactService)
}

func TestArtifactService(t *testing.T) {
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
	artifactServiceFactory := NewArtifactServiceFactory()
	err = query.Q.Transaction(func(tx *query.Query) error {
		namespaceService := namespaceServiceFactory.New(tx)
		namespaceObj := &models.Namespace{Name: "test"}
		err = namespaceService.Create(ctx, namespaceObj)
		assert.NoError(t, err)

		repositoryService := repositoryServiceFactory.New(tx)
		repositoryObj := &models.Repository{Name: "test/busybox", NamespaceID: namespaceObj.ID}
		err = repositoryService.Create(ctx, repositoryObj)
		assert.NoError(t, err)

		artifactService := artifactServiceFactory.New(tx)
		artifactObj := &models.Artifact{
			RepositoryID: repositoryObj.ID,
			Digest:       "sha256:xxxx",
			Size:         123,
			ContentType:  "test",
			Raw:          "test",
		}
		err = artifactService.Save(ctx, artifactObj)
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
				Raw:          "test",
			},
		}
		tagObj1, err := tagService.Save(ctx, tagObj)
		assert.NoError(t, err)
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
		assert.Equal(t, artifact1.PullTimes, uint64(1))
		assert.True(t, artifact1.LastPull.Valid)

		return nil
	})
	assert.NoError(t, err)
}
