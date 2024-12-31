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
	"fmt"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/docker/cli/cli/config/configfile"
	dockertypes "github.com/docker/cli/cli/config/types"
	"github.com/go-git/go-git/v5/plumbing/transport"
	buildkitdconfig "github.com/moby/buildkit/cmd/buildkitd/config"
	resolverconfig "github.com/moby/buildkit/util/resolver/config"
	"github.com/rs/zerolog/log"

	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

// initToken init git clone token and buildkit push token
func (b Builder) initToken() error {
	if b.ScmCredentialType != nil && ptr.To(b.ScmCredentialType) == enums.ScmCredentialTypeSsh {
		keyScan, err := exec.LookPath("ssh-keyscan")
		if err != nil {
			return fmt.Errorf("ssh-keyscan binary not found in path: %v", err)
		}
		endpoint, err := transport.NewEndpoint(ptr.To(b.ScmRepository))
		if err != nil {
			return fmt.Errorf("transport.NewEndpoint failed: %v", err)
		}
		cmd := exec.Command(keyScan)
		if endpoint.Port != 0 {
			cmd.Args = append(cmd.Args, "-p", strconv.Itoa(endpoint.Port))
		}
		cmd.Args = append(cmd.Args, endpoint.Host)
		log.Info().Str("command", cmd.String()).Msg("Running ssh-keyscan")

		if utils.IsFile(path.Join(homeSigma, knownHosts)) {
			err = os.Remove(path.Join(homeSigma, knownHosts))
			if err != nil {
				return fmt.Errorf("Remove known hosts file failed")
			}
		}
		knownHostsFileObj, err := os.Create(path.Join(homeSigma, knownHosts))
		if err != nil {
			return fmt.Errorf("create file failed: %v", err)
		}
		defer func() {
			_ = knownHostsFileObj.Close() // nolint: errcheck
		}()
		cmd.Stdout = knownHostsFileObj
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			return fmt.Errorf("ssh-keyscan failed: %v", err)
		}

		privateKeyObj, err := os.Create(path.Join(homeSigma, privateKey))
		if err != nil {
			return fmt.Errorf("create file failed: %v", err)
		}
		defer func() {
			_ = privateKeyObj.Close() // nolint: errcheck
		}()
		_, err = privateKeyObj.WriteString(ptr.To(b.ScmSshKey))
		if err != nil {
			return fmt.Errorf("Write private key failed: %v", err)
		}
	}
	{
		if utils.IsFile(path.Join(homeSigma, dockerConfig)) {
			err := os.Remove(path.Join(homeSigma, dockerConfig))
			if err != nil {
				return fmt.Errorf("Remove docker config file failed")
			}
		}
		dockerConfigObj, err := os.Create(path.Join(homeSigma, dockerConfig))
		if err != nil {
			return fmt.Errorf("Create file failed: %v", dockerConfigObj)
		}
		defer func() {
			_ = dockerConfigObj.Close() // nolint: errcheck
		}()
		cf := configfile.ConfigFile{}
		cf.AuthConfigs = make(map[string]dockertypes.AuthConfig)
		for index, domain := range b.OciRegistryDomain {
			if len(b.OciRegistryUsername[index]) != 0 || len(b.OciRegistryPassword[index]) != 0 {
				authConfig := dockertypes.AuthConfig{}
				if len(b.Authorization) > 0 {
					authConfig.RegistryToken = b.Authorization
				}
				if len(b.OciRegistryUsername[index]) > 0 {
					authConfig.Username = b.OciRegistryUsername[index]
				}
				if len(b.OciRegistryPassword[index]) > 0 {
					authConfig.Password = b.OciRegistryPassword[index]
				}
				cf.AuthConfigs[domain] = authConfig
			}
		}
		cf.AuthConfigs[utils.TrimHTTP(b.Endpoint)] = dockertypes.AuthConfig{
			RegistryToken: b.Authorization,
		}
		err = cf.SaveToWriter(dockerConfigObj)
		if err != nil {
			return fmt.Errorf("Save docker config failed: %v", err)
		}
	}
	var btConfig buildkitdconfig.Config
	if len(b.BuildkitInsecureRegistries) > 0 {
		btConfig.Registries = make(map[string]resolverconfig.RegistryConfig, len(b.BuildkitInsecureRegistries))
		for _, registry := range b.BuildkitInsecureRegistries {
			if registry == "" {
				continue
			}
			if strings.HasSuffix(registry, "@http") {
				btConfig.Registries[strings.TrimSuffix(registry, "@http")] = resolverconfig.RegistryConfig{PlainHTTP: ptr.Of(true)}
			} else {
				btConfig.Registries[registry] = resolverconfig.RegistryConfig{Insecure: ptr.Of(true)}
			}
		}
	}
	btConfig.Workers.OCI.Enabled = ptr.Of(true)
	btConfig.Workers.OCI.Snapshotter = "auto"
	btConfig.Workers.OCI.NoProcessSandbox = true
	btConfig.Workers.OCI.GC = ptr.Of(true)
	btConfig.Workers.OCI.GCReservedSpace.Bytes = 10 << 30 // 10GB
	btConfig.Workers.OCI.MaxParallelism = 4
	btConfig.Workers.OCI.CNIPoolSize = 16
	btConfig.Workers.OCI.Rootless = true
	if utils.IsFile(path.Join(homeSigma, buildkitdConfigFilename)) {
		err := os.Remove(path.Join(homeSigma, buildkitdConfigFilename))
		if err != nil {
			return fmt.Errorf("Remove knownHosts file failed")
		}
	}
	btConfigObj, err := os.Create(path.Join(homeSigma, buildkitdConfigFilename))
	if err != nil {
		return fmt.Errorf("Create buildkitd config failed: %v", err)
	}
	defer func() {
		_ = btConfigObj.Close() // nolint: errcheck
	}()
	err = toml.NewEncoder(btConfigObj).Encode(btConfig)
	if err != nil {
		return fmt.Errorf("Marshal buildkitd config failed: %v", err)
	}
	return nil
}
