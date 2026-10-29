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
	"bytes"
	"compress/gzip"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/dal/models"
	reponamespacemocks "github.com/go-sigma/sigma/pkg/dal/repository/namespace/mocks"
	reporegistrymocks "github.com/go-sigma/sigma/pkg/dal/repository/registry/mocks"
	storagemocks "github.com/go-sigma/sigma/pkg/storage/mocks"
)

func TestListArtifactVulnerabilitiesReturnsTheScanStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	namespaceRepository := reponamespacemocks.NewMockNamespaceRepository(ctrl)
	repositoryRepository := reporegistrymocks.NewMockRepositoryRepository(ctrl)
	artifactRepository := reporegistrymocks.NewMockArtifactRepository(ctrl)

	namespaceRepository.EXPECT().Get(gomock.Any(), "namespace-1").
		Return(&models.Namespace{ID: "namespace-1", Name: "library"}, nil)
	repositoryRepository.EXPECT().Get(gomock.Any(), "repository-1").
		Return(&models.Repository{ID: "repository-1", NamespaceID: "namespace-1", Name: "alpine"}, nil)
	artifactRepository.EXPECT().GetByDigestWithReports(gomock.Any(), "repository-1", "sha256:digest").
		Return(&models.Artifact{
			ID:     "artifact-1",
			Digest: "sha256:digest",
			Vulnerability: models.ArtifactVulnerability{
				Status: enums.TaskCommonStatusPending,
			},
		}, nil)

	report, err := (&service{
		RepoNs:       namespaceRepository,
		RepoRegistry: repositoryRepository,
		RepoArtifact: artifactRepository,
	}).ListArtifactVulnerabilities(t.Context(), "namespace-1", "repository-1", "sha256:digest")

	require.NoError(t, err)
	require.Equal(t, string(enums.TaskCommonStatusPending), report.Status)
	require.Zero(t, report.Total)
	require.Empty(t, report.Items)
}

func TestListArtifactVulnerabilitiesReadsTheStoredReport(t *testing.T) {
	ctrl := gomock.NewController(t)
	namespaceRepository := reponamespacemocks.NewMockNamespaceRepository(ctrl)
	repositoryRepository := reporegistrymocks.NewMockRepositoryRepository(ctrl)
	artifactRepository := reporegistrymocks.NewMockArtifactRepository(ctrl)
	storageDriver := storagemocks.NewMockStorageDriver(ctrl)

	namespaceRepository.EXPECT().Get(gomock.Any(), "namespace-1").
		Return(&models.Namespace{ID: "namespace-1", Name: "library"}, nil)
	repositoryRepository.EXPECT().Get(gomock.Any(), "repository-1").
		Return(&models.Repository{ID: "repository-1", NamespaceID: "namespace-1", Name: "alpine"}, nil)
	artifactRepository.EXPECT().GetByDigestWithReports(gomock.Any(), "repository-1", "sha256:digest").
		Return(&models.Artifact{
			ID:     "artifact-1",
			Digest: "sha256:digest",
			Vulnerability: models.ArtifactVulnerability{
				Status: enums.TaskCommonStatusSuccess,
				Raw:    []byte("sigma/vulnerabilities/artifact-1.json.gz"),
			},
		}, nil)
	storageDriver.EXPECT().Reader(gomock.Any(), "sigma/vulnerabilities/artifact-1.json.gz").
		Return(io.NopCloser(bytes.NewReader(gzipReport(t, grypeReportJSON))), nil)

	report, err := (&service{
		RepoNs:       namespaceRepository,
		RepoRegistry: repositoryRepository,
		RepoArtifact: artifactRepository,
		Storage:      storageDriver,
	}).ListArtifactVulnerabilities(t.Context(), "namespace-1", "repository-1", "sha256:digest")

	require.NoError(t, err)
	require.Equal(t, string(enums.TaskCommonStatusSuccess), report.Status)
	require.Equal(t, "2025-01-02 03:04:05", report.DBBuiltAt)
	require.Equal(t, 2, report.Total)
	require.Equal(t, 1, report.Critical)
	require.Equal(t, 1, report.Low)
	require.Len(t, report.Items, 2)
	require.Equal(t, "CVE-2023-4863", report.Items[0].ID)
	require.Equal(t, "libwebp", report.Items[0].Package)
	require.Equal(t, []string{"sha256:layer-1"}, report.Items[0].Layers)
	require.Equal(t, "CVE-2023-0001", report.Items[1].ID)
}

// gzipReport compresses a grype report the way the scan daemon stores it.
func gzipReport(t *testing.T, report string) []byte {
	t.Helper()

	var compressed bytes.Buffer
	writer := gzip.NewWriter(&compressed)
	_, err := writer.Write([]byte(report))
	require.NoError(t, err)
	require.NoError(t, writer.Close())
	return compressed.Bytes()
}

