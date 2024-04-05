// Copyright 2024 sigma
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
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"

	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/dal/redis"
	"github.com/go-sigma/sigma/pkg/types/enums"
)

func TestRedisAcquire(t *testing.T) {
	miniRedis := miniredis.RunT(t)
	config := configs.Configuration{
		Redis: configs.ConfigurationRedis{
			Type: enums.RedisTypeExternal,
			Url:  "redis://" + miniRedis.Addr(),
		},
	}

	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	err := redis.Initialize(ctx, config)
	assert.NoError(t, err)

	c, err := New(config)
	assert.NoError(t, err)

	const key = "test-redis-lock"
	var concurrency uint64 = 500

	var wg sync.WaitGroup
	var cnt uint64 = 0
	for i := uint64(0); i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			l, err := c.Acquire(ctx, key, time.Second*1, time.Second*3)
			assert.Equal(t, true, err == nil || errors.Is(err, context.DeadlineExceeded))
			if l != nil {
				<-time.After(time.Millisecond * 100)
				defer l.Unlock(ctx) // nolint: errcheck
			}
			atomic.AddUint64(&cnt, 1)
		}()
	}
	wg.Wait()
	assert.True(t, true, concurrency > cnt && cnt > 1)
}

func TestRedisAcquireWithRenew(t *testing.T) {
	miniRedis := miniredis.RunT(t)
	config := configs.Configuration{
		Redis: configs.ConfigurationRedis{
			Type: enums.RedisTypeExternal,
			Url:  "redis://" + miniRedis.Addr(),
		},
	}

	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	err := redis.Initialize(ctx, config)
	assert.NoError(t, err)

	c, err := New(config)
	assert.NoError(t, err)

	const key = "test-redis-lock"
	var concurrency uint64 = 500

	var wg sync.WaitGroup
	var cnt uint64 = 0
	for i := uint64(0); i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := c.AcquireWithRenew(ctx, key, time.Second*1, time.Second*3)
			if errors.Is(err, context.DeadlineExceeded) {
				atomic.AddUint64(&cnt, 1)
			}
		}()
	}
	wg.Wait()
	assert.Equal(t, cnt, concurrency-1)
}
