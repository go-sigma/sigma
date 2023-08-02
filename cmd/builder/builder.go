package main

import (
	"flag"
	"fmt"
	"math/big"
	"net/url"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/caarlos0/env/v9"
	"github.com/docker/cli/cli/config/configfile"
	"github.com/docker/cli/cli/config/types"
	"github.com/dustin/go-humanize"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/mholt/archiver/v3"
	buildkitdconfig "github.com/moby/buildkit/cmd/buildkitd/config"
	resolverconfig "github.com/moby/buildkit/util/resolver/config"
	"github.com/rs/zerolog/log"

	"github.com/go-sigma/sigma/pkg/logger"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils"
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
	dockerConfig            = "docker_config.json"
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
	checkErr(builder.initCache())
	checkErr(builder.initToken())
	checkErr(builder.gitClone())
	checkErr(builder.build())
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
	ScmCredentialType enums.ScmCredentialType `env:"SCM_CREDENTIAL_TYPE,notEmpty"`
	ScmSshKey         string                  `env:"SCM_SSH_KEY"`
	ScmToken          string                  `env:"SCM_TOKEN"`
	ScmUsername       string                  `env:"SCM_USERNAME"`
	ScmPassword       string                  `env:"SCM_PASSWORD"`
	ScmProvider       enums.ScmProvider       `env:"SCM_PROVIDER,notEmpty"`
	ScmRepository     string                  `env:"SCM_REPOSITORY,notEmpty"`
	ScmBranch         string                  `env:"SCM_BRANCH" envDefault:"main"`
	ScmDepth          int                     `env:"SCM_DEPTH" envDefault:"0"`
	ScmSubModule      bool                    `env:"SCM_SUBMODULE" envDefault:"false"`

	OciRegistryDomain   string `env:"OCI_REGISTRY_DOMAIN,notEmpty"`
	OciRegistryUsername string `env:"OCI_REGISTRY_USERNAME"`
	OciRegistryPassword string `env:"OCI_REGISTRY_PASSWORD"`
	OciName             string `env:"OCI_NAME,notEmpty"`

	BuildkitInsecureRegistries []string            `env:"BUILDKIT_INSECURE_REGISTRIES" envSeparator:","`
	BuildkitCacheDir           string              `env:"BUILDKIT_CACHE_DIR" envDefault:"/tmp/buildkit"`
	BuildkitContext            string              `env:"BUILDKIT_CONTEXT" envDefault:"."`
	BuildkitDockerfile         string              `env:"BUILDKIT_DOCKERFILE" envDefault:"Dockerfile"`
	BuildkitPlatforms          []enums.OciPlatform `env:"BUILDKIT_PLATFORMS" envSeparator:","`
}

