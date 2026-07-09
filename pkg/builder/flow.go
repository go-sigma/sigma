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
	"log/slog"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/caarlos0/env/v9"

	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/builder/source"
	"github.com/go-sigma/sigma/pkg/distribution/signing"
	"github.com/go-sigma/sigma/pkg/utils"
)

// BuildFlow 构建流程编排器，管理镜像构建的完整生命周期
type BuildFlow struct {
	Config
	api apiClient
}

// Run 执行完整的构建流程
func (f *BuildFlow) Run() error {
	if err := f.initialize(); err != nil {
		return err
	}
	if err := env.Parse(f); err != nil {
		return err
	}
	if err := f.checker(); err != nil {
		return err
	}
	f.api = newAPIClient(f.Authorization, f.Endpoint)
	if err := f.initCache(); err != nil {
		return err
	}
	if err := f.initToken(); err != nil {
		return err
	}

	var src source.Source
	if f.Source == enums.BuilderSourceDockerfile { // nolint: staticcheck
		src = source.DockerfileSource{
			Dockerfile: f.Dockerfile,
		}
	} else {
		src = source.GitSource{
			ScmProvider:       f.ScmProvider,
			ScmCredentialType: f.ScmCredentialType,
			ScmSshKey:         f.ScmSshKey,
			ScmToken:          f.ScmToken,
			ScmUsername:       f.ScmUsername,
			ScmPassword:       f.ScmPassword,
			ScmRepository:     f.ScmRepository,
			ScmBranch:         f.ScmBranch,
			ScmDepth:          f.ScmDepth,
			ScmSubmodule:      f.ScmSubmodule,
		}
	}
	if err := src.Prepare(); err != nil {
		return err
	}

	imageName, err := f.genTag()
	if err != nil {
		return err
	}
	if err := f.build(imageName); err != nil {
		return err
	}
	if err := f.sign(imageName); err != nil {
		return err
	}
	return f.exportCache()
}

// initialize 创建必需的目录
func (f *BuildFlow) initialize() error {
	var dirs = []string{homeSigma, cache}
	for _, dir := range dirs {
		if !utils.IsDir(dir) {
			err := os.MkdirAll(dir, 0755)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// build 使用 buildctl-daemonless.sh 构建镜像
func (f *BuildFlow) build(imageName string) error {
	slog.Info("start to build image")
	buildCtl, err := exec.LookPath("buildctl-daemonless.sh")
	if err != nil {
		return fmt.Errorf("cannot find the buildctl-daemonless.sh: %v", err)
	}
	cmd := exec.Command(buildCtl, "build")
	cmd.Args = append(cmd.Args, "--local", fmt.Sprintf("context=%s", path.Join(workspace, f.BuildkitContext)))
	cmd.Args = append(cmd.Args, "--local", fmt.Sprintf("dockerfile=%s", path.Join(workspace, f.BuildkitContext)))
	cmd.Args = append(cmd.Args, "--progress", "plain")
	if len(f.BuildkitPlatforms) > 0 {
		var platforms = make([]string, 0, len(f.BuildkitPlatforms))
		for _, platform := range f.BuildkitPlatforms {
			platforms = append(platforms, platform.String())
		}
		cmd.Args = append(cmd.Args, "--opt", fmt.Sprintf("platform=%s", strings.Join(platforms, ",")))
	}
	cmd.Args = append(cmd.Args, "--frontend", "gateway.v0", "--opt", "source=docker/dockerfile")
	cmd.Args = append(cmd.Args, "--export-cache", fmt.Sprintf("type=local,mode=max,compression=gzip,dest=%s", cacheOut))
	cmd.Args = append(cmd.Args, "--import-cache", fmt.Sprintf("type=local,src=%s", cacheIn))

	if len(f.BuildkitPlatforms) > 1 {
		cmd.Args = append(cmd.Args, "--output",
			fmt.Sprintf("type=image,name=%s,annotation-index.org.opencontainers.sigma.builder_id=%s,annotation-index.org.opencontainers.sigma.runner_id=%s,push=true,oci-mediatypes=true", imageName, f.BuilderID, f.RunnerID))
	} else {
		cmd.Args = append(cmd.Args, "--output",
			fmt.Sprintf("type=image,name=%s,annotation.org.opencontainers.sigma.builder_id=%s,annotation.org.opencontainers.sigma.runner_id=%s,push=true,oci-mediatypes=true", imageName, f.BuilderID, f.RunnerID))
	}

	buildkitdFlags := ""
	if utils.IsFile(path.Join(homeSigma, buildkitdConfigFilename)) {
		buildkitdFlags += fmt.Sprintf("--config=%s", path.Join(homeSigma, buildkitdConfigFilename))
	}
	cmd.Env = append(os.Environ(), fmt.Sprintf("BUILDKITD_FLAGS=%s", buildkitdFlags))
	cmd.Env = append(cmd.Env, fmt.Sprintf("DOCKER_CONFIG=%s", homeSigma))

	slog.Info("building image", "command", cmd.String(), "env", cmd.Env)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("build image failed: %v", err)
	}
	slog.Info("finished build image")
	return nil
}

// sign 使用 cosign 对构建产物签名
func (f *BuildFlow) sign(imageName string) error {
	s, err := signing.NewSigning(signing.Options{
		Type:      enums.SigningTypeCosign,
		Http:      strings.HasPrefix(f.Endpoint, "http://"),
		MultiArch: len(f.BuildkitPlatforms) > 1,
	})
	if err != nil {
		return err
	}
	return s.Sign(context.Background(), f.Authorization, f.SigningPrivateKey, imageName)
}
