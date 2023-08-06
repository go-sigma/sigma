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

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/rs/zerolog/log"

	builderlogger "github.com/go-sigma/sigma/pkg/builder/logger"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/types/enums"
)

func (i *instance) informer(ctx context.Context) {
	go func(ctx context.Context) {
		eventsOpt := types.EventsOptions{}
		events, errs := i.client.Events(ctx, eventsOpt)
		for {
			select {
			case <-ctx.Done():
				return
			case event := <-events:
				switch event.Type { // nolint: gocritic
				case "container":
					switch event.Action {
					case "start":
						container, err := i.client.ContainerInspect(ctx, event.Actor.ID)
						if err != nil {
							log.Error().Err(err).Str("id", event.Actor.ID).Msg("Inspect container failed")
							continue
						}
						if container.Config != nil && container.Config.Labels != nil {
							if container.Config.Labels["oci-image-builder"] != consts.AppName ||
								container.Config.Labels["builder-id"] == "" ||
								container.Config.Labels["runner-id"] == "" {
								log.Debug().Msgf("Container not controlled by %s", consts.AppName)
								continue
							}
						}
						if container.ContainerJSONBase != nil && container.ContainerJSONBase.State != nil &&
							(container.ContainerJSONBase.State.Running ||
								container.ContainerJSONBase.State.Status == "running" || // TODO: we should test all case
								container.ContainerJSONBase.State.Status == "exited") {
							log.Info().Str("id", event.Actor.ID).Str("name", container.ContainerJSONBase.Name).Msg("Builder container started")
							builderID, runnerID, err := i.getBuilderTaskID(container.ContainerJSONBase.Name)
							if err != nil {
								log.Error().Err(err).Str("container", container.ContainerJSONBase.Name).Msg("Parse builder task id failed")
								continue
							}
							go func(id string) {
								err := i.logStore(ctx, event.Actor.ID, builderID, runnerID)
								if err != nil {
									log.Error().Err(err).Str("id", id).Msg("Get container log failed")
								}
							}(event.Actor.ID)
						}
					case "die":
						container, err := i.client.ContainerInspect(ctx, event.Actor.ID)
						if err != nil {
							log.Error().Err(err).Str("id", event.Actor.ID).Msg("Inspect container failed")
							continue
						}
						if container.Config != nil && container.Config.Labels != nil && container.ContainerJSONBase != nil {
							if container.Config.Labels["oci-image-builder"] != consts.AppName ||
								container.Config.Labels["builder-id"] == "" ||
								container.Config.Labels["runner-id"] == "" {
								log.Debug().Msgf("Container not controlled by %s", consts.AppName)
								continue
							}
						}
						i.controlled.Remove(event.Actor.ID)

						builderID, runnerID, err := i.getBuilderTaskID(container.ContainerJSONBase.Name)
						if err != nil {
							log.Error().Err(err).Str("container", container.ContainerJSONBase.Name).Msg("Parse builder task id failed")
							continue
						}

						builderService := i.builderServiceFactory.New()
						updates := make(map[string]any, 1)
						if container.ContainerJSONBase != nil && container.ContainerJSONBase.State != nil {
							if container.ContainerJSONBase.State.ExitCode == 0 {
								updates = map[string]any{query.BuilderRunner.Status.ColumnName().String(): enums.BuildStatusSuccess}
								log.Info().Str("id", event.Actor.ID).Str("name", container.ContainerJSONBase.Name).Msg("Builder container succeed")
							} else {
								updates = map[string]any{query.BuilderRunner.Status.ColumnName().String(): enums.BuildStatusFailed}
								log.Error().Int("ExitCode", container.ContainerJSONBase.State.ExitCode).
									Str("Error", container.ContainerJSONBase.State.Error).
									Bool("OOMKilled", container.ContainerJSONBase.State.OOMKilled).
									Msg("Builder container exited")
							}
						}
						err = builderService.UpdateRunner(ctx, builderID, runnerID, updates)
						if err != nil {
							log.Error().Err(err).Msg("Update runner failed")
						}
					}
				}
			case err := <-errs:
				log.Error().Err(err).Msg("Docker event error")
			}
		}
	}(ctx)
}

func (i *instance) logStore(ctx context.Context, containerID string, builderID, runnerID int64) error {
	ok := i.controlled.Add(containerID)
	if !ok {
		log.Error().Str("container", containerID).Int64("builder", builderID).Int64("runner", runnerID).Msg("Add container id to controlled array failed")
		return fmt.Errorf("Add container id to controlled array failed")
	}
	reader, err := i.client.ContainerLogs(ctx, containerID, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
	})
	if err != nil {
		return fmt.Errorf("Get container logs failed: %v", err)
	}

	writer := builderlogger.Driver.Write(builderID, runnerID)
	_, err = stdcopy.StdCopy(writer, writer, reader)
	if err != nil {
		return fmt.Errorf("Copy container logs failed: %v", err)
	}
	err = writer.Close()
	if err != nil {
		return fmt.Errorf("Close container logs failed: %v", err)
	}

	err = i.client.ContainerRemove(ctx, containerID, types.ContainerRemoveOptions{})
	if err != nil {
		log.Error().Err(err).Str("container", containerID).Int64("builder", builderID).Int64("runner", runnerID).Msg("Remove container failed")
		return fmt.Errorf("Remove container failed: %v", err)
	}

	return err
}
