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
	"fmt"
	"io"
	"time"

	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/background/build/logger"
)

const (
	builderRunnerStatusColumn    = "status"
	builderRunnerStartedAtColumn = "started_at"
	builderRunnerEndedAtColumn   = "ended_at"
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

// MarkBuilding marks a runner as actively building and records its start time.
func (c Coordinator) MarkBuilding(ctx context.Context, builderID, runnerID string) error {
	return c.builderRepository.UpdateRunner(ctx, builderID, runnerID, map[string]any{
		builderRunnerStatusColumn:    enums.BuildStatusBuilding,
		builderRunnerStartedAtColumn: time.Now().UnixMilli(),
	})
}

// MarkStopped marks a runner as stopped unless the stop operation failed unexpectedly.
func (c Coordinator) MarkStopped(ctx context.Context, builderID, runnerID string, stopErr error, isExpectedStopErr func(error) bool) error {
	status := enums.BuildStatusStopped
	if stopErr != nil && (isExpectedStopErr == nil || !isExpectedStopErr(stopErr)) {
		status = enums.BuildStatusFailed
	}
	return c.builderRepository.UpdateRunner(ctx, builderID, runnerID, map[string]any{
		builderRunnerStatusColumn:  status,
		builderRunnerEndedAtColumn: time.Now().UnixMilli(),
	})
}

// MarkCompleted marks a runner as successful when the runtime exits with code zero.
func (c Coordinator) MarkCompleted(ctx context.Context, builderID, runnerID string, exitCode int) error {
	status := enums.BuildStatusSuccess
	if exitCode != 0 {
		status = enums.BuildStatusFailed
	}
	return c.builderRepository.UpdateRunner(ctx, builderID, runnerID, map[string]any{
		builderRunnerStatusColumn:  status,
		builderRunnerEndedAtColumn: time.Now().UnixMilli(),
	})
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
