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
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/docker/cli/cli/config/configfile"
	dockertypes "github.com/docker/cli/cli/config/types"
	"github.com/go-git/go-git/v6/plumbing/transport"
	buildkitdconfig "github.com/moby/buildkit/cmd/buildkitd/config"
	resolverconfig "github.com/moby/buildkit/util/resolver/config"

	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

const (
	home                    = "/opt"
	homeSigma               = "/opt/sigma"
	knownHosts              = "known_hosts"
	privateKey              = "private_key"
	dockerConfig            = "config.json"
	buildkitdConfigFilename = "buildkitd.toml"
)

// initToken 初始化 git clone token 和 buildkit push token
func (f *BuildFlow) initToken() error {
	if f.ScmCredentialType != nil && ptr.To(f.ScmCredentialType) == enums.ScmCredentialTypeSsh {
		keyScan, err := exec.LookPath("ssh-keyscan")
		if err != nil {
			return fmt.Errorf("ssh-keyscan binary not found in path: %v", err)
		}
		endpoint, err := transport.ParseURL(ptr.To(f.ScmRepository))
		if err != nil {
			return fmt.Errorf("transport.NewEndpoint failed: %v", err)
		}
		cmd := exec.Command(keyScan)
		if endpoint.Port() != "" {
			cmd.Args = append(cmd.Args, "-p", endpoint.Port())
		}
		cmd.Args = append(cmd.Args, endpoint.Host)
		slog.Info("running ssh-keyscan", "command", cmd.String())

		if utils.IsFile(path.Join(homeSigma, knownHosts)) {
			err = os.Remove(path.Join(homeSigma, knownHosts))
			if err != nil {
				return fmt.Errorf("remove known hosts file failed: %v", err)
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
		_, err = privateKeyObj.WriteString(ptr.To(f.ScmSshKey))
		if err != nil {
			return fmt.Errorf("write private key failed: %v", err)
		}
	}
	{
		if utils.IsFile(path.Join(homeSigma, dockerConfig)) {
			e := os.Remove(path.Join(homeSigma, dockerConfig))
			if e != nil {
				return fmt.Errorf("remove docker config file failed: %v", e)
			}
		}
		dockerConfigObj, err := os.Create(path.Join(homeSigma, dockerConfig))
		if err != nil {
			return fmt.Errorf("create file failed: %v", err)
		}
		defer func() {
			_ = dockerConfigObj.Close() // nolint: errcheck
		}()
		cf := configfile.ConfigFile{}
		cf.AuthConfigs = make(map[string]dockertypes.AuthConfig)
		for index, domain := range f.OciRegistryDomain {
			if len(f.OciRegistryUsername[index]) != 0 || len(f.OciRegistryPassword[index]) != 0 {
				authConfig := dockertypes.AuthConfig{}
				if len(f.Authorization) > 0 {
					authConfig.RegistryToken = f.Authorization
				}
				if len(f.OciRegistryUsername[index]) > 0 {
					authConfig.Username = f.OciRegistryUsername[index]
				}
				if len(f.OciRegistryPassword[index]) > 0 {
					authConfig.Password = f.OciRegistryPassword[index]
				}
				cf.AuthConfigs[domain] = authConfig
			}
		}
		cf.AuthConfigs[utils.TrimHTTP(f.Endpoint)] = dockertypes.AuthConfig{
			RegistryToken: f.Authorization,
		}
		err = cf.SaveToWriter(dockerConfigObj)
		if err != nil {
			return fmt.Errorf("save docker config failed: %v", err)
		}
	}
	var btConfig buildkitdconfig.Config
	if len(f.BuildkitInsecureRegistries) > 0 {
		btConfig.Registries = make(map[string]resolverconfig.RegistryConfig, len(f.BuildkitInsecureRegistries))
		for _, registry := range f.BuildkitInsecureRegistries {
			if registry == "" {
				continue
			}
			if before, ok := strings.CutSuffix(registry, "@http"); ok {
				btConfig.Registries[before] = resolverconfig.RegistryConfig{PlainHTTP: new(true)}
			} else {
				btConfig.Registries[registry] = resolverconfig.RegistryConfig{Insecure: new(true)}
			}
		}
	}
	btConfig.Workers.OCI.Enabled = new(true)
	btConfig.Workers.OCI.Snapshotter = "auto"
	btConfig.Workers.OCI.NoProcessSandbox = true
	btConfig.Workers.OCI.GC = new(true)
	btConfig.Workers.OCI.GCReservedSpace.Bytes = 10 << 30 // 10GB
	btConfig.Workers.OCI.MaxParallelism = 4
	btConfig.Workers.OCI.CNIPoolSize = 16
	btConfig.Workers.OCI.Rootless = true
	if utils.IsFile(path.Join(homeSigma, buildkitdConfigFilename)) {
		e := os.Remove(path.Join(homeSigma, buildkitdConfigFilename))
		if e != nil {
			return fmt.Errorf("remove buildkitd config file failed: %v", e)
		}
	}
	btConfigObj, err := os.Create(path.Join(homeSigma, buildkitdConfigFilename))
	if err != nil {
		return fmt.Errorf("create buildkitd config failed: %v", err)
	}
	defer func() {
		_ = btConfigObj.Close() // nolint: errcheck
	}()
	err = toml.NewEncoder(btConfigObj).Encode(btConfig)
	if err != nil {
		return fmt.Errorf("marshal buildkitd config failed: %v", err)
	}

	// 注入 ExtraHosts 到 /etc/hosts，确保 buildctl-daemonless.sh 能解析私有仓库
	if len(f.ExtraHosts) > 0 {
		err = f.appendExtraHosts()
		if err != nil {
			return fmt.Errorf("append extra hosts failed: %v", err)
		}
	}

	return nil
}

// appendExtraHosts 将 ExtraHosts 条目追加到 /etc/hosts
func (f *BuildFlow) appendExtraHosts() error {
	hostsFile, err := osOpenFile("/etc/hosts", os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open /etc/hosts failed: %v", err)
	}
	defer func() {
		_ = hostsFile.Close() // nolint: errcheck
	}()
	for _, host := range f.ExtraHosts {
		parts := strings.SplitN(host, ":", 2)
		if len(parts) != 2 {
			slog.Warn("invalid extra_hosts entry, expected hostname:ip", "entry", host)
			continue
		}
		_, err := fmt.Fprintf(hostsFile, "%s\t%s\n", parts[1], parts[0])
		if err != nil {
			return fmt.Errorf("write extra hosts failed: %v", err)
		}
	}
	return nil
}
