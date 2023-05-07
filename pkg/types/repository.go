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

// RepositoryItem represents a repository.
type RepositoryItem struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`

	ArtifactCount int64 `json:"artifact_count"`

	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// ListRepositoryRequest represents the request to list repositories.
type ListRepositoryRequest struct {
	Pagination

	Namespace string `json:"namespace" param:"namespace"`
}

// GetRepositoryRequest represents the request to get a repository.
type GetRepositoryRequest struct {
	ID        uint64 `json:"name" param:"id" validate:"required,number"`
	Namespace string `json:"namespace" param:"namespace" validate:"required,min=2,max=20"`
}

// DeleteRepositoryRequest represents the request to delete a repository.
type DeleteRepositoryRequest struct {
	ID uint64 `json:"id" param:"id" validate:"required,number"`
}
