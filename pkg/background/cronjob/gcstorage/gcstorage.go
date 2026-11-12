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

package gcstorage

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/background/cronjob"
	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/consts"
	repodaemon "github.com/go-sigma/sigma/pkg/dal/repository/daemon"
	"github.com/go-sigma/sigma/pkg/infra/lock"
	"github.com/go-sigma/sigma/pkg/infra/timewheel"
)

const (
	defaultLockExpire      = 30 * time.Second
	defaultLockWaitTimeout = 100 * time.Millisecond
)

var cleanupTw timewheel.TimeWheel

func init() {
	cronjob.Starter = append(cronjob.Starter, gcStorageCleanupJob)
	cronjob.Stopper = append(cronjob.Stopper, func() {
		if cleanupTw != nil {
			cleanupTw.Stop()
		}
	})
}

type cleanupParams struct {
	dig.In

	Config           *config.Configuration
	DaemonRepository repodaemon.DaemonRepository
	Locker           lock.Locker
}

func gcStorageCleanupJob(digCon *dig.Container) error {
	var params cleanupParams
	if err := digCon.Invoke(func(p cleanupParams) {
		params = p
	}); err != nil {
		return err
	}

	cfg := params.Config.Daemon.GC.Cleanup
	cleanupTw = timewheel.NewTimeWheel(context.Background(), cfg.Interval)
	cleanupTw.AddRunner(cleanupRunner{
		retention:        cfg.Retention,
		batchSize:        cfg.BatchSize,
		daemonRepository: params.DaemonRepository,
		locker:           params.Locker,
	}.runner)
	return nil
}

// deletionTaskRepository is the subset of the daemon repository used by the
// cleanup job, kept narrow so the runner is easy to test.
type deletionTaskRepository interface {
	DeleteGcStorageDeletionTasksBefore(context.Context, int64, int) (int64, error)
}

type cleanupRunner struct {
	retention        time.Duration
	batchSize        int
	daemonRepository deletionTaskRepository
	locker           lock.Locker
}

func (r cleanupRunner) runner(ctx context.Context, _ timewheel.TimeWheel) {
	lockCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	if err := r.locker.AcquireWithRenew(lockCtx, consts.LockerCronjobGcStorage, defaultLockExpire, defaultLockWaitTimeout); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			slog.Info("skip gc storage deletion task cleanup because another runner holds the lock")
			return
		}
		slog.Error("acquire gc storage deletion task cleanup lock failed", "err", err)
		return
	}

	before := time.Now().Add(-r.retention).UnixMilli()
	for {
		deleted, err := r.daemonRepository.DeleteGcStorageDeletionTasksBefore(lockCtx, before, r.batchSize)
		if err != nil {
			slog.Error("gc storage deletion task cleanup failed", "err", err)
			return
		}
		if deleted > 0 {
			slog.Info("gc storage deletion task cleanup batch finished", "deleted", deleted)
		}
		if deleted < int64(r.batchSize) {
			return
		}
	}
}
