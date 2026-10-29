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
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/server/errcode"
	artifactsmocks "github.com/go-sigma/sigma/pkg/service/artifacts/mocks"
)

func TestGetArtifactLayers(t *testing.T) {
	ctrl := gomock.NewController(t)
	service := artifactsmocks.NewMockArtifactService(ctrl)
	service.EXPECT().GetArtifactLayers(gomock.Any(), "namespace-1", "repository-1", "sha256:abc").
		Return(&api.ArtifactLayers{
			Digest:     "sha256:abc",
			Repository: "library/alpine",
			Layers: []api.ArtifactLayer{
				{Index: 1, Digest: "sha256:layer-1", Size: 10},
			},
		}, nil)
	recorder, c := newArtifactContext()

	(&handler{ArtifactSvc: service}).GetArtifactLayers(c, &api.GetArtifactLayersRequest{
		NamespaceID:  "namespace-1",
		RepositoryID: "repository-1",
		Digest:       "sha256:abc",
	})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Contains(t, recorder.Body.String(), `"repository":"library/alpine"`)
	require.Contains(t, recorder.Body.String(), `"digest":"sha256:layer-1"`)
}

func TestGetArtifactLayersReturnsNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	service := artifactsmocks.NewMockArtifactService(ctrl)
	service.EXPECT().GetArtifactLayers(gomock.Any(), "namespace-1", "repository-1", "sha256:abc").
		Return(nil, errcode.HTTPErrCodeNotFound.Detail("artifact not found"))
	recorder, c := newArtifactContext()

	(&handler{ArtifactSvc: service}).GetArtifactLayers(c, &api.GetArtifactLayersRequest{
		NamespaceID:  "namespace-1",
		RepositoryID: "repository-1",
		Digest:       "sha256:abc",
	})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusNotFound, recorder.Code)
}
