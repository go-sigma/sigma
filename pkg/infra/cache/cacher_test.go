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

package cacher

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"

	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/config"
)

func TestMemoryCacheSingleflight(t *testing.T) {
	cfg := testCacheConfig(enums.CacherTypeInmemory)
	var calls atomic.Int64
	block := make(chan struct{})

	cacheCli, err := NewWithOptions(Params{Config: cfg}, "singleflight", func(_ context.Context, key string) (string, error) {
		calls.Add(1)
		<-block
		return "value:" + key, nil
	}, Options{})
	require.NoError(t, err)

	const workers = 8
	var wg sync.WaitGroup
	results := make([]string, workers)
	errs := make([]error, workers)
	wg.Add(workers)
	for i := range workers {
		go func(i int) {
			defer wg.Done()
			results[i], errs[i] = cacheCli.Get(t.Context(), "key")
		}(i)
	}
	time.Sleep(50 * time.Millisecond)
	close(block)
	wg.Wait()

	require.Equal(t, int64(1), calls.Load())
	for i := range workers {
		require.NoError(t, errs[i])
		require.Equal(t, "value:key", results[i])
	}
}

func TestMemoryCacheTTL(t *testing.T) {
	cfg := testCacheConfig(enums.CacherTypeInmemory)
	cacheCli, err := NewWithOptions(Params{Config: cfg}, "ttl", func(_ context.Context, _ string) (string, error) {
		return "value-" + time.Now().Format(time.RFC3339Nano), nil
	}, Options{TTL: 20 * time.Millisecond})
	require.NoError(t, err)

	first, err := cacheCli.Get(t.Context(), "key")
	require.NoError(t, err)

	second, err := cacheCli.Get(t.Context(), "key")
	require.NoError(t, err)
	require.Equal(t, first, second)

	require.Eventually(t, func() bool {
		third, err := cacheCli.Get(t.Context(), "key")
		require.NoError(t, err)
		return third != first
	}, time.Second, 20*time.Millisecond)
}

func TestMemoryCacheNegativeCache(t *testing.T) {
	notFoundErr := errors.New("not found")
	cfg := testCacheConfig(enums.CacherTypeInmemory)
	var calls atomic.Int64
	cacheCli, err := NewWithOptions(Params{Config: cfg}, "negative", func(context.Context, string) (string, error) {
		calls.Add(1)
		return "", notFoundErr
	}, Options{
		NegativeTTL: 100 * time.Millisecond,
		IsNotFound: func(err error) bool {
			return errors.Is(err, notFoundErr)
		},
	})
	require.NoError(t, err)

	_, err = cacheCli.Get(t.Context(), "missing")
	require.ErrorIs(t, err, ErrNotFound)
	_, err = cacheCli.Get(t.Context(), "missing")
	require.ErrorIs(t, err, ErrNotFound)
	require.Equal(t, int64(1), calls.Load())
}

func TestRedisCacheNegativeCache(t *testing.T) {
	notFoundErr := errors.New("not found")
	miniRedis := miniredis.RunT(t)
	cfg := testCacheConfig(enums.CacherTypeRedis)
	cfg.Redis.URL = "redis://" + miniRedis.Addr()
	var calls atomic.Int64

	cacheCli, err := NewWithOptions(Params{
		Config: cfg,
		RedisClientFactory: func() (redis.UniversalClient, error) {
			return redis.NewClient(&redis.Options{Addr: miniRedis.Addr()}), nil
		},
	}, "redis-negative", func(context.Context, string) (string, error) {
		calls.Add(1)
		return "", notFoundErr
	}, Options{
		NegativeTTL: time.Minute,
		IsNotFound: func(err error) bool {
			return errors.Is(err, notFoundErr)
		},
	})
	require.NoError(t, err)

	_, err = cacheCli.Get(t.Context(), "missing")
	require.ErrorIs(t, err, ErrNotFound)
	_, err = cacheCli.Get(t.Context(), "missing")
	require.ErrorIs(t, err, ErrNotFound)
	require.Equal(t, int64(1), calls.Load())
}

func testCacheConfig(cacheType enums.CacherType) *config.Configuration {
	cfg := &config.Configuration{}
	cfg.Cache.Type = cacheType
	cfg.Cache.WithDefaults()
	return cfg
}
