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
	"github.com/go-sigma/sigma/pkg/dal/dao"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/modules/workq"
	"github.com/go-sigma/sigma/pkg/modules/workq/definition"
	"github.com/go-sigma/sigma/pkg/storage"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

func init() {
	workq.TopicHandlers[enums.DaemonGcBlob.String()] = definition.Consumer{
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

	namespaceServiceFactory  dao.NamespaceServiceFactory
	repositoryServiceFactory dao.RepositoryServiceFactory
	artifactServiceFactory   dao.ArtifactServiceFactory
	blobServiceFactory       dao.BlobServiceFactory
	daemonServiceFactory     dao.DaemonServiceFactory
	storageDriverFactory     storage.StorageDriverFactory

	deleteBlobChan        chan blobTask
	deleteBlobChanOnce    *sync.Once
	collectRecordChan     chan blobTaskCollectRecord
	collectRecordChanOnce *sync.Once

	runnerChan chan decoratorStatus

	waitAllDone *sync.WaitGroup
}

// Run ...
func (g gcBlob) Run(runnerID int64) error {
	defer close(g.runnerChan)
	g.runnerChan <- decoratorStatus{Daemon: enums.DaemonGcRepository, Status: enums.TaskCommonStatusDoing}

	runnerObj, err := g.daemonServiceFactory.New().GetGcBlobRunner(g.ctx, runnerID)
	if err != nil {
		g.runnerChan <- decoratorStatus{Daemon: enums.DaemonGcBlob, Status: enums.TaskCommonStatusFailed, Message: fmt.Sprintf("Get gc blob runner failed: %v", err), Ended: true}
		return fmt.Errorf("get gc blob runner failed: %v", err)
	}

	blobService := g.blobServiceFactory.New()

	timeTarget := time.Now()
	if runnerObj.Rule.RetentionDay > 0 {
		timeTarget = time.Now().Add(-1 * time.Duration(runnerObj.Rule.RetentionDay) * 24 * time.Hour)
	}

	g.deleteBlobChanOnce.Do(g.deleteBlob)
	g.collectRecordChanOnce.Do(g.collectRecord)
	g.waitAllDone.Add(2)

	var curIndex int64
	for {
		blobs, err := blobService.FindWithLastPull(g.ctx, timeTarget, curIndex, pagination)
		if err != nil {
			g.runnerChan <- decoratorStatus{Daemon: enums.DaemonGcBlob, Status: enums.TaskCommonStatusFailed, Message: fmt.Sprintf("Get blob with last pull failed: %v", err), Ended: true}
			return fmt.Errorf("get blob with last pull failed: %v", err)
		}
		var ids []int64
		for _, blob := range blobs {
			ids = append(ids, blob.ID)
		}
		associateBlobIDs, err := blobService.FindAssociateWithArtifact(g.ctx, ids)
		if err != nil {
			g.runnerChan <- decoratorStatus{Daemon: enums.DaemonGcBlob, Status: enums.TaskCommonStatusFailed, Message: fmt.Sprintf("Check blob associate with artifact failed: %v", err), Ended: true}
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
					g.deleteBlobChan <- blobTask{Runner: ptr.To(runnerObj), Blob: ptr.To(blob)}
				}
			}
		}
		if len(blobs) < pagination {
			break
		}
		curIndex = blobs[len(blobs)-1].ID
	}
	close(g.deleteBlobChan)
	g.waitAllDone.Wait()

	g.runnerChan <- decoratorStatus{Daemon: enums.DaemonGcTag, Status: enums.TaskCommonStatusSuccess, Ended: true}

	return nil
}

func (g gcBlob) deleteBlob() {
	defer close(g.collectRecordChan)
	for task := range g.deleteBlobChan {
		err := query.Q.Transaction(func(tx *query.Query) error {
			err := g.blobServiceFactory.New(tx).DeleteByID(g.ctx, task.Blob.ID)
			if err != nil {
				return err
			}
			err = g.daemonServiceFactory.New(tx).CreateGcBlobRecords(g.ctx, []*models.DaemonGcBlobRecord{{
				RunnerID: task.Runner.ID,
				Digest:   task.Blob.Digest,
			}})
			if err != nil {
				return err
			}
			err = g.storageDriverFactory.New().Delete(g.ctx, utils.GenPathByDigest(digest.Digest(task.Blob.Digest)))
			if err != nil {
				return err
			}
			log.Info().Str("digest", task.Blob.Digest).Msg("Delete blob success")
			return nil
		})
		if err != nil {
			log.Error().Err(err).Interface("blob", task).Msgf("Delete blob failed: %v", err)
			g.collectRecordChan <- blobTaskCollectRecord{
				Status:  enums.GcRecordStatusFailed,
				Blob:    task.Blob,
				Runner:  task.Runner,
				Message: ptr.Of(fmt.Sprintf("Delete repository by id failed: %v", err)),
			}
			continue
		}
		g.collectRecordChan <- blobTaskCollectRecord{Status: enums.GcRecordStatusSuccess, Blob: task.Blob, Runner: task.Runner}
	}
}

func (g gcBlob) collectRecord() {
	var successCount, failedCount int64
	daemonService := g.daemonServiceFactory.New()
	go func() {
		defer g.waitAllDone.Done()
		defer func() {
			g.runnerChan <- decoratorStatus{Daemon: enums.DaemonGcBlob, Status: enums.TaskCommonStatusDoing, Updates: map[string]any{
				"success_count": successCount,
				"failed_count":  failedCount,
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
				successCount++
			} else {
				failedCount++
			}
		}
	}()
}
