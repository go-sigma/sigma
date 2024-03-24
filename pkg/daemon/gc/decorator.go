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
	"github.com/go-sigma/sigma/pkg/modules/workq"
	"github.com/go-sigma/sigma/pkg/modules/workq/definition"
	"github.com/go-sigma/sigma/pkg/storage"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

const pagination = 10

// decoratorStatus is a status for decorator
type decoratorStatus struct {
	Daemon  enums.Daemon
	Status  enums.TaskCommonStatus
	Message string
	Started bool
	Ended   bool
	Updates map[string]any
}

// decoratorWebhook used for webhook trigger
type decoratorWebhook struct {
	NamespaceID *int64
	Meta        types.WebhookPayload
	WebhookObj  any
}

type inject struct {
	daemonServiceFactory dao.DaemonServiceFactory
	storageDriverFactory storage.StorageDriverFactory
	producerClient       definition.WorkQueueProducer
}

// Runner ...
type Runner interface {
	// Run ...
	Run(runnerID int64) error
}

// decorator is a decorator for daemon gc task runners
func decorator(daemon enums.Daemon, injects ...inject) func(context.Context, []byte) error { // nolint: unparam
	return func(ctx context.Context, payload []byte) error {
		ctx = log.Logger.WithContext(ctx)
		id := gjson.GetBytes(payload, "runner_id").Int()

		producerClient := workq.ProducerClient
		daemonServiceFactory := dao.NewDaemonServiceFactory()
		if len(injects) > 0 {
			ij := injects[0]
			if ij.producerClient != nil {
				producerClient = ij.producerClient
			}
			if ij.daemonServiceFactory != nil {
				daemonServiceFactory = ij.daemonServiceFactory
			}
		}

		var runnerChan = make(chan decoratorStatus, 3)
		var webhookChan = make(chan decoratorWebhook, 3)
		var gc = initGc(ctx, daemon, runnerChan, webhookChan, injects...)
		if gc == nil {
			return fmt.Errorf("daemon %s not support", daemon.String())
		}

		daemonService := daemonServiceFactory.New()

		var waitAllEvents = &sync.WaitGroup{}
		waitAllEvents.Add(2)
		go func() {
			defer waitAllEvents.Done()

			var err error
			var startedAt time.Time

			for status := range runnerChan {
				var updates = map[string]any{
					"status":  status.Status,
					"message": status.Message,
				}
				if status.Started {
					startedAt = time.Now()
					updates["started_at"] = startedAt.UnixMilli()
				}
				if status.Ended {
					endedAt := time.Now()
					updates["ended_at"] = endedAt.UnixMilli()
					updates["duration"] = endedAt.Sub(startedAt).Milliseconds()
				}
				if len(status.Updates) != 0 {
					for key, val := range status.Updates {
						updates[key] = val
					}
				}
				switch status.Daemon {
				case enums.DaemonGcRepository:
					err = daemonService.UpdateGcRepositoryRunner(ctx, id, updates)
				case enums.DaemonGcTag:
					err = daemonService.UpdateGcTagRunner(ctx, id, updates)
				case enums.DaemonGcArtifact:
					err = daemonService.UpdateGcArtifactRunner(ctx, id, updates)
				case enums.DaemonGcBlob:
					err = daemonService.UpdateGcBlobRunner(ctx, id, updates)
				default:
					continue
				}
				if err != nil {
					log.Error().Err(err).Msg("Update gc builder status failed")
				}
			}
		}()

		go func() {
			defer waitAllEvents.Done()
			for webhook := range webhookChan {
				err := triggerWebhook(ctx, webhook, producerClient)
				if err != nil {
					log.Error().Err(err).Msg("Webhook event produce failed")
				}
			}
		}()

		err := gc.Run(id)
		if err != nil {
			return fmt.Errorf("gc runner(%s) failed: %v", daemon.String(), err)
		}

		waitAllEvents.Wait()

		return nil
	}
}

