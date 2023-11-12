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

	"github.com/rs/zerolog/log"

	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/dal/dao"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/modules/workq"
	"github.com/go-sigma/sigma/pkg/modules/workq/definition"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

// deleteRepositoryWithNamespace -> deleteRepositoryCheckEmpty -> deleteRepository

func init() {
	workq.TopicHandlers[enums.DaemonGcRepository.String()] = definition.Consumer{
		Handler:     decorator(enums.DaemonGcRepository),
		MaxRetry:    6,
		Concurrency: 10,
		Timeout:     time.Minute * 10,
	}
}

type repositoryWithNamespaceTask struct {
	Runner      models.DaemonGcRepositoryRunner
	NamespaceID int64
}

// repositoryTask ...
type repositoryTask struct {
	Runner       models.DaemonGcRepositoryRunner
	RepositoryID int64
}

type gcRepository struct {
	ctx    context.Context
	config configs.Configuration

	namespaceServiceFactory  dao.NamespaceServiceFactory
	repositoryServiceFactory dao.RepositoryServiceFactory
	tagServiceFactory        dao.TagServiceFactory
	daemonServiceFactory     dao.DaemonServiceFactory

	deleteRepositoryWithNamespaceChan       chan repositoryWithNamespaceTask
	deleteRepositoryWithNamespaceChanOnce   *sync.Once
	deleteRepositoryCheckRepositoryChan     chan repositoryTask
	deleteRepositoryCheckRepositoryChanOnce *sync.Once
	deleteRepositoryChan                    chan repositoryTask
	deleteRepositoryChanOnce                *sync.Once

	waitAllDone *sync.WaitGroup
}

// Run ...
func (g gcRepository) Run(ctx context.Context, runnerID int64, statusChan chan decoratorStatus) error {
	defer close(statusChan)
	statusChan <- decoratorStatus{Daemon: enums.DaemonGcRepository, Status: enums.TaskCommonStatusDoing}
	runnerObj, err := g.daemonServiceFactory.New().GetGcRepositoryRunner(ctx, runnerID)
	if err != nil {
		statusChan <- decoratorStatus{Daemon: enums.DaemonGcRepository, Status: enums.TaskCommonStatusFailed, Message: fmt.Sprintf("Get gc repository runner failed: %v", err)}
		return fmt.Errorf("get gc repository runner failed: %v", err)
	}

	g.deleteRepositoryWithNamespaceChanOnce.Do(g.deleteRepositoryWithNamespace)
	g.deleteRepositoryCheckRepositoryChanOnce.Do(g.deleteRepositoryCheck)
	g.deleteRepositoryChanOnce.Do(g.deleteRepository)
	g.waitAllDone.Add(3)

	namespaceService := g.namespaceServiceFactory.New()

	if runnerObj.Rule.NamespaceID != nil {
		g.deleteRepositoryWithNamespaceChan <- repositoryWithNamespaceTask{Runner: ptr.To(runnerObj), NamespaceID: ptr.To(runnerObj.Rule.NamespaceID)}
	} else {
		var namespaceCurIndex int64
		for {
			namespaceObjs, err := namespaceService.FindWithCursor(ctx, pagination, namespaceCurIndex)
			if err != nil {
				return err
			}
			for _, nsObj := range namespaceObjs {
				g.deleteRepositoryWithNamespaceChan <- repositoryWithNamespaceTask{Runner: ptr.To(runnerObj), NamespaceID: nsObj.ID}
			}
			if len(namespaceObjs) < pagination {
				break
			}
			namespaceCurIndex = namespaceObjs[len(namespaceObjs)-1].ID
		}
	}
	close(g.deleteRepositoryWithNamespaceChan)
	g.waitAllDone.Wait()

	statusChan <- decoratorStatus{Daemon: enums.DaemonGcTag, Status: enums.TaskCommonStatusSuccess, Ended: true}

	return nil
}

func (g gcRepository) deleteRepositoryWithNamespace() {
	repositoryService := g.repositoryServiceFactory.New()
	go func() {
		defer g.waitAllDone.Done()
		for task := range g.deleteRepositoryWithNamespaceChan {
			var repositoryCurIndex int64
			for {
				repositoryObjs, err := repositoryService.FindAll(g.ctx, task.NamespaceID, pagination, repositoryCurIndex)
				if err != nil {
					log.Error().Err(err).Int64("namespaceID", task.NamespaceID).Msg("List repository failed")
					continue
				}
				for _, repositoryObj := range repositoryObjs {
					g.deleteRepositoryCheckRepositoryChan <- repositoryTask{Runner: task.Runner, RepositoryID: repositoryObj.ID}
				}
				if len(repositoryObjs) < pagination {
					break
				}
				repositoryCurIndex = repositoryObjs[len(repositoryObjs)-1].ID
			}
		}
		close(g.deleteRepositoryCheckRepositoryChan)
	}()
}

func (g gcRepository) deleteRepositoryCheck() {
	tagService := g.tagServiceFactory.New()
	repositoryService := g.repositoryServiceFactory.New()
	go func() {
		defer g.waitAllDone.Done()
		for task := range g.deleteRepositoryCheckRepositoryChan {
			count, err := tagService.CountByRepository(g.ctx, task.RepositoryID)
			if err != nil {
				log.Error().Err(err).Int64("RepositoryID", task.RepositoryID).Msg("Get repository tag count failed")
				continue
			}
			if count > 0 {
				continue
			}
			if task.Runner.Rule.RetentionDay == 0 {
				repositoryObj, err := repositoryService.Get(g.ctx, task.RepositoryID)
				if err != nil {
					log.Error().Err(err).Int64("RepositoryID", task.RepositoryID).Msg("Get repository by id failed")
					continue
				}
				if !repositoryObj.UpdatedAt.Before(time.Now().Add(-1 * 24 * time.Duration(task.Runner.Rule.RetentionDay) * time.Hour)) {
					continue
				}
			}
			g.deleteRepositoryChan <- task
		}
	}()
}

func (g gcRepository) deleteRepository() {
	repositoryService := g.repositoryServiceFactory.New()
	go func() {
		defer g.waitAllDone.Done()
		for task := range g.deleteRepositoryChan {
			err := repositoryService.DeleteByID(g.ctx, task.RepositoryID)
			if err != nil {
				log.Error().Err(err).Int64("RepositoryID", task.RepositoryID).Msg("Delete repository by id failed")
			}
		}
	}()
}
