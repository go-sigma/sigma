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
	"time"

	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/dal/dao"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/modules/workq"
	"github.com/go-sigma/sigma/pkg/modules/workq/definition"
	"github.com/go-sigma/sigma/pkg/storage"
	"github.com/go-sigma/sigma/pkg/types/enums"
)

func init() {
	workq.TopicHandlers[enums.DaemonGcRepository.String()] = definition.Consumer{
		Handler:     decorator(enums.DaemonGcRepository),
		MaxRetry:    6,
		Concurrency: 10,
		Timeout:     time.Minute * 10,
	}
}

type gcRepository struct {
	namespaceServiceFactory  dao.NamespaceServiceFactory
	repositoryServiceFactory dao.RepositoryServiceFactory
	artifactServiceFactory   dao.ArtifactServiceFactory
	blobServiceFactory       dao.BlobServiceFactory
	daemonServiceFactory     dao.DaemonServiceFactory
	storageDriverFactory     storage.StorageDriverFactory
	config                   configs.Configuration
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
	err = query.Q.Transaction(func(tx *query.Query) error {
		repositoryService := g.repositoryServiceFactory.New(tx)
		deletedRepositoryObjs, err := repositoryService.DeleteEmpty(ctx, runnerObj.Rule.NamespaceID)
		if err != nil {
			return err
		}
		daemonService := g.daemonServiceFactory.New(tx)
		daemonLogs := make([]*models.DaemonGcRepositoryRecord, 0, len(deletedRepositoryObjs))
		for _, obj := range deletedRepositoryObjs {
			daemonLogs = append(daemonLogs, &models.DaemonGcRepositoryRecord{RunnerID: runnerID, Repository: obj})
		}
		err = daemonService.CreateGcRepositoryRecords(ctx, daemonLogs)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		statusChan <- decoratorStatus{Daemon: enums.DaemonGcRepository, Status: enums.TaskCommonStatusFailed, Message: fmt.Sprintf("Gc empty repository failed: %v", err)}
		return fmt.Errorf("gc empty repository failed: %v", err)
	}
	statusChan <- decoratorStatus{Daemon: enums.DaemonGcRepository, Status: enums.TaskCommonStatusSuccess}
	return nil
}
