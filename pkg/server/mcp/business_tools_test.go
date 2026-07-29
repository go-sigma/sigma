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

package mcpserver

import (
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/dal/models"
	svcartifacts "github.com/go-sigma/sigma/pkg/service/artifacts"
	svcnamespaces "github.com/go-sigma/sigma/pkg/service/namespaces"
	svcrepositories "github.com/go-sigma/sigma/pkg/service/repositories"
	svctags "github.com/go-sigma/sigma/pkg/service/tags"
)

func TestNamespaceTools(t *testing.T) {
	ctrl := gomock.NewController(t)
	namespaceSvc := svcnamespaces.NewMockNamespaceService(ctrl)
	server := &Server{
		authorizer:   fakeAuthorizer{namespaceAllowed: true},
		namespaceSvc: namespaceSvc,
	}
	ctx := withUser(t.Context(), &models.User{ID: "user-1"})

	namespaceSvc.EXPECT().
		ListNamespaces(gomock.Any(), "user-1", nil, api.Pagination{}, api.Sortable{}).
		Return([]*models.Namespace{{ID: "namespace-1"}}, int64(1), nil)
	result, err := server.namespaceList(ctx, toolRequest(nil))
	require.NoError(t, err)
	require.Equal(t, int64(1), result.(map[string]any)["total"])

	request := api.PostNamespaceRequest{Name: "sigma"}
	namespaceSvc.EXPECT().
		CreateNamespace(gomock.Any(), "user-1", request).
		Return(&models.Namespace{ID: "namespace-1"}, nil)
	result, err = server.namespaceCreate(ctx, toolRequest(map[string]any{"name": "sigma"}))
	require.NoError(t, err)
	require.Equal(t, map[string]string{"id": "namespace-1"}, result)

	namespaceSvc.EXPECT().
		DeleteNamespace(gomock.Any(), "user-1", "namespace-1").
		Return(nil)
	result, err = server.namespaceDelete(ctx, toolRequest(map[string]any{"id": "namespace-1"}))
	require.NoError(t, err)
	require.Equal(t, okResponse(), result)
}

func TestRepositoryTools(t *testing.T) {
	ctrl := gomock.NewController(t)
	repositorySvc := svcrepositories.NewMockRepositoryService(ctrl)
	server := &Server{
		authorizer:    fakeAuthorizer{namespaceAllowed: true, repositoryAllowed: true},
		repositorySvc: repositorySvc,
	}
	ctx := withUser(t.Context(), &models.User{ID: "user-1"})

	request := api.CreateRepositoryRequest{NamespaceID: "namespace-1", Name: "sigma/demo"}
	repositorySvc.EXPECT().
		CreateRepository(gomock.Any(), "user-1", request).
		Return(&models.Repository{ID: "repository-1"}, nil)
	result, err := server.repositoryCreate(ctx, toolRequest(map[string]any{
		"namespace_id": "namespace-1",
		"name":         "sigma/demo",
	}))
	require.NoError(t, err)
	require.Equal(t, map[string]string{"id": "repository-1"}, result)

	repositorySvc.EXPECT().
		GetRepository(gomock.Any(), "repository-1").
		Return(&models.Repository{ID: "repository-1"}, nil)
	result, err = server.repositoryGet(ctx, toolRequest(map[string]any{"id": "repository-1"}))
	require.NoError(t, err)
	require.Equal(t, &models.Repository{ID: "repository-1"}, result)
}

func TestTagManifestRawGet(t *testing.T) {
	ctrl := gomock.NewController(t)
	tagSvc := svctags.NewMockTagService(ctrl)
	server := &Server{
		authorizer: fakeAuthorizer{repositoryAllowed: true},
		tagSvc:     tagSvc,
	}
	ctx := withUser(t.Context(), &models.User{ID: "user-1"})

	tagSvc.EXPECT().
		GetArtifactRaw(gomock.Any(), "sha256:digest").
		Return([]byte(`{"schemaVersion":2}`), nil)
	result, err := server.tagManifestRawGet(ctx, toolRequest(map[string]any{
		"repository_id": "repository-1",
		"digest":        "sha256:digest",
	}))
	require.NoError(t, err)
	require.Equal(t, map[string]string{"raw": `{"schemaVersion":2}`}, result)
}

func TestArtifactTools(t *testing.T) {
	ctrl := gomock.NewController(t)
	repositorySvc := svcrepositories.NewMockRepositoryService(ctrl)
	artifactSvc := svcartifacts.NewMockArtifactService(ctrl)
	server := &Server{
		authorizer:    fakeAuthorizer{repositoryAllowed: true},
		repositorySvc: repositorySvc,
		artifactSvc:   artifactSvc,
	}
	ctx := withUser(t.Context(), &models.User{ID: "user-1"})

	repositorySvc.EXPECT().
		GetRepositoryByName(gomock.Any(), "sigma/demo").
		Return(&models.Repository{ID: "repository-1"}, nil)
	artifactSvc.EXPECT().
		GetArtifact(gomock.Any(), "sigma/demo", "sha256:digest").
		Return(&models.Artifact{ID: "artifact-1"}, nil)
	result, err := server.artifactGet(ctx, toolRequest(map[string]any{
		"repository": "sigma/demo",
		"digest":     "sha256:digest",
	}))
	require.NoError(t, err)
	require.Equal(t, &models.Artifact{ID: "artifact-1"}, result)

	repositorySvc.EXPECT().
		GetRepositoryByName(gomock.Any(), "sigma/demo").
		Return(&models.Repository{ID: "repository-1"}, nil)
	artifactSvc.EXPECT().
		DeleteArtifact(gomock.Any(), "sigma/demo", "sha256:digest").
		Return(nil)
	result, err = server.artifactDelete(ctx, toolRequest(map[string]any{
		"repository": "sigma/demo",
		"digest":     "sha256:digest",
	}))
	require.NoError(t, err)
	require.Equal(t, okResponse(), result)
}

func toolRequest(arguments map[string]any) mcp.CallToolRequest {
	return mcp.CallToolRequest{
		Params: mcp.CallToolParams{Arguments: arguments},
	}
}
