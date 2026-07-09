// Copyright 2026 sigma
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

package app

import (
	"context"
	"log/slog"

	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/config"
	reponamespace "github.com/go-sigma/sigma/pkg/dal/repository/namespace"
	reporegistry "github.com/go-sigma/sigma/pkg/dal/repository/registry"
)

type cachePrewarmParams struct {
	dig.In

	Config               *config.Configuration
	NamespaceRepository  reponamespace.NamespaceRepository
	RepositoryRepository reporegistry.RepositoryRepository
}

// PrewarmMetadataCache warms namespace and repository metadata cache.
func PrewarmMetadataCache(ctx context.Context, digCon *dig.Container) {
	if err := digCon.Invoke(func(params cachePrewarmParams) {
		prewarmMetadataCache(ctx, params)
	}); err != nil {
		slog.Warn("invoke metadata cache prewarm failed", "err", err)
	}
}

func prewarmMetadataCache(ctx context.Context, params cachePrewarmParams) {
	cfg := params.Config.Cache.Prewarm
	if !cfg.IsEnabled() || cfg.Limit <= 0 {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, cfg.Timeout)
	defer cancel()

	namespaces, err := params.NamespaceRepository.FindRecentlyUpdated(ctx, cfg.Limit)
	if err != nil {
		slog.Warn("prewarm namespace cache failed", "err", err)
	} else {
		slog.Info("prewarmed namespace cache", "count", len(namespaces))
	}

	repositories, err := params.RepositoryRepository.FindRecentlyUpdated(ctx, cfg.Limit)
	if err != nil {
		slog.Warn("prewarm repository cache failed", "err", err)
	} else {
		slog.Info("prewarmed repository cache", "count", len(repositories))
	}
}
