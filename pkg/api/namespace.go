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

// NamespaceItem represents a namespace returned by the API.
type NamespaceItem struct {
	ID              string           `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name            string           `json:"name" example:"test"`
	Description     *string          `json:"description,omitempty" example:"i am just description"`
	Overview        *string          `json:"overview,omitempty" example:"i am just overview"`
	Visibility      enums.Visibility `json:"visibility" example:"private"`
	RepositoryLimit int64            `json:"repository_limit" example:"1000"`
	RepositoryCount int64            `json:"repository_count" example:"100"`
	TagLimit        int64            `json:"tag_limit" example:"1000"`
	TagCount        int64            `json:"tag_count" example:"100"`
	Size            int64            `json:"size" example:"10000"`
	SizeLimit       int64            `json:"size_limit" example:"10000"`

	CreatedAt string `json:"created_at" example:"2006-01-02 15:04:05"`
	UpdatedAt string `json:"updated_at" example:"2006-01-02 15:04:05"`
}

// ListNamespaceRequest is the request for listing namespaces.
type ListNamespaceRequest struct {
	Pagination
	Sortable

	// Name filters namespaces by name.
	Name *string `json:"name" query:"name" example:"test"`
}

// PostNamespaceRequest is the request for creating a namespace.
type PostNamespaceRequest struct {
	Name            string            `json:"name" validate:"required,min=2,max=20,is_valid_namespace" example:"test"`
	Description     *string           `json:"description,omitempty" validate:"omitempty,max=30" example:"i am just description"`
	SizeLimit       *int64            `json:"size_limit,omitempty" validate:"omitempty" example:"10000"`
	RepositoryLimit *int64            `json:"repository_limit,omitempty" validate:"omitempty" example:"10000"`
	TagLimit        *int64            `json:"tag_limit,omitempty" validate:"omitempty" example:"10000"`
	Visibility      *enums.Visibility `json:"visibility,omitempty" validate:"omitempty,is_valid_visibility" example:"public"`
}

// PostNamespaceResponse is the response returned after creating a namespace.
type PostNamespaceResponse struct {
	ID string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// GetNamespaceRequest is the request for getting a namespace.
type GetNamespaceRequest struct {
	ID string `json:"id" param:"namespace_id" validate:"required"`
}

// DeleteNamespaceRequest is the request for deleting a namespace.
type DeleteNamespaceRequest struct {
	ID string `json:"id" param:"namespace_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// UpdateNamespaceRequest is the request for updating a namespace.
type UpdateNamespaceRequest struct {
	ID string `json:"id" param:"namespace_id" validate:"required" swaggerignore:"true"`

	SizeLimit       *int64            `json:"size_limit,omitempty" validate:"omitempty" example:"10000"`
	RepositoryLimit *int64            `json:"repository_limit" validate:"omitempty" example:"10000"`
	TagLimit        *int64            `json:"tag_limit,omitempty" validate:"omitempty" example:"10000"`
	Visibility      *enums.Visibility `json:"visibility,omitempty" validate:"omitempty,is_valid_visibility" example:"public"`
	Description     *string           `json:"description,omitempty" validate:"omitempty,max=30" example:"i am just description"`
	Overview        *string           `json:"overview,omitempty" validate:"omitempty,max=100000" example:"i am just overview"`
}

// AddNamespaceMemberRequest is the request for adding a namespace member.
type AddNamespaceMemberRequest struct {
	NamespaceID string `json:"namespace_id" param:"namespace_id" validate:"required" swaggerignore:"true"`

	UserID string              `json:"user_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Role   enums.NamespaceRole `json:"role" validate:"is_valid_namespace_role" example:"NamespaceReader"`
}

// AddNamespaceMemberResponse is the response returned after adding a namespace member.
type AddNamespaceMemberResponse struct {
	ID string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// UpdateNamespaceMemberRequest is the request for updating a namespace member.
type UpdateNamespaceMemberRequest struct {
	NamespaceID string `json:"namespace_id" param:"namespace_id" validate:"required" swaggerignore:"true"`
	UserID      string `json:"user_id" param:"user_id" swaggerignore:"true"`

	Role enums.NamespaceRole `json:"role" validate:"is_valid_namespace_role" example:"NamespaceReader"`
}

// DeleteNamespaceMemberRequest is the request for removing a namespace member.
type DeleteNamespaceMemberRequest struct {
	NamespaceID string `json:"namespace_id" param:"namespace_id" validate:"required" swaggerignore:"true"`
	UserID      string `json:"user_id" param:"user_id" example:"550e8400-e29b-41d4-a716-446655440000" swaggerignore:"true"`
}

// ListNamespaceMemberRequest is the request for listing namespace members.
type ListNamespaceMemberRequest struct {
	NamespaceID string `json:"namespace_id" param:"namespace_id" validate:"required" swaggerignore:"true"`

	// Name filters namespace members by username.
	Name *string `json:"name" query:"name" example:"test" swaggerignore:"true"`

	Pagination
	Sortable
}

// NamespaceMemberItem represents a namespace member.
type NamespaceMemberItem struct {
	ID       string              `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Username string              `json:"username" example:"admin"`
	UserID   string              `json:"user_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Role     enums.NamespaceRole `json:"role" example:"NamespaceAdmin"`

	CreatedAt string `json:"created_at" example:"2006-01-02 15:04:05"`
	UpdatedAt string `json:"updated_at" example:"2006-01-02 15:04:05"`
}

// GetNamespaceMemberSelfRequest is the request for getting the caller's namespace membership.
type GetNamespaceMemberSelfRequest struct {
	NamespaceID string `json:"namespace_id" param:"namespace_id" validate:"required" swaggerignore:"true"`
}
