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
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/dao"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/modules/workq"
	"github.com/go-sigma/sigma/pkg/modules/workq/definition"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

func init() {
	workq.TopicHandlers[enums.DaemonGcArtifact] = definition.Consumer{
		Handler:     decorator(enums.DaemonGcArtifact),
		MaxRetry:    6,
		Concurrency: 10,
		Timeout:     time.Minute * 10,
	}
}

type artifactWithNamespaceTask struct {
	Runner      models.DaemonGcArtifactRunner
	NamespaceID int64
}

type artifactTask struct {
	Runner   models.DaemonGcArtifactRunner
	Artifact models.Artifact
}

type artifactTaskCollectRecord struct {
	Status   enums.GcRecordStatus
	Runner   models.DaemonGcArtifactRunner
	Artifact models.Artifact
	Message  *string
}

type gcArtifact struct {
	ctx    context.Context
	config configs.Configuration

	runnerObj *models.DaemonGcArtifactRunner

	successCount int64
	failedCount  int64

	namespaceServiceFactory  dao.NamespaceServiceFactory
	repositoryServiceFactory dao.RepositoryServiceFactory
	tagServiceFactory        dao.TagServiceFactory
	artifactServiceFactory   dao.ArtifactServiceFactory
	daemonServiceFactory     dao.DaemonServiceFactory

	deleteArtifactWithNamespaceChan     chan artifactWithNamespaceTask
	deleteArtifactWithNamespaceChanOnce *sync.Once
	deleteArtifactCheckChan             chan artifactTask
	deleteArtifactCheckChanOnce         *sync.Once
	deleteArtifactChan                  chan artifactTask
	deleteArtifactChanOnce              *sync.Once
	collectRecordChan                   chan artifactTaskCollectRecord
	collectRecordChanOnce               *sync.Once

	runnerChan  chan decoratorStatus
	webhookChan chan decoratorWebhook

	waitAllDone *sync.WaitGroup
}

func (g gcArtifact) Run(runnerID int64) error {
	defer close(g.runnerChan)
	defer close(g.webhookChan)
	g.runnerChan <- decoratorStatus{Daemon: enums.DaemonGcArtifact, Status: enums.TaskCommonStatusDoing, Started: true}

	var err error
	g.runnerObj, err = g.daemonServiceFactory.New().GetGcArtifactRunner(g.ctx, runnerID)
	if err != nil {
		g.runnerChan <- decoratorStatus{
			Daemon:  enums.DaemonGcArtifact,
			Status:  enums.TaskCommonStatusFailed,
			Message: fmt.Sprintf("Get gc artifact runner failed: %v", err),
			Ended:   true,
		}
		return fmt.Errorf("get gc artifact runner failed: %v", err)
	}

	g.webhookChan <- decoratorWebhook{Meta: types.WebhookPayload{
		ResourceType: enums.WebhookResourceTypeDaemonTaskGcBlobRunner,
		Action:       enums.WebhookActionStarted,
	}, WebhookObj: g.packWebhookObj(enums.WebhookActionStarted)}

	namespaceService := g.namespaceServiceFactory.New()

	g.deleteArtifactWithNamespaceChanOnce.Do(g.deleteArtifactWithNamespace)
	g.deleteArtifactCheckChanOnce.Do(g.deleteArtifactCheck)
	g.deleteArtifactChanOnce.Do(g.deleteArtifact)
	g.collectRecordChanOnce.Do(g.collectRecord)
	g.waitAllDone.Add(4)

	if g.runnerObj.Rule.NamespaceID != nil {
		g.deleteArtifactWithNamespaceChan <- artifactWithNamespaceTask{Runner: ptr.To(g.runnerObj), NamespaceID: ptr.To(g.runnerObj.Rule.NamespaceID)}
	} else {
		var namespaceCurIndex int64
		for {
			namespaceObjs, err := namespaceService.FindWithCursor(g.ctx, pagination, namespaceCurIndex)
			if err != nil {
				g.runnerChan <- decoratorStatus{
					Daemon:  enums.DaemonGcArtifact,
					Status:  enums.TaskCommonStatusFailed,
					Message: fmt.Sprintf("Get namespace with cursor failed: %v", err),
					Ended:   true,
				}
				g.webhookChan <- decoratorWebhook{Meta: types.WebhookPayload{
					ResourceType: enums.WebhookResourceTypeDaemonTaskGcArtifactRunner,
					Action:       enums.WebhookActionFinished,
				}, WebhookObj: g.packWebhookObj(enums.WebhookActionFinished)}
				return fmt.Errorf("get namespace with cursor failed: %v", err)
			}
			for _, ns := range namespaceObjs {
				g.deleteArtifactWithNamespaceChan <- artifactWithNamespaceTask{Runner: ptr.To(g.runnerObj), NamespaceID: ns.ID}
			}
			if len(namespaceObjs) < pagination {
				break
			}
			namespaceCurIndex = namespaceObjs[len(namespaceObjs)-1].ID
		}
	}

	close(g.deleteArtifactWithNamespaceChan)
	g.waitAllDone.Wait()

	g.runnerChan <- decoratorStatus{Daemon: enums.DaemonGcArtifact, Status: enums.TaskCommonStatusSuccess, Ended: true}
	g.webhookChan <- decoratorWebhook{Meta: types.WebhookPayload{
		ResourceType: enums.WebhookResourceTypeDaemonTaskGcArtifactRunner,
		Action:       enums.WebhookActionFinished,
	}, WebhookObj: g.packWebhookObj(enums.WebhookActionFinished)}

	return nil
}

