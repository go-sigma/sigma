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

import (
	"github.com/go-sigma/sigma/pkg/types/enums"
)

// Builder config for builder
type Builder struct {
	BuilderID int64 `env:"ID,notEmpty"`
	RunnerID  int64 `env:"RUNNER_ID,notEmpty"`

	ScmCredentialType enums.ScmCredentialType `env:"SCM_CREDENTIAL_TYPE,notEmpty"`
	ScmSshKey         string                  `env:"SCM_SSH_KEY"`
	ScmToken          string                  `env:"SCM_TOKEN"`
	ScmUsername       string                  `env:"SCM_USERNAME"`
	ScmPassword       string                  `env:"SCM_PASSWORD"`
	ScmProvider       enums.ScmProvider       `env:"SCM_PROVIDER,notEmpty"`
	ScmRepository     string                  `env:"SCM_REPOSITORY,notEmpty"`
	ScmBranch         string                  `env:"SCM_BRANCH" envDefault:"main"`
	ScmDepth          int                     `env:"SCM_DEPTH" envDefault:"0"`
	ScmSubmodule      bool                    `env:"SCM_SUBMODULE" envDefault:"false"`

	OciRegistryDomain   []string `env:"OCI_REGISTRY_DOMAIN" envSeparator:","`
	OciRegistryUsername []string `env:"OCI_REGISTRY_USERNAME" envSeparator:","`
	OciRegistryPassword []string `env:"OCI_REGISTRY_PASSWORD" envSeparator:","`
	OciName             string   `env:"OCI_NAME,notEmpty"`

	BuildkitInsecureRegistries []string            `env:"BUILDKIT_INSECURE_REGISTRIES" envSeparator:","`
	BuildkitCacheDir           string              `env:"BUILDKIT_CACHE_DIR" envDefault:"/tmp/buildkit"`
	BuildkitContext            string              `env:"BUILDKIT_CONTEXT" envDefault:"."`
	BuildkitDockerfile         string              `env:"BUILDKIT_DOCKERFILE" envDefault:"Dockerfile"`
	BuildkitPlatforms          []enums.OciPlatform `env:"BUILDKIT_PLATFORMS" envSeparator:","`
}

// GetBuilderRequest represents the request to get a builder.
type GetBuilderRequest struct {
	RepositoryID int64 `json:"repository_id" param:"repository_id" validate:"required,number" example:"10"`
	ID           int64 `json:"id" param:"id" validate:"required,number" example:"10"`
}

// BuilderItem ...
type BuilderItem struct {
	ID int64 `json:"id" example:"10"`

	RepositoryID int64 `json:"repository_id" example:"10"`

	Source enums.BuilderSource `json:"source" example:"Dockerfile"`

	// source CodeRepository
	CodeRepositoryID *int64 `json:"code_repository_id" example:"10"`
	// source Dockerfile
	Dockerfile *string `json:"dockerfile" example:"xxx"`
	// source SelfCodeRepository
	ScmRepository     *string                  `json:"scm_repository" example:"https://github.com/go-sigma/sigma.git"`
	ScmCredentialType *enums.ScmCredentialType `json:"scm_credential_type" example:"ssh"`
	ScmSshKey         *string                  `json:"scm_ssh_key" example:"xxxx"`
	ScmToken          *string                  `json:"scm_token" example:"xxxx"`
	ScmUsername       *string                  `json:"scm_username" example:"sigma"`
	ScmPassword       *string                  `json:"scm_password" example:"sigma"`

	ScmBranch *string `json:"scm_branch" example:"main"`

	ScmDepth     *int  `json:"scm_depth" example:"0"`
	ScmSubmodule *bool `json:"scm_submodule" example:"false"`

	CronRule        *string `json:"cron_rule" example:"* * * * *"`
	CronBranch      *string `json:"cron_branch" example:"main"`
	CronTagTemplate *string `json:"cron_tag_template" example:"{.Ref}"`

	WebhookBranchName        *string `json:"webhook_branch_name" example:"main"`
	WebhookBranchTagTemplate *string `json:"webhook_branch_tag_template" example:"{.Ref}"`
	WebhookTagTagTemplate    *string `json:"webhook_tag_tag_template" example:"{.Ref}"`

	BuildkitInsecureRegistries []string            `json:"buildkit_insecure_registries,omitempty" example:"test.com,xxx.com@http"`
	BuildkitContext            *string             `json:"buildkit_context"`
	BuildkitDockerfile         *string             `json:"buildkit_dockerfile"`
	BuildkitPlatforms          []enums.OciPlatform `json:"buildkit_platforms" example:"linux/amd64"`
	BuildkitBuildArgs          *string             `json:"buildkit_build_args" example:"a=b,c=d"`
}

