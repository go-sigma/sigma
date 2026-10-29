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
	"compress/gzip"
	"context"
	"encoding/json/v2"
	"fmt"
	"log/slog"
	"slices"
	"strings"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/server/errcode"
)

// grypeDocument is the subset of the grype JSON report sigma renders.
type grypeDocument struct {
	Matches    []grypeMatch `json:"matches"`
	Descriptor struct {
		DB struct {
			Status struct {
				Built string `json:"built"`
			} `json:"status"`
		} `json:"db"`
	} `json:"descriptor"`
}

// grypeMatch is one package vulnerability match of a grype report.
type grypeMatch struct {
	Vulnerability grypeVulnerability `json:"vulnerability"`
	Artifact      grypePackage       `json:"artifact"`
	MatchDetails  []struct {
		Fix *struct {
			SuggestedVersion string `json:"suggestedVersion"`
		} `json:"fix"`
	} `json:"matchDetails"`
}

// grypeVulnerability is the vulnerability side of a grype match.
type grypeVulnerability struct {
	ID          string   `json:"id"`
	DataSource  string   `json:"dataSource"`
	Namespace   string   `json:"namespace"`
	Severity    string   `json:"severity"`
	URLs        []string `json:"urls"`
	Description string   `json:"description"`
	Fix         struct {
		Versions  []string            `json:"versions"`
		State     string              `json:"state"`
		Available []grypeFixAvailable `json:"available"`
	} `json:"fix"`
	Cvss []struct {
		Version string `json:"version"`
		Vector  string `json:"vector"`
		Metrics struct {
			BaseScore float64 `json:"baseScore"`
		} `json:"metrics"`
	} `json:"cvss"`
	EPSS []struct {
		EPSS       float64 `json:"epss"`
		Percentile float64 `json:"percentile"`
	} `json:"epss"`
	CWEs []struct {
		CWE string `json:"cwe"`
	} `json:"cwes"`
	KnownExploited []struct {
		DateAdded string `json:"dateAdded"`
	} `json:"knownExploited"`
}

// grypeFixAvailable is one release that carries the fix of a vulnerability.
type grypeFixAvailable struct {
	Version string `json:"version"`
	Date    string `json:"date"`
}

// grypePackage is the vulnerable package side of a grype match.
type grypePackage struct {
	Name      string          `json:"name"`
	Version   string          `json:"version"`
	Type      string          `json:"type"`
	Language  string          `json:"language"`
	PURL      string          `json:"purl"`
	Locations []grypeLocation `json:"locations"`
}

// grypeLocation is where grype found the vulnerable package inside the image.
type grypeLocation struct {
	Path    string `json:"path"`
	LayerID string `json:"layerID"`
}

// ListArtifactVulnerabilities gets every vulnerability grype reported for an artifact of a
// repository. The report is returned without pagination: the artifact page sorts and pages it.
func (s *service) ListArtifactVulnerabilities(ctx context.Context, namespaceID, repositoryID, digest string) (*api.ArtifactVulnerabilities, error) {
	artifactObj, _, err := s.getArtifactWithReports(ctx, namespaceID, repositoryID, digest)
	if err != nil {
		return nil, err
	}

	vulnerabilityObj := artifactObj.Vulnerability
	resp := &api.ArtifactVulnerabilities{
		Status:    string(vulnerabilityObj.Status),
		ScannedAt: formatArtifactTime(vulnerabilityObj.UpdatedAt),
		Items:     []api.ArtifactVulnerability{},
	}
	if vulnerabilityObj.Status != enums.TaskCommonStatusSuccess || len(vulnerabilityObj.Raw) == 0 {
		return resp, nil
	}

	report, err := s.readVulnerabilityReport(ctx, string(vulnerabilityObj.Raw))
	if err != nil {
		return nil, err
	}
	resp.DBBuiltAt = formatConfigTime(report.Descriptor.DB.Status.Built)

	for _, match := range report.Matches {
		item := grypeVulnerabilityItem(match)
		resp.Items = append(resp.Items, item)
		resp.Total++
		switch strings.ToUpper(item.Severity) {
		case "CRITICAL":
			resp.Critical++
		case "HIGH":
			resp.High++
		case "MEDIUM":
			resp.Medium++
		case "LOW":
			resp.Low++
		case "NEGLIGIBLE":
			resp.Negligible++
		default:
			resp.Unknown++
		}
	}
	return resp, nil
}

