// Copyright 2026 sigma
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

package build

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/background/build/logger"
)

const (
	builderRunnerStatusColumn        = "status"
	builderRunnerStatusMessageColumn = "status_message"
	builderRunnerStartedAtColumn     = "started_at"
	builderRunnerEndedAtColumn       = "ended_at"
)

// markBuildingAttempts and markBuildingRetryInterval bound the wait for the runner row of a build
// that the api process has just created.
var (
	markBuildingAttempts      = 5
	markBuildingRetryInterval = 200 * time.Millisecond
)

// RunnerRepository persists build runner state transitions.
type RunnerRepository interface {
	// UpdateRunner persists runner field updates for a build run.
	UpdateRunner(ctx context.Context, builderID, runnerID string, updates map[string]any) error
}

// Coordinator owns build-run side effects that are shared across runtime drivers.
type Coordinator struct {
	builderRepository RunnerRepository
	logStore          logger.LogStore
}

// NewCoordinator creates a coordinator for runner state transitions and log storage.
func NewCoordinator(builderRepository RunnerRepository, logStore logger.LogStore) Coordinator {
	return Coordinator{
		builderRepository: builderRepository,
		logStore:          logStore,
	}
}

// MarkBuilding marks a runner as actively building and records its start time. The runner row is
// written by the api process, so the update can arrive before that insert is committed; a short
// retry covers that window instead of failing the whole build with "record not found".
func (c Coordinator) MarkBuilding(ctx context.Context, builderID, runnerID string) error {
	startedAt := time.Now().UnixMilli()
	updates := map[string]any{
		builderRunnerStatusColumn:    enums.BuildStatusBuilding,
		builderRunnerStartedAtColumn: startedAt,
	}

	var err error
	for attempt := range markBuildingAttempts {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(markBuildingRetryInterval):
			}
		}
		err = c.builderRepository.UpdateRunner(ctx, builderID, runnerID, updates)
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
	}
	return err
}

// MarkStopped marks a runner as stopped unless the stop operation failed unexpectedly.
func (c Coordinator) MarkStopped(ctx context.Context, builderID, runnerID string, stopErr error, isExpectedStopErr func(error) bool) error {
	status := enums.BuildStatusStopped
	// The message describes this transition, so an older reason cannot linger on the runner.
	message := ""
	if stopErr != nil && (isExpectedStopErr == nil || !isExpectedStopErr(stopErr)) {
		status = enums.BuildStatusFailed
		message = stopErr.Error()
	}
	return c.builderRepository.UpdateRunner(ctx, builderID, runnerID, map[string]any{
		builderRunnerStatusColumn:        status,
		builderRunnerEndedAtColumn:       time.Now().UnixMilli(),
		builderRunnerStatusMessageColumn: message,
	})
}

// MarkCompleted marks a runner as successful when the runtime exits with code zero. A failing
// exit is marked as failed and stores message as the reason shown in the runner list. The message
// always follows the exit, so a successful run clears the reason of an earlier attempt.
func (c Coordinator) MarkCompleted(ctx context.Context, builderID, runnerID string, exitCode int, message string) error {
	updates := map[string]any{
		builderRunnerStatusColumn:        enums.BuildStatusSuccess,
		builderRunnerEndedAtColumn:       time.Now().UnixMilli(),
		builderRunnerStatusMessageColumn: message,
	}
	if exitCode != 0 {
		updates[builderRunnerStatusColumn] = enums.BuildStatusFailed
	}
	return c.builderRepository.UpdateRunner(ctx, builderID, runnerID, updates)
}

// MarkFailed marks a runner as failed without changing its end time.
func (c Coordinator) MarkFailed(ctx context.Context, builderID, runnerID string) error {
	return c.builderRepository.UpdateRunner(ctx, builderID, runnerID, map[string]any{
		builderRunnerStatusColumn: enums.BuildStatusFailed,
	})
}

// StoreLogs copies runtime log output into the configured build log store.
func (c Coordinator) StoreLogs(builderID, runnerID string, copyLogs func(stdout, stderr io.Writer) error) error {
	if c.logStore == nil {
		return fmt.Errorf("builder log store is not initialized")
	}
	writer := c.logStore.Write(builderID, runnerID)
	err := copyLogs(writer, writer)
	if err != nil {
		return err
	}
	err = writer.Close()
	if err != nil {
		return fmt.Errorf("close container logs failed: %v", err)
	}
	return nil
}
