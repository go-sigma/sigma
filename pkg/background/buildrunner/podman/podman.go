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

package podman

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"path"
	"reflect"
	"sync"

	mapset "github.com/deckarep/golang-set/v2"
	"go.podman.io/podman/v6/pkg/bindings"
	"go.podman.io/podman/v6/pkg/bindings/containers"
	"go.podman.io/podman/v6/pkg/specgen"

	"github.com/go-sigma/sigma/pkg/background/buildrunner"
	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/consts"
	repobuilder "github.com/go-sigma/sigma/pkg/dal/repository/builder"
)

func init() {
	buildrunner.DriverFactories[path.Base(reflect.TypeFor[factory]().PkgPath())] = &factory{}
}

type factory struct{}

var _ buildrunner.Factory = factory{}

type instance struct {
	conn              context.Context
	config            *config.Configuration
	controlled        mapset.Set[string] // the controlled container in docker container
	builderRepository repobuilder.BuilderRepository
}

// New returns a new filesystem storage driver
func (f factory) New(config *config.Configuration) (buildrunner.Builder, error) {
	ctx, err := bindings.NewConnection(context.Background(), "unix:///run/podman/podman.sock")
	if err != nil {
		return nil, fmt.Errorf("create docker client failed: %v", err)
	}

	i := &instance{
		conn:              ctx,
		config:            config,
		controlled:        mapset.NewSet[string](),
		builderRepository: repobuilder.NewBuilderRepository(),
	}
	return i, nil
}

// Start start a container to build oci image and push to registry
func (i instance) Start(ctx context.Context, builderConfig buildrunner.BuilderConfig) error {
	envs, err := buildrunner.BuildEnvMap(builderConfig)
	if err != nil {
		return err
	}
	s := specgen.NewSpecGenerator(i.config.Daemon.Builder.Image, false)
	s.Name = buildrunner.GenContainerID(builderConfig.BuilderID, builderConfig.RunnerID)
	s.Env = envs
	s.Entrypoint = []string{}
	s.Command = []string{"sigma-builder"}
	s.Labels = map[string]string{
		"oci-image-builder": consts.AppName,
		"builder-id":        builderConfig.BuilderID,
		"runner-id":         builderConfig.RunnerID,
	}
	s.ContainerSecurityConfig.SeccompPolicy = "unconfined"   // nolint: staticcheck
	s.ContainerSecurityConfig.ApparmorProfile = "unconfined" // nolint: staticcheck
	if len(builderConfig.ExtraHosts) > 0 {
		s.HostAdd = builderConfig.ExtraHosts
	}
	createResponse, err := containers.CreateWithSpec(i.conn, s, nil)
	if err != nil {
		return fmt.Errorf("create container failed: %v", err)
	}
	err = containers.Start(i.conn, createResponse.ID, nil)
	if err != nil {
		return fmt.Errorf("start container failed: %v", err)
	}
	return nil
}

// Stop stop the container
func (i instance) Stop(ctx context.Context, builderID, runnerID string) error {
	name := buildrunner.GenContainerID(builderID, runnerID)
	signal := "SIGKILL"
	err := containers.Kill(i.conn, name, &containers.KillOptions{Signal: &signal})
	if err != nil {
		return fmt.Errorf("kill container failed: %v", err)
	}
	ignore := true
	rmReports, err := containers.Remove(i.conn, name, &containers.RemoveOptions{Ignore: &ignore})
	if err != nil {
		return fmt.Errorf("remove container failed")
	}
	for _, rmReport := range rmReports {
		if rmReport.Err != nil {
			return fmt.Errorf("remove container with something error: %v", rmReport.Err)
		}
	}
	return nil
}

// Restart wrap stop and start
func (i instance) Restart(ctx context.Context, builderConfig buildrunner.BuilderConfig) error {
	return nil
}

// LogStream get the real time log stream
func (i instance) LogStream(ctx context.Context, builderID, runnerID string, writer io.Writer) error {
	var name = buildrunner.GenContainerID(builderID, runnerID)
	var stdoutChan = make(chan string, 10)
	var stderrChan = make(chan string, 10)

	var err error
	var wg = &sync.WaitGroup{}
	wg.Go(func() {
		follow := true
		stderr := false
		stdout := true
		err = containers.Logs(i.conn, name, &containers.LogOptions{
			Follow: &follow,
			Stderr: &stderr,
			Stdout: &stdout,
		}, stdoutChan, stderrChan)
		if err != nil {
			err = fmt.Errorf("get container(%s) log stream failed: %v", name, err)
		}
	})

	var wgStd = &sync.WaitGroup{}
	wgStd.Add(2)
	go func() {
		defer wgStd.Done()
		for s := range stdoutChan {
			_, err := writer.Write([]byte(s))
			if err != nil {
				slog.Error("write stdout to writer failed", "err", err, "Name", name, "Msg", s)
			}
		}
	}()

	go func() {
		defer wgStd.Done()
		for s := range stderrChan {
			slog.Debug("container stderr output", "Name", name, "Msg", s)
		}
	}()

	wg.Wait()
	close(stdoutChan)
	close(stderrChan)

	if err != nil {
		return err
	}

	wgStd.Wait()

	return nil
}
