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

// PostBuilderRequest ...
type PostBuilderRequest struct {
	RepositoryID int64 `json:"repository_id" param:"repository_id" example:"10"`

	ScmCredentialType enums.ScmCredentialType `json:"scm_credential_type" example:"ssh"`
	ScmSshKey         string                  `json:"scm_ssh_key" example:"xxxx"`
	ScmToken          string                  `json:"scm_token" example:"xxxx"`
	ScmUsername       string                  `json:"scm_username" example:"sigma"`
	ScmPassword       string                  `json:"scm_password" example:"sigma"`
	// ScmProvider       enums.ScmProvider       `json:"scm_provider"`
	ScmRepository string `json:"scm_repository" example:"https://github.com/go-sigma/sigma.git"`
	ScmBranch     string `json:"scm_branch" example:"main"`
	ScmDepth      int    `json:"scm_depth" example:"0"`
	ScmSubmodule  bool   `json:"scm_submodule" example:"false"`
}

// PostBuilderRequestSwagger ...
type PostBuilderRequestSwagger struct {
	ScmCredentialType enums.ScmCredentialType `json:"scm_credential_type" example:"ssh"`
	ScmSshKey         string                  `json:"scm_ssh_key" example:"xxxx"`
	ScmToken          string                  `json:"scm_token" example:"xxxx"`
	ScmUsername       string                  `json:"scm_username" example:"sigma"`
	ScmPassword       string                  `json:"scm_password" example:"sigma"`
	// ScmProvider       enums.ScmProvider       `json:"scm_provider"`
	ScmRepository string `json:"scm_repository" example:"https://github.com/go-sigma/sigma.git"`
	ScmBranch     string `json:"scm_branch" example:"main"`
	ScmDepth      int    `json:"scm_depth" example:"0"`
	ScmSubmodule  bool   `json:"scm_submodule" example:"false"`
}
