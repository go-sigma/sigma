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

package tagpushed

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/distribution/distribution/v3"
	"github.com/rs/zerolog/log"

	"github.com/go-sigma/sigma/pkg/dal/dao"
	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/modules/workq"
	"github.com/go-sigma/sigma/pkg/modules/workq/definition"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/types/enums"
)

func init() {
	workq.TopicHandlers[enums.DaemonTagPushed.String()] = definition.Consumer{
		Handler:     builderRunner,
		MaxRetry:    6,
		Concurrency: 10,
		Timeout:     time.Minute * 10,
	}
}

func builderRunner(ctx context.Context, data []byte) error {
	var payload types.DaemonTagPushedPayload
	err := json.Unmarshal(data, &payload)
	if err != nil {
		return fmt.Errorf("Unmarshal payload failed: %v", err)
	}
	r := runner{
		builderServiceFactory:  dao.NewBuilderServiceFactory(),
		tagServiceFactory:      dao.NewTagServiceFactory(),
		artifactServiceFactory: dao.NewArtifactServiceFactory(),
	}
	return r.run(ctx, payload)
}

type runner struct {
	builderServiceFactory  dao.BuilderServiceFactory
	tagServiceFactory      dao.TagServiceFactory
	artifactServiceFactory dao.ArtifactServiceFactory
}

func (r runner) run(ctx context.Context, payload types.DaemonTagPushedPayload) error {
	tagService := r.tagServiceFactory.New()
	tagObj, err := tagService.GetByName(ctx, payload.RepositoryID, payload.Tag)
	if err != nil {
		log.Error().Err(err).Int64("repository_id", payload.RepositoryID).Str("tag", payload.Tag).Msg("Get tag by name failed")
		return fmt.Errorf("Get tag by name failed: %v", err)
	}
	artifactService := r.artifactServiceFactory.New()
	artifactObj, err := artifactService.Get(ctx, tagObj.ArtifactID)
	if err != nil {
		log.Error().Err(err).Int64("artifact_id", tagObj.ArtifactID).Msg("Get artifact by id failed")
		return fmt.Errorf("Get artifact by id failed: %v", err)
	}

	manifest, descriptor, err := distribution.UnmarshalManifest(artifactObj.ContentType, artifactObj.Raw)
	if err != nil {
		log.Error().Err(err).Int64("artifact_id", tagObj.ArtifactID).Str("content_type", artifactObj.ContentType).Msg("Unmarshal manifest failed")
		return fmt.Errorf("Unmarshal manifest failed: %v", err)
	}

	log.Info().Interface("descriptor", descriptor).Str("raw", string(artifactObj.Raw)).Interface("ref", manifest.References()).Msg("Unmarshal manifest success")

	builderIDStr, ok := descriptor.Annotations["org.opencontainers.sigma.builder_id"]
	if !ok {
		log.Error().Msg("Annotation not have specific key 'org.opencontainers.sigma.builder_id'")
		return nil
	}
	builderID, err := strconv.ParseInt(builderIDStr, 10, 64)
	if err != nil {
		log.Error().Err(err).Msg("Annotation not have specific key 'org.opencontainers.sigma.builder_id'")
		return fmt.Errorf("Annotation 'org.opencontainers.sigma.builder_id' convert failed: %v", err)
	}
	runnerIDStr, ok := descriptor.Annotations["org.opencontainers.sigma.runner_id"]
	if !ok {
		log.Error().Msg("Annotation not have specific key 'org.opencontainers.sigma.runner_id'")
		return nil
	}
	runnerID, err := strconv.ParseInt(runnerIDStr, 10, 64)
	if err != nil {
		log.Error().Err(err).Msg("Annotation not have specific key 'org.opencontainers.sigma.runner_id'")
		return fmt.Errorf("Annotation 'org.opencontainers.sigma.runner_id' convert failed: %v", err)
	}

	builderService := r.builderServiceFactory.New()
	err = builderService.UpdateRunner(ctx, builderID, runnerID, map[string]any{
		query.BuilderRunner.Tag.ColumnName().String(): tagObj.Name,
	})
	if err != nil {
		log.Error().Err(err).Msg("Runner update tag failed")
		return fmt.Errorf("Runner update tag failed: %v", err)
	}

	return nil
}
