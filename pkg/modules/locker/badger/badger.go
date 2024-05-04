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

package badger

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/go-sigma/sigma/pkg/configs"
	rBadger "github.com/go-sigma/sigma/pkg/dal/badger"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/modules/locker/definition"
	"github.com/go-sigma/sigma/pkg/utils"
)

type lockerDatabase struct {
	db *badger.DB
}

func New(config configs.Configuration) (definition.Locker, error) {
	return &lockerDatabase{
		db: rBadger.Client,
	}, nil
}

type lock struct {
	db         *badger.DB
	key, value string
	expire     time.Duration
}

// Lock ...
func (l lockerDatabase) Acquire(ctx context.Context, key string, expire, waitTimeout time.Duration) (definition.Lock, error) {
	if expire < 100*time.Millisecond {
		return nil, definition.ErrLockTooShort
	}
	ddlCtx, cancel := context.WithTimeout(ctx, waitTimeout)
	defer cancel()
	ticker := time.NewTicker(time.Duration(100) * time.Millisecond)
	defer func() {
		ticker.Stop()
	}()
	var err error
	for {
		select {
		case <-ddlCtx.Done():
			if err != nil {
				log.Error().Err(err).Msg("Acquire lock failed")
			}
			return nil, ddlCtx.Err()
		case <-ticker.C:
		}
		value := fmt.Sprintf("%s-%d", uuid.NewString(), time.Now().Nanosecond()) // nolint: gosec
		txn := l.db.NewTransaction(true)
		var res *badger.Item
		res, err = txn.Get([]byte(key))
		if err == badger.ErrKeyNotFound {
			err = txn.Set([]byte(key), utils.MustMarshal(models.Locker{Key: key, Value: value,
				Expire: time.Now().Add(expire).UnixMilli()}))
			if err != nil {
				continue
			}
		} else {
			var val []byte
			val, err = res.ValueCopy(nil)
			if err != nil {
				continue
			}
			var v models.Locker
			err = json.Unmarshal(val, &v)
			if err != nil {
				continue
			}
			if v.Expire > time.Now().UnixMilli() {
				continue
			} else {
				err = txn.Delete([]byte(key))
				if err != nil {
					continue
				}
				err = txn.Set([]byte(key), utils.MustMarshal(models.Locker{Key: key, Value: value,
					Expire: time.Now().Add(expire).UnixMilli()}))
				if err != nil {
					continue
				}
			}
		}
		err = txn.Commit()
		if err != nil {
			continue
		}
		return &lock{
			db:     l.db,
			key:    key,
			value:  value,
			expire: expire,
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
		ticker := time.NewTicker(time.Duration(100) * time.Millisecond)
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

	ddlCtx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	ticker := time.NewTicker(time.Duration(100) * time.Millisecond)
	defer func() {
		ticker.Stop()
	}()

	var err error
	for {
		select {
		case <-ddlCtx.Done():
			if err != nil {
				return err
			}
			return ddlCtx.Err()
		case <-ticker.C:
		}
		txn := l.db.NewTransaction(true)
		var val []byte
		val, err = getByKey(txn, l.key)
		if err != nil {
			continue
		}
		var v models.Locker
		err = json.Unmarshal(val, &v)
		if err != nil {
			continue
		}
		if v.Value != l.value {
			return definition.ErrLockNotHeld
		}
		if v.Expire < time.Now().UnixMilli() {
			return definition.ErrLockAlreadyExpired
		}
		err = txn.Set([]byte(l.key), utils.MustMarshal(models.Locker{Key: l.key, Value: l.value,
			Expire: time.Now().Add(expire).UnixMilli()}))
		if err != nil {
			continue
		}
		break
	}

	return nil
}

// Unlock ...
func (l *lock) Unlock(ctx context.Context) error {
	ddlCtx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	ticker := time.NewTicker(time.Duration(100) * time.Millisecond)
	defer func() {
		ticker.Stop()
	}()

	var err error
	for {
		select {
		case <-ddlCtx.Done():
			if err != nil {
				return err
			}
			return ddlCtx.Err()
		case <-ticker.C:
		}
		txn := l.db.NewTransaction(true)
		var val []byte
		val, err = getByKey(txn, l.key)
		if err != nil {
			continue
		}
		var v models.Locker
		err = json.Unmarshal(val, &v)
		if err != nil {
			continue
		}
		if v.Value != l.value {
			return definition.ErrLockNotHeld
		}
		err = txn.Delete([]byte(l.key))
		if err != nil {
			continue
		}
		err = txn.Commit()
		if err != nil {
			return err
		}
		break
	}

	return nil
}

func getByKey(txn *badger.Txn, key string) ([]byte, error) {
	if txn == nil {
		return nil, fmt.Errorf("txn is nil")
	}
	item, err := txn.Get([]byte(key))
	if err != nil {
		return nil, err
	}
	val, err := item.ValueCopy(nil)
	if err != nil {
		return nil, err
	}
	return val, nil
}
