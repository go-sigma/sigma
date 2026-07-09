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

// CodeRepositoryItem represents a synchronized source code repository.
type CodeRepositoryItem struct {
	ID           string            `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	RepositoryID string            `json:"repository_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Provider     enums.ScmProvider `json:"provider" example:"github"`
	Name         string            `json:"name" example:"sigma"`
	OwnerID      string            `json:"owner_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Owner        string            `json:"owner" example:"go-sigma"`
	IsOrg        bool              `json:"is_org" example:"true"`
	CloneUrl     string            `json:"clone_url" example:"https://github.com/go-sigma/sigma.git"`
	SshUrl       string            `json:"ssh_url" example:"git@github.com:go-sigma/sigma.git"`
	OciRepoCount int64             `json:"oci_repo_count" example:"5"`
	CreatedAt    string            `json:"created_at" example:"2006-01-02 15:04:05"`
	UpdatedAt    string            `json:"updated_at" example:"2006-01-02 15:04:05"`
}

// CodeRepositoryOwnerItem represents a source code repository owner or organization.
type CodeRepositoryOwnerItem struct {
	ID        string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	OwnerID   string `json:"owner_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Owner     string `json:"owner" example:"go-sigma"`
	IsOrg     bool   `json:"is_org" example:"true"`
	CreatedAt string `json:"created_at" example:"2006-01-02 15:04:05"`
	UpdatedAt string `json:"updated_at" example:"2006-01-02 15:04:05"`
}

// ListCodeRepositoryRequest is the request for listing source code repositories.
type ListCodeRepositoryRequest struct {
	Pagination
	Sortable

	Provider enums.Provider `json:"provider" param:"provider" validate:"required,is_valid_provider"`
	Owner    *string        `json:"owner,omitempty" query:"owner" validate:"omitempty,min=1"`
	Name     *string        `json:"name,omitempty" query:"name" validate:"omitempty,min=1"`
}

// GetCodeRepositoryRequest is the request for getting a source code repository.
type GetCodeRepositoryRequest struct {
	Provider enums.Provider `json:"provider" param:"provider" validate:"required,is_valid_provider"`
	ID       string         `json:"id" param:"id" validate:"required"`
}

// ListCodeRepositoryOwnerRequest is the request for listing source code repository owners.
type ListCodeRepositoryOwnerRequest struct {
	Provider enums.Provider `json:"provider" param:"provider" validate:"required,is_valid_provider"`
	Name     *string        `json:"name,omitempty" query:"name" validate:"omitempty,min=1"`
}

// ListCodeRepositoryBranchesRequest is the request for listing source code repository branches.
type ListCodeRepositoryBranchesRequest struct {
	ID string `json:"id" param:"id" validate:"required"`
}

// CodeRepositoryBranchItem represents a source code repository branch.
type CodeRepositoryBranchItem struct {
	ID        string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name      string `json:"name" example:"main"`
	CreatedAt string `json:"created_at" example:"2006-01-02 15:04:05"`
	UpdatedAt string `json:"updated_at" example:"2006-01-02 15:04:05"`
}

// GetCodeRepositoryBranchRequest is the request for getting a source code repository branch.
type GetCodeRepositoryBranchRequest struct {
	ID   string `json:"id" param:"id" validate:"required"`
	Name string `json:"name" param:"name" validate:"required"`
}

// PostCodeRepositorySetupBuilder is the request for creating a builder from a source code repository.
type PostCodeRepositorySetupBuilder struct {
	ID string `json:"id" param:"id" validate:"required"`

	NamespaceID    string `json:"namespace_id" validate:"required"`
	RepositoryName string `json:"repository_name" validate:"required,is_valid_repository" example:"library/test"`

	ScmBranch    string `json:"scm_branch" example:"main"`
	ScmDepth     int    `json:"scm_depth" example:"0"`
	ScmSubmodule bool   `json:"scm_submodule" example:"false"`

	BuildkitContext    string              `json:"buildkit_context" example:"."`
	BuildkitDockerfile string              `json:"buildkit_dockerfile" example:"Dockerfile"`
	BuildkitPlatforms  []enums.OciPlatform `json:"buildkit_platforms" example:"linux/amd64"`
}

// PostCodeRepositorySetupBuilderSwagger describes the Swagger body for setting up a repository builder.
type PostCodeRepositorySetupBuilderSwagger struct {
	NamespaceID    string `json:"namespace_id" validate:"required"`
	RepositoryName string `json:"repository_name" validate:"required,is_valid_repository" example:"library/test"`

	ScmBranch    string `json:"scm_branch" example:"main"`
	ScmDepth     int    `json:"scm_depth" example:"0"`
	ScmSubmodule bool   `json:"scm_submodule" example:"false"`

	BuildkitContext    string              `json:"buildkit_context" example:"."`
	BuildkitDockerfile string              `json:"buildkit_dockerfile" example:"Dockerfile"`
	BuildkitPlatforms  []enums.OciPlatform `json:"buildkit_platforms" example:"linux/amd64"`
}
