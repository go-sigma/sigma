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
	"encoding/json"
	"fmt"
	"regexp"
	"sync"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/dal/dao"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/modules/workq"
	"github.com/go-sigma/sigma/pkg/modules/workq/definition"
	"github.com/go-sigma/sigma/pkg/storage"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

// deleteTagWithNamespace -> deleteTagWithRepository -> deleteTagCheckPattern -> deleteTag

func init() {
	workq.TopicHandlers[enums.DaemonGcTag.String()] = definition.Consumer{
		Handler:     decorator(enums.DaemonGcTag),
		MaxRetry:    6,
		Concurrency: 10,
		Timeout:     time.Minute * 10,
	}
}

// tagWithNamespaceTask ...
type tagWithNamespaceTask struct {
	Runner      models.DaemonGcTagRunner
	NamespaceID int64
}

// tagWithRepositoryTask ...
type tagWithRepositoryTask struct {
	Runner       models.DaemonGcTagRunner
	RepositoryID int64
}

// tagTask ...
type tagTask struct {
	Runner models.DaemonGcTagRunner
	Tag    models.Tag
}

type gcTag struct {
	namespaceServiceFactory  dao.NamespaceServiceFactory
	repositoryServiceFactory dao.RepositoryServiceFactory
	tagServiceFactory        dao.TagServiceFactory
	artifactServiceFactory   dao.ArtifactServiceFactory
	blobServiceFactory       dao.BlobServiceFactory
	daemonServiceFactory     dao.DaemonServiceFactory
	storageDriverFactory     storage.StorageDriverFactory
	config                   configs.Configuration

	deleteTagWithNamespaceChan      chan tagWithNamespaceTask
	deleteTagWithNamespaceChanOnce  *sync.Once
	deleteTagWithRepositoryChan     chan tagWithRepositoryTask
	deleteTagWithRepositoryChanOnce *sync.Once
	deleteTagCheckPatternChan       chan tagTask
	deleteTagCheckPatternChanOnce   *sync.Once
	deleteTagChan                   chan tagTask
	deleteTagChanOnce               *sync.Once

	waitAllDone *sync.WaitGroup
}

// Run ...
func (g gcTag) Run(ctx context.Context, runnerID int64, statusChan chan decoratorStatus) error {
	defer close(statusChan)
	statusChan <- decoratorStatus{Daemon: enums.DaemonGcTag, Status: enums.TaskCommonStatusDoing, Started: true}
	runnerObj, err := g.daemonServiceFactory.New().GetGcTagRunner(ctx, runnerID)
	if err != nil {
		statusChan <- decoratorStatus{Daemon: enums.DaemonGcTag, Status: enums.TaskCommonStatusFailed, Message: fmt.Sprintf("Get gc tag runner failed: %v", err), Ended: true}
		return fmt.Errorf("get gc tag runner failed: %v", err)
	}

	if ptr.To(runnerObj.Rule.RetentionRuleType) != enums.RetentionRuleTypeDay && ptr.To(runnerObj.Rule.RetentionRuleType) != enums.RetentionRuleTypeQuantity {
		log.Error().Err(err).Interface("RetentionRuleType", ptr.To(runnerObj.Rule.RetentionRuleType))
		return fmt.Errorf("gc tag rule retention type is invalid: %v", ptr.To(runnerObj.Rule.RetentionRuleType))
	}

	namespaceService := g.namespaceServiceFactory.New()

	g.deleteTagWithNamespaceChanOnce.Do(g.deleteTagWithNamespace)
	g.deleteTagWithRepositoryChanOnce.Do(g.deleteTagWithRepository)
	g.deleteTagCheckPatternChanOnce.Do(g.deleteTagCheckPattern)
	g.deleteTagChanOnce.Do(g.deleteTag)
	g.waitAllDone.Add(4)

	if runnerObj.Rule.NamespaceID != nil {
		g.deleteTagWithNamespaceChan <- tagWithNamespaceTask{Runner: ptr.To(runnerObj), NamespaceID: ptr.To(runnerObj.Rule.NamespaceID)}
	} else {
		var namespaceCurIndex int64
		for {
			namespaceObjs, err := namespaceService.FindWithCursor(ctx, pagination, namespaceCurIndex)
			if err != nil {
				return err
			}
			for _, nsObj := range namespaceObjs {
				g.deleteTagWithNamespaceChan <- tagWithNamespaceTask{Runner: ptr.To(runnerObj), NamespaceID: nsObj.ID}
			}
			if len(namespaceObjs) < pagination {
				break
			}
			namespaceCurIndex = namespaceObjs[len(namespaceObjs)-1].ID
		}
	}
	close(g.deleteTagWithNamespaceChan)
	g.waitAllDone.Wait()

	statusChan <- decoratorStatus{Daemon: enums.DaemonGcTag, Status: enums.TaskCommonStatusSuccess, Ended: true}

	return nil
}

