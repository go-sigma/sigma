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
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/go-sigma/sigma/pkg/api"
	artifactsmocks "github.com/go-sigma/sigma/pkg/service/artifacts/mocks"
)

func TestListArtifactVulnerabilities(t *testing.T) {
	ctrl := gomock.NewController(t)
	service := artifactsmocks.NewMockArtifactService(ctrl)
	service.EXPECT().ListArtifactVulnerabilities(gomock.Any(), "namespace-1", "repository-1", "sha256:abc").
		Return(&api.ArtifactVulnerabilities{
			Total:    1,
			Critical: 1,
			Items: []api.ArtifactVulnerability{
				{ID: "CVE-2023-4863", Severity: "Critical", Package: "libwebp"},
			},
		}, nil)
	recorder, c := newArtifactContext()

	(&handler{ArtifactSvc: service}).ListArtifactVulnerabilities(c, &api.ListArtifactVulnerabilitiesRequest{
		NamespaceID:  "namespace-1",
		RepositoryID: "repository-1",
		Digest:       "sha256:abc",
	})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Contains(t, recorder.Body.String(), `"id":"CVE-2023-4863"`)
	require.Contains(t, recorder.Body.String(), `"critical":1`)
}

func TestListArtifactVulnerabilitiesReturnsInternalError(t *testing.T) {
	ctrl := gomock.NewController(t)
	service := artifactsmocks.NewMockArtifactService(ctrl)
	service.EXPECT().ListArtifactVulnerabilities(gomock.Any(), "namespace-1", "repository-1", "sha256:abc").
		Return(nil, errors.New("failed"))
	recorder, c := newArtifactContext()

	(&handler{ArtifactSvc: service}).ListArtifactVulnerabilities(c, &api.ListArtifactVulnerabilitiesRequest{
		NamespaceID:  "namespace-1",
		RepositoryID: "repository-1",
		Digest:       "sha256:abc",
	})
	c.Writer.WriteHeaderNow()

	require.Equal(t, http.StatusInternalServerError, recorder.Code)
}
