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
	"fmt"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/redis/go-redis/v9"

	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/modules/cacher/definition"
)

type cacher[T any] struct {
	redisCli redis.UniversalClient
	prefix   string
	fetcher  definition.Fetcher[T]
	config   configs.Configuration
}

// New returns a new Cacher.
func New[T any](config configs.Configuration, prefix string, fetcher definition.Fetcher[T]) (definition.Cacher[T], error) {
	redisOpt, err := redis.ParseURL(config.Redis.Url)
	if err != nil {
		return nil, err
	}
	return &cacher[T]{
		redisCli: redis.NewClient(redisOpt),
		prefix:   prefix,
		fetcher:  fetcher,
		config:   config,
	}, nil
}

// Set sets the value of given key if it is new to the cache.
// Param val should not be nil.
func (c *cacher[T]) Set(ctx context.Context, key string, val T, ttls ...time.Duration) error {
	content, err := jsoniter.MarshalToString(val)
	if err != nil {
		return fmt.Errorf("marshal value failed: %w", err)
	}
	var ttl = c.config.Cache.Redis.Ttl
	if len(ttls) > 0 {
		ttl = ttls[0]
	}
	return c.redisCli.Set(ctx, definition.GenKey(c.config, c.prefix, key), content, ttl).Err()
}

// Get tries to fetch a value corresponding to the given key from the cache.
// If error occurs during the first time fetching, it will be cached until the
// sequential fetching triggered by the refresh goroutine succeed.
func (c *cacher[T]) Get(ctx context.Context, key string) (T, error) {
	var result T
	content, err := c.redisCli.Get(ctx, definition.GenKey(c.config, c.prefix, key)).Result()
	if err != nil {
		if err == redis.Nil {
			if c.fetcher == nil {
				return result, definition.ErrNotFound
			}
			result, err = c.fetcher(key)
			if err != nil {
				return result, err
			}
			err = c.Set(ctx, key, result)
			if err != nil {
				return result, fmt.Errorf("set value failed: %w", err)
			}
			return result, nil
		}
		return result, fmt.Errorf("get value failed: %w", err)
	}

	err = jsoniter.UnmarshalFromString(content, &result)
	if err != nil {
		return result, fmt.Errorf("unmarshal value failed: %w", err)
	}
	return result, nil
}

// Del deletes the value corresponding to the given key from the cache.
func (c *cacher[T]) Del(ctx context.Context, key string) error {
	return c.redisCli.Del(ctx, definition.GenKey(c.config, c.prefix, key)).Err()
}
