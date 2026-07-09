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

	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/background/buildrunner"
	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/query"
	repobuilder "github.com/go-sigma/sigma/pkg/dal/repository/builder"
)

func init() {
	buildrunner.DriverFactories[path.Base(reflect.TypeFor[factory]().PkgPath())] = &factory{}
}

type factory struct{}

var _ buildrunner.Factory = factory{}

// New returns a new filesystem storage driver
func (f factory) New(config *config.Configuration) (buildrunner.Builder, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, fmt.Errorf("create docker client failed: %v", err)
	}
	i := &instance{
		config:            config,
		client:            cli,
		controlled:        mapset.NewSet[string](),
		builderRepository: repobuilder.NewBuilderRepository(),
	}
	err = i.cacheList(context.Background())
	if err != nil {
		return nil, err
	}

	go i.informer(context.Background())

	return i, nil
}

type instance struct {
	config            *config.Configuration
	client            *client.Client
	controlled        mapset.Set[string] // the controlled container in docker container
	builderRepository repobuilder.BuilderRepository
}

var _ buildrunner.Builder = instance{}

// Start start a container to build oci image and push to registry
func (i instance) Start(ctx context.Context, builderConfig buildrunner.BuilderConfig) error {
	envs, err := buildrunner.BuildEnv(builderConfig)
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
	_, err = i.client.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, buildrunner.GenContainerID(builderConfig.BuilderID, builderConfig.RunnerID))

	if err != nil {
		return fmt.Errorf("create container failed: %v", err)
	}

	err = i.client.ContainerStart(ctx, buildrunner.GenContainerID(builderConfig.BuilderID, builderConfig.RunnerID), container.StartOptions{})
	if err != nil {
		return fmt.Errorf("start container failed: %v", err)
	}
	builderRepository := i.builderRepository
	err = builderRepository.UpdateRunner(ctx, builderConfig.BuilderID, builderConfig.RunnerID, map[string]any{
		query.BuilderRunner.Status.ColumnName().String():    enums.BuildStatusBuilding,
		query.BuilderRunner.StartedAt.ColumnName().String(): time.Now().UnixMilli(),
	})
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
		status := enums.BuildStatusStopped

		if err != nil {
			if !(strings.Contains(err.Error(), "No such container") || strings.Contains(err.Error(), "is not running")) { // nolint: staticcheck
				status = enums.BuildStatusFailed
			}
		}

		builderRepository := i.builderRepository
		err := builderRepository.UpdateRunner(ctx, builderID, runnerID, map[string]any{
			query.BuilderRunner.Status.ColumnName().String():  status,
			query.BuilderRunner.EndedAt.ColumnName().String(): time.Now().UnixMilli(),
		})

		if err != nil {
			slog.Error("update runner status failed", "err", err)
		}
	}()

	err = i.client.ContainerKill(ctx, buildrunner.GenContainerID(builderID, runnerID), "SIGKILL")
	if err != nil {
		if strings.Contains(err.Error(), "No such container") || strings.Contains(err.Error(), "is not running") {
			slog.Info("container is not running or container is not exist", "id", buildrunner.GenContainerID(builderID, runnerID))
			return nil
		}

		slog.Error("kill container failed", "err", err, "id", buildrunner.GenContainerID(builderID, runnerID))

		return fmt.Errorf("kill container failed: %v", err)
	}

	err = i.client.ContainerRemove(ctx, buildrunner.GenContainerID(builderID, runnerID), container.RemoveOptions{})
	if err != nil {
		slog.Error("remove container failed", "err", err, "id", buildrunner.GenContainerID(builderID, runnerID))
		return fmt.Errorf("remove container failed: %v", err)
	}

	for range retryMax {
		_, err = i.client.ContainerInspect(ctx, buildrunner.GenContainerID(builderID, runnerID))
		if err != nil {
			if strings.Contains(err.Error(), fmt.Sprintf("No such container: %s", buildrunner.GenContainerID(builderID, runnerID))) {
				return nil
			}
			return fmt.Errorf("inspect container with error: %v", err)
		}

		<-time.After(retryDuration)
	}

	return fmt.Errorf("container %s still exists after %d retries", buildrunner.GenContainerID(builderID, runnerID), retryMax)
}

// Restart wrap stop and start
func (i instance) Restart(ctx context.Context, builderConfig buildrunner.BuilderConfig) error {
	err := i.Stop(ctx, builderConfig.BuilderID, builderConfig.RunnerID)
	if err != nil {
		return err
	}
	return i.Start(ctx, builderConfig)
}

// LogStream get the real time log stream
func (i instance) LogStream(ctx context.Context, builderID, runnerID string, writer io.Writer) error {
	reader, err := i.client.ContainerLogs(ctx, buildrunner.GenContainerID(builderID, runnerID),
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
