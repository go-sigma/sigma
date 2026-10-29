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

package api

// ArtifactItem represents an artifact returned by the API.
type ArtifactItem struct {
	ID        string `json:"id"`
	Digest    string `json:"digest"`
	ConfigRaw string `json:"config_raw"`
	Size      int64  `json:"size"`
	BlobSize  int64  `json:"blob_size"`
	LastPull  string `json:"last_pull"`
	PushedAt  string `json:"pushed_at"`
	PullTimes int64  `json:"pull_times"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// ArtifactLayer is one layer of an image manifest, enriched with the command that produced it.
type ArtifactLayer struct {
	Index     int    `json:"index" example:"1"`
	Digest    string `json:"digest" example:"sha256:31e352740f534f9ad170f75378a84fe453d6156e40700b882d737a8f4a6988a3"`
	DiffID    string `json:"diff_id" example:"sha256:78a822fe2a2d2c84f3de4a403188c45f623017d6a4521d23047c9fbb0801794c"`
	MediaType string `json:"media_type" example:"application/vnd.oci.image.layer.v1.tar+gzip"`
	Size      int64  `json:"size" example:"3397879"`
	Command   string `json:"command" example:"/bin/sh -c #(nop) ADD file:1da756d12551a0e3e793e02ef87432d69d4968937bd11bed0af215db19dd94cd in / "`
	CreatedAt string `json:"created_at" example:"2006-01-02 15:04:05"`
}

// ArtifactLayers is the artifact detail the artifact page renders: the image reference data and
// every layer of the manifest.
type ArtifactLayers struct {
	ID           string          `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Digest       string          `json:"digest" example:"sha256:87508bf3e050b975770b142e62db72eeb345a67d82d36ca166300d8b27e45744"`
	MediaType    string          `json:"media_type" example:"application/vnd.oci.image.manifest.v1+json"`
	Repository   string          `json:"repository" example:"library/alpine"`
	Size         int64           `json:"size" example:"10201"`
	BlobSize     int64           `json:"blob_size" example:"100210"`
	OS           string          `json:"os" example:"linux"`
	Architecture string          `json:"architecture" example:"amd64"`
	Distro       string          `json:"distro" example:"alpine 3.18.2"`
	PushedAt     string          `json:"pushed_at" example:"2006-01-02 15:04:05"`
	PullTimes    int64           `json:"pull_times" example:"100"`
	Layers       []ArtifactLayer `json:"layers"`
}

// ArtifactVulnerability is one vulnerability grype found in an artifact.
type ArtifactVulnerability struct {
	ID               string   `json:"id" example:"CVE-2023-4863"`
	Severity         string   `json:"severity" example:"Critical"`
	Package          string   `json:"package" example:"libwebp"`
	PackageVersion   string   `json:"package_version" example:"1.2.4-0.1"`
	PackageType      string   `json:"package_type" example:"deb"`
	PackageLanguage  string   `json:"package_language" example:"go"`
	PURL             string   `json:"purl" example:"pkg:deb/debian/libwebp@1.2.4-0.1?arch=amd64"`
	FixState         string   `json:"fix_state" example:"fixed"`
	FixVersions      []string `json:"fix_versions"`
	FixDate          string   `json:"fix_date" example:"2023-09-12"`
	SuggestedVersion string   `json:"suggested_version" example:"1.2.4-0.1+deb12u1"`
	Description      string   `json:"description"`
	DataSource       string   `json:"data_source" example:"https://security-tracker.debian.org/tracker/CVE-2023-4863"`
	Namespace        string   `json:"namespace" example:"debian:distro:debian:12"`
	URLs             []string `json:"urls"`
	CvssVersion      string   `json:"cvss_version" example:"3.1"`
	CvssVector       string   `json:"cvss_vector" example:"CVSS:3.1/AV:N/AC:L/PR:N/UI:R/S:U/C:H/I:H/A:H"`
	CvssScore        float64  `json:"cvss_score" example:"8.8"`
	EpssScore        float64  `json:"epss_score" example:"0.0213"`
	EpssPercentile   float64  `json:"epss_percentile" example:"0.8"`
	CWEs             []string `json:"cwes"`
	KnownExploited   string   `json:"known_exploited" example:"2023-09-13"`
	Layers           []string `json:"layers"`
	Paths            []string `json:"paths"`
}

// ArtifactVulnerabilities is the full vulnerability report of one artifact. It is returned without
// pagination so the page can sort and page the rows itself.
type ArtifactVulnerabilities struct {
	Status     string                  `json:"status" example:"Success"`
	ScannedAt  string                  `json:"scanned_at" example:"2006-01-02 15:04:05"`
	DBBuiltAt  string                  `json:"db_built_at" example:"2006-01-02 15:04:05"`
	Total      int                     `json:"total" example:"42"`
	Critical   int                     `json:"critical" example:"1"`
	High       int                     `json:"high" example:"2"`
	Medium     int                     `json:"medium" example:"3"`
	Low        int                     `json:"low" example:"4"`
	Negligible int                     `json:"negligible" example:"0"`
	Unknown    int                     `json:"unknown" example:"0"`
	Items      []ArtifactVulnerability `json:"items"`
}

// ListArtifactRequest is the request for listing artifacts.
type ListArtifactRequest struct {
	Pagination

	Namespace  string `json:"namespace" param:"namespace_id" validate:"required,min=2,max=20,is_valid_namespace"`
	Repository string `json:"repository" query:"repository" validate:"required,is_valid_repository"`
}

// GetArtifactRequest is the request for getting an artifact.
type GetArtifactRequest struct {
	Namespace  string `json:"namespace" param:"namespace_id" validate:"required,min=2,max=20,is_valid_namespace"`
	Repository string `json:"repository" query:"repository" validate:"required,is_valid_repository"`
	Digest     string `json:"digest" param:"digest" validate:"required,is_valid_digest"`
}

// GetArtifactLayersRequest is the request for getting an artifact with its layers.
type GetArtifactLayersRequest struct {
	NamespaceID  string `json:"namespace_id" param:"namespace_id" validate:"required"`
	RepositoryID string `json:"repository_id" param:"repository_id" validate:"required"`
	Digest       string `json:"digest" param:"digest" validate:"required,is_valid_digest"`
}

// ListArtifactVulnerabilitiesRequest is the request for getting the vulnerability report of an artifact.
type ListArtifactVulnerabilitiesRequest struct {
	NamespaceID  string `json:"namespace_id" param:"namespace_id" validate:"required"`
	RepositoryID string `json:"repository_id" param:"repository_id" validate:"required"`
	Digest       string `json:"digest" param:"digest" validate:"required,is_valid_digest"`
}

// DeleteArtifactRequest is the request for deleting an artifact.
type DeleteArtifactRequest struct {
	Namespace  string `json:"namespace" param:"namespace_id" validate:"required,min=2,max=20,is_valid_namespace"`
	Repository string `json:"repository" query:"repository" validate:"required,is_valid_repository"`
	Digest     string `json:"digest" param:"digest" validate:"required,is_valid_digest"`
}
