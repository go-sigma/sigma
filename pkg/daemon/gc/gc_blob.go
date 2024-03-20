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

import (
	"context"
	"fmt"
	"sync"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/opencontainers/go-digest"
	"github.com/rs/zerolog/log"

	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/dao"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/modules/workq"
	"github.com/go-sigma/sigma/pkg/modules/workq/definition"
	"github.com/go-sigma/sigma/pkg/storage"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

func init() {
	workq.TopicHandlers[enums.DaemonGcBlob] = definition.Consumer{
		Handler:     decorator(enums.DaemonGcBlob),
		MaxRetry:    6,
		Concurrency: 10,
		Timeout:     time.Minute * 10,
	}
}

type blobTask struct {
	Runner models.DaemonGcBlobRunner
	Blob   models.Blob
}

type blobTaskCollectRecord struct {
	Status  enums.GcRecordStatus
	Runner  models.DaemonGcBlobRunner
	Blob    models.Blob
	Message *string
}

type gcBlob struct {
	ctx    context.Context
	config configs.Configuration

	runnerObj *models.DaemonGcBlobRunner

	successCount int64
	failedCount  int64

	blobServiceFactory   dao.BlobServiceFactory
	daemonServiceFactory dao.DaemonServiceFactory
	storageDriverFactory storage.StorageDriverFactory

	deleteBlobChan        chan blobTask
	deleteBlobChanOnce    *sync.Once
	collectRecordChan     chan blobTaskCollectRecord
	collectRecordChanOnce *sync.Once

	runnerChan  chan decoratorStatus
	webhookChan chan decoratorWebhook

	waitAllDone *sync.WaitGroup
}

// Run ...
func (g gcBlob) Run(runnerID int64) error {
	defer close(g.runnerChan)
	g.runnerChan <- decoratorStatus{Daemon: enums.DaemonGcBlob, Status: enums.TaskCommonStatusDoing, Started: true}

	var err error
	g.runnerObj, err = g.daemonServiceFactory.New().GetGcBlobRunner(g.ctx, runnerID)
	if err != nil {
		g.runnerChan <- decoratorStatus{
			Daemon:  enums.DaemonGcBlob,
			Status:  enums.TaskCommonStatusFailed,
			Message: fmt.Sprintf("Get gc blob runner failed: %v", err),
			Ended:   true,
		}
		return fmt.Errorf("get gc blob runner failed: %v", err)
	}

	g.webhookChan <- decoratorWebhook{Meta: types.WebhookPayload{
		ResourceType: enums.WebhookResourceTypeDaemonTaskGcBlobRunner,
		Action:       enums.WebhookActionStarted,
	}, WebhookObj: g.packWebhookObj(enums.WebhookActionStarted)}

	blobService := g.blobServiceFactory.New()

	timeTarget := time.Now().UnixMilli()
	if g.runnerObj.Rule.RetentionDay > 0 {
		timeTarget = time.Now().Add(-1 * time.Duration(g.runnerObj.Rule.RetentionDay) * 24 * time.Hour).UnixMilli()
	}

	g.deleteBlobChanOnce.Do(g.deleteBlob)
	g.collectRecordChanOnce.Do(g.collectRecord)
	g.waitAllDone.Add(2)

	var curIndex int64
	for {
		blobs, err := blobService.FindWithLastPull(g.ctx, timeTarget, curIndex, pagination)
		if err != nil {
			g.runnerChan <- decoratorStatus{
				Daemon:  enums.DaemonGcBlob,
				Status:  enums.TaskCommonStatusFailed,
				Message: fmt.Sprintf("Get blob with last pull failed: %v", err),
				Ended:   true,
			}
			g.webhookChan <- decoratorWebhook{Meta: types.WebhookPayload{
				ResourceType: enums.WebhookResourceTypeDaemonTaskGcBlobRunner,
				Action:       enums.WebhookActionFinished,
			}, WebhookObj: g.packWebhookObj(enums.WebhookActionFinished)}
			return fmt.Errorf("get blob with last pull failed: %v", err)
		}
		if len(blobs) == 0 {
			break
		}
		var ids []int64
		for _, blob := range blobs {
			ids = append(ids, blob.ID)
		}
		associateBlobIDs, err := blobService.FindAssociateWithArtifact(g.ctx, ids)
		if err != nil {
			g.runnerChan <- decoratorStatus{Daemon: enums.DaemonGcBlob, Status: enums.TaskCommonStatusFailed, Message: fmt.Sprintf("Check blob associate with artifact failed: %v", err), Ended: true}
			g.webhookChan <- decoratorWebhook{Meta: types.WebhookPayload{
				ResourceType: enums.WebhookResourceTypeDaemonTaskGcBlobRunner,
				Action:       enums.WebhookActionFinished,
			}, WebhookObj: g.packWebhookObj(enums.WebhookActionFinished)}
			return fmt.Errorf("check blob associate with artifact failed: %v", err)
		}
		notAssociateBlobIDs := mapset.NewSet(ids...)
		notAssociateBlobIDs.RemoveAll(associateBlobIDs...)
		notAssociateBlobSlice := notAssociateBlobIDs.ToSlice()
		if len(notAssociateBlobSlice) > 0 {
			var notAssociateBlobObjs = make([]*models.Blob, 0, pagination)
			for _, id := range notAssociateBlobSlice {
				for _, blob := range blobs {
					if blob.ID == id {
						notAssociateBlobObjs = append(notAssociateBlobObjs, blob)
					}
				}
			}
			if len(notAssociateBlobObjs) > 0 {
				for _, blob := range notAssociateBlobObjs {
					g.deleteBlobChan <- blobTask{Runner: ptr.To(g.runnerObj), Blob: ptr.To(blob)}
				}
			}
		}
		if len(blobs) < pagination {
			break
		}
		log.Info().Interface("blob", blobs[len(blobs)-1]).Send()
		curIndex = blobs[len(blobs)-1].ID
	}
	close(g.deleteBlobChan)
	g.waitAllDone.Wait()

	g.runnerChan <- decoratorStatus{Daemon: enums.DaemonGcBlob, Status: enums.TaskCommonStatusSuccess, Ended: true}
	g.webhookChan <- decoratorWebhook{Meta: types.WebhookPayload{
		ResourceType: enums.WebhookResourceTypeDaemonTaskGcBlobRunner,
		Action:       enums.WebhookActionFinished,
	}, WebhookObj: g.packWebhookObj(enums.WebhookActionFinished)}

	return nil
}

func (g gcBlob) deleteBlob() {
	blobService := g.blobServiceFactory.New()
	go func() {
		defer g.waitAllDone.Done()
		defer close(g.collectRecordChan)
		for task := range g.deleteBlobChan {
			// TODO: we should set a lock for the delete action
			err := blobService.DeleteByID(g.ctx, task.Blob.ID)
			if err != nil {
				log.Error().Err(err).Interface("Task", task).Msgf("Delete blob failed: %v", err)
				g.collectRecordChan <- blobTaskCollectRecord{
					Status:  enums.GcRecordStatusFailed,
					Blob:    task.Blob,
					Runner:  task.Runner,
					Message: ptr.Of(fmt.Sprintf("Delete blob failed: %v", err)),
				}
				continue
			}
			err = g.storageDriverFactory.New().Delete(g.ctx, utils.GenPathByDigest(digest.Digest(task.Blob.Digest)))
			if err != nil {
				log.Error().Err(err).Interface("blob", task).Msgf("Delete blob in obs failed: %v", err)
			}
			// TODO: if we delete the file in obs failed, just ignore the error.
			// so we should check each file in obs associate with database record.
			g.collectRecordChan <- blobTaskCollectRecord{Status: enums.GcRecordStatusSuccess, Blob: task.Blob, Runner: task.Runner}
		}
	}()
}

func (g gcBlob) collectRecord() {
	daemonService := g.daemonServiceFactory.New()
	go func() {
		defer g.waitAllDone.Done()
		defer func() {
			g.runnerChan <- decoratorStatus{Daemon: enums.DaemonGcBlob, Status: enums.TaskCommonStatusDoing, Updates: map[string]any{
				"success_count": g.successCount,
				"failed_count":  g.failedCount,
			}}
		}()
		for task := range g.collectRecordChan {
			err := daemonService.CreateGcBlobRecords(g.ctx, []*models.DaemonGcBlobRecord{
				{
					RunnerID: task.Runner.ID,
					Digest:   task.Blob.Digest,
					Status:   task.Status,
					Message:  []byte(ptr.To(task.Message)),
				},
			})
			if err != nil {
				log.Error().Err(err).Msg("Create gc blob record failed")
				continue
			}
			if task.Status == enums.GcRecordStatusSuccess {
				g.successCount++
			} else {
				g.failedCount++
			}
		}
	}()
}

func (g gcBlob) packWebhookObj(action enums.WebhookAction) types.WebhookPayloadGcBlob {
	payload := types.WebhookPayloadGcBlob{
		WebhookPayload: types.WebhookPayload{
			ResourceType: enums.WebhookResourceTypeDaemonTaskGcBlobRunner,
			Action:       action,
		},
		OperateType:  g.runnerObj.OperateType,
		SuccessCount: g.successCount,
		FailedCount:  g.failedCount,
	}
	if g.runnerObj.OperateType == enums.OperateTypeManual && g.runnerObj.OperateUser != nil {
		payload.OperateUser = &types.WebhookPayloadUser{
			ID:        g.runnerObj.OperateUser.ID,
			Username:  g.runnerObj.OperateUser.Username,
			Email:     ptr.To(g.runnerObj.OperateUser.Email),
			Status:    g.runnerObj.OperateUser.Status,
			LastLogin: time.Unix(0, int64(time.Millisecond)*g.runnerObj.OperateUser.LastLogin).UTC().Format(consts.DefaultTimePattern),
			CreatedAt: time.Unix(0, int64(time.Millisecond)*g.runnerObj.OperateUser.CreatedAt).UTC().Format(consts.DefaultTimePattern),
			UpdatedAt: time.Unix(0, int64(time.Millisecond)*g.runnerObj.OperateUser.CreatedAt).UTC().Format(consts.DefaultTimePattern),
		}
	}
	return payload
}
