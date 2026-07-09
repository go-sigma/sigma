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
	"log/slog"
	"sync"
	"time"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	repodaemon "github.com/go-sigma/sigma/pkg/dal/repository/daemon"
	reponamespace "github.com/go-sigma/sigma/pkg/dal/repository/namespace"
	reporegistry "github.com/go-sigma/sigma/pkg/dal/repository/registry"
	"github.com/go-sigma/sigma/pkg/infra/lock"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
	"github.com/go-sigma/sigma/pkg/utils/uuid"
)

type gcRepository struct {
	ctx    context.Context
	config *config.Configuration

	runnerObj *models.DaemonGcRepositoryRunner

	successCount int64
	failedCount  int64

	namespaceRepository  reponamespace.NamespaceRepository
	repositoryRepository reporegistry.RepositoryRepository
	tagRepository        reporegistry.TagRepository
	daemonRepository     repodaemon.DaemonRepository
	locker               lock.Locker
}

// Run ...
func (g *gcRepository) Run(ctx context.Context, runner *runnerContext, runnerID string) error {
	if err := runner.start(); err != nil {
		return fmt.Errorf("start gc repository runner failed: %v", err)
	}

	daemonRepository := g.daemonRepository
	var err error
	g.runnerObj, err = daemonRepository.GetGcRepositoryRunner(ctx, runnerID)
	if err != nil {
		_ = runner.finish(enums.TaskCommonStatusFailed, fmt.Sprintf("get gc repository runner failed: %v", err), g.successCount, g.failedCount)
		return fmt.Errorf("get gc repository runner failed: %v", err)
	}

	lockCtx, cancel, err := acquireRuleLock(ctx, g.locker, g.config.Daemon.GC.LockExpire, g.config.Daemon.GC.LockWaitTimeout, enums.DaemonGcRepository, g.runnerObj.RuleID)
	if err != nil {
		_ = runner.finish(enums.TaskCommonStatusFailed, fmt.Sprintf("acquire gc repository lock failed: %v", err), g.successCount, g.failedCount)
		return fmt.Errorf("acquire gc repository lock failed: %v", err)
	}
	defer cancel()
	g.ctx = lockCtx

	_ = runner.emitWebhook(api.WebhookPayload{
		ResourceType: enums.WebhookResourceTypeDaemonTaskGcRepositoryRunner,
		Action:       enums.WebhookActionStarted,
	}, g.packWebhookObj(enums.WebhookActionStarted))

	namespaceRepository := g.namespaceRepository
	repositoryRepository := g.repositoryRepository
	timeTarget := time.Now().Add(-24 * time.Hour * time.Duration(g.runnerObj.Rule.RetentionDay)).UTC().UnixMilli()

	if g.runnerObj.Rule.NamespaceID != nil {
		if err := g.deleteRepositoriesInNamespace(lockCtx, daemonRepository, repositoryRepository, ptr.To(g.runnerObj.Rule.NamespaceID), timeTarget); err != nil {
			_ = runner.finish(enums.TaskCommonStatusFailed, err.Error(), g.successCount, g.failedCount)
			return err
		}
	} else {
		var namespaceCurIndex string
		for {
			namespaceObjs, err := namespaceRepository.FindWithCursor(lockCtx, int64(g.config.Daemon.GC.BatchSize), namespaceCurIndex)
			if err != nil {
				_ = runner.finish(enums.TaskCommonStatusFailed, fmt.Sprintf("get namespace with cursor failed: %v", err), g.successCount, g.failedCount)
				_ = runner.emitWebhook(api.WebhookPayload{ResourceType: enums.WebhookResourceTypeDaemonTaskGcRepositoryRunner, Action: enums.WebhookActionFinished}, g.packWebhookObj(enums.WebhookActionFinished))
				return fmt.Errorf("get namespace with cursor failed: %v", err)
			}
			for _, nsObj := range namespaceObjs {
				if err := g.deleteRepositoriesInNamespace(lockCtx, daemonRepository, repositoryRepository, nsObj.ID, timeTarget); err != nil {
					_ = runner.finish(enums.TaskCommonStatusFailed, err.Error(), g.successCount, g.failedCount)
					return err
				}
			}
			if len(namespaceObjs) < g.config.Daemon.GC.BatchSize {
				break
			}
			namespaceCurIndex = namespaceObjs[len(namespaceObjs)-1].ID
		}
	}

	if err := runner.finish(enums.TaskCommonStatusSuccess, "", g.successCount, g.failedCount); err != nil {
		return fmt.Errorf("finish gc repository runner failed: %v", err)
	}
	_ = runner.emitWebhook(api.WebhookPayload{
		ResourceType: enums.WebhookResourceTypeDaemonTaskGcRepositoryRunner,
		Action:       enums.WebhookActionFinished,
	}, g.packWebhookObj(enums.WebhookActionFinished))

	return nil
}

func (g *gcRepository) deleteRepositoriesInNamespace(ctx context.Context, daemonRepository repodaemon.DaemonRepository, repositoryRepository reporegistry.RepositoryRepository, namespaceID string, before int64) error {
	var curIndex string
	for {
		repositories, err := repositoryRepository.FindDeletableWithCursor(ctx, namespaceID, before, curIndex, int64(g.config.Daemon.GC.BatchSize))
		if err != nil {
			return fmt.Errorf("find deletable repository failed: %v", err)
		}
		var mu sync.Mutex
		if err := runWorkers(ctx, g.config.Daemon.GC.WorkerCount, repositories, func(ctx context.Context, repositoryObj *models.Repository) error {
			record := &models.DaemonGcRepositoryRecord{ID: uuid.NewV7String(), RunnerID: g.runnerObj.ID, Repository: repositoryObj.Name, Status: enums.GcRecordStatusSuccess}
			if err := repositoryRepository.DeleteByID(ctx, repositoryObj.ID); err != nil {
				record.Status = enums.GcRecordStatusFailed
				record.Message = []byte(err.Error())
				mu.Lock()
				g.failedCount++
				mu.Unlock()
			} else {
				mu.Lock()
				g.successCount++
				mu.Unlock()
			}
			if err := daemonRepository.CreateGcRepositoryRecords(ctx, []*models.DaemonGcRepositoryRecord{record}); err != nil {
				slog.Error("create gc repository record failed", "err", err)
			}
			return nil
		}); err != nil {
			return err
		}
		if len(repositories) < g.config.Daemon.GC.BatchSize {
			return nil
		}
		curIndex = repositories[len(repositories)-1].ID
	}
}

func (g *gcRepository) packWebhookObj(action enums.WebhookAction) api.WebhookPayloadGcRepository {
	payload := api.WebhookPayloadGcRepository{
		WebhookPayload: api.WebhookPayload{
			ResourceType: enums.WebhookResourceTypeDaemonTaskGcRepositoryRunner,
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