func initGc(ctx context.Context, daemon enums.Daemon, runnerChan chan decoratorStatus, webhookChan chan decoratorWebhook, injects ...inject) Runner {
	switch daemon {
	case enums.DaemonGcRepository:
		return &gcRepository{
			ctx:    log.Logger.WithContext(ctx),
			config: ptr.To(configs.GetConfiguration()),

			daemonServiceFactory:     dao.NewDaemonServiceFactory(),
			namespaceServiceFactory:  dao.NewNamespaceServiceFactory(),
			repositoryServiceFactory: dao.NewRepositoryServiceFactory(),
			tagServiceFactory:        dao.NewTagServiceFactory(),

			deleteRepositoryWithNamespaceChan:       make(chan repositoryWithNamespaceTask, pagination),
			deleteRepositoryWithNamespaceChanOnce:   &sync.Once{},
			deleteRepositoryCheckRepositoryChan:     make(chan repositoryTask, pagination),
			deleteRepositoryCheckRepositoryChanOnce: &sync.Once{},
			deleteRepositoryChan:                    make(chan repositoryTask, pagination),
			deleteRepositoryChanOnce:                &sync.Once{},
			collectRecordChan:                       make(chan repositoryTaskCollectRecord, pagination),
			collectRecordChanOnce:                   &sync.Once{},

			runnerChan:  runnerChan,
			webhookChan: webhookChan,

			waitAllDone: &sync.WaitGroup{},
		}
	case enums.DaemonGcArtifact:
		return &gcArtifact{
			ctx:    log.Logger.WithContext(ctx),
			config: ptr.To(configs.GetConfiguration()),

			namespaceServiceFactory:  dao.NewNamespaceServiceFactory(),
			repositoryServiceFactory: dao.NewRepositoryServiceFactory(),
			tagServiceFactory:        dao.NewTagServiceFactory(),
			artifactServiceFactory:   dao.NewArtifactServiceFactory(),
			daemonServiceFactory:     dao.NewDaemonServiceFactory(),

			deleteArtifactWithNamespaceChan:     make(chan artifactWithNamespaceTask, pagination),
			deleteArtifactWithNamespaceChanOnce: &sync.Once{},
			deleteArtifactCheckChan:             make(chan artifactTask, pagination),
			deleteArtifactCheckChanOnce:         &sync.Once{},
			deleteArtifactChan:                  make(chan artifactTask, pagination),
			deleteArtifactChanOnce:              &sync.Once{},
			collectRecordChan:                   make(chan artifactTaskCollectRecord, pagination),
			collectRecordChanOnce:               &sync.Once{},

			runnerChan:  runnerChan,
			webhookChan: webhookChan,

			waitAllDone: &sync.WaitGroup{},
		}
	case enums.DaemonGcTag:
		return &gcTag{
			ctx:    log.Logger.WithContext(ctx),
			config: ptr.To(configs.GetConfiguration()),

			daemonServiceFactory:     dao.NewDaemonServiceFactory(),
			namespaceServiceFactory:  dao.NewNamespaceServiceFactory(),
			repositoryServiceFactory: dao.NewRepositoryServiceFactory(),
			tagServiceFactory:        dao.NewTagServiceFactory(),
			artifactServiceFactory:   dao.NewArtifactServiceFactory(),
			blobServiceFactory:       dao.NewBlobServiceFactory(),

			deleteTagWithNamespaceChan:      make(chan tagWithNamespaceTask, pagination),
			deleteTagWithNamespaceChanOnce:  &sync.Once{},
			deleteTagWithRepositoryChan:     make(chan tagWithRepositoryTask, pagination),
			deleteTagWithRepositoryChanOnce: &sync.Once{},
			deleteTagCheckPatternChan:       make(chan tagTask, pagination),
			deleteTagCheckPatternChanOnce:   &sync.Once{},
			deleteTagChan:                   make(chan tagTask, pagination),
			deleteTagChanOnce:               &sync.Once{},
			collectRecordChan:               make(chan tagTaskCollectRecord, pagination),
			collectRecordChanOnce:           &sync.Once{},

			runnerChan:  runnerChan,
			webhookChan: webhookChan,

			waitAllDone: &sync.WaitGroup{},
		}
	case enums.DaemonGcBlob:
		runner := &gcBlob{
			ctx:    log.Logger.WithContext(ctx),
			config: ptr.To(configs.GetConfiguration()),

			blobServiceFactory:   dao.NewBlobServiceFactory(),
			daemonServiceFactory: dao.NewDaemonServiceFactory(),
			storageDriverFactory: storage.NewStorageDriverFactory(),

			deleteBlobChan:        make(chan blobTask, pagination),
			deleteBlobChanOnce:    &sync.Once{},
			collectRecordChan:     make(chan blobTaskCollectRecord, pagination),
			collectRecordChanOnce: &sync.Once{},

			runnerChan:  runnerChan,
			webhookChan: webhookChan,

			waitAllDone: &sync.WaitGroup{},
		}
		if len(injects) > 0 {
			ij := injects[0]
			if ij.storageDriverFactory != nil {
				runner.storageDriverFactory = ij.storageDriverFactory
			}
		}
		return runner
	default:
		return nil
	}
}

func triggerWebhook(ctx context.Context, webhook decoratorWebhook, producerClient definition.WorkQueueProducer) error {
	err := producerClient.Produce(ctx, enums.DaemonWebhook, types.DaemonWebhookPayload{
		NamespaceID:  webhook.NamespaceID,
		Action:       webhook.Meta.Action,
		Type:         enums.WebhookTypeSend,
		ResourceType: webhook.Meta.ResourceType,
		Payload:      utils.MustMarshal(webhook.WebhookObj),
	}, definition.ProducerOption{})
	if err != nil {
		return fmt.Errorf("Webhook event produce failed: %v", err)
	}
	return nil
}