func (g gcArtifact) deleteArtifactWithNamespace() {
	repositoryService := g.repositoryServiceFactory.New()
	artifactService := g.artifactServiceFactory.New()
	go func() {
		defer g.waitAllDone.Done()
		defer close(g.deleteArtifactCheckChan)
		for task := range g.deleteArtifactWithNamespaceChan {
			var repositoryCurIndex int64
			timeTarget := time.Now().UnixMilli()
			if g.runnerObj.Rule.RetentionDay > 0 {
				timeTarget = time.Now().Add(-1 * time.Duration(g.runnerObj.Rule.RetentionDay) * 24 * time.Hour).UnixMilli()
			}
			for {
				repositoryObjs, err := repositoryService.FindAll(g.ctx, task.NamespaceID, pagination, repositoryCurIndex)
				if err != nil {
					log.Error().Err(err).Int64("namespaceID", task.NamespaceID).Msg("List repository failed")
					continue
				}
				for _, repositoryObj := range repositoryObjs {
					var artifactCurIndex int64
					for {
						artifactObjs, err := artifactService.FindWithLastPull(g.ctx, repositoryObj.ID, timeTarget, pagination, artifactCurIndex)
						if err != nil {
							log.Error().Err(err).Msg("List artifact failed")
							continue
						}
						for _, a := range artifactObjs {
							g.deleteArtifactCheckChan <- artifactTask{Runner: task.Runner, Artifact: ptr.To(a)}
						}
						if len(artifactObjs) < pagination {
							break
						}
						artifactCurIndex = artifactObjs[len(artifactObjs)-1].ID
					}
				}
				if len(repositoryObjs) < pagination {
					break
				}
				repositoryCurIndex = repositoryObjs[len(repositoryObjs)-1].ID
			}
		}
	}()
}

