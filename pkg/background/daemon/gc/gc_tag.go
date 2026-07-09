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
	"regexp"
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

type gcTag struct {
	ctx    context.Context
	config *config.Configuration

	runnerObj *models.DaemonGcTagRunner

	successCount int64
	failedCount  int64

	namespaceRepository  reponamespace.NamespaceRepository
	repositoryRepository reporegistry.RepositoryRepository
	tagRepository        reporegistry.TagRepository
	artifactRepository   reporegistry.ArtifactRepository
	blobRepository       reporegistry.BlobRepository
	daemonRepository     repodaemon.DaemonRepository
	locker               lock.Locker
}

// Run ...
func (g *gcTag) Run(ctx context.Context, runner *runnerContext, runnerID string) error {
	if err := runner.start(); err != nil {
		return fmt.Errorf("start gc tag runner failed: %v", err)
	}

	daemonRepository := g.daemonRepository
	var err error
	g.runnerObj, err = daemonRepository.GetGcTagRunner(ctx, runnerID)
	if err != nil {
		_ = runner.finish(enums.TaskCommonStatusFailed, fmt.Sprintf("get gc tag runner failed: %v", err), g.successCount, g.failedCount)
		return fmt.Errorf("get gc tag runner failed: %v", err)
	}

	lockCtx, cancel, err := acquireRuleLock(ctx, g.locker, g.config.Daemon.GC.LockExpire, g.config.Daemon.GC.LockWaitTimeout, enums.DaemonGcTag, g.runnerObj.RuleID)
	if err != nil {
		_ = runner.finish(enums.TaskCommonStatusFailed, fmt.Sprintf("acquire gc tag lock failed: %v", err), g.successCount, g.failedCount)
		return fmt.Errorf("acquire gc tag lock failed: %v", err)
	}
	defer cancel()
	g.ctx = lockCtx

	_ = runner.emitWebhook(api.WebhookPayload{
		ResourceType: enums.WebhookResourceTypeDaemonTaskGcTagRunner,
		Action:       enums.WebhookActionStarted,
	}, g.packWebhookObj(enums.WebhookActionStarted))

	if g.runnerObj.Rule.RetentionRuleType != enums.RetentionRuleTypeDay && g.runnerObj.Rule.RetentionRuleType != enums.RetentionRuleTypeQuantity {
		_ = runner.finish(enums.TaskCommonStatusFailed, fmt.Sprintf("gc tag rule retention type(%s) is invalid", g.runnerObj.Rule.RetentionRuleType), g.successCount, g.failedCount)
		_ = runner.emitWebhook(api.WebhookPayload{ResourceType: enums.WebhookResourceTypeDaemonTaskGcTagRunner, Action: enums.WebhookActionFinished}, g.packWebhookObj(enums.WebhookActionFinished))
		return fmt.Errorf("gc tag rule retention type is invalid: %v", g.runnerObj.Rule.RetentionRuleType)
	}

	var pattern *regexp.Regexp
	if g.runnerObj.Rule.RetentionPattern != nil {
		pattern, err = regexp.Compile(ptr.To(g.runnerObj.Rule.RetentionPattern))
		if err != nil {
			_ = runner.finish(enums.TaskCommonStatusFailed, fmt.Sprintf("compile gc tag pattern failed: %v", err), g.successCount, g.failedCount)
			_ = runner.emitWebhook(api.WebhookPayload{ResourceType: enums.WebhookResourceTypeDaemonTaskGcTagRunner, Action: enums.WebhookActionFinished}, g.packWebhookObj(enums.WebhookActionFinished))
			return fmt.Errorf("compile gc tag pattern failed: %v", err)
		}
	}

	namespaceRepository := g.namespaceRepository
	repositoryRepository := g.repositoryRepository
	tagRepository := g.tagRepository

	if g.runnerObj.Rule.NamespaceID != nil {
		if err := g.deleteTagsInNamespace(lockCtx, daemonRepository, repositoryRepository, tagRepository, ptr.To(g.runnerObj.Rule.NamespaceID), pattern); err != nil {
			_ = runner.finish(enums.TaskCommonStatusFailed, err.Error(), g.successCount, g.failedCount)
			return err
		}
	} else {
		var namespaceCurIndex string
		for {
			namespaceObjs, err := namespaceRepository.FindWithCursor(lockCtx, int64(g.config.Daemon.GC.BatchSize), namespaceCurIndex)
			if err != nil {
				_ = runner.finish(enums.TaskCommonStatusFailed, fmt.Sprintf("get namespace with cursor failed: %v", err), g.successCount, g.failedCount)
				_ = runner.emitWebhook(api.WebhookPayload{ResourceType: enums.WebhookResourceTypeDaemonTaskGcTagRunner, Action: enums.WebhookActionFinished}, g.packWebhookObj(enums.WebhookActionFinished))
				return fmt.Errorf("get namespace with cursor failed: %v", err)
			}
			for _, nsObj := range namespaceObjs {
				if err := g.deleteTagsInNamespace(lockCtx, daemonRepository, repositoryRepository, tagRepository, nsObj.ID, pattern); err != nil {
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
		return fmt.Errorf("finish gc tag runner failed: %v", err)
	}
	_ = runner.emitWebhook(api.WebhookPayload{
		ResourceType: enums.WebhookResourceTypeDaemonTaskGcTagRunner,
		Action:       enums.WebhookActionFinished,
	}, g.packWebhookObj(enums.WebhookActionFinished))

	return nil
}

func (g *gcTag) deleteTagsInNamespace(ctx context.Context, daemonRepository repodaemon.DaemonRepository, repositoryRepository reporegistry.RepositoryRepository, tagRepository reporegistry.TagRepository, namespaceID string, pattern *regexp.Regexp) error {
	var repositoryCurIndex string
	for {
		repositories, err := repositoryRepository.FindAll(ctx, namespaceID, int64(g.config.Daemon.GC.BatchSize), repositoryCurIndex)
		if err != nil {
			return fmt.Errorf("list repository failed: %v", err)
		}
		for _, repositoryObj := range repositories {
			if err := g.deleteTagsInRepository(ctx, daemonRepository, tagRepository, repositoryObj.ID, pattern); err != nil {
				return err
			}
		}
		if len(repositories) < g.config.Daemon.GC.BatchSize {
			return nil
		}
		repositoryCurIndex = repositories[len(repositories)-1].ID
	}
}

func (g *gcTag) deleteTagsInRepository(ctx context.Context, daemonRepository repodaemon.DaemonRepository, tagRepository reporegistry.TagRepository, repositoryID string, pattern *regexp.Regexp) error {
	var tagCurIndex string
	for {
		var tags []*models.Tag
		var err error
		switch g.runnerObj.Rule.RetentionRuleType {
		case enums.RetentionRuleTypeDay:
			tags, err = tagRepository.FindWithDayCursor(ctx, repositoryID, int(g.runnerObj.Rule.RetentionRuleAmount), g.config.Daemon.GC.BatchSize, tagCurIndex)
		case enums.RetentionRuleTypeQuantity:
			tags, err = tagRepository.FindWithQuantityCursor(ctx, repositoryID, int(g.runnerObj.Rule.RetentionRuleAmount), g.config.Daemon.GC.BatchSize, tagCurIndex)
		}
		if err != nil {
			return fmt.Errorf("find deletable tag failed: %v", err)
		}
		var mu sync.Mutex
		if err := runWorkers(ctx, g.config.Daemon.GC.WorkerCount, tags, func(ctx context.Context, tagObj *models.Tag) error {
			if pattern != nil && !pattern.MatchString(tagObj.Name) {
				return nil
			}
			record := &models.DaemonGcTagRecord{ID: uuid.NewV7String(), RunnerID: g.runnerObj.ID, Tag: tagObj.Name, Status: enums.GcRecordStatusSuccess}
			if err := tagRepository.DeleteByID(ctx, tagObj.ID); err != nil {
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
			if err := daemonRepository.CreateGcTagRecords(ctx, []*models.DaemonGcTagRecord{record}); err != nil {
				slog.Error("create gc tag record failed", "err", err)
			}
			return nil
		}); err != nil {
			return err
		}
		if len(tags) < g.config.Daemon.GC.BatchSize {
			return nil
		}
		tagCurIndex = tags[len(tags)-1].ID
	}
}

func (g *gcTag) packWebhookObj(action enums.WebhookAction) api.WebhookPayloadGcTag {
	payload := api.WebhookPayloadGcTag{
		WebhookPayload: api.WebhookPayload{
			ResourceType: enums.WebhookResourceTypeDaemonTaskGcTagRunner,
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
