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

import (
	"github.com/go-sigma/sigma/pkg/api/enums"
)

// Builder describes the environment configuration consumed by a build runner.
type Builder struct {
	BuilderID string `env:"BUILDER_ID,notEmpty"`
	RunnerID  string `env:"RUNNER_ID,notEmpty"`

	Authorization string `env:"AUTHORIZATION,notEmpty"`
	Endpoint      string `env:"ENDPOINT,notEmpty"`
	Repository    string `env:"REPOSITORY,notEmpty"`
	Tag           string `env:"TAG,notEmpty"`

	Source enums.BuilderSource `env:"SOURCE,notEmpty"`

	Dockerfile *string `env:"DOCKERFILE"`

	ScmProvider       *enums.ScmProvider       `env:"SCM_PROVIDER"`
	ScmCredentialType *enums.ScmCredentialType `env:"SCM_CREDENTIAL_TYPE"`
	ScmSshKey         *string                  `env:"SCM_SSH_KEY"`
	ScmToken          *string                  `env:"SCM_TOKEN"`
	ScmUsername       *string                  `env:"SCM_USERNAME"`
	ScmPassword       *string                  `env:"SCM_PASSWORD"`
	ScmRepository     *string                  `env:"SCM_REPOSITORY"`
	ScmBranch         *string                  `env:"SCM_BRANCH" envDefault:"main"`
	ScmDepth          *int                     `env:"SCM_DEPTH" envDefault:"0"`
	ScmSubmodule      *bool                    `env:"SCM_SUBMODULE" envDefault:"false"`

	OciRegistryDomain   []string `env:"OCI_REGISTRY_DOMAIN" envSeparator:","`
	OciRegistryUsername []string `env:"OCI_REGISTRY_USERNAME" envSeparator:","`
	OciRegistryPassword []string `env:"OCI_REGISTRY_PASSWORD" envSeparator:","`

	BuildkitInsecureRegistries []string            `env:"BUILDKIT_INSECURE_REGISTRIES" envSeparator:","`
	BuildkitCacheDir           string              `env:"BUILDKIT_CACHE_DIR" envDefault:"/tmp/buildkit"`
	BuildkitContext            string              `env:"BUILDKIT_CONTEXT" envDefault:"."`
	BuildkitDockerfile         string              `env:"BUILDKIT_DOCKERFILE" envDefault:"Dockerfile"`
	BuildkitPlatforms          []enums.OciPlatform `env:"BUILDKIT_PLATFORMS" envSeparator:","`
	BuildkitBuildArgs          []string            `env:"BUILDKIT_BUILD_ARGS" envSeparator:","`

	SigningPrivateKey string `env:"SIGNING_PRIVATE_KEY,notEmpty"`
}

