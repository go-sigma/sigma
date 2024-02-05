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
	"path"
	"reflect"
	"strconv"
	"sync"

	"github.com/containers/podman/v4/pkg/bindings"
	"github.com/containers/podman/v4/pkg/bindings/containers"
	"github.com/containers/podman/v4/pkg/specgen"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/rs/zerolog/log"

	"github.com/go-sigma/sigma/pkg/builder"
	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/dao"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

func init() {
	builder.DriverFactories[path.Base(reflect.TypeOf(factory{}).PkgPath())] = &factory{}
}

type factory struct{}

var _ builder.Factory = factory{}

type instance struct {
	conn                  context.Context
	config                configs.Configuration
	controlled            mapset.Set[string] // the controlled container in docker container
	builderServiceFactory dao.BuilderServiceFactory
}

// New returns a new filesystem storage driver
func (f factory) New(config configs.Configuration) (builder.Builder, error) {
	ctx, err := bindings.NewConnection(context.Background(), "unix:///run/podman/podman.sock")
	if err != nil {
		return nil, fmt.Errorf("Create docker client failed: %v", err)
	}

	i := &instance{
		conn:                  ctx,
		config:                config,
		controlled:            mapset.NewSet[string](),
		builderServiceFactory: dao.NewBuilderServiceFactory(),
	}
	return i, nil
}

// Start start a container to build oci image and push to registry
func (i instance) Start(ctx context.Context, builderConfig builder.BuilderConfig) error {
	envs, err := builder.BuildEnvMap(builderConfig)
	if err != nil {
		return err
	}
	s := specgen.NewSpecGenerator(i.config.Daemon.Builder.Image, false)
	s.Name = i.genContainerID(builderConfig.BuilderID, builderConfig.RunnerID)
	s.Env = envs
	s.Entrypoint = []string{}
	s.Command = []string{"sigma-builder"}
	s.Labels = map[string]string{
		"oci-image-builder": consts.AppName,
		"builder-id":        strconv.FormatInt(builderConfig.BuilderID, 10),
		"runner-id":         strconv.FormatInt(builderConfig.RunnerID, 10),
	}
	s.ContainerSecurityConfig.SeccompPolicy = "unconfined"
	s.ContainerSecurityConfig.ApparmorProfile = "unconfined"
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
func (i instance) Stop(ctx context.Context, builderID, runnerID int64) error {
	name := i.genContainerID(builderID, runnerID)
	err := containers.Kill(i.conn, name, &containers.KillOptions{Signal: ptr.Of("SIGKILL")})
	if err != nil {
		return fmt.Errorf("kill container failed: %v", err)
	}
	rmReports, err := containers.Remove(i.conn, name, &containers.RemoveOptions{Ignore: ptr.Of(true)})
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
func (i instance) Restart(ctx context.Context, builderConfig builder.BuilderConfig) error {
	return nil
}

// LogStream get the real time log stream
func (i instance) LogStream(ctx context.Context, builderID, runnerID int64, writer io.Writer) error {
	var name = i.genContainerID(builderID, runnerID)
	var stdoutChan = make(chan string, 10)
	var stderrChan = make(chan string, 10)

	var err error
	var wg = &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		err = containers.Logs(i.conn, name, &containers.LogOptions{}, stdoutChan, stderrChan)
		if err != nil {
			err = fmt.Errorf("Get container(%s) log stream failed: %v", name, err)
		}
	}()

	var wgStd = &sync.WaitGroup{}
	wgStd.Add(2)
	go func() {
		defer wgStd.Done()
		for s := range stdoutChan {
			_, err := writer.Write([]byte(s))
			if err != nil {
				log.Error().Err(err).Str("Name", name).Str("Msg", s).Msg("Write stdout to writer failed")
			}
		}
	}()

	go func() {
		defer wgStd.Done()
		for s := range stderrChan {
			log.Debug().Str("Name", name).Str("Msg", s).Msg("container stderr output")
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

// genContainerID ...
func (i instance) genContainerID(builderID, runnerID int64) string {
	return fmt.Sprintf("sigma-builder-%d-%d", builderID, runnerID)
}
