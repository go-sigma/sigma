// Copyright 2025 sigma
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

package cacher

import (
	"context"
	"fmt"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
	"golang.org/x/sync/singleflight"

	"github.com/go-sigma/sigma/pkg/config"
)

type cacheEntry[T any] struct {
	value     T
	expiresAt time.Time
	negative  bool
}

type memoryCacher[T any] struct {
	config  *config.Configuration
	cache   *lru.TwoQueueCache[string, cacheEntry[T]]
	prefix  string
	fetcher Fetcher[T]
	options Options
	group   singleflight.Group
}

// newMemory returns a new Cacher.
func newMemory[T any](config *config.Configuration, prefix string, fetcher Fetcher[T], options Options) (Cacher[T], error) {
	cache, err := lru.New2Q[string, cacheEntry[T]](config.Cache.Inmemory.Size)
	if err != nil {
		return nil, err
	}
	return &memoryCacher[T]{
		config:  config,
		cache:   cache,
		prefix:  prefix,
		fetcher: fetcher,
		options: options,
	}, nil
}

// Set sets the value of given key if it is new to the cache.
// Param val should not be nil.
func (c *memoryCacher[T]) Set(_ context.Context, key string, val T, ttls ...time.Duration) error {
	c.cache.Add(genKey(c.config, c.prefix, key), cacheEntry[T]{
		value:     val,
		expiresAt: c.expiresAt(ttls...),
	})
	c.cache.Remove(genNegativeKey(c.config, c.prefix, key))
	return nil
}

// Get tries to fetch a value corresponding to the given key from the cache.
// If error occurs during the first time fetching, it will be cached until the
// sequential fetching triggered by the refresh goroutine succeed.
func (c *memoryCacher[T]) Get(ctx context.Context, key string) (T, error) {
	if result, ok, err := c.getCached(key); ok || err != nil {
		return result, err
	}

	recordCacheRequest(backendMemory, c.prefix, resultMiss)
	var zero T
	if c.fetcher == nil {
		return zero, ErrNotFound
	}

	value, err, _ := c.group.Do(genKey(c.config, c.prefix, key), func() (any, error) {
		if result, ok, err := c.getCached(key); ok || err != nil {
			return result, err
		}

		startedAt := time.Now()
		result, err := c.fetcher(ctx, key)
		if err != nil {
			observeCacheFetch(backendMemory, c.prefix, resultFetchError, startedAt)
			if c.options.IsNotFound(err) {
				if setErr := c.setNegative(key); setErr != nil {
					recordCacheRequest(backendMemory, c.prefix, resultSetError)
					return result, fmt.Errorf("set negative value failed: %w", setErr)
				}
				return result, ErrNotFound
			}
			return result, err
		}
		observeCacheFetch(backendMemory, c.prefix, resultMiss, startedAt)
		err = c.Set(ctx, key, result)
		if err != nil {
			recordCacheRequest(backendMemory, c.prefix, resultSetError)
			return result, fmt.Errorf("set value failed: %w", err)
		}
		return result, nil
	})
	if err != nil {
		return zero, err
	}
	result, ok := value.(T)
	if !ok {
		return zero, fmt.Errorf("cache value type mismatch")
	}
	return result, nil
}

// Del deletes the value corresponding to the given key from the cache.
func (c *memoryCacher[T]) Del(_ context.Context, key string) error {
	c.cache.Remove(genKey(c.config, c.prefix, key))
	c.cache.Remove(genNegativeKey(c.config, c.prefix, key))
	return nil
}

func (c *memoryCacher[T]) getCached(key string) (T, bool, error) {
	cacheKey := genKey(c.config, c.prefix, key)
	result, ok := c.cache.Get(cacheKey)
	if ok {
		if result.expired() {
			c.cache.Remove(cacheKey)
		} else {
			recordCacheRequest(backendMemory, c.prefix, resultHit)
			return result.value, true, nil
		}
	}

	negativeKey := genNegativeKey(c.config, c.prefix, key)
	result, ok = c.cache.Get(negativeKey)
	if ok {
		if result.expired() {
			c.cache.Remove(negativeKey)
		} else if result.negative {
			var zero T
			recordCacheRequest(backendMemory, c.prefix, resultNegativeHit)
			return zero, false, ErrNotFound
		}
	}

	var zero T
	return zero, false, nil
}

func (c *memoryCacher[T]) setNegative(key string) error {
	c.cache.Add(genNegativeKey(c.config, c.prefix, key), cacheEntry[T]{
		expiresAt: c.negativeExpiresAt(),
		negative:  true,
	})
	return nil
}

func (c *memoryCacher[T]) expiresAt(ttls ...time.Duration) time.Time {
	ttl := c.options.TTL
	if len(ttls) > 0 {
		ttl = ttls[0]
	}
	if ttl <= 0 {
		return time.Time{}
	}
	return time.Now().Add(ttl)
}

func (c *memoryCacher[T]) negativeExpiresAt() time.Time {
	if c.options.NegativeTTL <= 0 {
		return time.Time{}
	}
	return time.Now().Add(c.options.NegativeTTL)
}

func (e cacheEntry[T]) expired() bool {
	return !e.expiresAt.IsZero() && time.Now().After(e.expiresAt)
}
