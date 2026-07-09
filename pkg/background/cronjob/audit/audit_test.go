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
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/infra/lock"
	"github.com/go-sigma/sigma/pkg/infra/timewheel"
)

func TestNormalizeCleanupConfig(t *testing.T) {
	cleanupConfig := normalizeCleanupConfig(config.ConfigurationAuditCleanup{})
	require.Equal(t, defaultSoftDeleteAfterMonths, cleanupConfig.SoftDeleteAfterMonths)
	require.Equal(t, defaultHardDeleteAfterMonths, cleanupConfig.HardDeleteAfterMonths)
	require.Equal(t, defaultCleanupInterval, cleanupConfig.Interval)
	require.Equal(t, defaultCleanupBatchSize, cleanupConfig.BatchSize)
}

func TestCleanupRunnerRunner(t *testing.T) {
	repository := &fakeAuditRepository{
		hardIDs: makeIDs(700),
		softIDs: makeIDs(500),
	}
	tw := &fakeTimeWheel{}
	runner := cleanupRunner{
		config: config.ConfigurationAuditCleanup{
			SoftDeleteAfterMonths: 3,
			HardDeleteAfterMonths: 6,
			Interval:              time.Hour,
			BatchSize:             1000,
		},
		auditRepository: repository,
		locker:          &fakeLocker{},
	}

	runner.runner(t.Context(), tw)

	require.Len(t, repository.hardDeletedIDs, 700)
	require.Len(t, repository.softDeletedIDs, 500)
	require.Equal(t, 1000, repository.hardLimit)
	require.Equal(t, 1000, repository.softLimit)
	require.Zero(t, tw.tickNext)
}

func TestCleanupRunnerRunnerContinuesWhenHardDeleteFillsBatch(t *testing.T) {
	repository := &fakeAuditRepository{
		hardIDs: makeIDs(1000),
		softIDs: makeIDs(500),
	}
	tw := &fakeTimeWheel{}
	runner := cleanupRunner{
		config: config.ConfigurationAuditCleanup{
			SoftDeleteAfterMonths: 3,
			HardDeleteAfterMonths: 6,
			Interval:              time.Hour,
			BatchSize:             1000,
		},
		auditRepository: repository,
		locker:          &fakeLocker{},
	}

	runner.runner(t.Context(), tw)

	require.Len(t, repository.hardDeletedIDs, 1000)
	require.Len(t, repository.softDeletedIDs, 500)
	require.Equal(t, 1000, repository.softLimit)
	require.Zero(t, tw.tickNext)
}

func TestCleanupRunnerRunnerSkipsWhenLockIsHeld(t *testing.T) {
	repository := &fakeAuditRepository{
		hardIDs: makeIDs(1000),
		softIDs: makeIDs(500),
	}
	tw := &fakeTimeWheel{}
	runner := cleanupRunner{
		config: config.ConfigurationAuditCleanup{
			SoftDeleteAfterMonths: 3,
			HardDeleteAfterMonths: 6,
			Interval:              time.Hour,
			BatchSize:             1000,
		},
		auditRepository: repository,
		locker:          &fakeLocker{err: context.DeadlineExceeded},
	}

	runner.runner(t.Context(), tw)

	require.Empty(t, repository.hardDeletedIDs)
	require.Empty(t, repository.softDeletedIDs)
	require.Zero(t, repository.hardLimit)
	require.Zero(t, repository.softLimit)
	require.Zero(t, tw.tickNext)
}

type fakeAuditRepository struct {
	hardIDs []string
	softIDs []string

	hardLimit int
	softLimit int

	hardDeletedIDs []string
	softDeletedIDs []string
}

func (r *fakeAuditRepository) Create(_ context.Context, _ *models.Audit) error {
	return nil
}

func (r *fakeAuditRepository) SoftDeleteBefore(_ context.Context, _ int64, limit int) (int64, error) {
	r.softLimit = limit
	ids := popIDs(&r.softIDs, limit)
	r.softDeletedIDs = append(r.softDeletedIDs, ids...)
	return int64(len(ids)), nil
}

func (r *fakeAuditRepository) HardDeleteBefore(_ context.Context, _ int64, limit int) (int64, error) {
	r.hardLimit = limit
	ids := popIDs(&r.hardIDs, limit)
	r.hardDeletedIDs = append(r.hardDeletedIDs, ids...)
	return int64(len(ids)), nil
}

var _ lock.Locker = (*fakeLocker)(nil)

type fakeLocker struct {
	err error
}

func (l *fakeLocker) Acquire(_ context.Context, _ string, _, _ time.Duration) (lock.Lock, error) {
	return fakeLock{}, l.err
}

func (l *fakeLocker) AcquireWithRenew(_ context.Context, _ string, _, _ time.Duration) error {
	return l.err
}

type fakeLock struct{}

func (fakeLock) Unlock(context.Context) error {
	return nil
}

func (fakeLock) Renew(context.Context, ...time.Duration) error {
	return nil
}

type fakeTimeWheel struct {
	tickNext time.Duration
}

func (f *fakeTimeWheel) TickNext(ddl time.Duration) {
	f.tickNext = ddl
}

func (f *fakeTimeWheel) AddRunner(_ timewheel.Notify) {}

func (f *fakeTimeWheel) Stop() {}

func makeIDs(n int) []string {
	ids := make([]string, 0, n)
	for i := range n {
		ids = append(ids, strconv.Itoa(i))
	}
	return ids
}

func popIDs(ids *[]string, limit int) []string {
	if len(*ids) <= limit {
		popped := *ids
		*ids = nil
		return popped
	}
	popped := (*ids)[:limit]
	*ids = (*ids)[limit:]
	return popped
}
