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

package inmemory

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/infra/lock"
	"github.com/go-sigma/sigma/pkg/utils"
)

func init() {
	utils.PanicIf(lock.Factories.Register(enums.LockerTypeInmemory, &factory{}))
}

type lockerMemory struct {
	mu    sync.Mutex
	locks map[string]*memoryLock
}

type memoryLock struct {
	key      string
	value    string
	expireAt time.Time
	expire   time.Duration
	parent   *lockerMemory
}

type factory struct{}

var _ lock.Factory = factory{}

// New ...
func (factory) New(_ lock.Params) (lock.Locker, error) {
	return &lockerMemory{
		locks: make(map[string]*memoryLock),
	}, nil
}

func (l *lockerMemory) Acquire(ctx context.Context, key string, expire, waitTimeout time.Duration) (lock.Lock, error) {
	if expire < lock.MinLockExpire {
		return nil, lock.ErrLockTooShort
	}

	ddlCtx, cancel := context.WithTimeout(ctx, waitTimeout)
	defer cancel()

	val := fmt.Sprintf("%s-%d", utils.GenSecureID(16), time.Now().Nanosecond())
	tryAcquire := func() lock.Lock {
		l.mu.Lock()
		defer l.mu.Unlock()

		now := time.Now()
		if heldLock, exists := l.locks[key]; exists {
			if now.Before(heldLock.expireAt) {
				return nil
			}
			delete(l.locks, key)
		}

		newLock := &memoryLock{
			key:      key,
			value:    val,
			expireAt: now.Add(expire),
			expire:   expire,
			parent:   l,
		}
		l.locks[key] = newLock
		return newLock
	}

	if acquiredLock := tryAcquire(); acquiredLock != nil {
		return acquiredLock, nil
	}

	ticker := time.NewTicker(time.Duration(100) * time.Millisecond)
	defer func() {
		ticker.Stop()
	}()

	for {
		select {
		case <-ddlCtx.Done():
			return nil, ddlCtx.Err()
		case <-ticker.C:
			if acquiredLock := tryAcquire(); acquiredLock != nil {
				return acquiredLock, nil
			}
		}
	}
}

// AcquireWithRenew acquire lock with renew the lock
func (l *lockerMemory) AcquireWithRenew(ctx context.Context, key string, expire, waitTimeout time.Duration) error {
	lock, err := l.Acquire(ctx, key, expire, waitTimeout)
	if err != nil {
		return err
	}

	go func() {
		ticker := time.NewTicker(time.Duration(100) * time.Millisecond)
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
				return
			}
		}
	}()
	return nil
}

// Renew renew the lock
func (l *memoryLock) Renew(ctx context.Context, ttls ...time.Duration) error {
	l.parent.mu.Lock()
	defer l.parent.mu.Unlock()

	var expire time.Duration
	if len(ttls) == 0 {
		expire = l.expire
	} else {
		expire = ttls[0]
	}

	if expire < lock.MinLockExpire {
		return lock.ErrLockTooShort
	}

	// Check if we still hold the lock
	if heldLock, exists := l.parent.locks[l.key]; exists {
		if heldLock.value != l.value {
			return lock.ErrLockNotHeld
		}
		if time.Now().After(heldLock.expireAt) {
			return lock.ErrLockAlreadyExpired
		}

		// Update expiration time
		l.expireAt = time.Now().Add(expire)
		l.expire = expire
		l.parent.locks[l.key] = l
		return nil
	}

	return lock.ErrLockNotHeld
}

// Unlock unlock the lock
func (l *memoryLock) Unlock(ctx context.Context) error {
	l.parent.mu.Lock()
	defer l.parent.mu.Unlock()

	// Check if we still hold the lock
	if heldLock, exists := l.parent.locks[l.key]; exists {
		if heldLock.value != l.value {
			return lock.ErrLockNotHeld
		}

		// Remove the lock
		delete(l.parent.locks, l.key)
		return nil
	}

	return lock.ErrLockNotHeld
}