// GetBuilderRequest represents the request to get a builder.
type GetBuilderRequest struct {
	Namespace    string `json:"namespace" param:"namespace" validate:"required,min=2,max=20,is_valid_namespace" example:"library"`
	RepositoryID string `json:"repository_id" param:"repository_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// BuilderItem represents a builder configuration returned by the API.
type BuilderItem struct {
	ID string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`

	RepositoryID string `json:"repository_id" example:"550e8400-e29b-41d4-a716-446655440000"`

	Source enums.BuilderSource `json:"source" example:"Dockerfile"`

	// source CodeRepository
	CodeRepositoryID *string `json:"code_repository_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	// source Dockerfile
	Dockerfile *string `json:"dockerfile" example:"xxx"`
	// source SelfCodeRepository
	ScmRepository     *string                  `json:"scm_repository" example:"https://github.com/go-sigma/sigma.git"`
	ScmCredentialType *enums.ScmCredentialType `json:"scm_credential_type" example:"ssh"`
	ScmSshKey         *string                  `json:"scm_ssh_key" example:"xxxx"`
	ScmToken          *string                  `json:"scm_token" example:"xxxx"`
	ScmUsername       *string                  `json:"scm_username" example:"sigma"`
	ScmPassword       *string                  `json:"scm_password" example:"sigma"`
	ScmProvider       *enums.ScmProvider       `json:"scm_provider" example:"github"`

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

// PostOrPutBuilderRequest is the shared request body for creating or updating a builder.
type PostOrPutBuilderRequest struct {
	Source      enums.BuilderSource `json:"source" example:"Dockerfile"`
	ScmProvider *enums.ScmProvider  `json:"scm_provider"`

	// source CodeRepository
	CodeRepositoryID *string `json:"code_repository_id" example:"550e8400-e29b-41d4-a716-446655440000"`
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

// CreateBuilderRequest is the request for creating a builder.
type CreateBuilderRequest struct {
	NamespaceID  string `json:"namespace_id" param:"namespace_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000" swaggerignore:"true"`
	RepositoryID string `json:"repository_id" param:"repository_id" example:"550e8400-e29b-41d4-a716-446655440000" swaggerignore:"true"`

	PostOrPutBuilderRequest
}

// UpdateBuilderRequest is the request for updating a builder.
type UpdateBuilderRequest struct {
	NamespaceID  string `json:"namespace_id" param:"namespace_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000" swaggerignore:"true"`
	RepositoryID string `json:"repository_id" param:"repository_id" example:"550e8400-e29b-41d4-a716-446655440000" swaggerignore:"true"`
	BuilderID    string `json:"builder_id" param:"builder_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000" swaggerignore:"true"`

	PostOrPutBuilderRequest
}

// ListBuilderRunnersRequest is the request for listing builder runners.
type ListBuilderRunnersRequest struct {
	Pagination
	Sortable

	NamespaceID  string `json:"namespace_id" param:"namespace_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	RepositoryID string `json:"repository_id" param:"repository_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	BuilderID    string `json:"builder_id" param:"builder_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// BuilderRunnerItem represents a single build runner execution.
type BuilderRunnerItem struct {
	ID            string            `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	BuilderID     string            `json:"builder_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Log           []byte            `json:"log" example:"log"`
	Status        enums.BuildStatus `json:"status" example:"Success"`
	StatusMessage *string           `json:"status_message" example:""`

	Tag         *string `json:"tag" example:"v1.0"`
	RawTag      string  `json:"raw_tag" example:"v1.0"`
	Description *string `json:"description" example:"description"`
	ScmBranch   *string `json:"scm_branch" example:"main"`

	StartedAt   *int64  `json:"started_at" example:"1702128050507"`
	EndedAt     *int64  `json:"ended_at" example:"1702128050507"`
	RawDuration *int64  `json:"raw_duration" example:"1000"`
	Duration    *string `json:"duration" example:"1h"`

	CreatedAt string `json:"created_at" example:"2006-01-02 15:04:05"`
	UpdatedAt string `json:"updated_at" example:"2006-01-02 15:04:05"`
}

// PostRunnerRun is the request for starting a builder runner.
type PostRunnerRun struct {
	NamespaceID  string `json:"namespace_id" param:"namespace_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	RepositoryID string `json:"repository_id" param:"repository_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	BuilderID    string `json:"builder_id" param:"builder_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`

	RawTag      string  `json:"raw_tag" example:"test"` // TODO: validate
	Description *string `json:"description,omitempty" validate:"omitempty,max=50"`
	ScmBranch   *string `json:"scm_branch,omitempty" validate:"omitempty,min=1,max=64" example:"main"`
}

// GetRunnerRerun is the request for rerunning a builder runner.
type GetRunnerRerun struct {
	NamespaceID  string `json:"namespace_id" param:"namespace_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	RepositoryID string `json:"repository_id" param:"repository_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	BuilderID    string `json:"builder_id" param:"builder_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	RunnerID     string `json:"runner_id" param:"runner_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// RunOrRerunRunnerResponse is the response returned after starting or rerunning a runner.
type RunOrRerunRunnerResponse struct {
	RunnerID string `json:"runner_id" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// GetRunnerStop is the request for stopping a builder runner.
type GetRunnerStop struct {
	NamespaceID  string `json:"namespace_id" param:"namespace_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	RepositoryID string `json:"repository_id" param:"repository_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	BuilderID    string `json:"builder_id" param:"builder_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	RunnerID     string `json:"runner_id" param:"runner_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// GetRunnerLog is the request for reading a builder runner log.
type GetRunnerLog struct {
	NamespaceID  string `json:"namespace_id" param:"namespace_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	RepositoryID string `json:"repository_id" param:"repository_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	BuilderID    string `json:"builder_id" param:"builder_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	RunnerID     string `json:"runner_id" param:"runner_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// GetRunner is the request for getting a builder runner.
type GetRunner struct {
	NamespaceID  string `json:"namespace_id" param:"namespace_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	RepositoryID string `json:"repository_id" param:"repository_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	BuilderID    string `json:"builder_id" param:"builder_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	RunnerID     string `json:"runner_id" param:"runner_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// BuildTagOption contains SCM values used to render build tag templates.
type BuildTagOption struct {
	ScmBranch string
	ScmTag    string
	ScmRef    string
}
