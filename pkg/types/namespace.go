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

// NamespaceItem represents a namespace.
type NamespaceItem struct {
	ID              int64            `json:"id" example:"1"`
	Name            string           `json:"name" example:"test"`
	Description     *string          `json:"description,omitempty" example:"i am just description"`
	Overview        *string          `json:"overview,omitempty" example:"i am just overview"`
	Visibility      enums.Visibility `json:"visibility" example:"private"`
	RepositoryLimit int64            `json:"repository_limit" example:"10"`
	RepositoryCount int64            `json:"repository_count" example:"10"`
	TagLimit        int64            `json:"tag_limit" example:"10"`
	TagCount        int64            `json:"tag_count" example:"10"`
	Size            int64            `json:"size" example:"10000"`
	SizeLimit       int64            `json:"size_limit" example:"10000"`

	CreatedAt string `json:"created_at" example:"2006-01-02 15:04:05"`
	UpdatedAt string `json:"updated_at" example:"2006-01-02 15:04:05"`
}

// ListNamespaceRequest represents the request to list namespaces.
type ListNamespaceRequest struct {
	Pagination
	Sortable

	// Name query the namespace by name.
	Name *string `json:"name" query:"name" example:"test"`
}

// PostNamespaceRequest represents the request to create a namespace.
type PostNamespaceRequest struct {
	Name            string            `json:"name" validate:"required,min=2,max=20,is_valid_namespace" example:"test"`
	Description     *string           `json:"description,omitempty" validate:"omitempty,max=30" example:"i am just description"`
	SizeLimit       *int64            `json:"size_limit,omitempty" validate:"omitempty,numeric" example:"10000"`
	RepositoryLimit *int64            `json:"repository_limit,omitempty" validate:"omitempty,numeric" example:"10000"`
	TagLimit        *int64            `json:"tag_limit,omitempty" validate:"omitempty,numeric" example:"10000"`
	Visibility      *enums.Visibility `json:"visibility,omitempty" validate:"omitempty,is_valid_visibility" example:"public"`
}

// PostNamespaceResponse represents the response to create a namespace.
type PostNamespaceResponse struct {
	ID int64 `json:"id" example:"21911"`
}

// GetNamespaceRequest represents the request to get a namespace.
type GetNamespaceRequest struct {
	ID int64 `json:"id" param:"id" validate:"required,number"`
}

// DeleteNamespaceRequest represents the request to delete a namespace.
type DeleteNamespaceRequest struct {
	ID int64 `json:"id" param:"id" validate:"required,number" example:"1"`
}

// UpdateNamespaceRequest represents the request to update a namespace.
type UpdateNamespaceRequest struct {
	ID int64 `json:"id" param:"id" validate:"required,number" swaggerignore:"true"`

	SizeLimit       *int64            `json:"size_limit,omitempty" validate:"omitempty,numeric" example:"10000"`
	RepositoryLimit *int64            `json:"repository_limit" validate:"omitempty,numeric" example:"10000"`
	TagLimit        *int64            `json:"tag_limit,omitempty" validate:"omitempty,numeric" example:"10000"`
	Visibility      *enums.Visibility `json:"visibility,omitempty" validate:"omitempty,is_valid_visibility" example:"public"`
	Description     *string           `json:"description,omitempty" validate:"omitempty,max=30" example:"i am just description"`
	Overview        *string           `json:"overview,omitempty" validate:"omitempty,max=100000" example:"i am just overview"`
}

// AddNamespaceMemberRequest ...
type AddNamespaceMemberRequest struct {
	NamespaceID int64 `json:"namespace_id" param:"namespace_id" validate:"required,number" swaggerignore:"true"`

	UserID int64               `json:"user_id" example:"10"`
	Role   enums.NamespaceRole `json:"role" validate:"is_valid_namespace_role" example:"NamespaceReader"`
}

// AddNamespaceMemberResponse ...
type AddNamespaceMemberResponse struct {
	ID int64 `json:"id" example:"10"`
}

// UpdateNamespaceMemberRequest ...
type UpdateNamespaceMemberRequest struct {
	NamespaceID int64 `json:"namespace_id" param:"namespace_id" validate:"required,number" swaggerignore:"true"`
	UserID      int64 `json:"user_id" param:"user_id" swaggerignore:"true"`

	Role enums.NamespaceRole `json:"role" validate:"is_valid_namespace_role" example:"NamespaceReader"`
}

// DeleteNamespaceMemberRequest ...
type DeleteNamespaceMemberRequest struct {
	NamespaceID int64 `json:"namespace_id" param:"namespace_id" validate:"required,number" swaggerignore:"true"`
	UserID      int64 `json:"user_id" param:"user_id" example:"10" swaggerignore:"true"`
}

// ListNamespaceMemberRequest represents the request to list namespace members.
type ListNamespaceMemberRequest struct {
	NamespaceID int64 `json:"namespace_id" param:"namespace_id" validate:"required,number" swaggerignore:"true"`

	// Name query the namespace member by name.
	Name *string `json:"name" query:"name" example:"test" swaggerignore:"true"`

	Pagination
	Sortable
}

// NamespaceMemberItem ...
type NamespaceMemberItem struct {
	ID       int64               `json:"id" example:"1"`
	Username string              `json:"username" example:"admin"`
	UserID   int64               `json:"user_id" example:"1"`
	Role     enums.NamespaceRole `json:"role" example:"NamespaceAdmin"`

	CreatedAt string `json:"created_at" example:"2006-01-02 15:04:05"`
	UpdatedAt string `json:"updated_at" example:"2006-01-02 15:04:05"`
}

// GetNamespaceMemberSelfRequest ...
type GetNamespaceMemberSelfRequest struct {
	NamespaceID int64 `json:"namespace_id" param:"namespace_id" validate:"required,number" swaggerignore:"true"`
}