func (b Builder) initCache() error {
	if utils.IsFile(path.Join(cache, compressedCache)) {
		log.Info().Msg("Start to decompress cache")
		err := archiver.Unarchive(path.Join(cache, compressedCache), home)
		if err != nil {
			return fmt.Errorf("Decompress cache failed: %v", err)
		}
		fileInfo, err := os.Stat(path.Join(cache, compressedCache))
		if err != nil {
			return fmt.Errorf("Read compressed file failed: %v", err)
		}
		err = os.Rename(cacheOut, cacheIn)
		if err != nil {
			return fmt.Errorf("Rename cache_out to cache_in failed: %v", err)
		}
		log.Info().Str("size", humanize.BigBytes(big.NewInt(fileInfo.Size()))).Msg("Decompress cache success")
	}
	var dirs = []string{cacheOut, cacheIn}
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

// initToken init git clone token and buildkit push token
func (b Builder) initToken() error {
	if b.ScmCredentialType == enums.ScmCredentialTypeSsh {
		keyScan, err := exec.LookPath("ssh-keyscan")
		if err != nil {
			return fmt.Errorf("ssh-keyscan binary not found in path: %v", err)
		}
		endpoint, err := transport.NewEndpoint(b.ScmRepository)
		if err != nil {
			return fmt.Errorf("transport.NewEndpoint failed: %v", err)
		}
		cmd := exec.Command(keyScan)
		if endpoint.Port != 0 {
			cmd.Args = append(cmd.Args, "-p", strconv.Itoa(endpoint.Port))
		}
		cmd.Args = append(cmd.Args, endpoint.Host)
		log.Info().Str("command", cmd.String()).Msg("Running ssh-keyscan")

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
		if !utils.IsFile(path.Join(homeSigma, privateKey)) {
			_, err = privateKeyObj.WriteString(b.ScmSshKey)
			if err != nil {
				return fmt.Errorf("Write private key failed: %v", err)
			}
		}
	}
	if b.OciRegistryPassword != "" && b.OciRegistryUsername != "" {
		dockerConfigObj, err := os.Create(path.Join(homeSigma, dockerConfig))
		if err != nil {
			return fmt.Errorf("Create file failed: %v", dockerConfigObj)
		}
		defer func() {
			_ = dockerConfigObj.Close() // nolint: errcheck
		}()
		cf := configfile.ConfigFile{
			AuthConfigs: map[string]types.AuthConfig{
				b.OciRegistryDomain: {
					Username: b.OciRegistryUsername,
					Password: b.OciRegistryPassword,
				},
			},
		}
		if !utils.IsFile(path.Join(homeSigma, dockerConfig)) {
			err = cf.SaveToWriter(dockerConfigObj)
			if err != nil {
				return fmt.Errorf("Save docker config failed: %v", err)
			}
		}
	}
	var btConfig buildkitdconfig.Config
	if len(b.BuildkitInsecureRegistries) > 0 {
		btConfig.Registries = make(map[string]resolverconfig.RegistryConfig, len(b.BuildkitInsecureRegistries))
		for _, registry := range b.BuildkitInsecureRegistries {
			if strings.HasSuffix(registry, "@http") {
				btConfig.Registries[strings.TrimSuffix(registry, "@http")] = resolverconfig.RegistryConfig{PlainHTTP: ptr.Of(true)}
			} else {
				btConfig.Registries[strings.TrimSuffix(registry, "@http")] = resolverconfig.RegistryConfig{Insecure: ptr.Of(true)}
			}
		}
	}
	btConfig.Workers.OCI.Enabled = ptr.Of(true)
	btConfig.Workers.OCI.Snapshotter = "auto"
	btConfig.Workers.OCI.NoProcessSandbox = true
	btConfig.Workers.OCI.GC = ptr.Of(true)
	btConfig.Workers.OCI.GCKeepStorage.Bytes = 10 << 30 // 10GB
	btConfig.Workers.OCI.MaxParallelism = 4
	btConfig.Workers.OCI.CNIPoolSize = 16
	btConfig.Workers.OCI.Rootless = true
	if !utils.IsFile(path.Join(homeSigma, buildkitdConfigFilename)) {
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
	cmd := exec.Command(git, "clone", "--branch", b.ScmBranch)
	if b.ScmDepth != 0 {
		cmd.Args = append(cmd.Args, "--depth", strconv.Itoa(b.ScmDepth))
	}
	if b.ScmSubModule {
		cmd.Args = append(cmd.Args, "--recurse-submodules")
	}
	if b.ScmCredentialType == enums.ScmCredentialTypeSsh {
		cmd.Args = append(cmd.Args, "-i", path.Join(homeSigma, privateKey))
		cmd.Env = append(os.Environ(), fmt.Sprintf("SSH_KNOWN_HOSTS=%s", path.Join(homeSigma, knownHosts)))
	}
	repository := b.ScmRepository
	if b.ScmCredentialType == enums.ScmCredentialTypeToken {
		u, err := url.Parse(repository)
		if err != nil {
			return fmt.Errorf("SCM_REPOSITORY parse with url failed: %v", err)
		}
		repository = fmt.Sprintf("%s//%s@%s/%s", u.Scheme, b.ScmToken, u.Host, u.Path)
	}
	if b.ScmCredentialType == enums.ScmCredentialTypeUsername {
		endpoint, err := transport.NewEndpoint(repository)
		if err != nil {
			return fmt.Errorf("transport.NewEndpoint failed: %v", err)
		}
		endpoint.User = b.ScmUsername
		endpoint.Password = b.ScmPassword
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

func (b Builder) build() error {
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
	cmd.Args = append(cmd.Args, "--frontend", "gateway.v0", "--opt", "source=docker/dockerfile")                         // TODO: set frontend
	cmd.Args = append(cmd.Args, "--output", fmt.Sprintf("type=image,name=%s,push=false", b.OciName))                     // TODO: set output push true
	cmd.Args = append(cmd.Args, "--export-cache", fmt.Sprintf("type=local,mode=max,compression=gzip,dest=%s", cacheOut)) // TODO: set cache volume
	cmd.Args = append(cmd.Args, "--import-cache", fmt.Sprintf("type=local,src=%s", cacheIn))                             // TODO: set cache volume

	buildkitdFlags := ""
	if utils.IsFile(path.Join(homeSigma, buildkitdConfigFilename)) {
		buildkitdFlags += fmt.Sprintf("--config=%s", path.Join(homeSigma, buildkitdConfigFilename))
	}
	cmd.Env = append(os.Environ(), fmt.Sprintf("BUILDKITD_FLAGS=%s", buildkitdFlags))

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

func (b Builder) exportCache() error {
	log.Info().Msg("Start to compress cache")
	tgz := archiver.NewTarGz()
	err := tgz.Archive([]string{path.Join(cacheOut)}, path.Join("/tmp", compressedCache))
	if err != nil {
		return fmt.Errorf("Compress cache failed: %v", err)
	}
	err = os.Rename(path.Join("/tmp", compressedCache), path.Join(cache, compressedCache))
	if err != nil {
		return fmt.Errorf("Move compressed file to dir failed")
	}
	fileInfo, err := os.Stat(path.Join(cache, compressedCache))
	if err != nil {
		return fmt.Errorf("Read compressed file failed: %v", err)
	}
	log.Info().Str("size", humanize.BigBytes(big.NewInt(fileInfo.Size()))).Msg("Export cache success")
	return nil
}

// docker run -it --rm --security-opt apparmor=unconfined -e SCM_CREDENTIAL_TYPE=none -e SCM_PROVIDER=github -e OCI_REGISTRY_DOMAIN=docker.com -e SCM_REPOSITORY=https://github.com/tosone/sudoku.git -e SCM_BRANCH=dev -e OCI_NAME=test:dev -e BUILDKIT_INSECURE_REGISTRIES="10.1.0.1:3000@http,docker.io@http,test.com" --entrypoint '' docker.io/library/builder:dev sh
// docker run -it --rm --security-opt apparmor=unconfined -e SCM_CREDENTIAL_TYPE=none -e SCM_PROVIDER=github -e OCI_REGISTRY_DOMAIN=docker.com -e SCM_REPOSITORY=https://github.com/tosone/sudoku.git -e SCM_BRANCH=master -e OCI_NAME=test:dev -e BUILDKIT_INSECURE_REGISTRIES="10.1.0.1:3000@http,docker.io@http,test.com" --entrypoint '' docker.io/library/builder:dev sh

// BUILDKITD_FLAGS="--config=/opt/sigma/buildkitd.toml" /usr/bin/buildctl-daemonless.sh build --local context=/code --local dockerfile=/code --progress plain --frontend gateway.v0 --opt source=docker/dockerfile:1.6 --output type=image,name=test:dev,push=false --export-cache type=local,mode=max,compression=gzip,dest=/opt/cache_out --import-cache type=local,src=/opt/cache_in
