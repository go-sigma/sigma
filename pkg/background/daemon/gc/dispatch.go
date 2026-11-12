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

// Runner runs a single GC task; each GC kind (artifact, repository, tag, blob)
// provides an implementation selected through runnerFactories.
type Runner interface {
	// Run executes the GC task, reporting progress and the final status through runner.
	Run(ctx context.Context, runner *runnerContext) error
}

// runnerFactory builds a fresh Runner for one GC message. A new instance is
// required per message because runners hold per-run state (counters, runner object).
type runnerFactory func(params) Runner

// runnerFactories maps each GC daemon to its constructor. Registration and
// dispatch share this table, so adding a GC kind only needs one entry here.
var runnerFactories = map[enums.Daemon]runnerFactory{
	enums.DaemonGcArtifact:   newGcArtifact,
	enums.DaemonGcRepository: newGcRepository,
	enums.DaemonGcTag:        newGcTag,
	enums.DaemonGcBlob:       newGcBlob,
}

// Initialize initializes the gc daemons, which registers the work queue
// consumers for all gc subtasks.
func (f factory) Initialize(digCon *dig.Container) error {
	var p params
	if err := digCon.Invoke(func(deps params) { p = deps }); err != nil {
		return err
	}
	for daemonType, newRunner := range runnerFactories {
		if err := p.HandlerRegistry.Register(daemonType, workq.Consumer{
			Handler:     handleGC(daemonType, p, newRunner),
			Concurrency: 10,
			Timeout:     time.Minute * 10,
		}); err != nil {
			return err
		}
	}
	return nil
}

func handleGC(daemonType enums.Daemon, p params, newRunner runnerFactory) func(context.Context, []byte) error {
	return func(ctx context.Context, payload []byte) error {
		id := gjson.GetBytes(payload, "runner_id").String()
		runner := newRunnerContext(ctx, id, p.DaemonRepository, p.Producer)
		if err := newRunner(p).Run(ctx, runner); err != nil {
			return fmt.Errorf("gc runner(%s) failed: %w", daemonType.String(), err)
		}
		return nil
	}
}

func newGcArtifact(p params) Runner {
	return &gcArtifact{
		config: p.Config,
		locker: p.Locker,

		namespaceRepository:  p.NamespaceRepository,
		repositoryRepository: p.RepositoryRepository,
		tagRepository:        p.TagRepository,
		artifactRepository:   p.ArtifactRepository,
		daemonRepository:     p.DaemonRepository,
	}
}

func newGcRepository(p params) Runner {
	return &gcRepository{
		config: p.Config,
		locker: p.Locker,

		daemonRepository:     p.DaemonRepository,
		namespaceRepository:  p.NamespaceRepository,
		repositoryRepository: p.RepositoryRepository,
		tagRepository:        p.TagRepository,
	}
}

func newGcTag(p params) Runner {
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
}

func newGcBlob(p params) Runner {
	return &gcBlob{
		config: p.Config,
		locker: p.Locker,

		blobRepository:   p.BlobRepository,
		daemonRepository: p.DaemonRepository,
		storageDriver:    p.StorageDriver,
	}
}

func triggerWebhook(ctx context.Context, meta api.WebhookPayload, obj any, producerClient workq.Producer) error {
	err := producerClient.Produce(ctx, enums.DaemonWebhook, api.DaemonWebhookPayload{
		Action:       meta.Action,
		Type:         enums.WebhookTypeSend,
		ResourceType: meta.ResourceType,
		Payload:      utils.MustMarshal(obj),
	})
	if err != nil {
		return fmt.Errorf("webhook event produce failed: %w", err)
	}
	return nil
}
