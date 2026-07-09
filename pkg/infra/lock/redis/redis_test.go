// Copyright 2024 sigma
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

package redis

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/require"
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/config"
	dalredis "github.com/go-sigma/sigma/pkg/dal/redis"
	"github.com/go-sigma/sigma/pkg/infra/lock"
	"github.com/go-sigma/sigma/pkg/logger"
)

func TestNew(t *testing.T) {
	logger.SetLevel("debug")

	tests := []struct {
		name      string
		newParams func(*testing.T) lock.Params
		wantErr   bool
	}{
		{
			name:      "normal",
			newParams: newParams,
			wantErr:   false,
		},
		{
			name: "missing redis client factory",
			newParams: func(_ *testing.T) lock.Params {
				return lock.Params{}
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := factory{}
			_, err := f.New(tt.newParams(t))
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestAcquireSerializesConcurrentAccess(t *testing.T) {
	lockerInst := newLocker(t)
	const (
		key        = "test-redis-lock-concurrent"
		concurrent = 10
	)
	var active int64
	var maxActive int64
	var entered int64
	var wg sync.WaitGroup
	start := make(chan struct{})

	for range concurrent {
		wg.Go(func() {
			<-start
			heldLock, err := lockerInst.Acquire(t.Context(), key, 500*time.Millisecond, 5*time.Second)
			require.NoError(t, err)

			current := atomic.AddInt64(&active, 1)
			for {
				observed := atomic.LoadInt64(&maxActive)
				if current <= observed || atomic.CompareAndSwapInt64(&maxActive, observed, current) {
					break
				}
			}

			time.Sleep(20 * time.Millisecond)
			atomic.AddInt64(&active, -1)
			require.NoError(t, heldLock.Unlock(t.Context()))
			atomic.AddInt64(&entered, 1)
		})
	}
	close(start)
	wg.Wait()

	require.EqualValues(t, concurrent, entered)
	require.EqualValues(t, 1, maxActive)
}

func TestAcquireTimesOutWhenLockIsHeld(t *testing.T) {
	lockerInst := newLocker(t)
	const key = "test-redis-lock-timeout"

	heldLock, err := lockerInst.Acquire(t.Context(), key, time.Second, time.Second)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, heldLock.Unlock(t.Context()))
	}()

	_, err = lockerInst.Acquire(t.Context(), key, time.Second, 150*time.Millisecond)
	require.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestUnlockAllowsAcquire(t *testing.T) {
	lockerInst := newLocker(t)
	const key = "test-redis-lock-unlock"

	heldLock, err := lockerInst.Acquire(t.Context(), key, time.Second, time.Second)
	require.NoError(t, err)
	require.NoError(t, heldLock.Unlock(t.Context()))

	nextLock, err := lockerInst.Acquire(t.Context(), key, time.Second, time.Second)
	require.NoError(t, err)
	require.NoError(t, nextLock.Unlock(t.Context()))
}

func TestAcquireWithRenewKeepsLockPastInitialTTL(t *testing.T) {
	lockerInst := newLocker(t)
	const key = "test-redis-lock-renew"

	ctx, cancel := context.WithCancel(t.Context())
	require.NoError(t, lockerInst.AcquireWithRenew(ctx, key, 300*time.Millisecond, time.Second))

	time.Sleep(700 * time.Millisecond)
	_, err := lockerInst.Acquire(t.Context(), key, 300*time.Millisecond, 150*time.Millisecond)
	require.ErrorIs(t, err, context.DeadlineExceeded)

	cancel()
	requireEventuallyAcquired(t, lockerInst, key)
}

func TestAcquireWithRenewReleasesAfterCancel(t *testing.T) {
	lockerInst := newLocker(t)
	const key = "test-redis-lock-cancel"

	ctx, cancel := context.WithCancel(t.Context())
	require.NoError(t, lockerInst.AcquireWithRenew(ctx, key, 300*time.Millisecond, time.Second))
	cancel()

	requireEventuallyAcquired(t, lockerInst, key)
}

func requireEventuallyAcquired(t *testing.T, lockerInst lock.Locker, key string) {
	t.Helper()
	require.Eventually(t, func() bool {
		heldLock, err := lockerInst.Acquire(t.Context(), key, 300*time.Millisecond, 500*time.Millisecond)
		if errors.Is(err, context.DeadlineExceeded) {
			return false
		}
		require.NoError(t, err)
		require.NoError(t, heldLock.Unlock(t.Context()))
		return true
	}, 2*time.Second, 50*time.Millisecond)
}

func TestRenewRejectsShortTTL(t *testing.T) {
	lockerInst := newLocker(t)
	const key = "test-redis-lock-short-renew"

	heldLock, err := lockerInst.Acquire(t.Context(), key, time.Second, time.Second)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, heldLock.Unlock(t.Context()))
	}()

	err = heldLock.Renew(t.Context(), lock.MinLockExpire-time.Millisecond)
	require.ErrorIs(t, err, lock.ErrLockTooShort)
}

func TestRenewUsesCustomTTL(t *testing.T) {
	lockerInst := newLocker(t)
	const key = "test-redis-lock-custom-renew"

	heldLock, err := lockerInst.Acquire(t.Context(), key, 300*time.Millisecond, time.Second)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, heldLock.Unlock(t.Context()))
	}()

	time.Sleep(150 * time.Millisecond)
	require.NoError(t, heldLock.Renew(t.Context(), time.Second))
	time.Sleep(500 * time.Millisecond)

	_, err = lockerInst.Acquire(t.Context(), key, 300*time.Millisecond, 150*time.Millisecond)
	require.ErrorIs(t, err, context.DeadlineExceeded)
}

func newLocker(t *testing.T) lock.Locker {
	t.Helper()
	lockerInst, err := factory{}.New(newParams(t))
	require.NoError(t, err)
	return lockerInst
}

func newParams(t *testing.T) lock.Params {
	t.Helper()
	digCon := dig.New()
	err := digCon.Provide(func() *config.Configuration {
		return &config.Configuration{
			Locker: config.ConfigurationLocker{
				Type:   enums.LockerTypeRedis,
				Prefix: "sigma-locker",
			},
			Redis: config.ConfigurationRedis{
				URL: "redis://" + miniredis.RunT(t).Addr(),
			},
		}
	})
	require.NoError(t, err)

	err = digCon.Provide(dalredis.NewClientFactory)
	require.NoError(t, err)

	var params lock.Params
	require.NoError(t, digCon.Invoke(func(p lock.Params) { params = p }))
	return params
}
