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

package runtime

import (
	"context"
	"fmt"
	"io"
	"strings"

	corev1 "k8s.io/api/core/v1"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/background/build"
	buildlogger "github.com/go-sigma/sigma/pkg/background/build/logger"
	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/consts"
	dalredis "github.com/go-sigma/sigma/pkg/dal/redis"
	repobuilder "github.com/go-sigma/sigma/pkg/dal/repository/builder"
	reposetting "github.com/go-sigma/sigma/pkg/dal/repository/setting"
	repouser "github.com/go-sigma/sigma/pkg/dal/repository/user"
	"github.com/go-sigma/sigma/pkg/service/token"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/crypt"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

//go:generate mockgen -destination=runtime_mocks.go -package=runtime github.com/go-sigma/sigma/pkg/background/build/runtime Builder

// Builder controls one runtime backend for executing build runners.
type Builder interface {
	// Start launches a build runner with the given build configuration.
	Start(ctx context.Context, builderConfig Config) error
	// Stop terminates the build runner identified by builderID and runnerID.
	Stop(ctx context.Context, builderID, runnerID string) error
	// Restart terminates an existing build runner and starts it again.
	Restart(ctx context.Context, builderConfig Config) error
	// LogStream streams live runner logs to writer.
	LogStream(ctx context.Context, builderID, runnerID string, writer io.Writer) error
}

// Config contains the build request and runtime-specific host settings.
type Config struct {
	api.Builder
	ExtraHosts []string
}

// Driver is the selected build runner runtime driver.
var Driver Builder

// Factory creates a build runner runtime driver.
type Factory interface {
	New(config *config.Configuration, coordinator build.Coordinator) (Builder, error)
}

// DriverFactories stores registered build runner runtime driver factories.
var DriverFactories = make(map[string]Factory)

// Initialize selects and initializes the configured build runner runtime.
func Initialize(config *config.Configuration) error {
	if !config.Daemon.Builder.Enabled {
		return nil
	}
	builderType := config.Daemon.Builder.Type
	factory, ok := DriverFactories[builderType.String()]
	if !ok {
		return fmt.Errorf("builder driver %s not registered", builderType.String())
	}
	var err error
	err = buildlogger.InitializeLogStore()
	if err != nil {
		return err
	}
	coordinator := build.NewCoordinator(repobuilder.NewBuilderRepository(), buildlogger.LogStoreDriver)
	Driver, err = factory.New(config, coordinator)
	if err != nil {
		return err
	}
	return nil
}

// BuildEnv builds the environment variables passed to a build runtime.
func BuildEnv(builderConfig Config) ([]string, error) {
	config := config.GetConfig()

	ctx := context.Background()

	userRepository := repouser.NewUserRepository()
	userObj, err := userRepository.GetByUsername(ctx, consts.UserInternal)
	if err != nil {
		return nil, err
	}
	tokenSvc, err := token.NewWithConfig(config, dalredis.NewClientFactory(config))
	if err != nil {
		return nil, err
	}
	authorization, err := tokenSvc.New(userObj.ID, config.Auth.Jwt.Ttl)
	if err != nil {
		return nil, err
	}

	if after, ok := strings.CutPrefix(config.HTTP.InternalEndpoint, "https://"); ok {
		builderConfig.BuildkitInsecureRegistries = append(builderConfig.BuildkitInsecureRegistries, after)
	} else if after, ok := strings.CutPrefix(config.HTTP.InternalEndpoint, "http://"); ok {
		builderConfig.BuildkitInsecureRegistries = append(builderConfig.BuildkitInsecureRegistries, fmt.Sprintf("%s@http", after))
	}

	settingRepository := reposetting.NewSettingRepository()
	privateKey, err := settingRepository.Get(ctx, consts.SettingSignPrivateKey)
	if err != nil {
		return nil, err
	}

	buildConfigEnvs := []string{
		fmt.Sprintf("BUILDER_ID=%s", builderConfig.BuilderID),
		fmt.Sprintf("RUNNER_ID=%s", builderConfig.RunnerID),

		fmt.Sprintf("ENDPOINT=%s", config.HTTP.InternalEndpoint),
		fmt.Sprintf("ENDPOINT_TLS_VERIFY=%t", config.Daemon.Builder.TLSVerifyEnabled()),
		fmt.Sprintf("AUTHORIZATION=%s", crypt.MustEncrypt(fmt.Sprintf("%s-%s", builderConfig.BuilderID, builderConfig.RunnerID), authorization)),
		fmt.Sprintf("REPOSITORY=%s", builderConfig.Repository),
		fmt.Sprintf("TAG=%s", builderConfig.Tag),

		fmt.Sprintf("SOURCE=%s", builderConfig.Source.String()),

		fmt.Sprintf("DOCKERFILE=%s", ptr.To(builderConfig.Dockerfile)),

		fmt.Sprintf("OCI_REGISTRY_DOMAIN=%s", strings.Join(builderConfig.OciRegistryDomain, ",")),
		fmt.Sprintf("OCI_REGISTRY_USERNAME=%s", strings.Join(builderConfig.OciRegistryUsername, ",")),

		fmt.Sprintf("BUILDKIT_INSECURE_REGISTRIES=%s", strings.Join(builderConfig.BuildkitInsecureRegistries, ",")),
		fmt.Sprintf("BUILDKIT_CACHE_DIR=%s", builderConfig.BuildkitCacheDir),
		fmt.Sprintf("BUILDKIT_CONTEXT=%s", builderConfig.BuildkitContext),
		fmt.Sprintf("BUILDKIT_DOCKERFILE=%s", builderConfig.BuildkitDockerfile),
		fmt.Sprintf("BUILDKIT_PLATFORMS=%s", utils.StringsJoin(builderConfig.BuildkitPlatforms, ",")),
		fmt.Sprintf("BUILDKIT_BUILD_ARGS=%s", strings.Join(builderConfig.BuildkitBuildArgs, ",")),

		fmt.Sprintf("SIGNING_PRIVATE_KEY=%s", crypt.MustEncrypt(fmt.Sprintf("%s-%s", builderConfig.BuilderID, builderConfig.RunnerID), string(privateKey.Val))),
	}
	if builderConfig.Dockerfile != nil {
		buildConfigEnvs = append(buildConfigEnvs, fmt.Sprintf("DOCKERFILE=%s", ptr.To(builderConfig.Dockerfile)))
	}
	if builderConfig.ScmCredentialType != nil {
		buildConfigEnvs = append(buildConfigEnvs, fmt.Sprintf("SCM_CREDENTIAL_TYPE=%s", builderConfig.ScmCredentialType.String()))
	}
	if builderConfig.ScmProvider != nil {
		buildConfigEnvs = append(buildConfigEnvs, fmt.Sprintf("SCM_PROVIDER=%s", builderConfig.ScmProvider.String()))
	}
	if builderConfig.ScmRepository != nil {
		buildConfigEnvs = append(buildConfigEnvs, fmt.Sprintf("SCM_REPOSITORY=%s", ptr.To(builderConfig.ScmRepository)))
	}
	if builderConfig.ScmBranch != nil {
		buildConfigEnvs = append(buildConfigEnvs, fmt.Sprintf("SCM_BRANCH=%s", ptr.To(builderConfig.ScmBranch)))
	}
	if builderConfig.ScmDepth != nil {
		buildConfigEnvs = append(buildConfigEnvs, fmt.Sprintf("SCM_DEPTH=%d", ptr.To(builderConfig.ScmDepth)))
	}
	if builderConfig.ScmSubmodule != nil {
		buildConfigEnvs = append(buildConfigEnvs, fmt.Sprintf("SCM_SUBMODULE=%t", ptr.To(builderConfig.ScmSubmodule)))
	}
	if builderConfig.ScmUsername != nil {
		buildConfigEnvs = append(buildConfigEnvs, fmt.Sprintf("SCM_USERNAME=%s", ptr.To(builderConfig.ScmUsername)))
	}
	if builderConfig.ScmPassword != nil {
		buildConfigEnvs = append(buildConfigEnvs, fmt.Sprintf("SCM_PASSWORD=%s", crypt.MustEncrypt(
			fmt.Sprintf("%s-%s", builderConfig.BuilderID, builderConfig.RunnerID), ptr.To(builderConfig.ScmPassword))))
	}
	if builderConfig.ScmSshKey != nil {
		buildConfigEnvs = append(buildConfigEnvs, fmt.Sprintf("SCM_SSH_KEY=%s", crypt.MustEncrypt(
			fmt.Sprintf("%s-%s", builderConfig.BuilderID, builderConfig.RunnerID), ptr.To(builderConfig.ScmSshKey))))
	}
	if builderConfig.ScmToken != nil {
		buildConfigEnvs = append(buildConfigEnvs, fmt.Sprintf("SCM_TOKEN=%s", crypt.MustEncrypt(
			fmt.Sprintf("%s-%s", builderConfig.BuilderID, builderConfig.RunnerID), ptr.To(builderConfig.ScmToken))))
	}
	if len(builderConfig.OciRegistryPassword) != 0 {
		var passwords []string
		for _, p := range builderConfig.OciRegistryPassword {
			passwords = append(passwords, crypt.MustEncrypt(fmt.Sprintf("%s-%s", builderConfig.BuilderID, builderConfig.RunnerID), p))
		}
		buildConfigEnvs = append(buildConfigEnvs, fmt.Sprintf("OCI_REGISTRY_PASSWORD=%s", strings.Join(passwords, ",")))
	}
	if len(builderConfig.ExtraHosts) > 0 {
		buildConfigEnvs = append(buildConfigEnvs,
			fmt.Sprintf("EXTRA_HOSTS=%s", strings.Join(builderConfig.ExtraHosts, ",")))
	}

	return buildConfigEnvs, nil
}

// BuildK8sEnv converts build runner environment variables to Kubernetes values.
func BuildK8sEnv(builderConfig Config) ([]corev1.EnvVar, error) {
	envs, err := BuildEnv(builderConfig)
	if err != nil {
		return nil, err
	}
	var k8sEnvs = make([]corev1.EnvVar, 0, len(envs))
	for _, env := range envs {
		s := strings.SplitN(env, "=", 2)
		k8sEnvs = append(k8sEnvs, corev1.EnvVar{
			Name:  s[0],
			Value: s[1],
		})
	}
	return k8sEnvs, nil
}

// BuildEnvMap converts build runner environment variables to a name-value map.
func BuildEnvMap(builderConfig Config) (map[string]string, error) {
	envs, err := BuildEnv(builderConfig)
	if err != nil {
		return nil, err
	}
	var res = make(map[string]string, len(envs))
	for _, env := range envs {
		s := strings.SplitN(env, "=", 2)
		res[s[0]] = s[1]
	}
	return res, nil
}

const (
	// ContainerPrefix is the name prefix used for build runner containers.
	ContainerPrefix = "sigma-builder-"
)

// GenContainerID returns the runtime container name for a builder runtime.
func GenContainerID(builderID, runnerID string) string {
	return fmt.Sprintf("%s%s_%s", ContainerPrefix, builderID, runnerID)
}

// ParseContainerID extracts builder and runner IDs from a runtime container name.
func ParseContainerID(containerName string) (string, string, error) {
	containerName = strings.TrimPrefix(containerName, "/")
	ids := strings.TrimPrefix(containerName, ContainerPrefix)
	builderIDStr, runnerIDStr, ok := strings.Cut(ids, "_")
	if !ok {
		return "", "", fmt.Errorf("parse builder task id(%s) failed", containerName)
	}
	return builderIDStr, runnerIDStr, nil
}
