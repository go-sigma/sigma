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
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"github.com/ximager/ximager/pkg/dal"
	"github.com/ximager/ximager/pkg/dal/models"
	"github.com/ximager/ximager/pkg/dal/query"
	"github.com/ximager/ximager/pkg/logger"
	"github.com/ximager/ximager/pkg/tests"
)

func TestBlobServiceFactory(t *testing.T) {
	f := NewBlobServiceFactory()
	blobService := f.New()
	assert.NotNil(t, blobService)
	blobService = f.New(query.Q)
	assert.NotNil(t, blobService)
}

func TestBlobService(t *testing.T) {
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
		assert.NoError(t, tests.DB.DeInit())
	}()

	ctx := log.Logger.WithContext(context.Background())

	f := NewBlobServiceFactory()
	err = query.Q.Transaction(func(tx *query.Query) error {
		blobService := f.New(tx)
		err = blobService.Create(ctx, &models.Blob{
			Digest:      "sha256:123",
			Size:        123,
			ContentType: "test",
		})
		assert.NoError(t, err)
		err = blobService.Create(ctx, &models.Blob{
			Digest:      "sha256:234",
			Size:        234,
			ContentType: "test",
		})
		assert.NoError(t, err)
		blob1, err := blobService.FindByDigest(ctx, "sha256:123")
		assert.NoError(t, err)
		assert.Equal(t, blob1.Size, int64(123))
		blobs1, err := blobService.FindByDigests(ctx, []string{"sha256:123", "sha256:234"})
		assert.NoError(t, err)
		assert.Equal(t, len(blobs1), int(2))

		time.Sleep(time.Second * 3)
		blobFindWithLastPull, err := blobService.FindWithLastPull(ctx, time.Now(), 0, 1000)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(blobFindWithLastPull))

		var ids []int64
		for _, blob := range blobFindWithLastPull {
			ids = append(ids, blob.ID)
		}
		rIds, err := blobService.FindAssociateWithArtifact(ctx, ids)
		assert.NoError(t, err)
		log.Info().Interface("ids", rIds).Msg("")

		exist, err := blobService.Exists(ctx, "sha256:123")
		assert.NoError(t, err)
		assert.True(t, exist)

		exist, err = blobService.Exists(ctx, "sha256:1231")
		assert.NoError(t, err)
		assert.False(t, exist)

		err = blobService.Incr(ctx, blob1.ID)
		assert.NoError(t, err)

		blob1, err = blobService.FindByDigest(ctx, "sha256:123")
		assert.NoError(t, err)
		assert.Equal(t, blob1.PullTimes, uint(1))
		assert.True(t, blob1.LastPull.Valid)

		err = blobService.DeleteByID(ctx, blob1.ID)
		assert.NoError(t, err)

		err = blobService.DeleteByID(ctx, 10)
		assert.ErrorIs(t, err, gorm.ErrRecordNotFound)

		return nil
	})
	assert.NoError(t, err)
}
