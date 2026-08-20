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

package builder

import (
	"github.com/go-sigma/sigma/pkg/api/enums"
)

// Config 构建流程运行时配置（从环境变量解析）
type Config struct {
	BuilderID string `env:"BUILDER_ID,notEmpty"`
	RunnerID  string `env:"RUNNER_ID,notEmpty"`

	Authorization     string `env:"AUTHORIZATION,notEmpty"`
	Endpoint          string `env:"ENDPOINT,notEmpty"`
	EndpointTLSVerify bool   `env:"ENDPOINT_TLS_VERIFY" envDefault:"true"`
	Repository        string `env:"REPOSITORY,notEmpty"`
	Tag               string `env:"TAG,notEmpty"`

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

	ExtraHosts []string `env:"EXTRA_HOSTS" envSeparator:","`
}
