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

	"github.com/stretchr/testify/require"

	"github.com/go-sigma/sigma/pkg/dal/dao"
	"github.com/go-sigma/sigma/pkg/dal/query"
)

func TestBlobUploadServiceFactory(t *testing.T) {
	f := dao.NewBlobUploadServiceFactory()
	require.NotNil(t, f.New())
	require.NotNil(t, f.New(query.Q))
}

// func TestBlobUploadService(t *testing.T) {
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

// 	blobUploadServiceFactory := dao.NewBlobUploadServiceFactory()
// 	blobUploadService := blobUploadServiceFactory.New()
// 	blobUploadObj := &models.BlobUpload{
// 		PartNumber: 1,
// 		UploadID:   "test1",
// 		Etag:       "test1",
// 		Repository: "test/busybox",
// 		FileID:     "test1",
// 		Size:       100,
// 	}
// 	assert.NoError(t, blobUploadService.Create(ctx, blobUploadObj))

// 	_, err := blobUploadService.TotalEtagsByUploadID(ctx, "test1")
// 	assert.NoError(t, err)

// 	blobUploadObj1 := &models.BlobUpload{
// 		PartNumber: 2,
// 		UploadID:   "test1",
// 		Etag:       "test2",
// 		Repository: "test/busybox",
// 		FileID:     "test2",
// 		Size:       100,
// 	}
// 	assert.NoError(t, blobUploadService.Create(ctx, blobUploadObj1))

// 	uploads1, err := blobUploadService.FindAllByUploadID(ctx, "test1")
// 	assert.NoError(t, err)
// 	assert.Len(t, uploads1, 2)

// 	upload1, err := blobUploadService.GetLastPart(ctx, "test1")
// 	assert.NoError(t, err)
// 	assert.Equal(t, blobUploadObj1.ID, upload1.ID)

// 	etags1, err := blobUploadService.TotalEtagsByUploadID(ctx, "test1")
// 	assert.NoError(t, err)
// 	assert.Len(t, etags1, 1)

// 	size, err := blobUploadService.TotalSizeByUploadID(ctx, "test1")
// 	assert.NoError(t, err)
// 	assert.Equal(t, int64(200), size)

// 	assert.NoError(t, blobUploadService.DeleteByUploadID(ctx, "test1"))
// }
