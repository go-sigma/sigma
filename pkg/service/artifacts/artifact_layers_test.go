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

package artifacts

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/dal/models"
	reponamespacemocks "github.com/go-sigma/sigma/pkg/dal/repository/namespace/mocks"
	reporegistrymocks "github.com/go-sigma/sigma/pkg/dal/repository/registry/mocks"
	"github.com/go-sigma/sigma/pkg/logger"
	"github.com/go-sigma/sigma/pkg/server/errcode"
	storagemocks "github.com/go-sigma/sigma/pkg/storage/mocks"
)

func TestGetArtifactLayersBuildsTheManifestLayers(t *testing.T) {
	ctrl := gomock.NewController(t)
	// storage keys the manifest by the artifact digest, so the test needs a well formed one.
	digest := "sha256:" + strings.Repeat("a", 64)
	namespaceRepository := reponamespacemocks.NewMockNamespaceRepository(ctrl)
	repositoryRepository := reporegistrymocks.NewMockRepositoryRepository(ctrl)
	artifactRepository := reporegistrymocks.NewMockArtifactRepository(ctrl)
	storageDriver := storagemocks.NewMockStorageDriver(ctrl)

	namespaceRepository.EXPECT().Get(gomock.Any(), "namespace-1").
		Return(&models.Namespace{ID: "namespace-1", Name: "library"}, nil)
	repositoryRepository.EXPECT().Get(gomock.Any(), "repository-1").
		Return(&models.Repository{ID: "repository-1", NamespaceID: "namespace-1", Name: "alpine"}, nil)
	artifactRepository.EXPECT().GetByDigestWithReports(gomock.Any(), "repository-1", digest).
		Return(&models.Artifact{
			ID:          "artifact-1",
			Digest:      digest,
			ContentType: "application/vnd.oci.image.manifest.v1+json",
			Size:        1234,
			BlobsSize:   5678,
			ConfigRaw: []byte(`{
				"os": "linux",
				"architecture": "amd64",
				"history": [
					{"created": "2025-01-02T03:04:05Z", "created_by": "ADD rootfs.tar /"},
					{"created": "2025-01-02T03:04:06Z", "created_by": "CMD [\"/bin/sh\"]", "empty_layer": true},
					{"created": "2025-01-02T03:04:07Z", "created_by": "RUN /install.sh"}
				],
				"rootfs": {"diff_ids": ["sha256:diff-1", "sha256:diff-2"]}
			}`),
			Sbom: models.ArtifactSbom{Result: []byte(`{"distro": {"name": "alpine", "version": "3.18.2"}}`)},
		}, nil)
	storageDriver.EXPECT().Reader(gomock.Any(), gomock.Any()).
		Return(io.NopCloser(strings.NewReader(`{
			"mediaType": "application/vnd.oci.image.manifest.v1+json",
			"layers": [
				{"mediaType": "application/vnd.oci.image.layer.v1.tar+gzip", "size": 10, "digest": "sha256:layer-1"},
				{"mediaType": "application/vnd.oci.image.layer.v1.tar+gzip", "size": 20, "digest": "sha256:layer-2"}
			]
		}`)), nil)

	layers, err := (&service{
		RepoNs:       namespaceRepository,
		RepoRegistry: repositoryRepository,
		RepoArtifact: artifactRepository,
		Storage:      storageDriver,
	}).GetArtifactLayers(t.Context(), "namespace-1", "repository-1", digest)

	require.NoError(t, err)
	require.Equal(t, "library/alpine", layers.Repository)
	require.Equal(t, "linux", layers.OS)
	require.Equal(t, "amd64", layers.Architecture)
	require.Equal(t, "alpine 3.18.2", layers.Distro)
	require.Equal(t, int64(5678), layers.BlobSize)
	require.Len(t, layers.Layers, 2)
	require.Equal(t, 1, layers.Layers[0].Index)
	require.Equal(t, "sha256:layer-1", layers.Layers[0].Digest)
	require.Equal(t, "ADD rootfs.tar /", layers.Layers[0].Command)
	require.Equal(t, "sha256:diff-1", layers.Layers[0].DiffID)
	require.Equal(t, "2025-01-02 03:04:05", layers.Layers[0].CreatedAt)
	require.Equal(t, "RUN /install.sh", layers.Layers[1].Command)
	require.Equal(t, "sha256:diff-2", layers.Layers[1].DiffID)
}

