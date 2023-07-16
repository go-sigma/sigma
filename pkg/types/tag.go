// Copyright 2023 XImager
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

// TagItemArtifact ...
type TagItemArtifact struct {
	ID            int64  `json:"id"`
	Digest        string `json:"digest"`
	Raw           string `json:"raw"`
	ConfigRaw     string `json:"config_raw"`
	Size          int64  `json:"size"`
	BlobSize      int64  `json:"blob_size"`
	LastPull      string `json:"last_pull"`
	PushedAt      string `json:"pushed_at"`
	PullTimes     int64  `json:"pull_times"`
	Vulnerability string `json:"vulnerability"`
	Sbom          string `json:"sbom"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

// TagItem represents an tag.
type TagItem struct {
	ID        int64             `json:"id"`
	Name      string            `json:"name"`
	PushedAt  string            `json:"pushed_at"`
	Artifact  TagItemArtifact   `json:"artifact"`
	Artifacts []TagItemArtifact `json:"artifacts"`
	CreatedAt string            `json:"created_at"`
	UpdatedAt string            `json:"updated_at"`
}

// ListTagRequest represents the request to list tags.
type ListTagRequest struct {
	Pagination
	Sortable

	Namespace  string `json:"namespace" param:"namespace" validate:"required,min=2,max=20,is_valid_namespace" example:"library"`
	Repository string `json:"repository" query:"repository" validate:"required,is_valid_repository" example:"library/busybox"`

	Name *string `json:"name" query:"name"`
}

// DeleteTagRequest represents the request to delete a tag.
type DeleteTagRequest struct {
	ID         int64  `param:"id" validate:"required,number"`
	Namespace  string `param:"namespace" validate:"required,min=2,max=20,is_valid_namespace"`
	Repository string `query:"repository" validate:"required,is_valid_repository"`
}

// GetTagRequest represents the request to get a tag.
type GetTagRequest struct {
	ID         int64  `json:"id" param:"id" validate:"required,number"`
	Namespace  string `json:"namespace" param:"namespace" validate:"required,min=2,max=20,is_valid_namespace"`
	Repository string `json:"repository" query:"repository" validate:"required,is_valid_repository"`
}
