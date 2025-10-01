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

	"github.com/bytedance/sonic"
	"github.com/redis/go-redis/v9"
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils"
)

type redisCacher[T any] struct {
	redisCli redis.UniversalClient
	prefix   string
	fetcher  Fetcher[T]
	config   configs.Configuration
}

// New returns a new Cacher.
func newRedis[T any](digCon *dig.Container, prefix string, fetcher Fetcher[T]) (Cacher[T], error) {
	config := utils.MustGetObjFromDigCon[configs.Configuration](digCon)
	if config.Redis.Type != enums.RedisTypeExternal {
		return nil, fmt.Errorf("cacher: please check redis configuration, it should be external")
	}
	redisOpt, err := redis.ParseURL(config.Redis.URL)
	if err != nil {
		return nil, err
	}
	return &redisCacher[T]{
		redisCli: redis.NewClient(redisOpt),
		prefix:   prefix,
		fetcher:  fetcher,
		config:   config,
	}, nil
}

// Set sets the value of given key if it is new to the cache.
// Param val should not be nil.
func (c *redisCacher[T]) Set(ctx context.Context, key string, val T, ttls ...time.Duration) error {
	content, err := sonic.MarshalString(val)
	if err != nil {
		return fmt.Errorf("marshal value failed: %w", err)
	}
	var ttl = c.config.Cache.Redis.Ttl
	if len(ttls) > 0 {
		ttl = ttls[0]
	}
	return c.redisCli.Set(ctx, genKey(c.config, c.prefix, key), content, ttl).Err()
}

// Get tries to fetch a value corresponding to the given key from the cache.
// If error occurs during the first time fetching, it will be cached until the
// sequential fetching triggered by the refresh goroutine succeed.
func (c *redisCacher[T]) Get(ctx context.Context, key string) (T, error) {
	var result T
	content, err := c.redisCli.Get(ctx, genKey(c.config, c.prefix, key)).Result()
	if err != nil {
		if err == redis.Nil {
			if c.fetcher == nil {
				return result, ErrNotFound
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

	err = sonic.UnmarshalString(content, &result)
	if err != nil {
		return result, fmt.Errorf("unmarshal value failed: %w", err)
	}
	return result, nil
}

// Del deletes the value corresponding to the given key from the cache.
func (c *redisCacher[T]) Del(ctx context.Context, key string) error {
	return c.redisCli.Del(ctx, genKey(c.config, c.prefix, key)).Err()
}
