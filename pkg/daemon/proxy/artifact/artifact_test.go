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

package artifact

import (
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/hibiken/asynq"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"

	"github.com/ximager/ximager/pkg/dal"
	"github.com/ximager/ximager/pkg/dal/dao"
	daomock "github.com/ximager/ximager/pkg/dal/dao/mocks"
	"github.com/ximager/ximager/pkg/dal/models"
	"github.com/ximager/ximager/pkg/dal/query"
	"github.com/ximager/ximager/pkg/logger"
	"github.com/ximager/ximager/pkg/tests"
	"github.com/ximager/ximager/pkg/validators"
)

func TestNewRunner(t *testing.T) {
	logger.SetLevel("debug")
	e := echo.New()
	validators.Initialize(e)
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

	proxyServiceFactory := dao.NewProxyServiceFactory()
	proxyService := proxyServiceFactory.New()
	err = proxyService.SaveProxyArtifact(ctx, &models.ProxyArtifactTask{
		Repository:  "library/busybox",
		Digest:      "sha256:f7d81d5be30e617068bf53a9b136400b13d91c0f54d097a72bf91127f43d0157",
		Size:        123,
		ContentType: "test",
		Raw:         []byte("test"),
		Blobs: []models.ProxyArtifactTaskBlob{
			{Blob: "sha256:123"},
			{Blob: "sha256:234"},
		},
	})
	assert.NoError(t, err)

	runner := newRunner()

	err = runner(ctx, asynq.NewTask("test", []byte(`{"blob_digest": "sha256:123"}`)))
	assert.NoError(t, err)

	blobServiceFactory := dao.NewBlobServiceFactory()
	blobService := blobServiceFactory.New()
	err = blobService.Create(ctx, &models.Blob{
		Digest:      "sha256:123",
		Size:        123,
		ContentType: "test",
		PushedAt:    time.Now(),
	})
	assert.NoError(t, err)
	err = blobService.Create(ctx, &models.Blob{
		Digest:      "sha256:234",
		Size:        234,
		ContentType: "test",
		PushedAt:    time.Now(),
	})
	assert.NoError(t, err)

	err = runner(ctx, asynq.NewTask("test", []byte(`{"blob_digest": "sha256:123"}`)))
	assert.NoError(t, err)

	artifact, err := query.Q.Artifact.WithContext(ctx).Where(query.Artifact.Digest.Eq("sha256:f7d81d5be30e617068bf53a9b136400b13d91c0f54d097a72bf91127f43d0157")).Preload(query.Artifact.Blobs).First()
	assert.NoError(t, err)
	assert.Equal(t, artifact.Digest, "sha256:f7d81d5be30e617068bf53a9b136400b13d91c0f54d097a72bf91127f43d0157")
	assert.Equal(t, len(artifact.Blobs), 2)
	assert.Equal(t, artifact.Raw, "test")

	err = runner(ctx, asynq.NewTask("test", []byte(`{"blob_digest": "sha256:345"}`)))
	assert.NoError(t, err)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	daoMockProxyService := daomock.NewMockProxyService(ctrl)
	daoMockProxyServiceTimes := 0
	daoMockProxyService.EXPECT().FindByBlob(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, _ string) ([]*models.ProxyArtifactTask, error) {
		daoMockProxyServiceTimes++
		if daoMockProxyServiceTimes == 1 {
			return nil, fmt.Errorf("test")
		}
		return []*models.ProxyArtifactTask{
			{
				Repository:  "library/busybox",
				Digest:      "sha256:f7d81d5be30e617068bf53a9b136400b13d91c0f54d097a72bf91127f43d0157",
				Size:        123,
				ContentType: "test",
				Raw:         []byte("test"),
				Blobs: []models.ProxyArtifactTaskBlob{
					{Blob: "sha256:123"},
					{Blob: "sha256:234"},
				},
			},
		}, nil
	}).Times(4)

	daoMockProxyServiceFactory := daomock.NewMockProxyServiceFactory(ctrl)
	daoMockProxyServiceFactory.EXPECT().New(gomock.Any()).DoAndReturn(func(txs ...*query.Query) dao.ProxyService {
		return daoMockProxyService
	}).Times(4)

	runner = newRunner(inject{proxyServiceFactory: daoMockProxyServiceFactory})

	err = runner(ctx, asynq.NewTask("test", []byte(`{"blob_digest": "sha256:123"}`)))
	assert.Error(t, err)

	daoMockBlobService := daomock.NewMockBlobService(ctrl)
	daoMockBlobServiceTimes := 0
	daoMockBlobService.EXPECT().FindByDigests(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, _ []string) ([]*models.Blob, error) {
		daoMockBlobServiceTimes++
		if daoMockBlobServiceTimes == 1 {
			return nil, fmt.Errorf("test")
		}
		return []*models.Blob{
			{
				Digest: "sha:123",
			},
			{
				Digest: "sha:234",
			},
		}, nil
	}).Times(3)

	daoMockBlobServiceFactory := daomock.NewMockBlobServiceFactory(ctrl)
	daoMockBlobServiceFactory.EXPECT().New(gomock.Any()).DoAndReturn(func(txs ...*query.Query) dao.BlobService {
		return daoMockBlobService
	}).Times(3)

	runner = newRunner(inject{proxyServiceFactory: daoMockProxyServiceFactory, blobServiceFactory: daoMockBlobServiceFactory})

	err = runner(ctx, asynq.NewTask("test", []byte(`{"blob_digest": "sha256:123"}`)))
	assert.Error(t, err)

	daoMockRepositoryService := daomock.NewMockRepositoryService(ctrl)
	daoMockRepositoryServiceTimes := 0
	daoMockRepositoryService.EXPECT().Save(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, _ *models.Repository) error {
		daoMockRepositoryServiceTimes++
		if daoMockRepositoryServiceTimes == 1 {
			return fmt.Errorf("test")
		}
		return nil
	}).Times(2)

	daoMockRepositoryServiceFactory := daomock.NewMockRepositoryServiceFactory(ctrl)
	daoMockRepositoryServiceFactory.EXPECT().New(gomock.Any()).DoAndReturn(func(txs ...*query.Query) dao.RepositoryService {
		return daoMockRepositoryService
	}).Times(2)

	runner = newRunner(inject{proxyServiceFactory: daoMockProxyServiceFactory, blobServiceFactory: daoMockBlobServiceFactory, repositoryServiceFactory: daoMockRepositoryServiceFactory})

	err = runner(ctx, asynq.NewTask("test", []byte(`{"blob_digest": "sha256:123"}`)))
	assert.Error(t, err)

	daoMockArtifactService := daomock.NewMockArtifactService(ctrl)
	daoMockArtifactService.EXPECT().Save(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, _ *models.Artifact) error {
		return fmt.Errorf("test")
	}).Times(1)

	daoMockArtifactServiceFactory := daomock.NewMockArtifactServiceFactory(ctrl)
	daoMockArtifactServiceFactory.EXPECT().New(gomock.Any()).DoAndReturn(func(txs ...*query.Query) dao.ArtifactService {
		return daoMockArtifactService
	}).Times(1)

	runner = newRunner(inject{proxyServiceFactory: daoMockProxyServiceFactory, blobServiceFactory: daoMockBlobServiceFactory, repositoryServiceFactory: daoMockRepositoryServiceFactory, artifactServiceFactory: daoMockArtifactServiceFactory})

	err = runner(ctx, asynq.NewTask("test", []byte(`{"blob_digest": "sha256:123"}`)))
	assert.Error(t, err)
}
