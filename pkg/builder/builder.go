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
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
	"go.uber.org/dig"
	corev1 "k8s.io/api/core/v1"

	"github.com/go-sigma/sigma/pkg/builder/logger"
	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/dao"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/crypt"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
	"github.com/go-sigma/sigma/pkg/utils/token"
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

// BuilderConfig ...
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

// Initialize ...
func Initialize(config configs.Configuration) error {
	if !config.Daemon.Builder.Enabled {
		return nil
	}
	typ := config.Daemon.Builder.Type
	factory, ok := DriverFactories[typ.String()]
	if !ok {
		return fmt.Errorf("builder driver %s not registered", typ.String())
	}
	var err error
	err = logger.Initialize()
	if err != nil {
		return err
	}
	Driver, err = factory.New(config)
	if err != nil {
		return err
	}
	return nil
}

// BuildEnv ...
func BuildEnv(builderConfig BuilderConfig) ([]string, error) {
	config := configs.GetConfiguration()

	ctx := log.Logger.WithContext(context.Background())

	userService := dao.NewUserServiceFactory().New()
	userObj, err := userService.GetByUsername(ctx, consts.UserInternal)
	if err != nil {
		return nil, err
	}
	tokenService, err := token.New(dig.New()) // TODO: dig
	if err != nil {
		return nil, err
	}
	authorization, err := tokenService.New(userObj.ID, config.Auth.Jwt.Ttl)
	if err != nil {
		return nil, err
	}

	if strings.HasPrefix(config.HTTP.InternalEndpoint, "https://") {
		builderConfig.BuildkitInsecureRegistries = append(builderConfig.BuildkitInsecureRegistries, strings.TrimPrefix(config.HTTP.InternalEndpoint, "https://"))
	} else if strings.HasPrefix(config.HTTP.InternalEndpoint, "http://") {
		builderConfig.BuildkitInsecureRegistries = append(builderConfig.BuildkitInsecureRegistries, fmt.Sprintf("%s@http", strings.TrimPrefix(config.HTTP.InternalEndpoint, "http://")))
	}

	settingService := dao.NewSettingServiceFactory().New()
	privateKey, err := settingService.Get(ctx, consts.SettingSignPrivateKey)
	if err != nil {
		return nil, err
	}

	buildConfigEnvs := []string{
		fmt.Sprintf("BUILDER_ID=%d", builderConfig.BuilderID),
		fmt.Sprintf("RUNNER_ID=%d", builderConfig.RunnerID),

		fmt.Sprintf("ENDPOINT=%s", config.HTTP.InternalEndpoint),
		fmt.Sprintf("AUTHORIZATION=%s", crypt.MustEncrypt(fmt.Sprintf("%d-%d", builderConfig.BuilderID, builderConfig.RunnerID), authorization)),
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

		fmt.Sprintf("SIGNING_PRIVATE_KEY=%s", crypt.MustEncrypt(fmt.Sprintf("%d-%d", builderConfig.BuilderID, builderConfig.RunnerID), string(privateKey.Val))),
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
			fmt.Sprintf("%d-%d", builderConfig.BuilderID, builderConfig.RunnerID), ptr.To(builderConfig.ScmPassword))))
	}
	if builderConfig.ScmSshKey != nil {
		buildConfigEnvs = append(buildConfigEnvs, fmt.Sprintf("SCM_SSH_KEY=%s", crypt.MustEncrypt(
			fmt.Sprintf("%d-%d", builderConfig.BuilderID, builderConfig.RunnerID), ptr.To(builderConfig.ScmSshKey))))
	}
	if builderConfig.ScmToken != nil {
		buildConfigEnvs = append(buildConfigEnvs, fmt.Sprintf("SCM_TOKEN=%s", crypt.MustEncrypt(
			fmt.Sprintf("%d-%d", builderConfig.BuilderID, builderConfig.RunnerID), ptr.To(builderConfig.ScmToken))))
	}
	if len(builderConfig.OciRegistryPassword) != 0 {
		var passwords []string
		for _, p := range builderConfig.OciRegistryPassword {
			passwords = append(passwords, crypt.MustEncrypt(fmt.Sprintf("%d-%d", builderConfig.BuilderID, builderConfig.RunnerID), p))
		}
		buildConfigEnvs = append(buildConfigEnvs, fmt.Sprintf("OCI_REGISTRY_PASSWORD=%s", strings.Join(passwords, ",")))
	}

	return buildConfigEnvs, nil
}

// BuildK8sEnv ...
func BuildK8sEnv(builderConfig BuilderConfig) ([]corev1.EnvVar, error) {
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

// BuildEnvMap ...
func BuildEnvMap(builderConfig BuilderConfig) (map[string]string, error) {
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
	// ContainerPrefix ...
	ContainerPrefix = "sigma-builder-"
)

// GenContainerID ...
func GenContainerID(builderID, runnerID int64) string {
	return fmt.Sprintf("%s%d-%d", ContainerPrefix, builderID, runnerID)
}

// ParseContainerID ...
func ParseContainerID(containerName string) (int64, int64, error) {
	containerName = strings.TrimPrefix(containerName, "/")
	ids := strings.TrimPrefix(containerName, ContainerPrefix)
	if len(strings.Split(ids, "-")) != 2 {
		return 0, 0, fmt.Errorf("Parse builder task id(%s) failed", containerName)
	}
	builderIDStr, runnerIDStr := strings.Split(ids, "-")[0], strings.Split(ids, "-")[1]
	builderID, err := strconv.ParseInt(builderIDStr, 10, 0)
	if err != nil {
		return 0, 0, fmt.Errorf("Parse builder task id(%s) failed: %v", containerName, err)
	}
	runnerID, err := strconv.ParseInt(runnerIDStr, 10, 0)
	if err != nil {
		return 0, 0, fmt.Errorf("Parse builder task id(%s) failed: %v", containerName, err)
	}
	return builderID, runnerID, nil
}
