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

package docker

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"path"
	"reflect"
	"strings"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"

	"github.com/go-sigma/sigma/pkg/background/build"
	"github.com/go-sigma/sigma/pkg/background/build/runtime"
	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/consts"
)

func init() {
	runtime.DriverFactories[path.Base(reflect.TypeFor[factory]().PkgPath())] = &factory{}
}

type factory struct{}

var _ runtime.Factory = factory{}

// New returns a new docker runtime.
func (f factory) New(config *config.Configuration, coordinator build.Coordinator) (runtime.Builder, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, fmt.Errorf("create docker client failed: %v", err)
	}
	i := &instance{
		config:      config,
		client:      cli,
		controlled:  mapset.NewSet[string](),
		coordinator: coordinator,
	}
	err = i.cacheList(context.Background())
	if err != nil {
		return nil, err
	}

	go i.informer(context.Background())

	return i, nil
}

type instance struct {
	config      *config.Configuration
	client      *client.Client
	controlled  mapset.Set[string] // the controlled containers in docker
	coordinator build.Coordinator
}

var _ runtime.Builder = instance{}

// Start start a container to build oci image and push to registry
func (i instance) Start(ctx context.Context, builderConfig runtime.Config) error {
	envs, err := runtime.BuildEnv(builderConfig)
	if err != nil {
		return err
	}

	containerConfig := &container.Config{
		Image:      i.config.Daemon.Builder.Image,
		Entrypoint: []string{},
		Cmd:        []string{"sigma-builder"},
		Env:        envs,
		Labels: map[string]string{
			"oci-image-builder": consts.AppName,
			"builder-id":        builderConfig.BuilderID,
			"runner-id":         builderConfig.RunnerID,
		},
	}
	hostConfig := &container.HostConfig{
		SecurityOpt: []string{"seccomp=unconfined", "apparmor=unconfined"},
		NetworkMode: container.NetworkMode(i.config.Daemon.Builder.Docker.Network),
		ExtraHosts:  builderConfig.ExtraHosts,
	}
	_, err = i.client.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, runtime.GenContainerID(builderConfig.BuilderID, builderConfig.RunnerID))

	if err != nil {
		return fmt.Errorf("create container failed: %v", err)
	}

	err = i.client.ContainerStart(ctx, runtime.GenContainerID(builderConfig.BuilderID, builderConfig.RunnerID), container.StartOptions{})
	if err != nil {
		return fmt.Errorf("start container failed: %v", err)
	}
	err = i.coordinator.MarkBuilding(ctx, builderConfig.BuilderID, builderConfig.RunnerID)
	if err != nil {
		return fmt.Errorf("update runner status failed: %v", err)
	}
	return nil
}

const (
	retryMax      = 10
	retryDuration = time.Second
)

// Stop stop the container
func (i instance) Stop(ctx context.Context, builderID, runnerID string) error {
	var err error
	defer func() {
		if e := i.coordinator.MarkStopped(ctx, builderID, runnerID, err, isExpectedStopError); e != nil {
			slog.Error("update runner status failed", "err", e)
		}
	}()

	err = i.client.ContainerKill(ctx, runtime.GenContainerID(builderID, runnerID), "SIGKILL")
	if err != nil {
		if strings.Contains(err.Error(), "No such container") || strings.Contains(err.Error(), "is not running") {
			slog.Info("container is not running or container is not exist", "id", runtime.GenContainerID(builderID, runnerID))
			return nil
		}

		slog.Error("kill container failed", "err", err, "id", runtime.GenContainerID(builderID, runnerID))

		return fmt.Errorf("kill container failed: %v", err)
	}

	err = i.client.ContainerRemove(ctx, runtime.GenContainerID(builderID, runnerID), container.RemoveOptions{})
	if err != nil {
		slog.Error("remove container failed", "err", err, "id", runtime.GenContainerID(builderID, runnerID))
		return fmt.Errorf("remove container failed: %v", err)
	}

	for range retryMax {
		_, err = i.client.ContainerInspect(ctx, runtime.GenContainerID(builderID, runnerID))
		if err != nil {
			if strings.Contains(err.Error(), fmt.Sprintf("No such container: %s", runtime.GenContainerID(builderID, runnerID))) {
				return nil
			}
			return fmt.Errorf("inspect container with error: %v", err)
		}

		<-time.After(retryDuration)
	}

	return fmt.Errorf("container %s still exists after %d retries", runtime.GenContainerID(builderID, runnerID), retryMax)
}

func isExpectedStopError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "No such container") || strings.Contains(err.Error(), "is not running")
}

// Restart wrap stop and start
func (i instance) Restart(ctx context.Context, builderConfig runtime.Config) error {
	err := i.Stop(ctx, builderConfig.BuilderID, builderConfig.RunnerID)
	if err != nil {
		return err
	}
	return i.Start(ctx, builderConfig)
}

// LogStream get the real time log stream
func (i instance) LogStream(ctx context.Context, builderID, runnerID string, writer io.Writer) error {
	reader, err := i.client.ContainerLogs(ctx, runtime.GenContainerID(builderID, runnerID),
		container.LogsOptions{
			ShowStdout: true,
			ShowStderr: false,
			Follow:     true,
		})
	if err != nil {
		return fmt.Errorf("get container logs failed: %v", err)
	}
	_, err = stdcopy.StdCopy(writer, nil, reader)
	return err
}
