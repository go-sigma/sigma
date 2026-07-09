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

package registry_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	reporegistry "github.com/go-sigma/sigma/pkg/dal/repository/registry"
	"github.com/go-sigma/sigma/pkg/testkit"
	"github.com/go-sigma/sigma/pkg/utils/uuid"
)

type registryFixture struct {
	Namespace  *models.Namespace
	Repository *models.Repository

	RepositoryRepository reporegistry.RepositoryRepository
	ArtifactRepository   reporegistry.ArtifactRepository
	BlobRepository       reporegistry.BlobRepository
	TagRepository        reporegistry.TagRepository
}

func newRegistryFixture(t *testing.T) registryFixture {
	t.Helper()

	testkit.InitRepository(t)
	ctx := t.Context()
	namespaceObj := &models.Namespace{
		ID:         uuid.NewV7String(),
		Name:       "registry-" + uuid.NewV7String()[:8],
		Visibility: enums.VisibilityPrivate,
	}
	repositoryObj := &models.Repository{
		ID:          uuid.NewV7String(),
		NamespaceID: namespaceObj.ID,
		Name:        namespaceObj.Name + "/repo",
	}
	require.NoError(t, query.Q.Namespace.WithContext(ctx).Create(namespaceObj))
	require.NoError(t, query.Q.Repository.WithContext(ctx).Create(repositoryObj))

	return registryFixture{
		Namespace:            namespaceObj,
		Repository:           repositoryObj,
		RepositoryRepository: reporegistry.NewRepositoryRepository(),
		ArtifactRepository:   reporegistry.NewArtifactRepository(),
		BlobRepository:       reporegistry.NewBlobRepository(),
		TagRepository:        reporegistry.NewTagRepository(),
	}
}

func registryPagination() api.Pagination {
	page := 1
	limit := 10
	return api.Pagination{Page: &page, Limit: &limit}
}

func registrySort(column string, method enums.SortMethod) api.Sortable {
	return api.Sortable{Sort: &column, Method: &method}
}

func newRegistryArtifact(namespaceID, repositoryID, digest string) *models.Artifact {
	return &models.Artifact{
		ID:           uuid.NewV7String(),
		NamespaceID:  namespaceID,
		RepositoryID: repositoryID,
		Digest:       digest,
		Size:         123,
		BlobsSize:    123,
		ContentType:  "application/vnd.oci.image.manifest.v1+json",
		Type:         enums.ArtifactTypeImage,
	}
}
