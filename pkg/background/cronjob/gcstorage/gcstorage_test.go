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
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/go-sigma/sigma/pkg/infra/lock"
	"github.com/go-sigma/sigma/pkg/infra/timewheel"
)

func TestCleanupRunnerDeletesInBatches(t *testing.T) {
	repository := &fakeDeletionTaskRepository{batches: []int64{3, 2}}
	runner := cleanupRunner{
		retention:        time.Hour,
		batchSize:        3,
		daemonRepository: repository,
		locker:           &fakeLocker{},
	}

	runner.runner(t.Context(), &fakeTimeWheel{})

	require.Equal(t, 2, repository.calls)
	require.Equal(t, 3, repository.limit)
	now := time.Now().UnixMilli()
	require.Less(t, repository.before, now)
	require.Greater(t, repository.before, now-(2*time.Hour).Milliseconds())
}

func TestCleanupRunnerSkipsWhenLockIsHeld(t *testing.T) {
	repository := &fakeDeletionTaskRepository{batches: []int64{3}}
	runner := cleanupRunner{
		retention:        time.Hour,
		batchSize:        3,
		daemonRepository: repository,
		locker:           &fakeLocker{err: context.DeadlineExceeded},
	}

	runner.runner(t.Context(), &fakeTimeWheel{})

	require.Zero(t, repository.calls)
}

type fakeDeletionTaskRepository struct {
	batches []int64
	calls   int
	before  int64
	limit   int
}

func (r *fakeDeletionTaskRepository) DeleteGcStorageDeletionTasksBefore(_ context.Context, before int64, limit int) (int64, error) {
	r.before = before
	r.limit = limit
	if r.calls >= len(r.batches) {
		return 0, nil
	}
	deleted := r.batches[r.calls]
	r.calls++
	return deleted, nil
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

type fakeTimeWheel struct{}

func (*fakeTimeWheel) TickNext(time.Duration) {}

func (*fakeTimeWheel) AddRunner(timewheel.Notify) {}

func (*fakeTimeWheel) Stop() {}
