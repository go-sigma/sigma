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
// 	"fmt"
// 	"os"
// 	"strings"
// 	"testing"

// 	"github.com/rs/zerolog/log"
// 	"github.com/stretchr/testify/assert"
// 	"go.uber.org/mock/gomock"
// 	"gorm.io/gorm"

// 	"github.com/go-sigma/sigma/pkg/dal"
// 	"github.com/go-sigma/sigma/pkg/dal/dao"
// 	"github.com/go-sigma/sigma/pkg/logger"
// 	"github.com/go-sigma/sigma/pkg/storage"
// 	storagemocks "github.com/go-sigma/sigma/pkg/storage/mocks"
// 	"github.com/go-sigma/sigma/pkg/tests"
// 	"github.com/go-sigma/sigma/pkg/types/enums"
// )

// func TestGcBlobNormal(t *testing.T) {
// 	logger.SetLevel("debug")
// 	assert.NoError(t, tests.Initialize(t))
// 	assert.NoError(t, tests.DB.Init())
// 	defer func() {
// 		conn, err := dal.DB.DB()
// 		assert.NoError(t, err)
// 		assert.NoError(t, conn.Close())
// 		assert.NoError(t, tests.DB.DeInit())
// 	}()

// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	ctx := log.Logger.WithContext(context.Background())

// 	sql, err := os.ReadFile(fmt.Sprintf("./testdata/gc_blob_normal.%s.sql", tests.DB.GetName()))
// 	assert.NoError(t, err)

// 	for _, s := range strings.Split(string(sql), ";\n") {
// 		s := strings.TrimSpace(s)
// 		if len(s) == 0 {
// 			continue
// 		}
// 		err = dal.DB.Debug().Exec(s).Error
// 		assert.NoError(t, err)
// 	}

// 	storageDriver := storagemocks.NewMockStorageDriver(ctrl)
// 	storageDriver.EXPECT().Delete(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, _ string) error {
// 		return nil
// 	}).Times(2)

// 	storageDriverFactory := storagemocks.NewMockStorageDriverFactory(ctrl)
// 	storageDriverFactory.EXPECT().New().DoAndReturn(func() storage.StorageDriver {
// 		return storageDriver
// 	}).Times(2)

// 	var runnerChan = make(chan decoratorStatus, 4)
// 	var webhookChan = make(chan decoratorWebhook, 4)

// 	runner := initGc(ctx, enums.DaemonGcBlob, runnerChan, webhookChan, inject{storageDriverFactory: storageDriverFactory})
// 	err = runner.Run(1)
// 	assert.NoError(t, err)

// 	var webhookArr = make([]string, 0, 10)
// 	for status := range webhookChan {
// 		webhookArr = append(webhookArr, string(status.Meta.Action))
// 	}
// 	assert.Equal(t, []string{"Started", "Finished"}, webhookArr)

// 	var statusArr = make([]string, 0, 10)
// 	for status := range runnerChan {
// 		statusArr = append(statusArr, string(status.Status))
// 	}
// 	assert.Equal(t, []string{"Doing", "Doing", "Success"}, statusArr)

// 	blobService := dao.NewBlobServiceFactory().New()

// 	_, err = blobService.FindByDigest(ctx, "sha256:c6b39de5b33961661dc939b997cc1d30cda01e38005a6c6625fd9c7e748bab44")
// 	assert.Equal(t, gorm.ErrRecordNotFound, err)

// 	_, err = blobService.FindByDigest(ctx, "sha256:33abbf0321492ff7379e60c252c05c4e7ed4dccf46fcca6c558067c25e76dc8b")
// 	assert.Equal(t, gorm.ErrRecordNotFound, err)

// 	blob1, err := blobService.FindByDigest(ctx, "sha256:5385a9a590c3e2872b3ed27554a56fb7ce544c694b41b9b95d70fa86f30b0566")
// 	assert.NoError(t, err)
// 	assert.Equal(t, int64(3258283), blob1.Size)

// 	blob2, err := blobService.FindByDigest(ctx, "sha256:f0fd8fe16bfa55179c65d208ce8abf58197e85136f6a1dc543d2136424fd665c")
// 	assert.NoError(t, err)
// 	assert.Equal(t, int64(1487), blob2.Size)
// }
