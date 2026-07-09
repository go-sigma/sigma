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
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"

	"github.com/go-sigma/sigma/pkg/config"
)

type redisCacher[T any] struct {
	redisCli redis.UniversalClient
	prefix   string
	fetcher  Fetcher[T]
	config   *config.Configuration
	options  Options
	group    singleflight.Group
}

// newRedis returns a new Cacher.
func newRedis[T any](params Params, prefix string, fetcher Fetcher[T], options Options) (Cacher[T], error) {
	config := params.Config
	if !config.Redis.Enabled() {
		return nil, fmt.Errorf("cacher: please check redis configuration, it should be configured")
	}
	if params.RedisClientFactory == nil {
		return nil, fmt.Errorf("cacher: redis client factory is required")
	}
	redisCli, err := params.RedisClientFactory()
	if err != nil {
		return nil, err
	}
	if redisCli == nil {
		return nil, fmt.Errorf("cacher: redis client is required")
	}
	return &redisCacher[T]{
		redisCli: redisCli,
		prefix:   prefix,
		fetcher:  fetcher,
		config:   config,
		options:  options,
	}, nil
}

// Set sets the value of given key if it is new to the cache.
// Param val should not be nil.
func (c *redisCacher[T]) Set(ctx context.Context, key string, val T, ttls ...time.Duration) error {
	content, err := json.Marshal(val)
	if err != nil {
		return fmt.Errorf("marshal value failed: %w", err)
	}
	var ttl = c.options.TTL
	if len(ttls) > 0 {
		ttl = ttls[0]
	}
	cacheKey := genKey(c.config, c.prefix, key)
	if err := c.redisCli.Set(ctx, cacheKey, string(content), ttl).Err(); err != nil {
		return err
	}
	return c.redisCli.Del(ctx, genNegativeKey(c.config, c.prefix, key)).Err()
}

// Get tries to fetch a value corresponding to the given key from the cache.
// If error occurs during the first time fetching, it will be cached until the
// sequential fetching triggered by the refresh goroutine succeed.
func (c *redisCacher[T]) Get(ctx context.Context, key string) (T, error) {
	if result, ok, err := c.getCached(ctx, key); ok || err != nil {
		return result, err
	}

	recordCacheRequest(backendRedis, c.prefix, resultMiss)
	var zero T
	if c.fetcher == nil {
		return zero, ErrNotFound
	}

	value, err, _ := c.group.Do(genKey(c.config, c.prefix, key), func() (any, error) {
		if result, ok, err := c.getCached(ctx, key); ok || err != nil {
			return result, err
		}

		startedAt := time.Now()
		result, err := c.fetcher(ctx, key)
		if err != nil {
			observeCacheFetch(backendRedis, c.prefix, resultFetchError, startedAt)
			if c.options.IsNotFound(err) {
				if setErr := c.setNegative(ctx, key); setErr != nil {
					recordCacheRequest(backendRedis, c.prefix, resultSetError)
					return result, fmt.Errorf("set negative value failed: %w", setErr)
				}
				return result, ErrNotFound
			}
			return result, err
		}
		observeCacheFetch(backendRedis, c.prefix, resultMiss, startedAt)
		err = c.Set(ctx, key, result)
		if err != nil {
			recordCacheRequest(backendRedis, c.prefix, resultSetError)
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
func (c *redisCacher[T]) Del(ctx context.Context, key string) error {
	return c.redisCli.Del(ctx, genKey(c.config, c.prefix, key), genNegativeKey(c.config, c.prefix, key)).Err()
}

func (c *redisCacher[T]) getCached(ctx context.Context, key string) (T, bool, error) {
	var result T
	content, err := c.redisCli.Get(ctx, genKey(c.config, c.prefix, key)).Result()
	if err == nil {
		if err = json.Unmarshal([]byte(content), &result); err != nil {
			return result, false, fmt.Errorf("unmarshal value failed: %w", err)
		}
		recordCacheRequest(backendRedis, c.prefix, resultHit)
		return result, true, nil
	}
	if err != redis.Nil {
		return result, false, fmt.Errorf("get value failed: %w", err)
	}

	exists, err := c.redisCli.Exists(ctx, genNegativeKey(c.config, c.prefix, key)).Result()
	if err != nil {
		return result, false, fmt.Errorf("get negative value failed: %w", err)
	}
	if exists > 0 {
		recordCacheRequest(backendRedis, c.prefix, resultNegativeHit)
		return result, false, ErrNotFound
	}
	return result, false, nil
}

func (c *redisCacher[T]) setNegative(ctx context.Context, key string) error {
	return c.redisCli.Set(ctx, genNegativeKey(c.config, c.prefix, key), "1", c.options.NegativeTTL).Err()
}
