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
	"time"

	lru "github.com/hashicorp/golang-lru/v2"

	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/modules/cacher/definition"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

// ValueWithTtl ...
type ValueWithTtl[T any] struct {
	Value T
	Ttl   *time.Time
}

type cacher[T any] struct {
	config  configs.Configuration
	cache   *lru.TwoQueueCache[string, ValueWithTtl[T]]
	prefix  string
	fetcher definition.Fetcher[T]
}

// New returns a new Cacher.
func New[T any](config configs.Configuration, prefix string, fetcher definition.Fetcher[T]) (definition.Cacher[T], error) {
	cache, err := lru.New2Q[string, ValueWithTtl[T]](config.Cache.Inmemory.Size)
	if err != nil {
		return nil, err
	}
	return &cacher[T]{
		config:  config,
		cache:   cache,
		prefix:  prefix,
		fetcher: fetcher,
	}, nil
}

// Set sets the value of given key if it is new to the cache.
// Param val should not be nil.
func (c *cacher[T]) Set(ctx context.Context, key string, val T, ttls ...time.Duration) error {
	value := ValueWithTtl[T]{
		Value: val,
		Ttl:   ptr.Of(time.Now().Add(c.config.Cache.Ttl)),
	}
	if len(ttls) > 0 {
		value.Ttl = ptr.Of(time.Now().Add(ttls[0]))
	}
	c.cache.Add(c.key(key), value)
	return nil
}

// Get tries to fetch a value corresponding to the given key from the cache.
// If error occurs during the first time fetching, it will be cached until the
// sequential fetching triggered by the refresh goroutine succeed.
func (c *cacher[T]) Get(ctx context.Context, key string) (T, error) {
	result, ok := c.cache.Get(c.key(key))
	if !ok {
		if c.fetcher == nil {
			return result.Value, definition.ErrNotFound
		}
		result, err := c.fetcher(key)
		if err != nil {
			return result, err
		}
		err = c.Set(ctx, key, result, c.config.Cache.Ttl)
		if err != nil {
			return result, fmt.Errorf("Set value failed: %w", err)
		}
		return result, nil
	}
	if result.Ttl == nil {
		return result.Value, nil
	}
	if result.Ttl.After(time.Now()) {
		return result.Value, nil
	}
	err := c.Del(ctx, key)
	if err != nil {
		return result.Value, err
	}
	return c.Get(ctx, key)
}

// Del deletes the value corresponding to the given key from the cache.
func (c *cacher[T]) Del(ctx context.Context, key string) error {
	c.cache.Remove(c.key(key))
	return nil
}

func (c *cacher[T]) key(key string) string {
	return fmt.Sprintf("%s:%s", c.prefix, key)
}
