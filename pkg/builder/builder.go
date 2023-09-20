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
	"context"
	"fmt"
	"io"
	"strings"

	corev1 "k8s.io/api/core/v1"

	builderlogger "github.com/go-sigma/sigma/pkg/builder/logger"
	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/crypt"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

// Builder ...
type Builder interface {
	// Start start a container to build oci image and push to registry
	Start(ctx context.Context, builderConfig BuilderConfig) error
	// Stop stop the container
	Stop(ctx context.Context, builderID, runnerID int64) error
	// Restart wrap stop and start
	Restart(ctx context.Context, builderConfig BuilderConfig) error
	// LogStream get the real time log stream
	LogStream(ctx context.Context, builderID, runnerID int64, writer io.Writer) error
}

type BuilderConfig struct {
	types.Builder
}

// Driver is the builder driver, maybe implement by docker, podman, k8s, etc.
var Driver Builder

// Factory is the interface for the builder driver factory
type Factory interface {
	New(config configs.Configuration) (Builder, error)
}

// DriverFactories ...
var DriverFactories = make(map[string]Factory)

func Initialize() error {
	typ := "docker"
	factory, ok := DriverFactories[typ]
	if !ok {
		return fmt.Errorf("builder driver %q not registered", typ)
	}
	var err error
	Driver, err = factory.New(configs.Configuration{})
	if err != nil {
		return err
	}
	return builderlogger.Initialize()
}

// BuildEnv ...
func BuildEnv(builderConfig BuilderConfig) []string {
	buildConfigEnvs := []string{
		fmt.Sprintf("ID=%d", builderConfig.BuilderID),
		fmt.Sprintf("RUNNER_ID=%d", builderConfig.RunnerID),

		fmt.Sprintf("SCM_CREDENTIAL_TYPE=%s", builderConfig.ScmCredentialType.String()),
		fmt.Sprintf("SCM_USERNAME=%s", builderConfig.ScmUsername),
		fmt.Sprintf("SCM_PROVIDER=%s", builderConfig.ScmProvider.String()),
		fmt.Sprintf("SCM_REPOSITORY=%s", builderConfig.ScmRepository),
		fmt.Sprintf("SCM_BRANCH=%s", ptr.To(builderConfig.ScmBranch)),
		fmt.Sprintf("SCM_DEPTH=%d", builderConfig.ScmDepth),
		fmt.Sprintf("SCM_SUBMODULE=%t", builderConfig.ScmSubmodule),

		fmt.Sprintf("OCI_REGISTRY_DOMAIN=%s", strings.Join(builderConfig.OciRegistryDomain, ",")),
		fmt.Sprintf("OCI_REGISTRY_USERNAME=%s", strings.Join(builderConfig.OciRegistryUsername, ",")),
		fmt.Sprintf("OCI_NAME=%s", builderConfig.OciName),

		fmt.Sprintf("BUILDKIT_INSECURE_REGISTRIES=%s", strings.Join(builderConfig.BuildkitInsecureRegistries, ",")),
		fmt.Sprintf("BUILDKIT_CACHE_DIR=%s", builderConfig.BuildkitCacheDir),
		fmt.Sprintf("BUILDKIT_CONTEXT=%s", builderConfig.BuildkitContext),
		fmt.Sprintf("BUILDKIT_DOCKERFILE=%s", builderConfig.BuildkitDockerfile),
		fmt.Sprintf("BUILDKIT_PLATFORMS=%s", utils.StringsJoin(builderConfig.BuildkitPlatforms, ",")),
	}
	if builderConfig.ScmPassword != "" {
		buildConfigEnvs = append(buildConfigEnvs, fmt.Sprintf("SCM_PASSWORD=%s", crypt.MustEncrypt(
			fmt.Sprintf("%d-%d", builderConfig.BuilderID, builderConfig.RunnerID), builderConfig.ScmPassword)))
	}
	if builderConfig.ScmSshKey != "" {
		buildConfigEnvs = append(buildConfigEnvs, fmt.Sprintf("SCM_SSH_KEY=%s", crypt.MustEncrypt(
			fmt.Sprintf("%d-%d", builderConfig.BuilderID, builderConfig.RunnerID), builderConfig.ScmSshKey)))
	}
	if builderConfig.ScmToken != "" {
		buildConfigEnvs = append(buildConfigEnvs, fmt.Sprintf("SCM_TOKEN=%s", crypt.MustEncrypt(
			fmt.Sprintf("%d-%d", builderConfig.BuilderID, builderConfig.RunnerID), builderConfig.ScmToken)))
	}
	if len(builderConfig.OciRegistryPassword) != 0 {
		var passwords []string
		for _, p := range builderConfig.OciRegistryPassword {
			passwords = append(passwords, crypt.MustEncrypt(fmt.Sprintf("%d-%d", builderConfig.BuilderID, builderConfig.RunnerID), p))
		}
		buildConfigEnvs = append(buildConfigEnvs, fmt.Sprintf("OCI_REGISTRY_PASSWORD=%s", strings.Join(passwords, ",")))
	}

	return buildConfigEnvs
}

// BuildK8sEnv ...
func BuildK8sEnv(builderConfig BuilderConfig) []corev1.EnvVar {
	envs := BuildEnv(builderConfig)
	var k8sEnvs = make([]corev1.EnvVar, 0, len(envs))
	for _, env := range envs {
		s := strings.SplitN(env, "=", 2)
		k8sEnvs = append(k8sEnvs, corev1.EnvVar{
			Name:  s[0],
			Value: s[1],
		})
	}
	return k8sEnvs
}
