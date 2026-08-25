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
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/pkg/stdcopy"

	"github.com/go-sigma/sigma/pkg/background/build/runtime"
	"github.com/go-sigma/sigma/pkg/consts"
)

func (i *instance) informer(ctx context.Context) {
	go func(ctx context.Context) {
		eventsOpt := events.ListOptions{
			Filters: filters.NewArgs(filters.KeyValuePair{Key: "label", Value: fmt.Sprintf("oci-image-builder=%s", consts.AppName)}),
		}
		evts, errs := i.client.Events(ctx, eventsOpt)
		for {
			select {
			case <-ctx.Done():
				return
			case evt := <-evts:
				switch evt.Type { // nolint: gocritic
				case events.ContainerEventType:
					slog.Debug("got a new docker event", "type", string(evt.Type), "action", string(evt.Action))
					switch evt.Action {
					case events.ActionStart:
						container, err := i.client.ContainerInspect(ctx, evt.Actor.ID)
						if err != nil {
							slog.Error("inspect container failed", "err", err, "id", evt.Actor.ID)
							continue
						}
						if container.Config != nil && container.Config.Labels != nil {
							if container.Config.Labels["oci-image-builder"] != consts.AppName ||
								container.Config.Labels["builder-id"] == "" ||
								container.Config.Labels["runner-id"] == "" {
								slog.Debug(fmt.Sprintf("Container not controlled by %s", consts.AppName))
								continue
							}
						}
						if container.ContainerJSONBase != nil && container.State != nil &&
							(container.State.Running ||
								container.State.Status == "running" ||
								// TODO: we should test all case
								container.State.Status == "exited") {
							slog.Info("builder container started", "id", evt.Actor.ID, "name", container.Name)
							builderID, runnerID, err := runtime.ParseContainerID(container.Name)
							if err != nil {
								slog.Error("parse builder task id failed", "err", err, "container", container.Name)
								continue
							}
							go func(id string) {
								err := i.logStore(ctx, id, builderID, runnerID)
								if err != nil {
									slog.Error("get container log failed", "err", err, "id", id)
								}
							}(evt.Actor.ID)
						}
					case events.ActionDie:
						container, err := i.client.ContainerInspect(ctx, evt.Actor.ID)
						if err != nil {
							slog.Error("inspect container failed", "err", err, "id", evt.Actor.ID)
							continue
						}
						if container.ContainerJSONBase == nil {
							slog.Debug("container JSON base is nil, skipping", "id", evt.Actor.ID)
							continue
						}
						if container.Config != nil && container.Config.Labels != nil {
							if container.Config.Labels["oci-image-builder"] != consts.AppName ||
								container.Config.Labels["builder-id"] == "" ||
								container.Config.Labels["runner-id"] == "" {
								slog.Debug(fmt.Sprintf("Container not controlled by %s", consts.AppName))
								continue
							}
						}

						builderID, runnerID, err := runtime.ParseContainerID(container.Name)
						if err != nil {
							slog.Error("parse builder task id failed", "err", err, "container", container.Name)
							continue
						}

						if !i.controlled.Contains(evt.Actor.ID) {
							err := i.logStore(ctx, evt.Actor.ID, builderID, runnerID)
							if err != nil {
								slog.Error("get container log failed", "err", err, "id", evt.Actor.ID)
							}
						}

						i.controlled.Remove(evt.Actor.ID)

						if container.ContainerJSONBase != nil && container.State != nil {
							if container.State.ExitCode == 0 {
								slog.Info("builder container succeed", "id", evt.Actor.ID, "name", container.Name)
							} else {
								slog.Error("builder container exited",
									"ExitCode", container.State.ExitCode,
									"Error", container.State.Error,
									"OOMKilled", container.State.OOMKilled)
							}
							err = i.coordinator.MarkCompleted(ctx, builderID, runnerID, container.State.ExitCode)
						} else {
							err = i.coordinator.MarkFailed(ctx, builderID, runnerID)
						}
						if err != nil {
							slog.Error("update runner failed", "err", err)
						}
					case events.ActionDestroy:
						i.controlled.Remove(evt.Actor.ID)
					}
				}
			case err := <-errs:
				slog.Error("docker event error", "err", err)
			}
		}
	}(ctx)
}

func (i *instance) logStore(ctx context.Context, containerID, builderID, runnerID string) error {
	ok := i.controlled.Add(containerID)
	if !ok {
		slog.Error("add container id to controlled array failed", "container", containerID, "builder", builderID, "runner", runnerID)
		return fmt.Errorf("add container id to controlled array failed")
	}
	err := i.coordinator.StoreLogs(builderID, runnerID, func(stdout, stderr io.Writer) error {
		reader, err := i.client.ContainerLogs(ctx, containerID, container.LogsOptions{
			ShowStdout: true,
			ShowStderr: true,
			Follow:     true,
		})
		if err != nil {
			return fmt.Errorf("get container logs failed: %v", err)
		}
		_, err = stdcopy.StdCopy(stdout, stderr, reader)
		if err != nil {
			return fmt.Errorf("copy container logs failed: %v", err)
		}
		return nil
	})
	if err != nil {
		return err
	}

	err = i.client.ContainerRemove(ctx, containerID, container.RemoveOptions{})
	if err != nil {
		slog.Error("remove container failed", "err", err, "container", containerID, "builder", builderID, "runner", runnerID)
		return fmt.Errorf("remove container failed: %v", err)
	}

	return nil
}

func (i *instance) cacheList(ctx context.Context) error {
	containers, err := i.client.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: filters.NewArgs(filters.KeyValuePair{Key: "label", Value: fmt.Sprintf("oci-image-builder=%s", consts.AppName)}),
	})
	if err != nil {
		slog.Error("list containers failed", "err", err)
		return err
	}
	for _, ctr := range containers {
		var name string
		if len(ctr.Names) > 0 {
			name = strings.TrimPrefix(ctr.Names[0], "/")
		} else {
			continue
		}
		builderID, runnerID, err := runtime.ParseContainerID(name)
		if err != nil {
			slog.Error("parse builder task id failed", "err", err)
			continue
		}
		con, err := i.client.ContainerInspect(ctx, ctr.ID)
		if err != nil {
			slog.Error("inspect container failed", "err", err, "id", ctr.ID)
			continue
		}
		go func(id, bID, rID string, con container.InspectResponse) {
			err := i.logStore(ctx, id, bID, rID)
			if err != nil {
				slog.Error("get container log failed", "err", err, "id", id)
			}
			if con.ContainerJSONBase != nil && con.State != nil {
				if con.State.ExitCode == 0 {
					slog.Info("builder container succeed", "id", id, "name", con.Name)
				} else {
					slog.Error("builder container exited",
						"ExitCode", con.State.ExitCode,
						"Error", con.State.Error,
						"OOMKilled", con.State.OOMKilled)
				}
				err = i.coordinator.MarkCompleted(ctx, bID, rID, con.State.ExitCode)
			} else {
				err = i.coordinator.MarkFailed(ctx, bID, rID)
			}
			if err != nil {
				slog.Error("update runner failed", "err", err)
			}
		}(ctr.ID, builderID, runnerID, con)
	}
	return nil
}
