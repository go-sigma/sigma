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
	"encoding/json"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/go-sigma/sigma/pkg/dal/dao"
	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/modules/workq"
	"github.com/go-sigma/sigma/pkg/modules/workq/definition"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/types/enums"
)

func init() {
	workq.TopicHandlers[enums.DaemonArtifactPushed] = definition.Consumer{
		Handler: func(ctx context.Context, data []byte) error {
			var payload types.DaemonArtifactPushedPayload
			err := json.Unmarshal(data, &payload)
			if err != nil {
				return fmt.Errorf("Unmarshal payload failed: %v", err)
			}
			r := runnerArtifact{
				namespaceServiceFactory:  dao.NewNamespaceServiceFactory(),
				repositoryServiceFactory: dao.NewRepositoryServiceFactory(),
				tagServiceFactory:        dao.NewTagServiceFactory(),
				artifactServiceFactory:   dao.NewArtifactServiceFactory(),
			}
			return r.run(ctx, payload)
		},
		MaxRetry:    1,
		Concurrency: 10,
		Timeout:     time.Minute * 10,
	}
}

type runnerArtifact struct {
	namespaceServiceFactory  dao.NamespaceServiceFactory
	repositoryServiceFactory dao.RepositoryServiceFactory
	tagServiceFactory        dao.TagServiceFactory
	artifactServiceFactory   dao.ArtifactServiceFactory
}

func (r runnerArtifact) run(ctx context.Context, payload types.DaemonArtifactPushedPayload) error {
	ctx = log.Logger.WithContext(ctx)

	repositoryService := r.repositoryServiceFactory.New()
	repositoryObj, err := repositoryService.Get(ctx, payload.RepositoryID)
	if err != nil {
		return err
	}
	artifactService := r.artifactServiceFactory.New()
	namespaceSize, err := artifactService.GetNamespaceSize(ctx, repositoryObj.NamespaceID)
	if err != nil {
		return err
	}
	namespaceService := r.namespaceServiceFactory.New()
	err = namespaceService.UpdateByID(ctx, repositoryObj.NamespaceID, map[string]any{
		query.Namespace.Size.ColumnName().String(): namespaceSize,
	})
	if err != nil {
		return err
	}
	repositorySize, err := artifactService.GetRepositorySize(ctx, repositoryObj.ID)
	if err != nil {
		return err
	}
	err = repositoryService.UpdateRepository(ctx, repositoryObj.ID, map[string]any{
		query.Repository.Size.ColumnName().String(): repositorySize,
	})
	if err != nil {
		log.Error().Err(err).Msg("Update repository failed")
		return err
	}
	return nil
}
