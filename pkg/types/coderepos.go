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

// CodeRepositoryItem ...
type CodeRepositoryItem struct {
	ID        int64  `json:"id" example:"1"`
	Name      string `json:"name" example:"sigma"`
	Owner     string `json:"owner" example:"go-sigma"`
	CloneUrl  string `json:"clone_url" example:"https://github.com/go-sigma/sigma.git"`
	SshUrl    string `json:"ssh_url" example:"git@github.com:go-sigma/sigma.git"`
	CreatedAt string `json:"created_at" example:"2006-01-02 15:04:05"`
	UpdatedAt string `json:"updated_at" example:"2006-01-02 15:04:05"`
}

// CodeRepositoryOwnerItem ...
type CodeRepositoryOwnerItem struct {
	ID        int64  `json:"id" example:"1"`
	Name      string `json:"name" example:"go-sigma"`
	CreatedAt string `json:"created_at" example:"2006-01-02 15:04:05"`
	UpdatedAt string `json:"updated_at" example:"2006-01-02 15:04:05"`
}

// ListCodeRepositoryRequest represents the request to list code repository.
type ListCodeRepositoryRequest struct {
	Pagination
	Sortable

	Provider enums.Provider `json:"provider" query:"provider" validate:"required,is_valid_provider"`
	Owner    *string        `json:"owner,omitempty" query:"owner" validate:"omitempty,min=1"`
	Name     *string        `json:"name,omitempty" query:"name" validate:"omitempty,min=1"`
}

// ListCodeRepositoryOwnerRequest ....
type ListCodeRepositoryOwnerRequest struct {
	Pagination
	Sortable

	Provider enums.Provider `json:"provider" query:"provider" validate:"required,is_valid_provider"`
	Name     *string        `json:"name,omitempty" query:"name" validate:"omitempty,min=1"`
}

// PostCodeRepositorySetupBuilder ...
type PostCodeRepositorySetupBuilder struct {
	ID int64 `json:"id" param:"id" validate:"required,number"`

	NamespaceID    int64  `json:"namespace_id" validate:"required,number"`
	RepositoryName string `json:"repository_name" validate:"required,is_valid_repository" example:"library/test"`

	ScmBranch    string `json:"scm_branch" example:"main"`
	ScmDepth     int    `json:"scm_depth" example:"0"`
	ScmSubmodule bool   `json:"scm_submodule" example:"false"`

	BuildkitContext    string              `json:"buildkit_context" example:"."`
	BuildkitDockerfile string              `json:"buildkit_dockerfile" example:"Dockerfile"`
	BuildkitPlatforms  []enums.OciPlatform `json:"buildkit_platforms" example:"linux/amd64"`
}

// PostCodeRepositorySetupBuilderSwagger ...
type PostCodeRepositorySetupBuilderSwagger struct {
	NamespaceID    int64  `json:"namespace_id" validate:"required,number"`
	RepositoryName string `json:"repository_name" validate:"required,is_valid_repository" example:"library/test"`

	ScmBranch    string `json:"scm_branch" example:"main"`
	ScmDepth     int    `json:"scm_depth" example:"0"`
	ScmSubmodule bool   `json:"scm_submodule" example:"false"`

	BuildkitContext    string              `json:"buildkit_context" example:"."`
	BuildkitDockerfile string              `json:"buildkit_dockerfile" example:"Dockerfile"`
	BuildkitPlatforms  []enums.OciPlatform `json:"buildkit_platforms" example:"linux/amd64"`
}
