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
	"path"
	"reflect"
	"strconv"
	"strings"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/rs/zerolog/log"

	"github.com/go-sigma/sigma/pkg/builder"
	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/dao"
	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/types/enums"
)

func init() {
	builder.DriverFactories[path.Base(reflect.TypeOf(factory{}).PkgPath())] = &factory{}
}

type factory struct{}

var _ builder.Factory = factory{}

// New returns a new filesystem storage driver
func (f factory) New(config configs.Configuration) (builder.Builder, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, fmt.Errorf("Create docker client failed: %v", err)
	}
	i := &instance{
		config:                config,
		client:                cli,
		controlled:            mapset.NewSet[string](),
		builderServiceFactory: dao.NewBuilderServiceFactory(),
	}
	err = i.cacheList(context.Background())
	if err != nil {
		return nil, err
	}
	go i.informer(context.Background())
	return i, nil
}

type instance struct {
	config                configs.Configuration
	client                *client.Client
	controlled            mapset.Set[string] // the controlled container in docker container
	builderServiceFactory dao.BuilderServiceFactory
}

var _ builder.Builder = instance{}

// Start start a container to build oci image and push to registry
func (i instance) Start(ctx context.Context, builderConfig builder.BuilderConfig) error {
	envs, err := builder.BuildEnv(builderConfig)
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
			"builder-id":        strconv.FormatInt(builderConfig.BuilderID, 10),
			"runner-id":         strconv.FormatInt(builderConfig.RunnerID, 10),
		},
	}
	hostConfig := &container.HostConfig{
		SecurityOpt: []string{"seccomp=unconfined", "apparmor=unconfined"},
		NetworkMode: container.NetworkMode(i.config.Daemon.Builder.Docker.Network),
	}
	_, err = i.client.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, builder.GenContainerID(builderConfig.BuilderID, builderConfig.RunnerID))
	if err != nil {
		return fmt.Errorf("Create container failed: %v", err)
	}
	err = i.client.ContainerStart(ctx, builder.GenContainerID(builderConfig.BuilderID, builderConfig.RunnerID), container.StartOptions{})
	if err != nil {
		return fmt.Errorf("Start container failed: %v", err)
	}
	builderService := i.builderServiceFactory.New()
	err = builderService.UpdateRunner(ctx, builderConfig.BuilderID, builderConfig.RunnerID, map[string]any{
		query.BuilderRunner.Status.ColumnName().String():    enums.BuildStatusBuilding,
		query.BuilderRunner.StartedAt.ColumnName().String(): time.Now().UnixMilli(),
	})
	if err != nil {
		return fmt.Errorf("Update runner status failed: %v", err)
	}
	return nil
}

const (
	retryMax      = 10
	retryDuration = time.Second
)

// Stop stop the container
func (i instance) Stop(ctx context.Context, builderID, runnerID int64) error {
	var err error
	defer func() {
		status := enums.BuildStatusStopped
		if err != nil {
			if !(strings.Contains(err.Error(), "No such container") || strings.Contains(err.Error(), "is not running")) {
				status = enums.BuildStatusFailed
			}
		}
		builderService := i.builderServiceFactory.New()
		err := builderService.UpdateRunner(ctx, builderID, runnerID, map[string]any{
			query.BuilderRunner.Status.ColumnName().String():  status,
			query.BuilderRunner.EndedAt.ColumnName().String(): time.Now().UnixMilli(),
		})
		if err != nil {
			log.Error().Err(err).Msg("Update runner status failed")
		}
	}()
	err = i.client.ContainerKill(ctx, builder.GenContainerID(builderID, runnerID), "SIGKILL")
	if err != nil {
		if strings.Contains(err.Error(), "No such container") || strings.Contains(err.Error(), "is not running") {
			log.Info().Str("id", builder.GenContainerID(builderID, runnerID)).Msg("Container is not running or container is not exist")
			return nil
		}
		log.Error().Err(err).Str("id", builder.GenContainerID(builderID, runnerID)).Msg("Kill container failed")
		return fmt.Errorf("Kill container failed: %v", err)
	}
	err = i.client.ContainerRemove(ctx, builder.GenContainerID(builderID, runnerID), container.RemoveOptions{})
	if err != nil {
		log.Error().Err(err).Str("id", builder.GenContainerID(builderID, runnerID)).Msg("Remove container failed")
		return fmt.Errorf("Remove container failed: %v", err)
	}
	for j := 0; j < retryMax; j++ {
		_, err = i.client.ContainerInspect(ctx, builder.GenContainerID(builderID, runnerID))
		if err != nil {
			if strings.Contains(err.Error(), fmt.Sprintf("No such container: %s", builder.GenContainerID(builderID, runnerID))) {
				return nil
			}
			return fmt.Errorf("Inspect container with error: %v", err)
		}
		<-time.After(retryDuration)
	}
	return nil
}

// Restart wrap stop and start
func (i instance) Restart(ctx context.Context, builderConfig builder.BuilderConfig) error {
	err := i.Stop(ctx, builderConfig.BuilderID, builderConfig.RunnerID)
	if err != nil {
		return err
	}
	return i.Start(ctx, builderConfig)
}

// LogStream get the real time log stream
func (i instance) LogStream(ctx context.Context, builderID, runnerID int64, writer io.Writer) error {
	reader, err := i.client.ContainerLogs(ctx, builder.GenContainerID(builderID, runnerID),
		container.LogsOptions{
			ShowStdout: true,
			ShowStderr: false,
			Follow:     true,
		})
	if err != nil {
		return fmt.Errorf("Get container logs failed: %v", err)
	}
	_, err = stdcopy.StdCopy(writer, nil, reader)
	return err
}
