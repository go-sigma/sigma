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

	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/modules/workq"
	"github.com/go-sigma/sigma/pkg/modules/workq/definition"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

func init() {
	workq.TopicHandlers[enums.DaemonGcArtifact.String()] = definition.Consumer{
		Handler:     decorator(enums.DaemonGcArtifact),
		MaxRetry:    6,
		Concurrency: 10,
		Timeout:     time.Minute * 10,
	}
}

func (g gc) gcArtifactRunner(ctx context.Context, runnerID int64, statusChan chan decoratorStatus) error {
	defer close(statusChan)
	statusChan <- decoratorStatus{Daemon: enums.DaemonGcArtifact, Status: enums.TaskCommonStatusDoing}
	runnerObj, err := g.daemonServiceFactory.New().GetGcArtifactRunner(ctx, runnerID)
	if err != nil {
		statusChan <- decoratorStatus{Daemon: enums.DaemonGcArtifact, Status: enums.TaskCommonStatusFailed, Message: fmt.Sprintf("Get gc artifact runner failed: %v", err)}
		return fmt.Errorf("Get gc artifact runner failed: %v", err)
	}

	namespaceService := g.namespaceServiceFactory.New()

	deleteArtifactWithNamespaceChanOnce.Do(g.deleteArtifactWithNamespace)
	deleteArtifactCheckChanOnce.Do(g.deleteArtifactCheck)
	deleteArtifactChanOnce.Do(g.deleteArtifact)

	if runnerObj.NamespaceID != nil {
		deleteArtifactWithNamespaceChan <- artifactWithNamespaceTask{RunnerID: runnerID, NamespaceID: ptr.To(runnerObj.NamespaceID)}
	} else {
		var namespaceCurIndex int64
		for {
			namespaceObjs, err := namespaceService.FindWithCursor(ctx, pagination, namespaceCurIndex)
			if err != nil {
				return err
			}
			for _, ns := range namespaceObjs {
				deleteArtifactWithNamespaceChan <- artifactWithNamespaceTask{RunnerID: runnerID, NamespaceID: ns.ID}
			}
			if len(namespaceObjs) < pagination {
				break
			}
			namespaceCurIndex = namespaceObjs[len(namespaceObjs)-1].ID
		}
	}
	return nil
}

type artifactWithNamespaceTask struct {
	RunnerID    int64
	NamespaceID int64
}

var deleteArtifactWithNamespaceChan = make(chan artifactWithNamespaceTask, 100)

var deleteArtifactWithNamespaceChanOnce = sync.Once{}

func (g gc) deleteArtifactWithNamespace() {
	ctx := log.Logger.WithContext(context.Background())
	repositoryService := g.repositoryServiceFactory.New()
	artifactService := g.artifactServiceFactory.New()
	go func() {
		for task := range deleteArtifactWithNamespaceChan {
			var repositoryCurIndex int64
			timeTarget := time.Now().Add(-1 * g.config.Daemon.Gc.Retention)
			for {
				repositoryObjs, err := repositoryService.FindAll(ctx, task.NamespaceID, pagination, repositoryCurIndex)
				if err != nil {
					log.Error().Err(err).Int64("namespaceID", task.NamespaceID).Msg("List repository failed")
					continue
				}
				for _, repositoryObj := range repositoryObjs {
					var artifactCurIndex int64
					for {
						artifactObjs, err := artifactService.FindWithLastPull(ctx, repositoryObj.ID, timeTarget, pagination, artifactCurIndex)
						if err != nil {
							log.Error().Err(err).Msg("List artifact failed")
							continue
						}
						for _, a := range artifactObjs {
							deleteArtifactCheckChan <- artifactTask{RunnerID: task.RunnerID, Artifact: ptr.To(a)}
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

type artifactTask struct {
	RunnerID int64
	Artifact models.Artifact
}

var deleteArtifactCheckChan = make(chan artifactTask, 100)

var deleteArtifactCheckChanOnce = sync.Once{}

func (g gc) deleteArtifactCheck() {
	ctx := log.Logger.WithContext(context.Background())
	artifactService := g.artifactServiceFactory.New()
	tagService := g.tagServiceFactory.New()
	go func() {
		for task := range deleteArtifactChan {
			// 1. check manifest referrer associate with another artifact
			if task.Artifact.ReferrerID != nil {
				continue
			}
			// 2. check tag associate with this artifact
			_, err := tagService.GetByArtifactID(ctx, task.Artifact.RepositoryID, task.Artifact.ID)
			if err != nil {
				if !errors.Is(err, gorm.ErrRecordNotFound) {
					log.Error().Err(err).Int64("repositoryID", task.Artifact.RepositoryID).Int64("artifactID", task.Artifact.ID).Msg("Get tag by artifact failed")
				}
				continue
			}
			// 3. check manifest index associate with this artifact
			err = artifactService.IsArtifactAssociatedWithArtifact(ctx, task.Artifact.ID)
			if err != nil {
				if !errors.Is(err, gorm.ErrRecordNotFound) {
					log.Error().Err(err).Int64("repositoryID", task.Artifact.RepositoryID).Int64("artifactID", task.Artifact.ID).Msg("Get manifest associated with manifest index failed")
				}
				continue
			}
			// 4. delete the artifact that referrer to this artifact
			delArtifacts, err := artifactService.GetReferrers(ctx, task.Artifact.RepositoryID, task.Artifact.Digest, nil)
			if err != nil {
				log.Error().Err(err).Int64("repositoryID", task.Artifact.RepositoryID).Int64("artifactID", task.Artifact.ID).Msg("Get artifact referrers failed")
				continue
			}
			for _, a := range delArtifacts {
				deleteArtifactChan <- artifactTask{RunnerID: task.RunnerID, Artifact: ptr.To(a)}
			}
			deleteArtifactChan <- task
		}
	}()
}

var deleteArtifactChan = make(chan artifactTask, 100)

var deleteArtifactChanOnce = sync.Once{}

func (g gc) deleteArtifact() {
	ctx := log.Logger.WithContext(context.Background())
	go func() {
		for task := range deleteArtifactChan {
			err := query.Q.Transaction(func(tx *query.Query) error {
				err := g.artifactServiceFactory.New(tx).DeleteByID(ctx, task.Artifact.ID)
				if err != nil {
					return err
				}
				err = g.daemonServiceFactory.New(tx).CreateGcArtifactRecords(ctx, []*models.DaemonGcArtifactRecord{{
					RunnerID: task.RunnerID,
					Digest:   task.Artifact.Digest,
				}})
				if err != nil {
					return err
				}
				log.Info().Str("artifact", task.Artifact.Digest).Msg("Delete artifact success")
				return nil
			})
			if err != nil {
				log.Error().Err(err).Interface("blob", task).Msgf("Delete blob failed: %v", err)
			}
		}
	}()
}
