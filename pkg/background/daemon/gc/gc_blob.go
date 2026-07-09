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
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/opencontainers/go-digest"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	repodaemon "github.com/go-sigma/sigma/pkg/dal/repository/daemon"
	reporegistry "github.com/go-sigma/sigma/pkg/dal/repository/registry"
	"github.com/go-sigma/sigma/pkg/infra/lock"
	"github.com/go-sigma/sigma/pkg/storage"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
	"github.com/go-sigma/sigma/pkg/utils/uuid"
)

type gcBlob struct {
	ctx    context.Context
	config *config.Configuration

	runnerObj *models.DaemonGcBlobRunner

	successCount int64
	failedCount  int64

	blobRepository   reporegistry.BlobRepository
	daemonRepository repodaemon.DaemonRepository
	storageDriver    storage.StorageDriver
	locker           lock.Locker
}

// Run ...
func (g *gcBlob) Run(ctx context.Context, runner *runnerContext, runnerID string) error {
	if err := runner.start(); err != nil {
		return fmt.Errorf("start gc blob runner failed: %v", err)
	}

	daemonRepository := g.daemonRepository
	var err error
	g.runnerObj, err = daemonRepository.GetGcBlobRunner(ctx, runnerID)
	if err != nil {
		_ = runner.finish(enums.TaskCommonStatusFailed, fmt.Sprintf("get gc blob runner failed: %v", err), g.successCount, g.failedCount)
		return fmt.Errorf("get gc blob runner failed: %v", err)
	}

	lockCtx, cancel, err := acquireRuleLock(ctx, g.locker, g.config.Daemon.GC.LockExpire, g.config.Daemon.GC.LockWaitTimeout, enums.DaemonGcBlob, g.runnerObj.RuleID)
	if err != nil {
		_ = runner.finish(enums.TaskCommonStatusFailed, fmt.Sprintf("acquire gc blob lock failed: %v", err), g.successCount, g.failedCount)
		return fmt.Errorf("acquire gc blob lock failed: %v", err)
	}
	defer cancel()
	g.ctx = lockCtx

	_ = runner.emitWebhook(api.WebhookPayload{
		ResourceType: enums.WebhookResourceTypeDaemonTaskGcBlobRunner,
		Action:       enums.WebhookActionStarted,
	}, g.packWebhookObj(enums.WebhookActionStarted))

	blobRepository := g.blobRepository

	timeTarget := time.Now().UnixMilli()
	if g.runnerObj.Rule.RetentionDay > 0 {
		timeTarget = time.Now().Add(-1 * time.Duration(g.runnerObj.Rule.RetentionDay) * 24 * time.Hour).UnixMilli()
	}

	var curIndex string
	for {
		blobs, err := blobRepository.FindDeletableWithCursor(lockCtx, timeTarget, curIndex, int64(g.config.Daemon.GC.BatchSize))
		if err != nil {
			_ = runner.finish(enums.TaskCommonStatusFailed, fmt.Sprintf("get deletable blob failed: %v", err), g.successCount, g.failedCount)
			_ = runner.emitWebhook(api.WebhookPayload{ResourceType: enums.WebhookResourceTypeDaemonTaskGcBlobRunner, Action: enums.WebhookActionFinished}, g.packWebhookObj(enums.WebhookActionFinished))
			return fmt.Errorf("get blob with last pull failed: %v", err)
		}
		if len(blobs) == 0 {
			break
		}
		var mu sync.Mutex
		if err := runWorkers(lockCtx, g.config.Daemon.GC.WorkerCount, blobs, func(ctx context.Context, blob *models.Blob) error {
			record := g.deleteBlobObject(lockCtx, ptr.To(g.runnerObj), ptr.To(blob))
			mu.Lock()
			defer mu.Unlock()
			if record.Status == enums.GcRecordStatusSuccess {
				g.successCount++
			} else {
				g.failedCount++
			}
			if err := daemonRepository.CreateGcBlobRecords(lockCtx, []*models.DaemonGcBlobRecord{record}); err != nil {
				slog.Error("create gc blob record failed", "err", err)
			}
			return nil
		}); err != nil {
			_ = runner.finish(enums.TaskCommonStatusFailed, err.Error(), g.successCount, g.failedCount)
			return err
		}
		if len(blobs) < g.config.Daemon.GC.BatchSize {
			break
		}
		curIndex = blobs[len(blobs)-1].ID
	}

	if err := runner.report(g.successCount, g.failedCount); err != nil {
		slog.Error("report gc blob count failed", "err", err)
	}
	if err := runner.finish(enums.TaskCommonStatusSuccess, "", g.successCount, g.failedCount); err != nil {
		return fmt.Errorf("finish gc blob runner failed: %v", err)
	}
	_ = runner.emitWebhook(api.WebhookPayload{
		ResourceType: enums.WebhookResourceTypeDaemonTaskGcBlobRunner,
		Action:       enums.WebhookActionFinished,
	}, g.packWebhookObj(enums.WebhookActionFinished))

	return nil
}