func TestGetArtifactLayersReturnsNotFoundForForeignRepository(t *testing.T) {
	ctrl := gomock.NewController(t)
	namespaceRepository := reponamespacemocks.NewMockNamespaceRepository(ctrl)
	repositoryRepository := reporegistrymocks.NewMockRepositoryRepository(ctrl)

	namespaceRepository.EXPECT().Get(gomock.Any(), "namespace-1").
		Return(&models.Namespace{ID: "namespace-1", Name: "library"}, nil)
	repositoryRepository.EXPECT().Get(gomock.Any(), "repository-1").
		Return(&models.Repository{ID: "repository-1", NamespaceID: "namespace-2", Name: "alpine"}, nil)

	_, err := (&service{
		RepoNs:       namespaceRepository,
		RepoRegistry: repositoryRepository,
	}).GetArtifactLayers(t.Context(), "namespace-1", "repository-1", "sha256:digest")

	require.Error(t, err)
	e, ok := errcode.AsType[errcode.ErrCode](err)
	require.True(t, ok)
	require.Equal(t, errcode.HTTPErrCodeNotFound, e)
}

func TestGetArtifactLayersReturnsArtifactNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	namespaceRepository := reponamespacemocks.NewMockNamespaceRepository(ctrl)
	repositoryRepository := reporegistrymocks.NewMockRepositoryRepository(ctrl)
	artifactRepository := reporegistrymocks.NewMockArtifactRepository(ctrl)

	namespaceRepository.EXPECT().Get(gomock.Any(), "namespace-1").
		Return(&models.Namespace{ID: "namespace-1", Name: "library"}, nil)
	repositoryRepository.EXPECT().Get(gomock.Any(), "repository-1").
		Return(&models.Repository{ID: "repository-1", NamespaceID: "namespace-1", Name: "alpine"}, nil)
	artifactRepository.EXPECT().GetByDigestWithReports(gomock.Any(), "repository-1", "sha256:digest").
		Return(nil, gorm.ErrRecordNotFound)

	_, err := (&service{
		RepoNs:       namespaceRepository,
		RepoRegistry: repositoryRepository,
		RepoArtifact: artifactRepository,
	}).GetArtifactLayers(t.Context(), "namespace-1", "repository-1", "sha256:digest")

	require.Error(t, err)
	e, ok := errcode.AsType[errcode.ErrCode](err)
	require.True(t, ok)
	require.Equal(t, errcode.HTTPErrCodeNotFound, e)
}

func TestLayerHistoriesPairsCommandsWithDiffIDs(t *testing.T) {
	var configObj imageConfigRaw
	configObj.History = []struct {
		Created    string `json:"created"`
		CreatedBy  string `json:"created_by"`
		EmptyLayer bool   `json:"empty_layer"`
	}{
		{Created: "2025-01-02T03:04:05Z", CreatedBy: "ADD rootfs.tar /"},
		{Created: "2025-01-02T03:04:06Z", CreatedBy: "CMD [\"/bin/sh\"]", EmptyLayer: true},
		{Created: "2025-01-02T03:04:07Z", CreatedBy: "RUN /install.sh"},
	}
	configObj.RootFS.DiffIDs = []string{"sha256:diff-1", "sha256:diff-2"}

	histories := layerHistories(configObj)

	require.Len(t, histories, 2)
	require.Equal(t, "ADD rootfs.tar /", histories[0].command)
	require.Equal(t, "sha256:diff-1", histories[0].diffID)
	require.Equal(t, "2025-01-02 03:04:05", histories[0].createdAt)
	require.Equal(t, "RUN /install.sh", histories[1].command)
	require.Equal(t, "sha256:diff-2", histories[1].diffID)
}

func TestLayerHistoriesKeepsLayersWithoutHistory(t *testing.T) {
	require.Empty(t, layerHistories(imageConfigRaw{}))
}

func TestFormatConfigTimeFallsBackToTheRawValue(t *testing.T) {
	logger.SetLevel("error")

	require.Empty(t, formatConfigTime(""))
	require.Equal(t, "not a time", formatConfigTime("not a time"))
	require.Equal(t, "2025-01-02 03:04:05",
		formatConfigTime("2025-01-02T03:04:05.123456789Z"))
}

func TestDistroNameJoinsNameAndVersion(t *testing.T) {
	var sbomResult sbomRaw
	require.Empty(t, distroName(sbomResult))

	sbomResult.Distro.Name = "alpine"
	require.Equal(t, "alpine", distroName(sbomResult))

	sbomResult.Distro.Version = "3.18.2"
	require.Equal(t, "alpine 3.18.2", distroName(sbomResult))
	require.False(t, strings.Contains(distroName(sbomResult), "  "))
}
