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
// 	"gorm.io/gorm"

// 	"github.com/go-sigma/sigma/pkg/dal"
// 	"github.com/go-sigma/sigma/pkg/dal/dao"
// 	"github.com/go-sigma/sigma/pkg/logger"
// 	"github.com/go-sigma/sigma/pkg/tests"
// 	"github.com/go-sigma/sigma/pkg/types/enums"
// )

// func TestGcArtifactNormal(t *testing.T) {
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

// 	sql, err := os.ReadFile(fmt.Sprintf("./testdata/gc_artifact_normal.%s.sql", tests.DB.GetName()))
// 	assert.NoError(t, err)

// 	for _, s := range strings.Split(string(sql), ";\n") {
// 		s := strings.TrimSpace(s)
// 		if len(s) == 0 {
// 			continue
// 		}
// 		err = dal.DB.Debug().Exec(s).Error
// 		assert.NoError(t, err)
// 	}

// 	var runnerChan = make(chan decoratorStatus, 4)
// 	var webhookChan = make(chan decoratorWebhook, 4)

// 	runner := initGc(ctx, enums.DaemonGcArtifact, runnerChan, webhookChan)
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

// 	artifactService := dao.NewArtifactServiceFactory().New()
// 	_, err = artifactService.Get(ctx, 5)
// 	assert.Equal(t, gorm.ErrRecordNotFound, err)
// 	artifact1, err := artifactService.Get(ctx, 1)
// 	assert.NoError(t, err)
// 	assert.Equal(t, "application/vnd.oci.image.manifest.v1+json", artifact1.ContentType)
// 	assert.Equal(t, int64(1), artifact1.ID)
// 	artifact2, err := artifactService.Get(ctx, 2)
// 	assert.NoError(t, err)
// 	assert.Equal(t, "application/vnd.oci.image.manifest.v1+json", artifact1.ContentType)
// 	assert.Equal(t, int64(2), artifact2.ID)
// 	artifact3, err := artifactService.Get(ctx, 3)
// 	assert.NoError(t, err)
// 	assert.Equal(t, "application/vnd.oci.image.manifest.v1+json", artifact1.ContentType)
// 	assert.Equal(t, int64(3), artifact3.ID)
// 	artifact4, err := artifactService.Get(ctx, 4)
// 	assert.NoError(t, err)
// 	assert.Equal(t, "application/vnd.oci.image.manifest.v1+json", artifact1.ContentType)
// 	assert.Equal(t, int64(4), artifact4.ID)

// 	runnerChan = make(chan decoratorStatus, 4)
// 	webhookChan = make(chan decoratorWebhook, 4)
// 	runner = initGc(ctx, enums.DaemonGcArtifact, runnerChan, webhookChan)
// 	err = runner.Run(1)
// 	assert.NoError(t, err)

// 	webhookArr = make([]string, 0, 10)
// 	for status := range webhookChan {
// 		webhookArr = append(webhookArr, string(status.Meta.Action))
// 	}
// 	assert.Equal(t, []string{"Started", "Finished"}, webhookArr)

// 	statusArr = make([]string, 0, 10)
// 	for status := range runnerChan {
// 		statusArr = append(statusArr, string(status.Status))
// 	}
// 	assert.Equal(t, []string{"Doing", "Doing", "Success"}, statusArr)

// 	_, err = artifactService.Get(ctx, 1)
// 	assert.Equal(t, gorm.ErrRecordNotFound, err)
// 	_, err = artifactService.Get(ctx, 2)
// 	assert.Equal(t, gorm.ErrRecordNotFound, err)
// 	_, err = artifactService.Get(ctx, 3)
// 	assert.Equal(t, gorm.ErrRecordNotFound, err)
// 	_, err = artifactService.Get(ctx, 4)
// 	assert.Equal(t, gorm.ErrRecordNotFound, err)
// }
