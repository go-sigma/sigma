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

func TestBlobUploadServiceFactory(t *testing.T) {
	f := NewBlobUploadServiceFactory()
	blobUploadService := f.New()
	assert.NotNil(t, blobUploadService)
	blobUploadService = f.New(query.Q)
	assert.NotNil(t, blobUploadService)
}

func TestBlobUploadService(t *testing.T) {
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

	blobUploadServiceFactory := NewBlobUploadServiceFactory()
	err = query.Q.Transaction(func(tx *query.Query) error {
		blobUploadService := blobUploadServiceFactory.New(tx)
		blobUploadObj := &models.BlobUpload{
			PartNumber: 1,
			UploadID:   "test1",
			Etag:       "test1",
			Repository: "test/busybox",
			FileID:     "test1",
			Size:       100,
		}
		err = blobUploadService.Create(ctx, blobUploadObj)
		assert.NoError(t, err)

		_, err = blobUploadService.TotalEtagsByUploadID(ctx, "test1")
		assert.NoError(t, err)

		blobUploadObj1 := &models.BlobUpload{
			PartNumber: 2,
			UploadID:   "test1",
			Etag:       "test2",
			Repository: "test/busybox",
			FileID:     "test2",
			Size:       100,
		}
		err = blobUploadService.Create(ctx, blobUploadObj1)
		assert.NoError(t, err)

		uploads1, err := blobUploadService.FindAllByUploadID(ctx, "test1")
		assert.NoError(t, err)
		assert.Len(t, uploads1, 2)

		upload1, err := blobUploadService.GetLastPart(ctx, "test1")
		assert.NoError(t, err)
		assert.Equal(t, blobUploadObj1.ID, upload1.ID)

		etags1, err := blobUploadService.TotalEtagsByUploadID(ctx, "test1")
		assert.NoError(t, err)
		assert.Len(t, etags1, 1)

		size, err := blobUploadService.TotalSizeByUploadID(ctx, "test1")
		assert.NoError(t, err)
		assert.Equal(t, int64(200), size)

		err = blobUploadService.DeleteByUploadID(ctx, "test1")
		assert.NoError(t, err)

		return nil
	})
	assert.NoError(t, err)
}
