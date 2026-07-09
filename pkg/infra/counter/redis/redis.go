// Copyright 2026 sigma
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
	"strconv"

	"github.com/redis/go-redis/v9"

	counter "github.com/go-sigma/sigma/pkg/infra/counter"
	"github.com/go-sigma/sigma/pkg/utils"
)

func init() {
	utils.PanicIf(counter.Factories.Register("redis", &factory{}))
}

type redisCounter struct {
	redisCli redis.UniversalClient
}

type factory struct{}

var _ counter.Factory = factory{}

// New creates a new Redis-backed counter.
func (factory) New(params counter.Params) (counter.Counter, error) {
	if params.RedisClientFactory == nil {
		return nil, fmt.Errorf("redis client factory is required for redis counter")
	}
	redisCli, err := params.RedisClientFactory()
	if err != nil {
		return nil, err
	}
	if redisCli == nil {
		return nil, fmt.Errorf("redis client is required for redis counter")
	}
	return &redisCounter{redisCli: redisCli}, nil
}

// HIncrBy increments hash fields by the given deltas.
func (c *redisCounter) HIncrBy(ctx context.Context, key string, fields map[string]int64) error {
	pipe := c.redisCli.Pipeline()
	for field, delta := range fields {
		pipe.HIncrBy(ctx, key, field, delta)
	}
	_, err := pipe.Exec(ctx)
	return err
}

// HGetAll returns all fields and values of a hash.
func (c *redisCounter) HGetAll(ctx context.Context, key string) (map[string]int64, error) {
	raw, err := c.redisCli.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	if len(raw) == 0 {
		return nil, nil
	}
	result := make(map[string]int64, len(raw))
	for field, val := range raw {
		n, parseErr := strconv.ParseInt(val, 10, 64)
		if parseErr != nil {
			return nil, fmt.Errorf("parse counter value for field %q: %w", field, parseErr)
		}
		result[field] = n
	}
	return result, nil
}

// Del removes one or more keys.
func (c *redisCounter) Del(ctx context.Context, keys ...string) error {
	return c.redisCli.Del(ctx, keys...).Err()
}

// SAdd adds members to a set.
func (c *redisCounter) SAdd(ctx context.Context, key string, members ...string) error {
	m := make([]any, len(members))
	for i, v := range members {
		m[i] = v
	}
	return c.redisCli.SAdd(ctx, key, m...).Err()
}

// SMembers returns all members of a set.
func (c *redisCounter) SMembers(ctx context.Context, key string) ([]string, error) {
	return c.redisCli.SMembers(ctx, key).Result()
}

// SRem removes members from a set.
func (c *redisCounter) SRem(ctx context.Context, key string, members ...string) error {
	m := make([]any, len(members))
	for i, v := range members {
		m[i] = v
	}
	return c.redisCli.SRem(ctx, key, m...).Err()
}

// Rename atomically renames a key. Returns nil if the target already exists.
func (c *redisCounter) Rename(ctx context.Context, oldKey, newKey string) error {
	exists, err := c.redisCli.Exists(ctx, newKey).Result()
	if err != nil {
		return err
	}
	if exists > 0 {
		return nil
	}
	exists, err = c.redisCli.Exists(ctx, oldKey).Result()
	if err != nil {
		return err
	}
	if exists == 0 {
		return nil
	}
	return c.redisCli.Rename(ctx, oldKey, newKey).Err()
}

// Exists returns the number of keys that exist among the given keys.
func (c *redisCounter) Exists(ctx context.Context, keys ...string) (int64, error) {
	return c.redisCli.Exists(ctx, keys...).Result()
}

// Shutdown gracefully shuts down the counter.
func (c *redisCounter) Shutdown(_ context.Context) error {
	return nil
}
