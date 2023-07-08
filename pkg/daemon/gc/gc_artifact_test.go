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

package gc

import (
	"context"
	"testing"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"github.com/ximager/ximager/pkg/dal"
	"github.com/ximager/ximager/pkg/dal/dao"
	"github.com/ximager/ximager/pkg/dal/models"
	"github.com/ximager/ximager/pkg/logger"
	"github.com/ximager/ximager/pkg/tests"
	"github.com/ximager/ximager/pkg/types/enums"
	"github.com/ximager/ximager/pkg/utils/ptr"
)

func TestGcArtifact(t *testing.T) {
	viper.SetDefault("log.level", "debug")
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
	repositoryServiceFactory := dao.NewRepositoryServiceFactory()
	artifactServiceFactory := dao.NewArtifactServiceFactory()
	userServiceFactory := dao.NewUserServiceFactory()

	userService := userServiceFactory.New()
	userObj := &models.User{Provider: enums.ProviderLocal, Username: "gc-artifact", Password: ptr.Of("test"), Email: ptr.Of("test@gmail.com")}
	err := userService.Create(ctx, userObj)
	assert.NoError(t, err)

	namespaceService := namespaceServiceFactory.New()
	namespaceObj := &models.Namespace{Name: "test", Visibility: enums.VisibilityPrivate}
	err = namespaceService.Create(ctx, namespaceObj)
	assert.NoError(t, err)

	repositoryService := repositoryServiceFactory.New()
	repositoryObj := &models.Repository{Name: "test/busybox", NamespaceID: namespaceObj.ID, Visibility: enums.VisibilityPrivate}
	err = repositoryService.Create(ctx, repositoryObj)
	assert.NoError(t, err)

	artifactService := artifactServiceFactory.New()
	artifactObj := &models.Artifact{
		RepositoryID: repositoryObj.ID,
		Digest:       "sha256:812535778d12027c8dd62a23e0547009560b2710c7da7ea2cd83a935ccb525ba",
		Size:         123,
		ContentType:  "test",
		Raw:          []byte("test"),
		CreatedAt:    time.Now().Add(time.Hour * 73 * -1),
		UpdatedAt:    time.Now().Add(time.Hour * 73 * -1),
	}
	err = artifactService.Create(ctx, artifactObj)
	assert.NoError(t, err)

	g := gc{
		namespaceServiceFactory:  namespaceServiceFactory,
		repositoryServiceFactory: repositoryServiceFactory,
		artifactServiceFactory:   artifactServiceFactory,
	}
	err = g.gcArtifact(ctx, "")
	assert.NoError(t, err)
}
