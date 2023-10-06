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
	"github.com/docker/docker/api/types"
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
		Image:      "docker.io/library/builder:dev",
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
	}
	_, err = i.client.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, i.genContainerID(builderConfig.BuilderID, builderConfig.RunnerID))
	if err != nil {
		return fmt.Errorf("Create container failed: %v", err)
	}
	err = i.client.ContainerStart(ctx, i.genContainerID(builderConfig.BuilderID, builderConfig.RunnerID), types.ContainerStartOptions{})
	if err != nil {
		return fmt.Errorf("Start container failed: %v", err)
	}
	builderService := i.builderServiceFactory.New()
	err = builderService.UpdateRunner(ctx, builderConfig.BuilderID, builderConfig.RunnerID, map[string]any{
		query.BuilderRunner.Status.ColumnName().String():    enums.BuildStatusBuilding,
		query.BuilderRunner.StartedAt.ColumnName().String(): time.Now(),
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
			query.BuilderRunner.Status.ColumnName().String():    status,
			query.BuilderRunner.StartedAt.ColumnName().String(): time.Now(),
		})
		if err != nil {
			log.Error().Err(err).Msg("Update runner status failed")
		}
	}()
	err = i.client.ContainerKill(ctx, i.genContainerID(builderID, runnerID), "SIGKILL")
	if err != nil {
		if strings.Contains(err.Error(), "No such container") || strings.Contains(err.Error(), "is not running") {
			log.Info().Str("id", i.genContainerID(builderID, runnerID)).Msg("Container is not running or container is not exist")
			return nil
		}
		log.Error().Err(err).Str("id", i.genContainerID(builderID, runnerID)).Msg("Kill container failed")
		return fmt.Errorf("Kill container failed: %v", err)
	}
	err = i.client.ContainerRemove(ctx, i.genContainerID(builderID, runnerID), types.ContainerRemoveOptions{})
	if err != nil {
		log.Error().Err(err).Str("id", i.genContainerID(builderID, runnerID)).Msg("Remove container failed")
		return fmt.Errorf("Remove container failed: %v", err)
	}
	for j := 0; j < retryMax; j++ {
		_, err = i.client.ContainerInspect(ctx, i.genContainerID(builderID, runnerID))
		if err != nil {
			if strings.Contains(err.Error(), fmt.Sprintf("No such container: %s", i.genContainerID(builderID, runnerID))) {
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
	err = i.Start(ctx, builderConfig)
	if err != nil {
		return err
	}
	return nil
}

// LogStream get the real time log stream
func (i instance) LogStream(ctx context.Context, builderID, runnerID int64, writer io.Writer) error {
	reader, err := i.client.ContainerLogs(ctx, i.genContainerID(builderID, runnerID), types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
	})
	if err != nil {
		return fmt.Errorf("Get container logs failed: %v", err)
	}
	_, err = stdcopy.StdCopy(writer, writer, reader)
	return err
}

// genContainerID ...
func (i instance) genContainerID(builderID, runnerID int64) string {
	return fmt.Sprintf("sigma-builder-%d-%d", builderID, runnerID)
}

// getBuilderTaskID ...
func (i instance) getBuilderTaskID(containerName string) (int64, int64, error) {
	containerName = strings.TrimPrefix(containerName, "/")
	ids := strings.TrimPrefix(containerName, "sigma-builder-")
	if len(strings.Split(ids, "-")) != 2 {
		return 0, 0, fmt.Errorf("Parse builder task id(%s) failed", containerName)
	}
	builderIDStr, runnerIDStr := strings.Split(ids, "-")[0], strings.Split(ids, "-")[1]
	builderID, err := strconv.ParseInt(builderIDStr, 10, 0)
	if err != nil {
		return 0, 0, fmt.Errorf("Parse builder task id(%s) failed: %v", containerName, err)
	}
	runnerID, err := strconv.ParseInt(runnerIDStr, 10, 0)
	if err != nil {
		return 0, 0, fmt.Errorf("Parse builder task id(%s) failed: %v", containerName, err)
	}
	return builderID, runnerID, nil
}
