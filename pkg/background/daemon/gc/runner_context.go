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

package gc

import (
	"context"
	"fmt"
	"time"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/api/enums"
	repodaemon "github.com/go-sigma/sigma/pkg/dal/repository/daemon"
	"github.com/go-sigma/sigma/pkg/infra/lock"
	"github.com/go-sigma/sigma/pkg/infra/workq"
)

type runnerContext struct {
	ctx              context.Context
	daemon           enums.Daemon
	runnerID         string
	startedAt        time.Time
	daemonRepository repodaemon.DaemonRepository
	producer         workq.Producer
}

func newRunnerContext(ctx context.Context, daemon enums.Daemon, runnerID string, daemonRepository repodaemon.DaemonRepository, producer workq.Producer) *runnerContext {
	return &runnerContext{
		ctx:              ctx,
		daemon:           daemon,
		runnerID:         runnerID,
		daemonRepository: daemonRepository,
		producer:         producer,
	}
}

func (r *runnerContext) start() error {
	r.startedAt = time.Now()
	return r.update(map[string]any{
		"status":     enums.TaskCommonStatusDoing,
		"message":    "",
		"started_at": r.startedAt.UnixMilli(),
	})
}

func (r *runnerContext) report(successCount, failedCount int64) error {
	return r.update(map[string]any{
		"status":        enums.TaskCommonStatusDoing,
		"success_count": successCount,
		"failed_count":  failedCount,
	})
}

func (r *runnerContext) finish(status enums.TaskCommonStatus, message string, successCount, failedCount int64) error {
	endedAt := time.Now()
	updates := map[string]any{
		"status":        status,
		"message":       message,
		"ended_at":      endedAt.UnixMilli(),
		"duration":      endedAt.Sub(r.startedAt).Milliseconds(),
		"success_count": successCount,
		"failed_count":  failedCount,
	}
	if r.startedAt.IsZero() {
		updates["duration"] = int64(0)
	}
	return r.update(updates)
}

func (r *runnerContext) emitWebhook(meta api.WebhookPayload, obj any) error {
	if r.producer == nil {
		return nil
	}
	return triggerWebhook(r.ctx, decoratorWebhook{
		NamespaceID: nil,
		Meta:        meta,
		WebhookObj:  obj,
	}, r.producer)
}

func (r *runnerContext) update(updates map[string]any) error {
	var err error
	switch r.daemon {
	case enums.DaemonGcRepository:
		err = r.daemonRepository.UpdateGcRepositoryRunner(r.ctx, r.runnerID, updates)
	case enums.DaemonGcTag:
		err = r.daemonRepository.UpdateGcTagRunner(r.ctx, r.runnerID, updates)
	case enums.DaemonGcArtifact:
		err = r.daemonRepository.UpdateGcArtifactRunner(r.ctx, r.runnerID, updates)
	case enums.DaemonGcBlob:
		err = r.daemonRepository.UpdateGcBlobRunner(r.ctx, r.runnerID, updates)
	default:
		err = fmt.Errorf("daemon %s not support", r.daemon.String())
	}
	return err
}

func acquireRuleLock(ctx context.Context, lockerClient lock.Locker, configLockExpire, configLockWaitTimeout time.Duration, daemon enums.Daemon, ruleID string) (context.Context, context.CancelFunc, error) {
	lockCtx, cancel := context.WithCancel(ctx)
	if lockerClient == nil {
		return lockCtx, cancel, nil
	}
	key := fmt.Sprintf("gc:%s:rule:%s", daemon.String(), ruleID)
	if err := lockerClient.AcquireWithRenew(lockCtx, key, configLockExpire, configLockWaitTimeout); err != nil {
		cancel()
		return nil, nil, err
	}
	return lockCtx, cancel, nil
}
