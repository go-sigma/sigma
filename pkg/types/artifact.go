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

// ArtifactItem represents an artifact.
type ArtifactItem struct {
	ID        int64  `json:"id"`
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

// ListArtifactRequest represents the request to list artifacts.
type ListArtifactRequest struct {
	Pagination

	Namespace  string `json:"namespace" param:"namespace" validate:"required,min=2,max=20,is_valid_namespace"`
	Repository string `json:"repository" query:"repository" validate:"required,is_valid_repository"`
}

// GetArtifactRequest represents the request to get an artifact.
type GetArtifactRequest struct {
	Namespace  string `json:"namespace" param:"namespace" validate:"required,min=2,max=20,is_valid_namespace"`
	Repository string `json:"repository" query:"repository" validate:"required,is_valid_repository"`
	Digest     string `json:"digest" param:"digest" validate:"required,is_valid_digest"`
}

// DeleteArtifactRequest represents the request to delete an artifact.
type DeleteArtifactRequest struct {
	Namespace  string `json:"namespace" param:"namespace" validate:"required,min=2,max=20,is_valid_namespace"`
	Repository string `json:"repository" query:"repository" validate:"required,is_valid_repository"`
	Digest     string `json:"digest" param:"digest" validate:"required,is_valid_digest"`
}
