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
	"errors"
	"fmt"
	"time"

	"github.com/dgraph-io/badger/v4"

	"github.com/go-sigma/sigma/pkg/configs"
	rBadger "github.com/go-sigma/sigma/pkg/dal/badger"
	"github.com/go-sigma/sigma/pkg/modules/cacher/definition"
)

type cacher[T any] struct {
	db      *badger.DB
	prefix  string
	fetcher definition.Fetcher[T]
	config  configs.Configuration
}

// New returns a new Cacher.
func New[T any](config configs.Configuration, prefix string, fetcher definition.Fetcher[T]) (definition.Cacher[T], error) {
	return &cacher[T]{
		db:      rBadger.Client,
		prefix:  prefix,
		fetcher: fetcher,
		config:  config,
	}, nil
}

// Set sets the value of given key if it is new to the cache.
// Param val should not be nil.
func (c *cacher[T]) Set(ctx context.Context, key string, val T, ttls ...time.Duration) error {
	content, err := json.Marshal(val)
	if err != nil {
		return fmt.Errorf("marshal value failed: %w", err)
	}
	var ttl = c.config.Cache.Badger.Ttl
	if len(ttls) > 0 {
		ttl = ttls[0]
	}
	return c.db.Update(func(txn *badger.Txn) error {
		e := badger.NewEntry([]byte(key), content).WithTTL(ttl)
		return txn.SetEntry(e)
	})
}

// Get tries to fetch a value corresponding to the given key from the cache.
// If error occurs during the first time fetching, it will be cached until the
// sequential fetching triggered by the refresh goroutine succeed.
func (c *cacher[T]) Get(ctx context.Context, key string) (T, error) {
	var val T
	var result []byte
	// var val []byte
	err := c.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(definition.GenKey(c.config, c.prefix, key)))
		if err != nil {
			return err
		}
		result, err = item.ValueCopy(nil)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			if c.fetcher == nil {
				return val, definition.ErrNotFound
			}
			val, err = c.fetcher(key)
			if err != nil {
				return val, err
			}
			err = c.Set(ctx, key, val)
			if err != nil {
				return val, err
			}
			return val, nil
		}
		return val, fmt.Errorf("get value failed: %w", err)
	}
	err = json.Unmarshal(result, &val)
	if err != nil {
		return val, fmt.Errorf("unmarshal value failed: %w", err)
	}
	return val, nil
}

// Del deletes the value corresponding to the given key from the cache.
func (c *cacher[T]) Del(ctx context.Context, key string) error {
	return c.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(definition.GenKey(c.config, c.prefix, key)))
	})
}