// readVulnerabilityReport downloads the gzipped grype report the scan daemon stored for the artifact.
func (s *service) readVulnerabilityReport(ctx context.Context, path string) (*grypeDocument, error) {
	reader, err := s.Storage.Reader(ctx, path)
	if err != nil {
		slog.Error("read vulnerability report failed", "err", err, "path", path)
		return nil, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Read vulnerability report failed: %v", err))
	}
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			slog.Warn("close vulnerability report failed", "err", closeErr, "path", path)
		}
	}()

	gzipReader, err := gzip.NewReader(reader)
	if err != nil {
		slog.Error("decompress vulnerability report failed", "err", err, "path", path)
		return nil, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Decompress vulnerability report failed: %v", err))
	}
	defer func() {
		if closeErr := gzipReader.Close(); closeErr != nil {
			slog.Warn("close vulnerability report reader failed", "err", closeErr, "path", path)
		}
	}()

	var report grypeDocument
	if err = json.UnmarshalRead(gzipReader, &report); err != nil {
		slog.Error("decode vulnerability report failed", "err", err, "path", path)
		return nil, errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Decode vulnerability report failed: %v", err))
	}
	return &report, nil
}

// grypeVulnerabilityItem flattens one grype match into the row the artifact page renders.
func grypeVulnerabilityItem(match grypeMatch) api.ArtifactVulnerability {
	vulnerability := match.Vulnerability
	item := api.ArtifactVulnerability{
		ID:              vulnerability.ID,
		Severity:        vulnerability.Severity,
		Package:         match.Artifact.Name,
		PackageVersion:  match.Artifact.Version,
		PackageType:     match.Artifact.Type,
		PackageLanguage: match.Artifact.Language,
		PURL:            match.Artifact.PURL,
		FixState:        vulnerability.Fix.State,
		FixVersions:     nonNilStrings(vulnerability.Fix.Versions),
		FixDate:         firstFixDate(vulnerability.Fix.Available),
		Description:     vulnerability.Description,
		DataSource:      vulnerability.DataSource,
		Namespace:       vulnerability.Namespace,
		URLs:            nonNilStrings(vulnerability.URLs),
		Layers:          locationLayerIDs(match.Artifact.Locations),
		Paths:           locationPaths(match.Artifact.Locations),
	}
	if len(match.MatchDetails) > 0 && match.MatchDetails[0].Fix != nil {
		item.SuggestedVersion = match.MatchDetails[0].Fix.SuggestedVersion
	}
	if len(vulnerability.Cvss) > 0 {
		item.CvssVersion = vulnerability.Cvss[0].Version
		item.CvssVector = vulnerability.Cvss[0].Vector
		item.CvssScore = vulnerability.Cvss[0].Metrics.BaseScore
	}
	if len(vulnerability.EPSS) > 0 {
		item.EpssScore = vulnerability.EPSS[0].EPSS
		item.EpssPercentile = vulnerability.EPSS[0].Percentile
	}
	if len(vulnerability.KnownExploited) > 0 {
		item.KnownExploited = vulnerability.KnownExploited[0].DateAdded
	}
	cwes := make([]string, 0, len(vulnerability.CWEs))
	for _, cwe := range vulnerability.CWEs {
		if cwe.CWE != "" {
			cwes = append(cwes, cwe.CWE)
		}
	}
	item.CWEs = cwes
	return item
}

// firstFixDate returns the date the earliest known fix of the vulnerability was published.
func firstFixDate(available []grypeFixAvailable) string {
	date := ""
	for _, item := range available {
		if item.Date == "" {
			continue
		}
		if date == "" || item.Date < date {
			date = item.Date
		}
	}
	return date
}

// locationLayerIDs collects the unique layers that carry the vulnerable package.
func locationLayerIDs(locations []grypeLocation) []string {
	layerIDs := make([]string, 0, len(locations))
	for _, location := range locations {
		if location.LayerID == "" || slices.Contains(layerIDs, location.LayerID) {
			continue
		}
		layerIDs = append(layerIDs, location.LayerID)
	}
	return layerIDs
}

// locationPaths collects the unique file paths of the vulnerable package.
func locationPaths(locations []grypeLocation) []string {
	paths := make([]string, 0, len(locations))
	for _, location := range locations {
		if location.Path == "" || slices.Contains(paths, location.Path) {
			continue
		}
		paths = append(paths, location.Path)
	}
	return paths
}

// nonNilStrings keeps the JSON shape predictable for the artifact page.
func nonNilStrings(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}
