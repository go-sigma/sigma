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
	"github.com/tidwall/gjson"

	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/dal/dao"
	"github.com/go-sigma/sigma/pkg/storage"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

const pagination = 1000

// decoratorStatus is a status for decorator
type decoratorStatus struct {
	Daemon  enums.Daemon
	Status  enums.TaskCommonStatus
	Message string
	Started bool
	Ended   bool
}

// decorator is a decorator for daemon gc task runners
func decorator(daemon enums.Daemon) func(context.Context, []byte) error {
	return func(ctx context.Context, payload []byte) error {
		ctx = log.Logger.WithContext(ctx)
		id := gjson.GetBytes(payload, "runner_id").Int()

		var gc = initGc(daemon)
		if gc == nil {
			return fmt.Errorf("daemon %s not support", daemon.String())
		}

		daemonService := dao.NewDaemonServiceFactory().New()

		var statusChan = make(chan decoratorStatus, 1)
		var waitAllEvents = &sync.WaitGroup{}
		waitAllEvents.Add(1)
		go func() {
			defer waitAllEvents.Done()

			var startedAt time.Time

			var err error
			for status := range statusChan {
				var updates = map[string]any{
					"status":  status.Status,
					"message": status.Message,
				}
				if status.Started {
					startedAt = time.Now()
					updates["started_at"] = startedAt
				}
				if status.Ended {
					endedAt := time.Now()
					updates["ended_at"] = endedAt
					updates["duration"] = endedAt.Sub(startedAt).Milliseconds()
				}
				switch status.Daemon {
				case enums.DaemonGcArtifact:
					err = daemonService.UpdateGcArtifactRunner(ctx, id, updates)
				case enums.DaemonGcRepository:
					err = daemonService.UpdateGcRepositoryRunner(ctx, id, updates)
				case enums.DaemonGcBlob:
					err = daemonService.UpdateGcRepositoryRunner(ctx, id, updates)
				case enums.DaemonGcTag:
					err = daemonService.UpdateGcTagRunner(ctx, id, updates)
				default:
					continue
				}
				if err != nil {
					log.Error().Err(err).Msg("Update gc builder status failed")
				}
			}
		}()

		err := gc.Run(ctx, id, statusChan)
		if err != nil {
			return fmt.Errorf("gc runner(%s) failed: %v", daemon.String(), err)
		}

		waitAllEvents.Wait()

		return nil
	}
}

// Runner ...
type Runner interface {
	// Run ...
	Run(ctx context.Context, runnerID int64, statusChan chan decoratorStatus) error
}

func initGc(daemon enums.Daemon) Runner {
	switch daemon {
	case enums.DaemonGcArtifact:
		return &gcArtifact{
			namespaceServiceFactory:  dao.NewNamespaceServiceFactory(),
			repositoryServiceFactory: dao.NewRepositoryServiceFactory(),
			artifactServiceFactory:   dao.NewArtifactServiceFactory(),
			blobServiceFactory:       dao.NewBlobServiceFactory(),
			daemonServiceFactory:     dao.NewDaemonServiceFactory(),
			storageDriverFactory:     storage.NewStorageDriverFactory(),
			config:                   ptr.To(configs.GetConfiguration()),

			deleteArtifactWithNamespaceChan:     make(chan artifactWithNamespaceTask, 100),
			deleteArtifactWithNamespaceChanOnce: &sync.Once{},
			deleteArtifactCheckChan:             make(chan artifactTask, 100),
			deleteArtifactCheckChanOnce:         &sync.Once{},
			deleteArtifactChan:                  make(chan artifactTask, 100),
			deleteArtifactChanOnce:              &sync.Once{},

			waitAllDone: &sync.WaitGroup{},
		}
	case enums.DaemonGcRepository:
		return &gcRepository{
			namespaceServiceFactory:  dao.NewNamespaceServiceFactory(),
			repositoryServiceFactory: dao.NewRepositoryServiceFactory(),
			artifactServiceFactory:   dao.NewArtifactServiceFactory(),
			blobServiceFactory:       dao.NewBlobServiceFactory(),
			daemonServiceFactory:     dao.NewDaemonServiceFactory(),
			storageDriverFactory:     storage.NewStorageDriverFactory(),
			config:                   ptr.To(configs.GetConfiguration()),
		}
	case enums.DaemonGcTag:
		return &gcTag{
			namespaceServiceFactory:  dao.NewNamespaceServiceFactory(),
			repositoryServiceFactory: dao.NewRepositoryServiceFactory(),
			artifactServiceFactory:   dao.NewArtifactServiceFactory(),
			blobServiceFactory:       dao.NewBlobServiceFactory(),
			daemonServiceFactory:     dao.NewDaemonServiceFactory(),
			storageDriverFactory:     storage.NewStorageDriverFactory(),
			config:                   ptr.To(configs.GetConfiguration()),
		}
	case enums.DaemonGcBlob:
		return &gcBlob{
			namespaceServiceFactory:  dao.NewNamespaceServiceFactory(),
			repositoryServiceFactory: dao.NewRepositoryServiceFactory(),
			artifactServiceFactory:   dao.NewArtifactServiceFactory(),
			blobServiceFactory:       dao.NewBlobServiceFactory(),
			daemonServiceFactory:     dao.NewDaemonServiceFactory(),
			storageDriverFactory:     storage.NewStorageDriverFactory(),
			config:                   ptr.To(configs.GetConfiguration()),

			deleteBlobChan:     make(chan blobTask, 100),
			deleteBlobChanOnce: &sync.Once{},
		}
	default:
		return nil
	}
}