func (g gcArtifact) deleteArtifactCheck() {
	artifactService := g.artifactServiceFactory.New()
	tagService := g.tagServiceFactory.New()
	go func() {
		defer g.waitAllDone.Done()
		defer close(g.deleteArtifactChan)
		for task := range g.deleteArtifactCheckChan {
			// 1. check manifest referrer associate with another artifact
			if task.Artifact.ReferrerID != nil {
				continue
			}
			// 2. check tag associate with this artifact
			_, err := tagService.GetByArtifactID(g.ctx, task.Artifact.RepositoryID, task.Artifact.ID)
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				log.Error().Err(err).Int64("repositoryID", task.Artifact.RepositoryID).Int64("artifactID", task.Artifact.ID).Msg("Get tag by artifact failed")
			}
			if err == nil {
				continue
			}
			// 3. check manifest index associate with this artifact
			if !(task.Artifact.ContentType == "application/vnd.docker.distribution.manifest.list.v2+json" ||
				task.Artifact.ContentType == "application/vnd.oci.image.index.v1+json") { // skip this check if artifact is manifest index
				err = artifactService.IsArtifactAssociatedWithArtifact(g.ctx, task.Artifact.ID)
				if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
					log.Error().Err(err).Int64("repositoryID", task.Artifact.RepositoryID).Int64("artifactID", task.Artifact.ID).Msg("Get manifest associated with manifest index failed")
				}
				if err == nil {
					continue
				}
			}
			// 4. delete the artifact that referrer to this artifact
			delArtifacts, err := artifactService.GetReferrers(g.ctx, task.Artifact.RepositoryID, task.Artifact.Digest, nil)
			if err != nil {
				log.Error().Err(err).Int64("repositoryID", task.Artifact.RepositoryID).Int64("artifactID", task.Artifact.ID).Msg("Get artifact referrers failed")
				continue
			}
			// TODO: maybe here should be submit a transaction clear all of the objects, include refers and index
			for _, a := range delArtifacts {
				g.deleteArtifactChan <- artifactTask{Runner: task.Runner, Artifact: ptr.To(a)}
			}
			g.deleteArtifactChan <- task
		}
	}()
}

func (g gcArtifact) deleteArtifact() {
	go func() {
		defer g.waitAllDone.Done()
		defer close(g.collectRecordChan)
		for task := range g.deleteArtifactChan {
			// TODO: we should set a lock for the delete action
			// otherwise, we should delete the artifact in goroutine
			err := query.Q.Transaction(func(tx *query.Query) error {
				err := g.artifactServiceFactory.New(tx).DeleteByID(g.ctx, task.Artifact.ID)
				if err != nil {
					return err
				}
				return nil
			})
			if err != nil {
				log.Error().Err(err).Interface("blob", task).Msgf("Delete blob failed: %v", err)
				g.collectRecordChan <- artifactTaskCollectRecord{
					Status:   enums.GcRecordStatusFailed,
					Artifact: task.Artifact,
					Runner:   task.Runner,
					Message:  ptr.Of(fmt.Sprintf("Delete blob failed: %v", err)),
				}
				continue
			}
			g.collectRecordChan <- artifactTaskCollectRecord{Status: enums.GcRecordStatusSuccess, Artifact: task.Artifact, Runner: task.Runner}
		}
	}()
}

func (g gcArtifact) collectRecord() {
	daemonService := g.daemonServiceFactory.New()
	go func() {
		defer g.waitAllDone.Done()
		defer func() {
			g.runnerChan <- decoratorStatus{Daemon: enums.DaemonGcArtifact, Status: enums.TaskCommonStatusDoing, Updates: map[string]any{
				"success_count": g.successCount,
				"failed_count":  g.failedCount,
			}}
		}()
		for task := range g.collectRecordChan {
			err := daemonService.CreateGcArtifactRecords(g.ctx, []*models.DaemonGcArtifactRecord{
				{
					RunnerID: task.Runner.ID,
					Digest:   task.Artifact.Digest,
					Status:   task.Status,
					Message:  []byte(ptr.To(task.Message)),
				},
			})
			if err != nil {
				log.Error().Err(err).Msg("Create gc repository record failed")
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

func (g gcArtifact) packWebhookObj(action enums.WebhookAction) types.WebhookPayloadGcArtifact {
	payload := types.WebhookPayloadGcArtifact{
		WebhookPayload: types.WebhookPayload{
			ResourceType: enums.WebhookResourceTypeDaemonTaskGcArtifactRunner,
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
