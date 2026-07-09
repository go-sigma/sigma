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

package source

import (
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"

	"github.com/go-git/go-git/v6/plumbing/transport"

	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

const (
	homeSigma  = "/opt/sigma"
	knownHosts = "known_hosts"
	privateKey = "private_key"
	workspace  = "/code"
)

// GitSource 通过 git clone 准备源码
type GitSource struct {
	ScmProvider       *enums.ScmProvider
	ScmCredentialType *enums.ScmCredentialType
	ScmSshKey         *string
	ScmToken          *string
	ScmUsername       *string
	ScmPassword       *string
	ScmRepository     *string
	ScmBranch         *string
	ScmDepth          *int
	ScmSubmodule      *bool
}

// Prepare 执行 git clone
func (s GitSource) Prepare() error {
	if utils.IsDir(path.Join(workspace, ".git")) {
		return nil
	}
	slog.Info("start to clone repository")
	git, err := exec.LookPath("git")
	if err != nil {
		return fmt.Errorf("git not found: %v", err)
	}
	cmd := exec.Command(git, "clone", "--branch", ptr.To(s.ScmBranch))
	if ptr.To(s.ScmDepth) != 0 {
		cmd.Args = append(cmd.Args, "--depth", strconv.Itoa(ptr.To(s.ScmDepth)))
	}
	if ptr.To(s.ScmSubmodule) {
		cmd.Args = append(cmd.Args, "--recurse-submodules")
	}
	if ptr.To(s.ScmCredentialType) == enums.ScmCredentialTypeSsh {
		cmd.Args = append(cmd.Args, "-i", path.Join(homeSigma, privateKey))
		cmd.Env = append(os.Environ(), fmt.Sprintf("SSH_KNOWN_HOSTS=%s", path.Join(homeSigma, knownHosts)))
	}
	repository := ptr.To(s.ScmRepository)
	if ptr.To(s.ScmCredentialType) == enums.ScmCredentialTypeToken {
		u, err := url.Parse(repository)
		if err != nil {
			return fmt.Errorf("SCM_REPOSITORY parse with url failed: %v", err)
		}
		repository = fmt.Sprintf("%s://%s@%s/%s", u.Scheme, ptr.To(s.ScmToken), strings.TrimSuffix(u.Host, "/"), strings.TrimPrefix(u.Path, "/"))
	}
	if ptr.To(s.ScmCredentialType) == enums.ScmCredentialTypeUsername {
		endpoint, err := transport.ParseURL(repository)
		if err != nil {
			return fmt.Errorf("transport.NewEndpoint failed: %v", err)
		}
		endpoint.User = url.UserPassword(ptr.To(s.ScmUsername), ptr.To(s.ScmPassword))
		repository = endpoint.String()
	}
	cmd.Args = append(cmd.Args, repository, workspace)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Dir = workspace
	slog.Info("running git clone", "command", cmd.String(), "dir", workspace, "env", cmd.Env)
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("clone repository failed: %v", err)
	}
	slog.Info("finished clone repository")
	return nil
}
