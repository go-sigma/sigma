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
	"io"
	"log/slog"

	"github.com/distribution/distribution/v3"
	"github.com/opencontainers/go-digest"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/dal/query"
	repobuilder "github.com/go-sigma/sigma/pkg/dal/repository/builder"
	reporegistry "github.com/go-sigma/sigma/pkg/dal/repository/registry"
	"github.com/go-sigma/sigma/pkg/storage"
	"github.com/go-sigma/sigma/pkg/utils"
)

func runnerTagHandler(p params) func(ctx context.Context, data []byte) error {
	return func(ctx context.Context, data []byte) error {
		var payload api.DaemonTagPushedPayload
		err := json.Unmarshal(data, &payload)
		if err != nil {
			return fmt.Errorf("unmarshal payload failed: %v", err)
		}
		r := runnerTag{
			builderRepository:  p.BuilderRepository,
			tagRepository:      p.TagRepository,
			artifactRepository: p.ArtifactRepository,
			storageDriver:      p.StorageDriver,
		}
		return r.run(ctx, payload)
	}
}

type runnerTag struct {
	builderRepository  repobuilder.BuilderRepository
	tagRepository      reporegistry.TagRepository
	artifactRepository reporegistry.ArtifactRepository
	storageDriver      storage.StorageDriver
}

func (r runnerTag) run(ctx context.Context, payload api.DaemonTagPushedPayload) error {
	tagRepository := r.tagRepository
	tagObj, err := tagRepository.GetByName(ctx, payload.RepositoryID, payload.Tag)
	if err != nil {
		slog.Error("get tag by name failed", "err", err, "repository_id", payload.RepositoryID, "tag", payload.Tag)
		return fmt.Errorf("get tag by name failed: %v", err)
	}
	artifactRepository := r.artifactRepository
	artifactObj, err := artifactRepository.Get(ctx, tagObj.ArtifactID)
	if err != nil {
		slog.Error("get artifact by id failed", "err", err, "artifact_id", tagObj.ArtifactID)
		return fmt.Errorf("get artifact by id failed: %v", err)
	}

	dgst, err := digest.Parse(artifactObj.Digest)
	if err != nil {
		slog.Error("parse artifact digest failed", "err", err, "artifact_id", tagObj.ArtifactID, "digest", artifactObj.Digest)
		return fmt.Errorf("parse artifact digest failed: %v", err)
	}
	reader, err := r.storageDriver.Reader(ctx, utils.GenManifestPathByDigest(dgst))
	if err != nil {
		slog.Error("read manifest failed", "err", err, "artifact_id", tagObj.ArtifactID, "digest", artifactObj.Digest)
		return fmt.Errorf("read manifest failed: %v", err)
	}
	defer reader.Close()
	raw, err := io.ReadAll(reader)
	if err != nil {
		slog.Error("read manifest failed", "err", err, "artifact_id", tagObj.ArtifactID, "digest", artifactObj.Digest)
		return fmt.Errorf("read manifest failed: %v", err)
	}

	manifest, descriptor, err := distribution.UnmarshalManifest(artifactObj.ContentType, raw)
	if err != nil {
		slog.Error("unmarshal manifest failed", "err", err, "artifact_id", tagObj.ArtifactID, "content_type", artifactObj.ContentType)
		return fmt.Errorf("unmarshal manifest failed: %v", err)
	}

	slog.Info("unmarshal manifest success", "descriptor", descriptor, "raw", string(raw), "ref", manifest.References())

	builderIDStr, ok := descriptor.Annotations["org.opencontainers.sigma.builder_id"]
	if !ok {
		slog.Error("annotation not have specific key 'org.opencontainers.sigma.builder_id'")
		return nil
	}
	runnerIDStr, ok := descriptor.Annotations["org.opencontainers.sigma.runner_id"]
	if !ok {
		slog.Error("annotation not have specific key 'org.opencontainers.sigma.runner_id'")
		return nil
	}

	builderRepository := r.builderRepository
	err = builderRepository.UpdateRunner(ctx, builderIDStr, runnerIDStr, map[string]any{
		query.BuilderRunner.Tag.ColumnName().String(): tagObj.Name,
	})
	if err != nil {
		slog.Error("runner update tag failed", "err", err)
		return fmt.Errorf("runner update tag failed: %v", err)
	}

	return nil
}
