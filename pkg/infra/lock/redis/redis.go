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

package redis

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"

	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/infra/lock"
	"github.com/go-sigma/sigma/pkg/utils"
)

const (
	backendRedis     = "redis"
	minRetryDelay    = 50 * time.Millisecond
	maxRetryDelay    = 250 * time.Millisecond
	minRenewInterval = 100 * time.Millisecond
)

func init() {
	utils.PanicIf(lock.Factories.Register(enums.LockerTypeRedis, &factory{}))
}

type lockerRedis struct {
	syncer *redsync.Redsync
}

type redisLock struct {
	mu       sync.Mutex
	syncer   *redsync.Redsync
	key      string
	mutex    *redsync.Mutex
	expire   time.Duration
	acquired time.Time
}

type factory struct{}

var _ lock.Factory = factory{}

// New ...
func (factory) New(params lock.Params) (lock.Locker, error) {
	if params.RedisClientFactory == nil {
		return nil, fmt.Errorf("redis client factory is required")
	}
	redisCli, err := params.RedisClientFactory()
	if err != nil {
		return nil, err
	}
	if redisCli == nil {
		return nil, fmt.Errorf("redis client is required")
	}
	return &lockerRedis{
		syncer: redsync.New(goredis.NewPool(redisCli)),
	}, nil
}

func (l lockerRedis) Acquire(ctx context.Context, key string, expire, waitTimeout time.Duration) (lock.Lock, error) {
	start := time.Now()
	result := lock.MetricResultFailure
	defer func() {
		lock.RecordLockAcquire(backendRedis, result, time.Since(start))
	}()

	if expire < lock.MinLockExpire {
		return nil, lock.ErrLockTooShort
	}
	ddlCtx, cancel := context.WithTimeout(ctx, waitTimeout)
	defer cancel()

	tries := 1
	if waitTimeout > 0 {
		tries = int(waitTimeout/minRetryDelay) + 2
	}
	mutex := newMutex(
		l.syncer,
		key,
		expire,
		redsync.WithTries(tries),
		redsync.WithRetryDelayFunc(retryDelay),
	)
	if err := mutex.LockContext(ddlCtx); err != nil {
		if ctxErr := ddlCtx.Err(); ctxErr != nil {
			return nil, ctxErr
		}
		return nil, err
	}

	result = lock.MetricResultSuccess
	return &redisLock{
		syncer:   l.syncer,
		key:      key,
		mutex:    mutex,
		expire:   expire,
		acquired: time.Now(),
	}, nil
}

// AcquireWithRenew acquire lock with renew the lock
func (l lockerRedis) AcquireWithRenew(ctx context.Context, key string, expire, waitTimeout time.Duration) error {
	lock, err := l.Acquire(ctx, key, expire, waitTimeout)
	if err != nil {
		return err
	}

	tick := max(expire/3, minRenewInterval)
	if tick >= expire {
		tick = expire / 2
	}

	go func() {
		ticker := time.NewTicker(tick)
		defer func() {
			ticker.Stop()
		}()
		for {
			select {
			case <-ctx.Done():
				err := lock.Unlock(context.WithoutCancel(ctx)) // should always release the locker
				if err != nil {
					slog.Error("release lock failed", "err", err)
				}
				return
			case <-ticker.C:
			}
			if err := lock.Renew(ctx, expire); err != nil {
				slog.Error("renew lock failed", "err", err, "key", key)
				return
			}
		}
	}()
	return nil
}

// Renew renew the lock
func (l *redisLock) Renew(ctx context.Context, ttls ...time.Duration) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	result := lock.MetricResultFailure
	defer func() {
		lock.RecordLockRenew(backendRedis, result)
	}()

	expire := l.expire
	if len(ttls) > 0 {
		expire = ttls[0]
	}
	if expire < lock.MinLockExpire {
		return lock.ErrLockTooShort
	}

	mutex := l.mutex
	if expire != l.expire {
		mutex = newMutex(l.syncer, l.key, expire, redsync.WithValue(l.mutex.Value()))
	}
	ok, err := mutex.ExtendContext(ctx)
	if err != nil {
		if isLockNotHeld(err) {
			return lock.ErrLockNotHeld
		}
		return err
	}
	if !ok {
		return lock.ErrLockNotHeld
	}
	l.mutex = mutex
	l.expire = expire
	result = lock.MetricResultSuccess
	return nil
}

// Unlock unlock the lock
func (l *redisLock) Unlock(ctx context.Context) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	result := lock.MetricResultFailure
	defer func() {
		lock.RecordLockHeld(backendRedis, result, time.Since(l.acquired))
	}()

	ok, err := l.mutex.UnlockContext(ctx)
	if err != nil {
		if isLockNotHeld(err) {
			return lock.ErrLockNotHeld
		}
		return err
	}
	if !ok {
		return lock.ErrLockNotHeld
	}
	result = lock.MetricResultSuccess
	return nil
}

func newMutex(syncer *redsync.Redsync, key string, expire time.Duration, options ...redsync.Option) *redsync.Mutex {
	opts := make([]redsync.Option, 0, 2+len(options))
	opts = append(opts, redsync.WithExpiry(expire), redsync.WithTimeoutFactor(0.5))
	opts = append(opts, options...)
	return syncer.NewMutex(key, opts...)
}

func retryDelay(try int) time.Duration {
	if try < 1 {
		try = 1
	}
	delay := minRetryDelay
	for range min(try-1, 5) {
		delay *= 2
	}
	delay = min(delay, maxRetryDelay)
	jitterRange := delay / 2
	if jitterRange <= 0 {
		return delay
	}
	jitter := time.Duration(time.Now().UnixNano() % int64(jitterRange))
	if jitter < 0 {
		jitter = -jitter
	}
	return delay + jitter
}

func isLockNotHeld(err error) bool {
	var errTaken *redsync.ErrTaken
	var errNodeTaken *redsync.ErrNodeTaken
	return errors.Is(err, redsync.ErrLockAlreadyExpired) ||
		errors.Is(err, redsync.ErrExtendFailed) ||
		errors.As(err, &errTaken) ||
		errors.As(err, &errNodeTaken)
}
