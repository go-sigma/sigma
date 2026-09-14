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

package lock

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/config"
	dalredis "github.com/go-sigma/sigma/pkg/dal/redis"
	"github.com/go-sigma/sigma/pkg/infra/registry"
)

//go:generate go tool mockgen -destination=locker_mocks.go -package=lock github.com/go-sigma/sigma/pkg/infra/lock Locker,Lock,Factory

const (
	// MinLockExpire is the shortest lease TTL accepted by Acquire and Renew; a shorter one fails with ErrLockTooShort.
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
	// Unlock releases the held lease, returning ErrLockNotHeld when it is no longer held because it expired or was taken over.
	Unlock(ctx context.Context) error
	// Renew extends the held lease by the given TTL, or by the original TTL when none is supplied, keeping the lock alive; it returns ErrLockTooShort for a TTL below MinLockExpire and ErrLockNotHeld once the lease is lost.
	Renew(ctx context.Context, ttls ...time.Duration) error
}

// Locker locker interface
type Locker interface {
	// Acquire acquire lock with the key, expire and wait timeout
	Acquire(ctx context.Context, key string, expire, waitTimeout time.Duration) (Lock, error)
	// AcquireWithRenew acquire lock with renew the lock
	AcquireWithRenew(ctx context.Context, key string, expire, waitTimeout time.Duration) error
}

// Factory is the interface for the storage driver factory
type Factory interface {
	New(params Params) (Locker, error)
}

// Params declares dependencies needed to construct a lock.
type Params struct {
	dig.In

	Config             *config.Configuration
	RedisClientFactory dalredis.ClientFactory `optional:"true"`
}

var Factories = make(registry.Factories[enums.LockerType, Factory])

// Initialize constructs the Locker for the driver named by Config.Locker.Type,
// returning an error when that driver has not been registered.
func Initialize(params Params) (Locker, error) {
	factory, ok := Factories[params.Config.Locker.Type]
	if !ok {
		return nil, fmt.Errorf("driver %q not registered", params.Config.Locker.Type)
	}
	return factory.New(params)
}
