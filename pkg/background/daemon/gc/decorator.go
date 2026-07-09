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
	"path"
	"reflect"
	"time"

	"github.com/tidwall/gjson"
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/background/daemon"
	"github.com/go-sigma/sigma/pkg/config"
	repodaemon "github.com/go-sigma/sigma/pkg/dal/repository/daemon"
	reponamespace "github.com/go-sigma/sigma/pkg/dal/repository/namespace"
	reporegistry "github.com/go-sigma/sigma/pkg/dal/repository/registry"
	"github.com/go-sigma/sigma/pkg/infra/lock"
	"github.com/go-sigma/sigma/pkg/infra/workq"
	"github.com/go-sigma/sigma/pkg/storage"
	"github.com/go-sigma/sigma/pkg/utils"
)

func init() {
	utils.PanicIf(daemon.Daemons.Register(path.Base(reflect.TypeFor[factory]().PkgPath()), &factory{}))
}

type factory struct{}

type params struct {
	dig.In

	Config               *config.Configuration
	HandlerRegistry      workq.HandlerRegistry
	DaemonRepository     repodaemon.DaemonRepository
	NamespaceRepository  reponamespace.NamespaceRepository
	RepositoryRepository reporegistry.RepositoryRepository
	TagRepository        reporegistry.TagRepository
	ArtifactRepository   reporegistry.ArtifactRepository
	BlobRepository       reporegistry.BlobRepository
	StorageDriver        storage.StorageDriver
	Locker               lock.Locker
	Producer             workq.Producer
}

// Initialize initializes the gc daemons, which registers the work queue
// consumers for all gc subtasks.
func (f factory) Initialize(digCon *dig.Container) error {
	var p params
	if err := digCon.Invoke(func(deps params) { p = deps }); err != nil {
		return err
	}
	if err := p.HandlerRegistry.Register(enums.DaemonGcArtifact, workq.Consumer{
		Handler:     handleGC(enums.DaemonGcArtifact, p),
		Concurrency: 10,
		Timeout:     time.Minute * 10,
	}); err != nil {
		return err
	}
	if err := p.HandlerRegistry.Register(enums.DaemonGcRepository, workq.Consumer{
		Handler:     handleGC(enums.DaemonGcRepository, p),
		Concurrency: 10,
		Timeout:     time.Minute * 10,
	}); err != nil {
		return err
	}
	if err := p.HandlerRegistry.Register(enums.DaemonGcTag, workq.Consumer{
		Handler:     handleGC(enums.DaemonGcTag, p),
		Concurrency: 10,
		Timeout:     time.Minute * 10,
	}); err != nil {
		return err
	}
	if err := p.HandlerRegistry.Register(enums.DaemonGcBlob, workq.Consumer{
		Handler:     handleGC(enums.DaemonGcBlob, p),
		Concurrency: 10,
		Timeout:     time.Minute * 10,
	}); err != nil {
		return err
	}
	return nil
}

// decoratorWebhook used for webhook trigger
type decoratorWebhook struct {
	NamespaceID *string
	Meta        api.WebhookPayload
	WebhookObj  any
}

// Runner ...
type Runner interface {
	// Run ...
	Run(ctx context.Context, runner *runnerContext, runnerID string) error
}

func handleGC(daemon enums.Daemon, p params) func(context.Context, []byte) error { // nolint: unparam
	return func(ctx context.Context, payload []byte) error {
		id := gjson.GetBytes(payload, "runner_id").String()
		gc := initGc(daemon, p)
		if gc == nil {
			return fmt.Errorf("daemon %s not support", daemon.String())
		}

		daemonRepository := p.DaemonRepository
		runner := newRunnerContext(ctx, daemon, id, daemonRepository, p.Producer)
		err := gc.Run(ctx, runner, id)
		if err != nil {
			return fmt.Errorf("gc runner(%s) failed: %v", daemon.String(), err)
		}
		return nil
	}
}

func initGc(daemon enums.Daemon, p params) Runner {
	switch daemon {
	case enums.DaemonGcRepository:
		return &gcRepository{
			config: p.Config,
			locker: p.Locker,

			daemonRepository:     p.DaemonRepository,
			namespaceRepository:  p.NamespaceRepository,
			repositoryRepository: p.RepositoryRepository,
			tagRepository:        p.TagRepository,
		}
	case enums.DaemonGcArtifact:
		return &gcArtifact{
			config: p.Config,
			locker: p.Locker,

			namespaceRepository:  p.NamespaceRepository,
			repositoryRepository: p.RepositoryRepository,
			tagRepository:        p.TagRepository,
			artifactRepository:   p.ArtifactRepository,
			daemonRepository:     p.DaemonRepository,
		}
	case enums.DaemonGcTag:
		return &gcTag{
			config: p.Config,
			locker: p.Locker,

			daemonRepository:     p.DaemonRepository,
			namespaceRepository:  p.NamespaceRepository,
			repositoryRepository: p.RepositoryRepository,
			tagRepository:        p.TagRepository,
			artifactRepository:   p.ArtifactRepository,
			blobRepository:       p.BlobRepository,
		}
	case enums.DaemonGcBlob:
		return &gcBlob{
			config: p.Config,
			locker: p.Locker,

			blobRepository:   p.BlobRepository,
			daemonRepository: p.DaemonRepository,
			storageDriver:    p.StorageDriver,
		}
	default:
		return nil
	}
}

func triggerWebhook(ctx context.Context, webhook decoratorWebhook, producerClient workq.Producer) error {
	err := producerClient.Produce(ctx, enums.DaemonWebhook, api.DaemonWebhookPayload{
		NamespaceID:  webhook.NamespaceID,
		Action:       webhook.Meta.Action,
		Type:         enums.WebhookTypeSend,
		ResourceType: webhook.Meta.ResourceType,
		Payload:      utils.MustMarshal(webhook.WebhookObj),
	})
	if err != nil {
		return fmt.Errorf("webhook event produce failed: %v", err)
	}
	return nil
}