const grypeReportJSON = `{
  "matches": [
    {
      "vulnerability": {
        "id": "CVE-2023-4863",
        "severity": "Critical",
        "fix": {"state": "fixed", "versions": ["1.2.4-0.1+deb12u1"]}
      },
      "artifact": {
        "name": "libwebp",
        "version": "1.2.4-0.1",
        "type": "deb",
        "locations": [{"path": "/usr/lib/libwebp.so", "layerID": "sha256:layer-1"}]
      }
    },
    {
      "vulnerability": {"id": "CVE-2023-0001", "severity": "Low", "fix": {"state": "wont-fix"}},
      "artifact": {"name": "zlib", "version": "1:1.2.13"}
    }
  ],
  "descriptor": {"db": {"status": {"built": "2025-01-02T03:04:05Z"}}}
}`

func TestGrypeVulnerabilityItemFlattensAMatch(t *testing.T) {
	var match grypeMatch
	match.Vulnerability.ID = "CVE-2023-4863"
	match.Vulnerability.Severity = "Critical"
	match.Vulnerability.DataSource = "https://security-tracker.debian.org/tracker/CVE-2023-4863"
	match.Vulnerability.Namespace = "debian:distro:debian:12"
	match.Vulnerability.URLs = []string{"https://example.com/cve"}
	match.Vulnerability.Fix.State = "fixed"
	match.Vulnerability.Fix.Versions = []string{"1.2.4-0.1+deb12u1"}
	match.Vulnerability.Fix.Available = []grypeFixAvailable{
		{Version: "1.2.4-0.1+deb12u2", Date: "2023-09-20"},
		{Version: "1.2.4-0.1+deb12u1", Date: "2023-09-12"},
	}
	match.Vulnerability.Cvss = []struct {
		Version string `json:"version"`
		Vector  string `json:"vector"`
		Metrics struct {
			BaseScore float64 `json:"baseScore"`
		} `json:"metrics"`
	}{{Version: "3.1", Vector: "CVSS:3.1/AV:N", Metrics: struct {
		BaseScore float64 `json:"baseScore"`
	}{BaseScore: 8.8}}}
	match.Vulnerability.EPSS = []struct {
		EPSS       float64 `json:"epss"`
		Percentile float64 `json:"percentile"`
	}{{EPSS: 0.5, Percentile: 0.9}}
	match.Vulnerability.CWEs = []struct {
		CWE string `json:"cwe"`
	}{{CWE: "CWE-787"}, {}}
	match.Vulnerability.KnownExploited = []struct {
		DateAdded string `json:"dateAdded"`
	}{{DateAdded: "2023-09-13"}}
	match.Artifact.Name = "libwebp"
	match.Artifact.Version = "1.2.4-0.1"
	match.Artifact.Type = "deb"
	match.Artifact.PURL = "pkg:deb/debian/libwebp@1.2.4-0.1"
	match.Artifact.Locations = []grypeLocation{
		{Path: "/usr/lib/libwebp.so", LayerID: "sha256:layer-1"},
		{Path: "/usr/lib/libwebp.so", LayerID: "sha256:layer-1"},
		{Path: "/usr/lib/libwebp.so.7", LayerID: "sha256:layer-2"},
	}
	match.MatchDetails = []struct {
		Fix *struct {
			SuggestedVersion string `json:"suggestedVersion"`
		} `json:"fix"`
	}{{Fix: &struct {
		SuggestedVersion string `json:"suggestedVersion"`
	}{SuggestedVersion: "1.2.4-0.1+deb12u1"}}}

	item := grypeVulnerabilityItem(match)

	require.Equal(t, "CVE-2023-4863", item.ID)
	require.Equal(t, "Critical", item.Severity)
	require.Equal(t, "libwebp", item.Package)
	require.Equal(t, "1.2.4-0.1", item.PackageVersion)
	require.Equal(t, "deb", item.PackageType)
	require.Equal(t, "fixed", item.FixState)
	require.Equal(t, []string{"1.2.4-0.1+deb12u1"}, item.FixVersions)
	// The earliest published fix is the date the page sorts the rows by.
	require.Equal(t, "2023-09-12", item.FixDate)
	require.Equal(t, "1.2.4-0.1+deb12u1", item.SuggestedVersion)
	require.Equal(t, 8.8, item.CvssScore)
	require.Equal(t, "CVSS:3.1/AV:N", item.CvssVector)
	require.Equal(t, 0.5, item.EpssScore)
	require.Equal(t, []string{"CWE-787"}, item.CWEs)
	require.Equal(t, "2023-09-13", item.KnownExploited)
	require.Equal(t, []string{"sha256:layer-1", "sha256:layer-2"}, item.Layers)
	require.Equal(t, []string{"/usr/lib/libwebp.so", "/usr/lib/libwebp.so.7"}, item.Paths)
}

func TestGrypeVulnerabilityItemKeepsEmptyListsAsEmptyArrays(t *testing.T) {
	item := grypeVulnerabilityItem(grypeMatch{})

	require.Empty(t, item.FixVersions)
	require.NotNil(t, item.FixVersions)
	require.NotNil(t, item.URLs)
	require.NotNil(t, item.CWEs)
	require.NotNil(t, item.Layers)
	require.NotNil(t, item.Paths)
}