func (g gcTag) deleteTagWithNamespace() {
	ctx := log.Logger.WithContext(context.Background())
	repositoryService := g.repositoryServiceFactory.New()
	go func() {
		defer g.waitAllDone.Done()
		for task := range g.deleteTagWithNamespaceChan {
			var repositoryCurIndex int64
			for {
				repositoryObjs, err := repositoryService.FindAll(ctx, task.NamespaceID, pagination, repositoryCurIndex)
				if err != nil {
					log.Error().Err(err).Int64("namespaceID", task.NamespaceID).Msg("List repository failed")
					continue
				}
				for _, repositoryObj := range repositoryObjs {
					g.deleteTagWithRepositoryChan <- tagWithRepositoryTask{Runner: task.Runner, RepositoryID: repositoryObj.ID}
				}
				if len(repositoryObjs) < pagination {
					break
				}
				repositoryCurIndex = repositoryObjs[len(repositoryObjs)-1].ID
			}
		}
		close(g.deleteTagWithRepositoryChan)
	}()
}

func (g gcTag) deleteTagWithRepository() {
	ctx := log.Logger.WithContext(context.Background())
	tagService := g.tagServiceFactory.New()
	go func() {
		defer g.waitAllDone.Done()
		for task := range g.deleteTagWithRepositoryChan {
			var artifactCurIndex int64
			for {
				var tagObjs []*models.Tag
				var err error
				if ptr.To(task.Runner.Rule.RetentionRuleType) == enums.RetentionRuleTypeQuantity {
					tagObjs, err = tagService.FindWithQuantityCursor(ctx, task.RepositoryID, int(ptr.To(task.Runner.Rule.RetentionRuleAmount)), pagination, artifactCurIndex)
				} else if ptr.To(task.Runner.Rule.RetentionRuleType) == enums.RetentionRuleTypeDay {
					tagObjs, err = tagService.FindWithDayCursor(ctx, task.RepositoryID, int(ptr.To(task.Runner.Rule.RetentionRuleAmount)), pagination, artifactCurIndex)
				}
				if err != nil {
					log.Error().Err(err).Msg("List artifact failed")
					continue
				}
				for _, tagObj := range tagObjs {
					g.deleteTagCheckPatternChan <- tagTask{Runner: task.Runner, Tag: ptr.To(tagObj)}
				}
				if len(tagObjs) < pagination {
					break
				}
				artifactCurIndex = tagObjs[len(tagObjs)-1].ID
			}
		}
		close(g.deleteTagCheckPatternChan)
	}()
}

func (g gcTag) deleteTagCheckPattern() {
	go func() {
		defer g.waitAllDone.Done()
		for task := range g.deleteTagCheckPatternChan {
			if len(task.Runner.Rule.RetentionPattern) == 0 {
				g.deleteTagChan <- tagTask{Runner: task.Runner, Tag: task.Tag}
				continue
			}
			var patternPayload types.RetentionPatternPayload
			err := json.Unmarshal(task.Runner.Rule.RetentionPattern, &patternPayload)
			if err != nil {
				log.Error().Err(err).Str("pattern", string(task.Runner.Rule.RetentionPattern)).Msg("Unmarshal payload failed")
				g.deleteTagChan <- tagTask{Runner: task.Runner, Tag: task.Tag}
				continue
			}
			if len(patternPayload.Patterns) == 0 {
				log.Error().Err(err).Msg("Unmarshal pattern payload length is zero")
				g.deleteTagChan <- tagTask{Runner: task.Runner, Tag: task.Tag}
				continue
			}
			var matched bool
			for _, pattern := range patternPayload.Patterns {
				if regexp.MustCompile(pattern).MatchString(task.Tag.Name) {
					matched = true
					break
				}
			}
			if !matched { // every pattern not match this tag, should delete the tag
				g.deleteTagChan <- tagTask{Runner: task.Runner, Tag: task.Tag}
			}
		}
		close(g.deleteTagChan)
	}()
}

func (g gcTag) deleteTag() {
	ctx := log.Logger.WithContext(context.Background())
	tagService := g.tagServiceFactory.New()
	go func() {
		defer g.waitAllDone.Done()
		for task := range g.deleteTagChan {
			err := tagService.DeleteByID(ctx, task.Tag.ID)
			if err != nil {
				log.Error().Err(err).Int64("id", task.Tag.ID).Msg("Delete tag by id failed")
			}
		}
	}()
}