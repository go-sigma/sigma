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

package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/caarlos0/env/v9"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/rs/zerolog/log"

	"github.com/go-sigma/sigma/pkg/logger"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/compress"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

const (
	home                    = "/opt"
	homeSsh                 = ".ssh"
	homeSigma               = "/opt/sigma"
	cache                   = "/opt/cache"
	cacheIn                 = "/opt/cache_in"
	cacheOut                = "/opt/cache_out"
	knownHosts              = "known_hosts"
	privateKey              = "private_key"
	dockerConfig            = "config.json"
	buildkitdConfigFilename = "buildkitd.toml"
	workspace               = "/code"
	compressedCache         = "cache.tgz"
)

func main() {
	var level string
	flag.StringVar(&level, "level", "info", "log level, available: debug, info, error")
	flag.Parse()

	logLevel, err := enums.ParseLogLevel(level)
	if err != nil {
		panic("level is invalid, available: debug, info, error")
	}
	logger.SetLevel(logLevel.String())

	checkErr(initialize())

	var builder Builder
	checkErr(env.Parse(&builder))
	checkErr(builder.checker())
	builder.api = NewAPI(builder.Authorization, builder.Endpoint)
	checkErr(builder.initCache())
	checkErr(builder.initToken())
	if builder.Builder.Source == enums.BuilderSourceDockerfile {
		checkErr(builder.writeDockerfile())
	} else {
		checkErr(builder.gitClone())
	}
	imageName, err := builder.genTag()
	checkErr(err)
	checkErr(builder.build(imageName))
	checkErr(builder.exportCache())
}

func checkErr(msg any) {
	if msg != nil {
		log.Fatal().Msgf("Something error occurred: %v", msg)
	}
}

func initialize() error {
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

// Builder config for builder
type Builder struct {
	types.Builder

	api api
}

func (b Builder) writeDockerfile() error {
	base64Bytes, err := base64.StdEncoding.DecodeString(ptr.To(b.Dockerfile))
	if err != nil {
		return err
	}
	dockerfileStr, err := compress.Decompress(base64Bytes)
	if err != nil {
		return err
	}
	file, err := os.Create(path.Join(workspace, "Dockerfile"))
	if err != nil {
		return err
	}
	_, err = file.WriteString(dockerfileStr)
	if err != nil {
		return err
	}
	return nil
}

func (b Builder) gitClone() error {
	if utils.IsDir(path.Join(workspace, ".git")) {
		return nil
	}
	log.Info().Msg("Start to clone repository")
	git, err := exec.LookPath("git")
	if err != nil {
		return fmt.Errorf("git not found: %v", err)
	}
	cmd := exec.Command(git, "clone", "--branch", ptr.To(b.ScmBranch))
	if ptr.To(b.ScmDepth) != 0 {
		cmd.Args = append(cmd.Args, "--depth", strconv.Itoa(ptr.To(b.ScmDepth)))
	}
	if ptr.To(b.ScmSubmodule) {
		cmd.Args = append(cmd.Args, "--recurse-submodules")
	}
	if ptr.To(b.ScmCredentialType) == enums.ScmCredentialTypeSsh {
		cmd.Args = append(cmd.Args, "-i", path.Join(homeSigma, privateKey))
		cmd.Env = append(os.Environ(), fmt.Sprintf("SSH_KNOWN_HOSTS=%s", path.Join(homeSigma, knownHosts)))
	}
	repository := ptr.To(b.ScmRepository)
	if ptr.To(b.ScmCredentialType) == enums.ScmCredentialTypeToken {
		u, err := url.Parse(repository)
		if err != nil {
			return fmt.Errorf("SCM_REPOSITORY parse with url failed: %v", err)
		}
		repository = fmt.Sprintf("%s//%s@%s/%s", u.Scheme, ptr.To(b.ScmToken), u.Host, u.Path)
	}
	if ptr.To(b.ScmCredentialType) == enums.ScmCredentialTypeUsername {
		endpoint, err := transport.NewEndpoint(repository)
		if err != nil {
			return fmt.Errorf("transport.NewEndpoint failed: %v", err)
		}
		endpoint.User = ptr.To(b.ScmUsername)
		endpoint.Password = ptr.To(b.ScmPassword)
		repository = endpoint.String()
	}
	cmd.Args = append(cmd.Args, repository, workspace)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Dir = workspace
	log.Info().Str("command", cmd.String()).Str("dir", workspace).Strs("env", cmd.Env).Msg("Running git clone")
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("Clone repository failed: %v", err)
	}
	log.Info().Msg("Finished clone repository")
	return nil
}

