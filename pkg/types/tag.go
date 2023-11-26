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

package types

import "github.com/go-sigma/sigma/pkg/types/enums"

// TagItemArtifact ...
type TagItemArtifact struct {
	ID              int64  `json:"id" example:"1"`
	Digest          string `json:"digest" example:"sha256:87508bf3e050b975770b142e62db72eeb345a67d82d36ca166300d8b27e45744"`
	MediaType       string `json:"media_type" example:"application/vnd.oci.image.manifest.v1+json"`
	Raw             string `json:"raw" example:"{\n   \"schemaVersion\": 2,\n   \"mediaType\": \"application/vnd.docker.distribution.manifest.v2+json\",\n   \"config\": {\n      \"mediaType\": \"application/vnd.docker.container.image.v1+json\",\n      \"size\": 1472,\n      \"digest\": \"sha256:c1aabb73d2339c5ebaa3681de2e9d9c18d57485045a4e311d9f8004bec208d67\"\n   },\n   \"layers\": [\n      {\n         \"mediaType\": \"application/vnd.docker.image.rootfs.diff.tar.gzip\",\n         \"size\": 3397879,\n         \"digest\": \"sha256:31e352740f534f9ad170f75378a84fe453d6156e40700b882d737a8f4a6988a3\"\n      }\n   ]\n}"`
	ConfigMediaType string `json:"config_media_type" example:"application/vnd.oci.image.config.v1+json"`
	ConfigRaw       string `json:"config_raw" example:"{\"architecture\":\"amd64\",\"config\":{\"Hostname\":\"\",\"Domainname\":\"\",\"User\":\"\",\"AttachStdin\":false,\"AttachStdout\":false,\"AttachStderr\":false,\"Tty\":false,\"OpenStdin\":false,\"StdinOnce\":false,\"Env\":[\"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin\"],\"Cmd\":[\"/bin/sh\"],\"Image\":\"sha256:5b8658701c96acefe1cd3a21b2a80220badf9124891ad440d95a7fa500d48765\",\"Volumes\":null,\"WorkingDir\":\"\",\"Entrypoint\":null,\"OnBuild\":null,\"Labels\":null},\"container\":\"bfc8078c169637d70e40ce591b5c2fe8d26329918dafcb96ebc9304ddff162ea\",\"container_config\":{\"Hostname\":\"bfc8078c1696\",\"Domainname\":\"\",\"User\":\"\",\"AttachStdin\":false,\"AttachStdout\":false,\"AttachStderr\":false,\"Tty\":false,\"OpenStdin\":false,\"StdinOnce\":false,\"Env\":[\"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin\"],\"Cmd\":[\"/bin/sh\",\"-c\",\"#(nop) \",\"CMD [\\\"/bin/sh\\\"]\"],\"Image\":\"sha256:5b8658701c96acefe1cd3a21b2a80220badf9124891ad440d95a7fa500d48765\",\"Volumes\":null,\"WorkingDir\":\"\",\"Entrypoint\":null,\"OnBuild\":null,\"Labels\":{}},\"created\":\"2023-06-14T20:41:59.079795125Z\",\"docker_version\":\"20.10.23\",\"history\":[{\"created\":\"2023-06-14T20:41:58.950178204Z\",\"created_by\":\"/bin/sh -c #(nop) ADD file:1da756d12551a0e3e793e02ef87432d69d4968937bd11bed0af215db19dd94cd in / \"},{\"created\":\"2023-06-14T20:41:59.079795125Z\",\"created_by\":\"/bin/sh -c #(nop)  CMD [\\\"/bin/sh\\\"]\",\"empty_layer\":true}],\"os\":\"linux\",\"rootfs\":{\"type\":\"layers\",\"diff_ids\":[\"sha256:78a822fe2a2d2c84f3de4a403188c45f623017d6a4521d23047c9fbb0801794c\"]}}"`
	Type            string `json:"type" example:"image"`
	Size            int64  `json:"size" example:"10201"`
	BlobSize        int64  `json:"blob_size" example:"100210"`
	LastPull        string `json:"last_pull" example:"2006-01-02 15:04:05"`
	PushedAt        string `json:"pushed_at" example:"2006-01-02 15:04:05"`
	PullTimes       int64  `json:"pull_times" example:"10"`
	Vulnerability   string `json:"vulnerability" example:"{\"critical\":0,\"high\":0,\"medium\":0,\"low\":0}"`
	Sbom            string `json:"sbom" example:"{\"distro\":{\"name\":\"alpine\",\"version\":\"3.18.2\"},\"os\":\"linux\",\"architecture\":\"amd64\"}"`
	CreatedAt       string `json:"created_at" example:"2006-01-02 15:04:05"`
	UpdatedAt       string `json:"updated_at" example:"2006-01-02 15:04:05"`
}

// TagItem represents an tag.
type TagItem struct {
	ID        int64             `json:"id" example:"1"`
	Name      string            `json:"name" example:"latest"`
	PushedAt  string            `json:"pushed_at" example:"2006-01-02 15:04:05"`
	PullTimes int64             `json:"pull_times" example:"10"`
	Artifact  TagItemArtifact   `json:"artifact"`
	Artifacts []TagItemArtifact `json:"artifacts"`
	CreatedAt string            `json:"created_at" example:"2006-01-02 15:04:05"`
	UpdatedAt string            `json:"updated_at" example:"2006-01-02 15:04:05"`
}

// ListTagRequest represents the request to list tags.
type ListTagRequest struct {
	NamespaceID  int64 `json:"namespace_id" param:"namespace_id" validate:"required" example:"10"`
	RepositoryID int64 `json:"repository_id" param:"repository_id" validate:"required" example:"10"`

	Pagination
	Sortable

	Name *string              `json:"name" query:"name"`
	Type []enums.ArtifactType `json:"type" query:"type"`
}

// DeleteTagRequest represents the request to delete a tag.
type DeleteTagRequest struct {
	NamespaceID  int64 `json:"namespace_id" param:"namespace_id" validate:"required" example:"10"`
	RepositoryID int64 `json:"repository_id" param:"repository_id" validate:"required" example:"10"`
	ID           int64 `param:"id" validate:"required,number"`
}

// GetTagRequest represents the request to get a tag.
type GetTagRequest struct {
	NamespaceID  int64  `json:"namespace_id" param:"namespace_id" validate:"required" example:"10"`
	RepositoryID string `json:"repository_id" param:"repository_id" validate:"required" example:"10"`

	ID int64 `json:"id" param:"id" validate:"required,number"`
}
