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

package builders

import (
	"context"
	"encoding/base64"
	"io"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/dal/models"
	reporegistry "github.com/go-sigma/sigma/pkg/dal/repository/registry"
	"github.com/go-sigma/sigma/pkg/utils/compress"
)

func TestCreateBuilderValidator(t *testing.T) {
	dockerfile := base64.StdEncoding.EncodeToString([]byte("FROM scratch"))

	require.NoError(t, createBuilderValidator(api.CreateBuilderRequest{
		Source:     enums.BuilderSourceDockerfile,
		Dockerfile: &dockerfile,
	}))
	require.Error(t, createBuilderValidator(api.CreateBuilderRequest{
		Source: enums.BuilderSourceDockerfile,
	}))
	require.Error(t, createBuilderValidator(api.CreateBuilderRequest{}))
}

func TestCompressDockerfile(t *testing.T) {
	data := base64.StdEncoding.EncodeToString([]byte("FROM scratch"))
	compressed, err := compressDockerfile(&data)
	require.NoError(t, err)

	got, err := compress.Decompress(compressed)
	require.NoError(t, err)
	require.Equal(t, "FROM scratch", got)

	_, err = compressDockerfile(new("invalid base64"))
	require.Error(t, err)

	compressed, err = compressDockerfile(nil)
	require.NoError(t, err)
	require.Nil(t, compressed)
}

// cachedRepositoryRepository is a repository implementation that records the cache invalidations so
// the builder service can be tested without a cache backend.
type cachedRepositoryRepository struct {
	reporegistry.RepositoryRepository
	invalidated []*models.Repository
}

func (c *cachedRepositoryRepository) InvalidateRepository(_ context.Context, repositoryObj *models.Repository) error {
	c.invalidated = append(c.invalidated, repositoryObj)
	return nil
}

func (c *cachedRepositoryRepository) InvalidateRepositoryID(_ context.Context, _ string) error {
	return nil
}

func TestInvalidateRepositoryCache(t *testing.T) {
	repositoryObj := &models.Repository{ID: "repository-1", Name: "library/test"}
	cached := &cachedRepositoryRepository{}

	(&service{RepoRegistry: cached}).invalidateRepositoryCache(t.Context(), repositoryObj)
	require.Equal(t, []*models.Repository{repositoryObj}, cached.invalidated)

	// A repository implementation without cache support is skipped instead of panicking.
	(&service{RepoRegistry: struct {
		reporegistry.RepositoryRepository
	}{}}).invalidateRepositoryCache(t.Context(), repositoryObj)

	// A nil repository cannot be invalidated and is ignored.
	(&service{}).invalidateRepositoryCache(t.Context(), nil)
}

func TestRunnerLogReaderForInactiveRunner(t *testing.T) {
	reader, compressed, err := (&service{}).RunnerLogReader(
		t.Context(), "builder-1", "runner-1", enums.BuildStatusPending,
	)
	require.NoError(t, err)
	require.False(t, compressed)
	data, err := io.ReadAll(reader)
	require.NoError(t, err)
	require.Empty(t, data)
}