func (b Builder) genTag() (string, error) {
	var buildTagOption = BuildTagOption{}
	if b.Source != enums.BuilderSourceDockerfile {
		r, err := git.PlainOpen(workspace)
		if err != nil {
			return "", err
		}
		tagRefs, err := r.Tags()
		if err != nil {
			return "", err
		}
		var latestTag = struct {
			ref  *plumbing.Reference
			when time.Time
		}{}
		err = tagRefs.ForEach(func(tagRef *plumbing.Reference) error {
			commitObj, err := r.CommitObject(tagRef.Hash())
			if err != nil {
				return err
			}
			if latestTag.ref == nil || commitObj.Committer.When.After(latestTag.when) {
				latestTag.ref = tagRef
				latestTag.when = commitObj.Committer.When
			}
			return nil
		})
		if err != nil {
			return "", err
		}

		ref, err := r.Head()
		if err != nil {
			return "", err
		}

		branchName := ref.Name().Short()

		buildTagOption = BuildTagOption{
			ScmBranch: branchName,
			ScmRef:    ref.Hash().String(),
		}
		if latestTag.ref != nil {
			buildTagOption.ScmTag = strings.TrimPrefix(latestTag.ref.String(), "refs/tags/")
		}
	}

	tagBytes, err := base64.StdEncoding.DecodeString(b.Tag)
	if err != nil {
		return "", err
	}
	tag, err := BuildTag(string(tagBytes), buildTagOption)
	if err != nil {
		return "", err
	}
	repositoryBytes, err := base64.StdEncoding.DecodeString(b.Repository)
	if err != nil {
		return "", err
	}
	domain := strings.TrimSuffix(strings.TrimPrefix(strings.TrimPrefix(b.Endpoint, "https://"), "http://"), "/")
	return fmt.Sprintf("%s/%s:%s", domain, string(repositoryBytes), tag), nil
}

func (b Builder) build(imageName string) error {
	log.Info().Msg("Start to build image")
	buildCtl, err := exec.LookPath("buildctl-daemonless.sh")
	if err != nil {
		return fmt.Errorf("Cannot find the buildctl-daemonless.sh: %v", err)
	}
	cmd := exec.Command(buildCtl, "build")
	cmd.Args = append(cmd.Args, "--local", fmt.Sprintf("context=%s", path.Join(workspace, b.BuildkitContext)))
	cmd.Args = append(cmd.Args, "--local", fmt.Sprintf("dockerfile=%s", path.Join(workspace, b.BuildkitContext)))
	cmd.Args = append(cmd.Args, "--progress", "plain")
	if len(b.BuildkitPlatforms) > 0 {
		var platforms = make([]string, 0, len(b.BuildkitPlatforms))
		for _, platform := range b.BuildkitPlatforms {
			platforms = append(platforms, platform.String())
		}
		cmd.Args = append(cmd.Args, "--opt", fmt.Sprintf("platform=%s", strings.Join(platforms, ",")))
	}
	cmd.Args = append(cmd.Args, "--frontend", "gateway.v0", "--opt", "source=docker/dockerfile") // TODO: set frontend
	cmd.Args = append(cmd.Args, "--export-cache", fmt.Sprintf("type=local,mode=max,compression=gzip,dest=%s", cacheOut))
	cmd.Args = append(cmd.Args, "--import-cache", fmt.Sprintf("type=local,src=%s", cacheIn))

	if len(b.BuildkitPlatforms) > 1 {
		cmd.Args = append(cmd.Args, "--output",
			fmt.Sprintf("type=image,name=%s,annotation-index.org.opencontainers.sigma.builder_id=%d,annotation-index.org.opencontainers.sigma.runner_id=%d,push=true,oci-mediatypes=true", imageName, b.BuilderID, b.RunnerID))
	} else {
		cmd.Args = append(cmd.Args, "--output",
			fmt.Sprintf("type=image,name=%s,annotation.org.opencontainers.sigma.builder_id=%d,annotation.org.opencontainers.sigma.runner_id=%d,push=true,oci-mediatypes=true", imageName, b.BuilderID, b.RunnerID))
	}

	buildkitdFlags := ""
	if utils.IsFile(path.Join(homeSigma, buildkitdConfigFilename)) {
		buildkitdFlags += fmt.Sprintf("--config=%s", path.Join(homeSigma, buildkitdConfigFilename))
	}
	cmd.Env = append(os.Environ(), fmt.Sprintf("BUILDKITD_FLAGS=%s", buildkitdFlags))
	cmd.Env = append(cmd.Env, fmt.Sprintf("DOCKER_CONFIG=%s", homeSigma))

	log.Info().Str("command", cmd.String()).Strs("env", cmd.Env).Msg("Building image")
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("Build image failed: %v", err)
	}
	log.Info().Msg("Finished build image")
	return nil
}

// docker run -it --rm --security-opt apparmor=unconfined -e SCM_CREDENTIAL_TYPE=none -e SCM_PROVIDER=github -e OCI_REGISTRY_DOMAIN=docker.com -e SCM_REPOSITORY=https://github.com/tosone/sudoku.git -e SCM_BRANCH=dev -e OCI_NAME=test:dev -e BUILDKIT_INSECURE_REGISTRIES="10.1.0.1:3000@http,docker.io@http,test.com" --entrypoint '' docker.io/library/builder:dev sh
// docker run -it --rm --security-opt apparmor=unconfined -e SCM_CREDENTIAL_TYPE=none -e SCM_PROVIDER=github -e OCI_REGISTRY_DOMAIN=docker.com -e SCM_REPOSITORY=https://github.com/tosone/sudoku.git -e SCM_BRANCH=master -e OCI_NAME=test:dev -e BUILDKIT_INSECURE_REGISTRIES="10.1.0.1:3000@http,docker.io@http,test.com" --entrypoint '' docker.io/library/builder:dev sh

// BUILDKITD_FLAGS="--config=/opt/sigma/buildkitd.toml" /usr/bin/buildctl-daemonless.sh build --local context=/code --local dockerfile=/code --progress plain --frontend gateway.v0 --opt source=docker/dockerfile:1.6 --output type=image,name=test:dev,push=false --export-cache type=local,mode=max,compression=gzip,dest=/opt/cache_out --import-cache type=local,src=/opt/cache_in

// Add anno to manifest
// https://github.com/moby/buildkit/blob/master/docs/annotations.md
