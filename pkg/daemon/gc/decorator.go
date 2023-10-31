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

	"github.com/rs/zerolog/log"
	"github.com/tidwall/gjson"

	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/dal/dao"
	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/storage"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

// decoratorStatus is a status for decorator
type decoratorStatus struct {
	Daemon  enums.Daemon
	Status  enums.TaskCommonStatus
	Message string
}

// decorator is a decorator for daemon task runners
func decorator(daemon enums.Daemon) func(context.Context, []byte) error {
	return func(ctx context.Context, payload []byte) error {
		ctx = log.Logger.WithContext(ctx)
		id := gjson.GetBytes(payload, "runner_id").Int()

		var g = gc{
			namespaceServiceFactory:  dao.NewNamespaceServiceFactory(),
			repositoryServiceFactory: dao.NewRepositoryServiceFactory(),
			artifactServiceFactory:   dao.NewArtifactServiceFactory(),
			blobServiceFactory:       dao.NewBlobServiceFactory(),
			daemonServiceFactory:     dao.NewDaemonServiceFactory(),
			storageDriverFactory:     storage.NewStorageDriverFactory(),
			config:                   ptr.To(configs.GetConfiguration()),
		}

		var statusChan = make(chan decoratorStatus, 1)
		var waitAllEvents = &sync.WaitGroup{}
		waitAllEvents.Add(1)
		go func() {
			defer waitAllEvents.Done()
			var err error
			for status := range statusChan {
				switch status.Daemon {
				case enums.DaemonGcRepository:
					err = g.daemonServiceFactory.New().UpdateGcRepositoryRunner(ctx, id,
						map[string]any{
							query.DaemonGcRepositoryRunner.Status.ColumnName().String():  status.Status,
							query.DaemonGcRepositoryRunner.Message.ColumnName().String(): status.Message,
						},
					)
				case enums.DaemonGcBlob:
					err = g.daemonServiceFactory.New().UpdateGcRepositoryRunner(ctx, id,
						map[string]any{
							query.DaemonGcRepositoryRunner.Status.ColumnName().String():  status.Status,
							query.DaemonGcRepositoryRunner.Message.ColumnName().String(): status.Message,
						},
					)
				default:
					continue
				}
				if err != nil {
					log.Error().Err(err).Msg("Update gc builder status failed")
				}
			}
		}()

		var err error
		switch daemon {
		case enums.DaemonGcRepository:
			err = g.gcRepositoryRunner(ctx, id, statusChan)
		case enums.DaemonGcBlob:
			err = g.gcBlobRunner(ctx, id, statusChan)
		case enums.DaemonGcArtifact:
			err = g.gcArtifactRunner(ctx, id, statusChan)
		case enums.DaemonGcTag:
			err = g.gcTagRunner(ctx, id, statusChan)
		default:
			return fmt.Errorf("Daemon %s is not support", daemon)
		}

		if err != nil {
			return fmt.Errorf("Gc runner(%s) failed: %v", daemon.String(), err)
		}

		waitAllEvents.Wait()

		return nil
	}
}