// PostOrPutBuilderRequest ...
type PostOrPutBuilderRequest struct {
	Source      enums.BuilderSource `json:"source" example:"Dockerfile"`
	ScmProvider *enums.ScmProvider  `json:"scm_provider"`

	// source CodeRepository
	CodeRepositoryID *int64 `json:"code_repository_id" example:"10"`
	// source Dockerfile
	Dockerfile *string `json:"dockerfile" example:"xxx"`
	// source SelfCodeRepository
	ScmRepository     *string                  `json:"scm_repository" example:"https://github.com/go-sigma/sigma.git"`
	ScmCredentialType *enums.ScmCredentialType `json:"scm_credential_type,omitempty" validate:"omitempty,is_valid_scm_credential_type" example:"ssh"`
	ScmSshKey         *string                  `json:"scm_ssh_key" example:"xxxx"`
	ScmToken          *string                  `json:"scm_token" example:"xxxx"`
	ScmUsername       *string                  `json:"scm_username" example:"sigma"`
	ScmPassword       *string                  `json:"scm_password" example:"sigma"`

	ScmBranch *string `json:"scm_branch,omitempty" validate:"omitempty,min=1,max=50" example:"main"`

	ScmDepth     *int  `json:"scm_depth,omitempty" validate:"omitempty,min=0" example:"0"`
	ScmSubmodule *bool `json:"scm_submodule,omitempty" example:"false"`

	CronRule        *string `json:"cron_rule" example:"* * * * *"` // TODO: validate
	CronBranch      *string `json:"cron_branch" example:"main"`
	CronTagTemplate *string `json:"cron_tag_template" example:"{.Ref}"`

	WebhookBranchName        *string `json:"webhook_branch_name" example:"main"`
	WebhookBranchTagTemplate *string `json:"webhook_branch_tag_template" example:"{.Ref}"`
	WebhookTagTagTemplate    *string `json:"webhook_tag_tag_template" example:"{.Ref}"` // TODO: validate

	BuildkitInsecureRegistries []string            `json:"buildkit_insecure_registries,omitempty" example:"test.com,xxx.com@http" validate:"omitempty,max=3"`
	BuildkitContext            *string             `json:"buildkit_context,omitempty" validate:"omitempty,min=1,max=255"`
	BuildkitDockerfile         *string             `json:"buildkit_dockerfile,omitempty" validate:"omitempty,min=1,max=255"`
	BuildkitPlatforms          []enums.OciPlatform `json:"buildkit_platforms" validate:"required,min=1,is_valid_oci_platforms" example:"linux/amd64"`
	BuildkitBuildArgs          *string             `json:"buildkit_build_args" example:"a=b,c=d"` // TODO: validate
}

// PostBuilderRequest ...
type PostBuilderRequest struct {
	RepositoryID int64 `json:"repository_id" param:"repository_id" example:"10"`

	PostOrPutBuilderRequest
}

// PostBuilderRequestSwagger ...
type PostBuilderRequestSwagger struct {
	PostOrPutBuilderRequest
}

// PutBuilderRequest represents the request to get a builder.
type PutBuilderRequest struct {
	ID           int64 `json:"id" param:"id" validate:"required,number"`
	RepositoryID int64 `json:"repository_id" param:"repository_id" example:"10"`

	PostOrPutBuilderRequest
}

// PutBuilderRequestSwagger
type PutBuilderRequestSwagger struct {
	PostOrPutBuilderRequest
}

// ListBuilderRunnersRequest ...
type ListBuilderRunnersRequest struct {
	Pagination
	Sortable

	ID           int64 `json:"id" param:"id" validate:"required,number"`
	RepositoryID int64 `json:"repository_id" param:"repository_id" example:"10"`
}

// BuilderRunnerItem ...
type BuilderRunnerItem struct {
	ID        int64 `json:"id" example:"10"`
	BuilderID int64
	Log       []byte
	Status    enums.BuildStatus

	Tag               string
	ScmBranch         string
	BuildkitPlatforms []enums.OciPlatform `json:"buildkit_platforms" example:"linux/amd64"`

	CreatedAt string `json:"created_at" example:"2006-01-02 15:04:05"`
	UpdatedAt string `json:"updated_at" example:"2006-01-02 15:04:05"`
}
