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

	"github.com/hibiken/asynq"

	"github.com/go-sigma/sigma/pkg/daemon"
	"github.com/go-sigma/sigma/pkg/dal/dao"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

func init() {
	utils.PanicIf(daemon.RegisterTask(enums.DaemonGcRepository, gcRepositoryRunner))
}

// gcRepositoryRunner ...
func gcRepositoryRunner(ctx context.Context, task *asynq.Task) error {
	var payload types.DaemonGcRepositoryPayload
	err := json.Unmarshal(task.Payload(), &payload)
	if err != nil {
		return fmt.Errorf("Unmarshal payload failed: %v", err)
	}
	gc := gcRepository{}
	return gc.runner(ctx, payload)
}

type gcRepository struct {
	namespaceServiceFactory  dao.NamespaceServiceFactory
	repositoryServiceFactory dao.RepositoryServiceFactory
	daemonServiceFactory     dao.DaemonServiceFactory
}

func (g gcRepository) runner(ctx context.Context, payload types.DaemonGcRepositoryPayload) error {
	var namespaceID *int64
	if payload.Scope != nil {
		namespaceService := g.namespaceServiceFactory.New()
		namespaceObj, err := namespaceService.GetByName(ctx, ptr.To(payload.Scope))
		if err != nil {
			return err
		}
		namespaceID = ptr.Of(namespaceObj.ID)
	}
	err := query.Q.Transaction(func(tx *query.Query) error {
		repositoryService := g.repositoryServiceFactory.New(tx)
		deletedRepositoryObjs, err := repositoryService.DeleteEmpty(ctx, namespaceID)
		if err != nil {
			return err
		}
		daemonService := g.daemonServiceFactory.New(tx)
		daemonLogs := make([]*models.DaemonLog, 0, len(deletedRepositoryObjs))
		for _, obj := range deletedRepositoryObjs {
			daemonLogs = append(daemonLogs, &models.DaemonLog{
				NamespaceID: namespaceID,
				Type:        enums.DaemonGcRepository,
				Action:      enums.AuditActionDelete,
				Resource:    obj,
				Status:      enums.TaskCommonStatusSuccess,
			})
		}
		err = daemonService.CreateMany(ctx, daemonLogs)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}
