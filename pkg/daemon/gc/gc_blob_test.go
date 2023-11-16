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

package gc

// import (
// 	"context"
// 	"testing"
// 	"time"

// 	"github.com/rs/zerolog/log"
// 	"github.com/spf13/viper"
// 	"github.com/stretchr/testify/assert"
// 	"go.uber.org/mock/gomock"

// 	"github.com/go-sigma/sigma/pkg/dal"
// 	"github.com/go-sigma/sigma/pkg/dal/dao"
// 	"github.com/go-sigma/sigma/pkg/dal/models"
// 	"github.com/go-sigma/sigma/pkg/logger"
// 	"github.com/go-sigma/sigma/pkg/storage"
// 	"github.com/go-sigma/sigma/pkg/storage/mocks"
// 	"github.com/go-sigma/sigma/pkg/tests"
// )

// func TestGcBlobs(t *testing.T) {
// 	viper.SetDefault("log.level", "debug")
// 	viper.SetDefault("daemon.gc.retention", "72h")
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

// 	blobServiceFactory := dao.NewBlobServiceFactory()
// 	blobService := blobServiceFactory.New()
// 	assert.NoError(t, blobService.Create(ctx, &models.Blob{
// 		Digest:      "sha256:812535778d12027c8dd62a23e0547009560b2710c7da7ea2cd83a935ccb525ba",
// 		Size:        123,
// 		ContentType: "test",
// 	}))
// 	assert.NoError(t, blobService.Create(ctx, &models.Blob{
// 		Digest:      "sha256:dd53a0648c2540d757c28393241492e45ef51eff032733da304010b3f616d660",
// 		Size:        234,
// 		ContentType: "test",
// 		CreatedAt:   time.Now().Add(time.Hour * 73 * -1),
// 		UpdatedAt:   time.Now().Add(time.Hour * 73 * -1),
// 	}))

// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	storageMockStorageDriver := mocks.NewMockStorageDriver(ctrl)
// 	storageMockStorageDriver.EXPECT().Delete(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, _ string) error {
// 		return nil
// 	}).Times(1)

// 	storageMockStorageDriverFactory := mocks.NewMockStorageDriverFactory(ctrl)
// 	storageMockStorageDriverFactory.EXPECT().New().DoAndReturn(func() storage.StorageDriver {
// 		return storageMockStorageDriver
// 	}).Times(1)

// 	g := gc{
// 		blobServiceFactory:   dao.NewBlobServiceFactory(),
// 		storageDriverFactory: storageMockStorageDriverFactory,
// 	}
// 	err := g.gcBlobRunner(ctx)
// 	assert.NoError(t, err)
// }
