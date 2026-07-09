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

import "github.com/go-sigma/sigma/pkg/api/enums"

// RepositoryItem represents an OCI repository returned by the API.
type RepositoryItem struct {
	ID          string           `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	NamespaceID string           `json:"namespace_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name        string           `json:"name" example:"busybox"`
	Description *string          `json:"description,omitempty" example:"i am just description"`
	Overview    *string          `json:"overview,omitempty" example:"i am just overview"`
	Visibility  enums.Visibility `json:"visibility" example:"private"`
	TagCount    int64            `json:"tag_count" example:"100"`
	TagLimit    *int64           `json:"tag_limit" example:"1000"`
	SizeLimit   *int64           `json:"size_limit" example:"10000"`
	Size        *int64           `json:"size" example:"10000"`

	Builder *BuilderItem `json:"builder"`

	CreatedAt string `json:"created_at" example:"2006-01-02 15:04:05"`
	UpdatedAt string `json:"updated_at" example:"2006-01-02 15:04:05"`
}

// ListRepositoryRequest is the request for listing repositories.
type ListRepositoryRequest struct {
	NamespaceID string `json:"namespace_id" param:"namespace_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`

	Pagination
	Sortable

	Name *string `json:"name" query:"name"`
}

// GetRepositoryRequest is the request for getting a repository.
type GetRepositoryRequest struct {
	NamespaceID string `json:"namespace_id" param:"namespace_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	ID          string `json:"id" param:"repository_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// DeleteRepositoryRequest is the request for deleting a repository.
type DeleteRepositoryRequest struct {
	NamespaceID string `json:"namespace_id" param:"namespace_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	ID          string `json:"id" param:"repository_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// CreateRepositoryRequest is the request for creating a repository.
type CreateRepositoryRequest struct {
	NamespaceID string `json:"namespace_id" param:"namespace_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000" swaggerignore:"true"`

	Name        string            `json:"name" validate:"required,is_valid_repository" example:"test"`
	Description *string           `json:"description,omitempty" validate:"omitempty,max=30" example:"i am just description"`
	Overview    *string           `json:"overview,omitempty" validate:"omitempty,max=3000" example:"i am just overview"`
	SizeLimit   *int64            `json:"size_limit,omitempty" validate:"omitempty" example:"10000"`
	TagLimit    *int64            `json:"tag_limit,omitempty" validate:"omitempty" example:"10000"`
	Visibility  *enums.Visibility `json:"visibility,omitempty" validate:"omitempty,is_valid_visibility" example:"public"`
}

// CreateRepositoryResponse is the response returned after creating a repository.
type CreateRepositoryResponse struct {
	ID string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// UpdateRepositoryRequest is the request for updating a repository.
type UpdateRepositoryRequest struct {
	NamespaceID string `json:"namespace_id" param:"namespace_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000" swaggerignore:"true"`

	ID          string  `json:"id" param:"repository_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000" swaggerignore:"true"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=300" example:"i am just description"`
	Overview    *string `json:"overview,omitempty" validate:"omitempty,max=100000" example:"i am just overview"`
	SizeLimit   *int64  `json:"size_limit,omitempty" validate:"omitempty" example:"10000"`
	TagLimit    *int64  `json:"tag_limit,omitempty" validate:"omitempty" example:"10000"`
}
