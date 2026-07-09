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

package scan

import (
	"context"
	"log/slog"
	"path"
	"reflect"
	"sync"
	"time"

	"github.com/tidwall/gjson"
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/background/daemon"
	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	reporegistry "github.com/go-sigma/sigma/pkg/dal/repository/registry"
	repouser "github.com/go-sigma/sigma/pkg/dal/repository/user"
	"github.com/go-sigma/sigma/pkg/infra/workq"
	"github.com/go-sigma/sigma/pkg/service/token"
	"github.com/go-sigma/sigma/pkg/storage"
	"github.com/go-sigma/sigma/pkg/utils"
)

func init() {
	utils.PanicIf(daemon.Daemons.Register(path.Base(reflect.TypeFor[factory]().PkgPath()), &factory{}))
}

type factory struct{}

type params struct {
	dig.In

	Config             *config.Configuration
	HandlerRegistry    workq.HandlerRegistry
	ArtifactRepository reporegistry.ArtifactRepository
	UserRepository     repouser.UserRepository
	TokenSvc           token.Service
	StorageDriver      storage.StorageDriver
}

// Initialize initializes the sbom scan daemon, which is used to scan the sbom of the artifacts.
func (f factory) Initialize(digCon *dig.Container) error {
	var p params
	if err := digCon.Invoke(func(deps params) { p = deps }); err != nil {
		return err
	}
	return p.HandlerRegistry.Register(enums.DaemonSbom, workq.Consumer{
		Handler:     decorator(runnerSbom, p),
		Concurrency: 10,
		Timeout:     time.Minute * 10,
	})
}

// decoratorArtifactStatus is a status for decorator
type decoratorArtifactStatus struct {
	Daemon  enums.Daemon
	Status  enums.TaskCommonStatus
	Raw     []byte
	Result  []byte
	Stdout  []byte
	Stderr  []byte
	Message string
}

// decorator is a decorator for scan task runners
func decorator(runner func(context.Context, params, *models.Artifact, chan decoratorArtifactStatus) error, p params) func(context.Context, []byte) error {
	return func(ctx context.Context, payload []byte) error {
		artifactRepository := p.ArtifactRepository

		id := gjson.GetBytes(payload, "artifact_id").String()
		artifact, err := artifactRepository.Get(ctx, id)
		if err != nil {
			return err
		}

		var waitAllEvents = &sync.WaitGroup{}
		waitAllEvents.Add(1)

		var statusChan = make(chan decoratorArtifactStatus, 1)
		go func() {
			defer waitAllEvents.Done()
			var err error
			for status := range statusChan {
				switch status.Daemon {
				case enums.DaemonVulnerability:
					err = artifactRepository.UpdateVulnerability(ctx, id,
						map[string]any{
							query.ArtifactVulnerability.Raw.ColumnName().String():     status.Raw,
							query.ArtifactVulnerability.Result.ColumnName().String():  status.Result,
							query.ArtifactVulnerability.Status.ColumnName().String():  status.Status,
							query.ArtifactVulnerability.Stdout.ColumnName().String():  status.Stdout,
							query.ArtifactVulnerability.Stderr.ColumnName().String():  status.Stderr,
							query.ArtifactVulnerability.Message.ColumnName().String(): status.Message,
						},
					)
				case enums.DaemonSbom:
					err = artifactRepository.UpdateSbom(ctx,
						id,
						map[string]any{
							query.ArtifactSbom.Raw.ColumnName().String():     status.Raw,
							query.ArtifactSbom.Result.ColumnName().String():  status.Result,
							query.ArtifactSbom.Status.ColumnName().String():  status.Status,
							query.ArtifactSbom.Stdout.ColumnName().String():  status.Stdout,
							query.ArtifactSbom.Stderr.ColumnName().String():  status.Stderr,
							query.ArtifactSbom.Message.ColumnName().String(): status.Message,
						},
					)
				default:
					continue
				}
				if err != nil {
					slog.Error("update artifact status failed", "err", err)
				}
			}
		}()

		err = runner(ctx, p, artifact, statusChan)
		if err != nil {
			return err
		}

		waitAllEvents.Wait()

		return nil
	}
}