func (g *gcBlob) deleteBlobObject(ctx context.Context, runner models.DaemonGcBlobRunner, blob models.Blob) *models.DaemonGcBlobRecord {
	daemonRepository := g.daemonRepository
	storagePath := utils.GenBlobPathByDigest(digest.Digest(blob.Digest))
	task := &models.GcStorageDeletionTask{
		ID:           uuid.NewV7String(),
		Daemon:       enums.DaemonGcBlob,
		RunnerID:     runner.ID,
		ResourceType: "blob",
		ResourceID:   blob.ID,
		StoragePath:  storagePath,
		Status:       enums.TaskCommonStatusPending,
	}
	if err := daemonRepository.UpsertGcStorageDeletionTask(ctx, task); err != nil {
		return &models.DaemonGcBlobRecord{ID: uuid.NewV7String(), RunnerID: runner.ID, Digest: blob.Digest, Status: enums.GcRecordStatusFailed, Message: fmt.Appendf(nil, "create storage deletion task failed: %v", err)}
	}
	if err := daemonRepository.UpdateGcStorageDeletionTask(ctx, task.ID, map[string]any{
		"status": enums.TaskCommonStatusDoing,
	}); err != nil {
		return &models.DaemonGcBlobRecord{ID: uuid.NewV7String(), RunnerID: runner.ID, Digest: blob.Digest, Status: enums.GcRecordStatusFailed, Message: fmt.Appendf(nil, "mark storage deletion task doing failed: %v", err)}
	}
	if err := g.storageDriver.Delete(ctx, storagePath); err != nil {
		_ = daemonRepository.UpdateGcStorageDeletionTask(ctx, task.ID, map[string]any{
			"status":   enums.TaskCommonStatusFailed,
			"attempts": task.Attempts + 1,
			"message":  []byte(err.Error()),
		})
		return &models.DaemonGcBlobRecord{ID: uuid.NewV7String(), RunnerID: runner.ID, Digest: blob.Digest, Status: enums.GcRecordStatusFailed, Message: fmt.Appendf(nil, "delete blob object failed: %v", err)}
	}
	err := query.Q.Transaction(func(tx *query.Query) error {
		if err := reporegistry.NewBlobRepository(tx).DeleteByID(ctx, blob.ID); err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		return repodaemon.NewDaemonRepository(tx).UpdateGcStorageDeletionTask(ctx, task.ID, map[string]any{
			"status": enums.TaskCommonStatusSuccess,
		})
	})
	if err != nil {
		return &models.DaemonGcBlobRecord{ID: uuid.NewV7String(), RunnerID: runner.ID, Digest: blob.Digest, Status: enums.GcRecordStatusFailed, Message: fmt.Appendf(nil, "finalize blob deletion failed: %v", err)}
	}
	return &models.DaemonGcBlobRecord{ID: uuid.NewV7String(), RunnerID: runner.ID, Digest: blob.Digest, Status: enums.GcRecordStatusSuccess}
}

func (g *gcBlob) packWebhookObj(action enums.WebhookAction) api.WebhookPayloadGcBlob {
	payload := api.WebhookPayloadGcBlob{
		WebhookPayload: api.WebhookPayload{
			ResourceType: enums.WebhookResourceTypeDaemonTaskGcBlobRunner,
			Action:       action,
		},
		OperateType:  g.runnerObj.OperateType,
		SuccessCount: g.successCount,
		FailedCount:  g.failedCount,
	}
	if g.runnerObj.OperateType == enums.OperateTypeManual && g.runnerObj.OperateUser != nil {
		payload.OperateUser = &api.WebhookPayloadUser{
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
