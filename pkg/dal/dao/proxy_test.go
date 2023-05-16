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
	"github.com/ximager/ximager/pkg/logger"
	"github.com/ximager/ximager/pkg/tests"
	"github.com/ximager/ximager/pkg/types"
)

func TestProxyArtifact(t *testing.T) {
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

	proxyService := NewProxyService()
	_, err = proxyService.SaveProxyArtifact(ctx, &models.ProxyArtifactTask{
		Status: types.TaskCommonStatusPending,
		Blobs: []models.ProxyArtifactBlob{
			{Blob: "sha256:123"},
			{Blob: "sha256:456"},
		}})
	assert.NoError(t, err)
	_, err = proxyService.SaveProxyArtifact(ctx, &models.ProxyArtifactTask{
		Status: types.TaskCommonStatusPending,
		Blobs: []models.ProxyArtifactBlob{
			{Blob: "sha256:789"},
			{Blob: "sha256:7891"},
		}})
	assert.NoError(t, err)
	findTasks, err := proxyService.FindByBlob(ctx, "sha256:123")
	assert.NoError(t, err)
	assert.Equal(t, len(findTasks), 1)
	err = proxyService.UpdateProxyArtifactStatus(ctx, findTasks[0].ID, types.TaskCommonStatusSuccess)
	assert.NoError(t, err)
}