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

package database

import (
	"context"
	"errors"
	"fmt"
	"math/rand" // nolint: gosec
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/dal/dao"
	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/modules/locker/definition"
)

type lockerDatabase struct {
	lockerServiceFactory dao.LockerServiceFactory
}

func New(config configs.Configuration) (definition.Locker, error) {
	return &lockerDatabase{
		lockerServiceFactory: dao.NewLockerServiceFactory(),
	}, nil
}

type lock struct {
	key, value           string
	expire               time.Duration
	lockerServiceFactory dao.LockerServiceFactory
}

// Lock ...
func (l lockerDatabase) Acquire(ctx context.Context, key string, expire, waitTimeout time.Duration) (definition.Lock, error) {
	if expire < 100*time.Millisecond {
		return nil, definition.ErrLockTooShort
	}
	ddlCtx, cancel := context.WithTimeout(ctx, waitTimeout)
	defer cancel()
	ticker := time.NewTicker(time.Duration(500) * time.Millisecond)
	defer func() {
		ticker.Stop()
	}()
	for {
		select {
		case <-ddlCtx.Done():
			return nil, ddlCtx.Err()
		case <-ticker.C:
		}
		val := fmt.Sprintf("%d-%d", rand.Int(), time.Now().Nanosecond()) // nolint: gosec
		err := query.Q.Transaction(func(tx *query.Query) error {
			lockerService := l.lockerServiceFactory.New(tx)
			return lockerService.Create(ctx, key, val, time.Now().Add(expire).UnixMilli())
		})
		if err != nil {
			log.Error().Err(err).Msg("Create locker failed, wait for retry")
			continue
		}
		return &lock{
			key:                  key,
			value:                val,
			expire:               expire,
			lockerServiceFactory: l.lockerServiceFactory,
		}, nil
	}
}

// AcquireWithRenew acquire lock with renew the lock
func (l lockerDatabase) AcquireWithRenew(ctx context.Context, key string, expire, waitTimeout time.Duration) error {
	lock, err := l.Acquire(ctx, key, expire, waitTimeout)
	if err != nil {
		return err
	}

	go func() {
		ticker := time.NewTicker(time.Duration(500) * time.Millisecond)
		defer func() {
			ticker.Stop()
		}()
		for {
			select {
			case <-ctx.Done():
				err := lock.Unlock(context.Background()) // should always release the locker
				if err != nil {
					log.Error().Err(err).Msg("release lock failed")
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

// Renew ...
func (l lock) Renew(ctx context.Context, ttls ...time.Duration) error {
	var expire time.Duration
	if len(ttls) == 0 {
		expire = l.expire
	} else {
		expire = ttls[0]
	}
	if expire < definition.MinLockExpire {
		return definition.ErrLockTooShort
	}
	err := query.Q.Transaction(func(tx *query.Query) error {
		lockerService := l.lockerServiceFactory.New(tx)
		return lockerService.Renew(ctx, l.key, l.value, time.Now().Add(expire).UnixMilli())
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) ||
			errors.Is(err, definition.ErrLockAlreadyExpired) {
			log.Error().Err(err).Msg("Locker already expired")
			return definition.ErrLockAlreadyExpired
		}
		if errors.Is(err, definition.ErrLockNotHeld) {
			log.Error().Err(err).Msg("Locker not held")
			return definition.ErrLockNotHeld
		}
		log.Error().Err(err).Msg("Renew locker failed")
		return fmt.Errorf("Renew locker failed")
	}
	return nil
}

// Unlock ...
func (l *lock) Unlock(ctx context.Context) error {
	err := query.Q.Transaction(func(tx *query.Query) error {
		lockerService := l.lockerServiceFactory.New(tx)
		return lockerService.Delete(ctx, l.key, l.value)
	})
	return err
}
