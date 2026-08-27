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

package pushed

import (
	"context"
	"encoding/json/v2"
	"fmt"
	"log/slog"
	"path"
	"reflect"
	"time"

	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/background/daemon"
	"github.com/go-sigma/sigma/pkg/dal/query"
	repobuilder "github.com/go-sigma/sigma/pkg/dal/repository/builder"
	reponamespace "github.com/go-sigma/sigma/pkg/dal/repository/namespace"
	reporegistry "github.com/go-sigma/sigma/pkg/dal/repository/registry"
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

	HandlerRegistry      workq.HandlerRegistry
	NamespaceRepository  reponamespace.NamespaceRepository
	RepositoryRepository reporegistry.RepositoryRepository
	TagRepository        reporegistry.TagRepository
	ArtifactRepository   reporegistry.ArtifactRepository
	BuilderRepository    repobuilder.BuilderRepository
	StorageDriver        storage.StorageDriver
}

// Initialize initializes the pushed daemon, which is used to update the
// size of namespace/repository after an artifact or tag is pushed.
func (f factory) Initialize(digCon *dig.Container) error {
	var p params
	if err := digCon.Invoke(func(deps params) { p = deps }); err != nil {
		return err
	}
	if err := p.HandlerRegistry.Register(enums.DaemonArtifactPushed, workq.Consumer{
		Handler:     runnerArtifactHandler(p),
		Concurrency: 10,
		Timeout:     time.Minute * 10,
	}); err != nil {
		return err
	}
	if err := p.HandlerRegistry.Register(enums.DaemonTagPushed, workq.Consumer{
		Handler:     runnerTagHandler(p),
		Concurrency: 10,
		Timeout:     time.Minute * 10,
	}); err != nil {
		return err
	}
	return nil
}

func runnerArtifactHandler(p params) func(ctx context.Context, data []byte) error {
	return func(ctx context.Context, data []byte) error {
		var payload api.DaemonArtifactPushedPayload
		err := json.Unmarshal(data, &payload)
		if err != nil {
			return fmt.Errorf("unmarshal payload failed: %v", err)
		}
		r := runnerArtifact{
			namespaceRepository:  p.NamespaceRepository,
			repositoryRepository: p.RepositoryRepository,
			tagRepository:        p.TagRepository,
			artifactRepository:   p.ArtifactRepository,
		}
		return r.run(ctx, payload)
	}
}

type runnerArtifact struct {
	namespaceRepository  reponamespace.NamespaceRepository
	repositoryRepository reporegistry.RepositoryRepository
	tagRepository        reporegistry.TagRepository
	artifactRepository   reporegistry.ArtifactRepository
}

func (r runnerArtifact) run(ctx context.Context, payload api.DaemonArtifactPushedPayload) error {
	repositoryRepository := r.repositoryRepository
	repositoryObj, err := repositoryRepository.Get(ctx, payload.RepositoryID)
	if err != nil {
		return err
	}
	artifactRepository := r.artifactRepository
	namespaceSize, err := artifactRepository.GetNamespaceSize(ctx, repositoryObj.NamespaceID)
	if err != nil {
		return err
	}
	namespaceRepository := r.namespaceRepository
	err = namespaceRepository.UpdateByID(ctx, repositoryObj.NamespaceID, map[string]any{
		query.Namespace.Size.ColumnName().String(): namespaceSize,
	})
	if err != nil {
		return err
	}
	repositorySize, err := artifactRepository.GetRepositorySize(ctx, repositoryObj.ID)
	if err != nil {
		return err
	}
	err = repositoryRepository.UpdateRepository(ctx, repositoryObj.ID, map[string]any{
		query.Repository.Size.ColumnName().String(): repositorySize,
	})
	if err != nil {
		slog.Error("update repository failed", "err", err)
		return err
	}
	return nil
}
