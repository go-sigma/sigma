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

type gcArtifact struct {
	ctx    context.Context
	config *config.Configuration

	runnerObj *models.DaemonGcArtifactRunner

	successCount int64
	failedCount  int64

	namespaceRepository  reponamespace.NamespaceRepository
	repositoryRepository reporegistry.RepositoryRepository
	tagRepository        reporegistry.TagRepository
	artifactRepository   reporegistry.ArtifactRepository
	daemonRepository     repodaemon.DaemonRepository
	locker               lock.Locker
}

func (g *gcArtifact) Run(ctx context.Context, runner *runnerContext, runnerID string) error {
	if err := runner.start(); err != nil {
		return fmt.Errorf("start gc artifact runner failed: %v", err)
	}

	daemonRepository := g.daemonRepository
	var err error
	g.runnerObj, err = daemonRepository.GetGcRunner(ctx, runnerID)
	if err != nil {
		_ = runner.finish(enums.TaskCommonStatusFailed, fmt.Sprintf("get gc artifact runner failed: %v", err), g.successCount, g.failedCount)
		return fmt.Errorf("get gc artifact runner failed: %v", err)
	}

	lockCtx, cancel, err := acquireRuleLock(ctx, g.locker, g.config.Daemon.GC.LockExpire, g.config.Daemon.GC.LockWaitTimeout, enums.DaemonGcArtifact, g.runnerObj.RuleID)
	if err != nil {
		_ = runner.finish(enums.TaskCommonStatusFailed, fmt.Sprintf("acquire gc artifact lock failed: %v", err), g.successCount, g.failedCount)
		return fmt.Errorf("acquire gc artifact lock failed: %v", err)
	}
	defer cancel()
	g.ctx = lockCtx

	_ = runner.emitWebhook(api.WebhookPayload{
		ResourceType: enums.WebhookResourceTypeDaemonTaskGcArtifactRunner,
		Action:       enums.WebhookActionStarted,
	}, g.packWebhookObj(enums.WebhookActionStarted))

	namespaceRepository := g.namespaceRepository
	repositoryRepository := g.repositoryRepository
	artifactRepository := g.artifactRepository
	timeTarget := time.Now().Add(-24 * time.Hour * time.Duration(g.runnerObj.Rule.RetentionDay)).UTC().UnixMilli()

	if g.runnerObj.Rule.NamespaceID != nil {
		if err := g.deleteArtifactsInNamespace(lockCtx, daemonRepository, repositoryRepository, artifactRepository, ptr.To(g.runnerObj.Rule.NamespaceID), timeTarget); err != nil {
			_ = runner.finish(enums.TaskCommonStatusFailed, err.Error(), g.successCount, g.failedCount)
			return err
		}
	} else {
		var namespaceCurIndex string
		for {
			namespaceObjs, err := namespaceRepository.FindWithCursor(lockCtx, int64(g.config.Daemon.GC.BatchSize), namespaceCurIndex)
			if err != nil {
				_ = runner.finish(enums.TaskCommonStatusFailed, fmt.Sprintf("get namespace with cursor failed: %v", err), g.successCount, g.failedCount)
				_ = runner.emitWebhook(api.WebhookPayload{ResourceType: enums.WebhookResourceTypeDaemonTaskGcArtifactRunner, Action: enums.WebhookActionFinished}, g.packWebhookObj(enums.WebhookActionFinished))
				return fmt.Errorf("get namespace with cursor failed: %v", err)
			}
			for _, ns := range namespaceObjs {
				if err := g.deleteArtifactsInNamespace(lockCtx, daemonRepository, repositoryRepository, artifactRepository, ns.ID, timeTarget); err != nil {
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
		return fmt.Errorf("finish gc artifact runner failed: %v", err)
	}
	_ = runner.emitWebhook(api.WebhookPayload{
		ResourceType: enums.WebhookResourceTypeDaemonTaskGcArtifactRunner,
		Action:       enums.WebhookActionFinished,
	}, g.packWebhookObj(enums.WebhookActionFinished))

	return nil
}

func (g *gcArtifact) deleteArtifactsInNamespace(ctx context.Context, daemonRepository repodaemon.DaemonRepository, repositoryRepository reporegistry.RepositoryRepository, artifactRepository reporegistry.ArtifactRepository, namespaceID string, before int64) error {
	var repositoryCurIndex string
	for {
		repositories, err := repositoryRepository.FindAll(ctx, namespaceID, int64(g.config.Daemon.GC.BatchSize), repositoryCurIndex)
		if err != nil {
			return fmt.Errorf("list repository failed: %v", err)
		}
		for _, repositoryObj := range repositories {
			if err := g.deleteArtifactsInRepository(ctx, daemonRepository, artifactRepository, repositoryObj, before); err != nil {
				return err
			}
		}
		if len(repositories) < g.config.Daemon.GC.BatchSize {
			return nil
		}
		repositoryCurIndex = repositories[len(repositories)-1].ID
	}
}

func (g *gcArtifact) deleteArtifactsInRepository(ctx context.Context, daemonRepository repodaemon.DaemonRepository, artifactRepository reporegistry.ArtifactRepository, repositoryObj *models.Repository, before int64) error {
	var artifactCurIndex string
	for {
		artifacts, err := artifactRepository.FindDeletableWithCursor(ctx, repositoryObj.ID, before, artifactCurIndex, int64(g.config.Daemon.GC.BatchSize))
		if err != nil {
			return fmt.Errorf("find deletable artifact failed: %v", err)
		}
		digests := make([]string, 0, len(artifacts))
		for _, artifactObj := range artifacts {
			digests = append(digests, artifactObj.Digest)
		}
		referrers, err := artifactRepository.FindReferrersBySubjects(ctx, repositoryObj.ID, digests)
		if err != nil {
			return fmt.Errorf("find artifact referrers failed: %v", err)
		}
		deleteIDs := make([]string, 0, len(artifacts)+len(referrers))
		records := make([]*models.DaemonGcRecord, 0, len(artifacts)+len(referrers))
		var deletedBlobsSize int64
		for _, artifactObj := range artifacts {
			deleteIDs = append(deleteIDs, artifactObj.ID)
			records = append(records, &models.DaemonGcRecord{ID: uuid.NewV7String(), RunnerID: g.runnerObj.ID, Resource: artifactObj.Digest, Status: enums.GcRecordStatusSuccess})
			deletedBlobsSize += artifactObj.BlobsSize
		}
		for _, artifactObj := range referrers {
			deleteIDs = append(deleteIDs, artifactObj.ID)
			records = append(records, &models.DaemonGcRecord{ID: uuid.NewV7String(), RunnerID: g.runnerObj.ID, Resource: artifactObj.Digest, Status: enums.GcRecordStatusSuccess})
			deletedBlobsSize += artifactObj.BlobsSize
		}
		if len(deleteIDs) > 0 {
			if err := artifactRepository.DeleteByIDs(ctx, deleteIDs); err != nil {
				for _, record := range records {
					record.Status = enums.GcRecordStatusFailed
					record.Message = []byte(err.Error())
				}
				g.failedCount += int64(len(records))
			} else {
				g.successCount += int64(len(records))
				if deletedBlobsSize > 0 {
					namespaceRepository := g.namespaceRepository
					if err := namespaceRepository.DecrementSize(ctx, repositoryObj.NamespaceID, deletedBlobsSize); err != nil {
						slog.Error("decrement namespace size after gc failed", "err", err, "namespace_id", repositoryObj.NamespaceID)
					}
					repositoryRepository := g.repositoryRepository
					if err := repositoryRepository.DecrementSize(ctx, repositoryObj.ID, deletedBlobsSize); err != nil {
						slog.Error("decrement repository size after gc failed", "err", err, "repository_id", repositoryObj.ID)
					}
				}
			}
			if err := daemonRepository.CreateGcRecords(ctx, records); err != nil {
				slog.Error("create gc artifact record failed", "err", err)
			}
		}
		if len(artifacts) < g.config.Daemon.GC.BatchSize {
			return nil
		}
		artifactCurIndex = artifacts[len(artifacts)-1].ID
	}
}

func (g *gcArtifact) packWebhookObj(action enums.WebhookAction) api.WebhookPayloadGcArtifact {
	payload := api.WebhookPayloadGcArtifact{
		WebhookPayload: api.WebhookPayload{
			ResourceType: enums.WebhookResourceTypeDaemonTaskGcArtifactRunner,
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
