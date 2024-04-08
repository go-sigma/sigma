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

package definition

//go:generate mockgen -destination=mocks/lock.go -package=mocks github.com/go-sigma/sigma/pkg/modules/locker/definition Lock
//go:generate mockgen -destination=mocks/locker.go -package=mocks github.com/go-sigma/sigma/pkg/modules/locker/definition Locker

import (
	"context"
	"errors"
	"time"
)

const (
	// MinLockExpire ...
	MinLockExpire = 100 * time.Millisecond
)

var (
	// ErrLockNotHeld is returned when trying to release an inactive lock.
	ErrLockNotHeld = errors.New("locker not held")
	// ErrLockTooShort expire should longer than 100ms
	ErrLockTooShort = errors.New("locker expire is too short")
	// ErrLockAlreadyExpired lock already expired
	ErrLockAlreadyExpired = errors.New("locker already expired")
)

// Lock lock interface
type Lock interface {
	// Unlock ...
	Unlock(ctx context.Context) error
	// Renew ...
	Renew(ctx context.Context, ttls ...time.Duration) error
}

// Locker locker interface
type Locker interface {
	// Acquire ...
	Acquire(ctx context.Context, key string, expire, waitTimeout time.Duration) (Lock, error)
	// AcquireWithRenew acquire lock with renew the lock
	AcquireWithRenew(ctx context.Context, key string, expire, waitTimeout time.Duration) error
}
