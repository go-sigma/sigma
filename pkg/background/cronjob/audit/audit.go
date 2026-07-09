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

package audit

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/background/cronjob"
	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/consts"
	repoaudit "github.com/go-sigma/sigma/pkg/dal/repository/audit"
	"github.com/go-sigma/sigma/pkg/infra/lock"
	"github.com/go-sigma/sigma/pkg/infra/timewheel"
)

const (
	defaultSoftDeleteAfterMonths = 3
	defaultHardDeleteAfterMonths = 6
	defaultCleanupInterval       = 7 * 24 * time.Hour
	defaultCleanupBatchSize      = 1000
	defaultLockExpire            = 30 * time.Second
	defaultLockWaitTimeout       = 100 * time.Millisecond
)

var auditTw timewheel.TimeWheel

func init() {
	cronjob.Starter = append(cronjob.Starter, auditCleanupJob)
	cronjob.Stopper = append(cronjob.Stopper, func() {
		if auditTw != nil {
			auditTw.Stop()
		}
	})
}

type cleanupParams struct {
	dig.In

	Config          *config.Configuration
	AuditRepository repoaudit.AuditRepository
	Locker          lock.Locker
}

func auditCleanupJob(digCon *dig.Container) error {
	var params cleanupParams
	if err := digCon.Invoke(func(p cleanupParams) {
		params = p
	}); err != nil {
		return err
	}

	cleanupConfig := normalizeCleanupConfig(params.Config.Audit.Cleanup)
	auditTw = timewheel.NewTimeWheel(context.Background(), cleanupConfig.Interval)
	auditTw.AddRunner(cleanupRunner{
		config:          cleanupConfig,
		auditRepository: params.AuditRepository,
		locker:          params.Locker,
	}.runner)
	return nil
}

type cleanupRunner struct {
	config          config.ConfigurationAuditCleanup
	auditRepository repoaudit.AuditRepository
	locker          lock.Locker
}

func (r cleanupRunner) runner(ctx context.Context, _ timewheel.TimeWheel) {
	lockCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	if err := r.locker.AcquireWithRenew(lockCtx, consts.LockerCronjobAudit, defaultLockExpire, defaultLockWaitTimeout); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			slog.Info("skip audit cleanup because another runner holds the lock")
			return
		}
		slog.Error("acquire audit cleanup lock failed", "err", err)
		return
	}

	now := time.Now()
	batchSize := r.config.BatchSize
	hardDeleteBefore := now.AddDate(0, -r.config.HardDeleteAfterMonths, 0).UnixMilli()
	softDeleteBefore := now.AddDate(0, -r.config.SoftDeleteAfterMonths, 0).UnixMilli()

	for {
		hardDeleted, err := r.auditRepository.HardDeleteBefore(ctx, hardDeleteBefore, batchSize)
		if err != nil {
			slog.Error("audit cleanup hard delete failed", "err", err)
			return
		}
		if hardDeleted > 0 {
			slog.Info("audit cleanup hard delete batch finished", "deleted", hardDeleted)
		}
		if hardDeleted < int64(batchSize) {
			break
		}
	}

	for {
		softDeleted, err := r.auditRepository.SoftDeleteBefore(ctx, softDeleteBefore, batchSize)
		if err != nil {
			slog.Error("audit cleanup soft delete failed", "err", err)
			return
		}
		if softDeleted > 0 {
			slog.Info("audit cleanup soft delete batch finished", "deleted", softDeleted)
		}
		if softDeleted < int64(batchSize) {
			return
		}
	}
}

func normalizeCleanupConfig(cleanupConfig config.ConfigurationAuditCleanup) config.ConfigurationAuditCleanup {
	if cleanupConfig.SoftDeleteAfterMonths <= 0 {
		cleanupConfig.SoftDeleteAfterMonths = defaultSoftDeleteAfterMonths
	}
	if cleanupConfig.HardDeleteAfterMonths <= 0 {
		cleanupConfig.HardDeleteAfterMonths = defaultHardDeleteAfterMonths
	}
	if cleanupConfig.HardDeleteAfterMonths <= cleanupConfig.SoftDeleteAfterMonths {
		cleanupConfig.HardDeleteAfterMonths = cleanupConfig.SoftDeleteAfterMonths + defaultSoftDeleteAfterMonths
	}
	if cleanupConfig.Interval <= 0 {
		cleanupConfig.Interval = defaultCleanupInterval
	}
	if cleanupConfig.BatchSize <= 0 {
		cleanupConfig.BatchSize = defaultCleanupBatchSize
	}
	return cleanupConfig
}
