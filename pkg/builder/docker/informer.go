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
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/rs/zerolog/log"

	"github.com/go-sigma/sigma/pkg/builder"
	"github.com/go-sigma/sigma/pkg/builder/logger"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/types/enums"
)

func (i *instance) informer(ctx context.Context) {
	go func(ctx context.Context) {
		eventsOpt := types.EventsOptions{
			Filters: filters.NewArgs(filters.KeyValuePair{Key: "label", Value: fmt.Sprintf("oci-image-builder=%s", consts.AppName)}),
		}
		events, errs := i.client.Events(ctx, eventsOpt)
		for {
			select {
			case <-ctx.Done():
				return
			case event := <-events:
				switch event.Type { // nolint: gocritic
				case "container":
					log.Debug().Str("type", string(event.Type)).Str("action", string(event.Action)).Msg("Got a new docker event")
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
							builderID, runnerID, err := builder.ParseContainerID(container.ContainerJSONBase.Name)
							if err != nil {
								log.Error().Err(err).Str("container", container.ContainerJSONBase.Name).Msg("Parse builder task id failed")
								continue
							}
							go func(id string) {
								err := i.logStore(ctx, id, builderID, runnerID)
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

						builderID, runnerID, err := builder.ParseContainerID(container.ContainerJSONBase.Name)
						if err != nil {
							log.Error().Err(err).Str("container", container.ContainerJSONBase.Name).Msg("Parse builder task id failed")
							continue
						}

						if !i.controlled.Contains(event.Actor.ID) {
							err := i.logStore(ctx, event.Actor.ID, builderID, runnerID)
							if err != nil {
								log.Error().Err(err).Str("id", event.Actor.ID).Msg("Get container log failed")
							}
						}

						i.controlled.Remove(event.Actor.ID)

						builderService := i.builderServiceFactory.New()
						updates := make(map[string]any, 1)
						if container.ContainerJSONBase != nil && container.ContainerJSONBase.State != nil {
							if container.ContainerJSONBase.State.ExitCode == 0 {
								updates = map[string]any{
									query.BuilderRunner.Status.ColumnName().String():  enums.BuildStatusSuccess,
									query.BuilderRunner.EndedAt.ColumnName().String(): time.Now().UnixMilli(),
								}
								log.Info().Str("id", event.Actor.ID).Str("name", container.ContainerJSONBase.Name).Msg("Builder container succeed")
							} else {
								updates = map[string]any{
									query.BuilderRunner.Status.ColumnName().String():  enums.BuildStatusFailed,
									query.BuilderRunner.EndedAt.ColumnName().String(): time.Now().UnixMilli(),
								}
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
					case "destroy":
						i.controlled.Remove(event.Actor.ID)
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
	reader, err := i.client.ContainerLogs(ctx, containerID, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
	})
	if err != nil {
		return fmt.Errorf("Get container logs failed: %v", err)
	}

	writer := logger.Driver.Write(builderID, runnerID)
	_, err = stdcopy.StdCopy(writer, writer, reader)
	if err != nil {
		return fmt.Errorf("Copy container logs failed: %v", err)
	}
	err = writer.Close()
	if err != nil {
		return fmt.Errorf("Close container logs failed: %v", err)
	}

	err = i.client.ContainerRemove(ctx, containerID, container.RemoveOptions{})
	if err != nil {
		log.Error().Err(err).Str("container", containerID).Int64("builder", builderID).Int64("runner", runnerID).Msg("Remove container failed")
		return fmt.Errorf("Remove container failed: %v", err)
	}

	return nil
}

func (i *instance) cacheList(ctx context.Context) error {
	containers, err := i.client.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: filters.NewArgs(filters.KeyValuePair{Key: "label", Value: fmt.Sprintf("oci-image-builder=%s", consts.AppName)}),
	})
	if err != nil {
		log.Error().Err(err).Msg("List containers failed")
		return err
	}
	for _, container := range containers {
		var name string
		if len(container.Names) > 0 {
			name = strings.TrimPrefix(container.Names[0], "/")
		} else {
			continue
		}
		builderID, runnerID, err := builder.ParseContainerID(name)
		if err != nil {
			log.Error().Err(err).Msg("Parse builder task id failed")
			continue
		}
		con, err := i.client.ContainerInspect(ctx, container.ID)
		if err != nil {
			log.Error().Err(err).Str("id", container.ID).Msg("Inspect container failed")
			continue
		}
		err = i.logStore(ctx, container.ID, builderID, runnerID)
		if err != nil {
			log.Error().Err(err).Str("id", container.ID).Msg("Get container log failed")
			continue
		}
		updates := map[string]any{query.BuilderRunner.Status.ColumnName().String(): enums.BuildStatusFailed}
		if con.ContainerJSONBase != nil && con.ContainerJSONBase.State != nil {
			if con.ContainerJSONBase.State.ExitCode == 0 {
				updates = map[string]any{
					query.BuilderRunner.Status.ColumnName().String():  enums.BuildStatusSuccess,
					query.BuilderRunner.EndedAt.ColumnName().String(): time.Now().UnixMilli(),
				}
				log.Info().Str("id", container.ID).Str("name", con.ContainerJSONBase.Name).Msg("Builder container succeed")
			} else {
				updates = map[string]any{
					query.BuilderRunner.Status.ColumnName().String():  enums.BuildStatusFailed,
					query.BuilderRunner.EndedAt.ColumnName().String(): time.Now().UnixMilli(),
				}
				log.Error().Int("ExitCode", con.ContainerJSONBase.State.ExitCode).
					Str("Error", con.ContainerJSONBase.State.Error).
					Bool("OOMKilled", con.ContainerJSONBase.State.OOMKilled).
					Msg("Builder container exited")
			}
		}
		builderService := i.builderServiceFactory.New()
		err = builderService.UpdateRunner(ctx, builderID, runnerID, updates)
		if err != nil {
			log.Error().Err(err).Msg("Update runner failed")
		}
	}
	return nil
}
