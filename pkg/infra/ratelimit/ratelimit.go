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

package ratelimit

import (
	"context"
	"time"

	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/infra/cache"
)

// cachePrefix namespaces failed-login records in the shared cache.
const cachePrefix = "login-failure"

// FailureRecord is the cached failed-attempt counter for a username.
type FailureRecord struct {
	Count int64 `json:"count"`
}

// Limiter tracks failed login attempts for a username. Records live in the
// shared infra cache so multi-replica deployments stay consistent.
type Limiter interface {
	// Count returns the current failure count for key.
	Count(ctx context.Context, key string) (int64, error)
	// Incr increments the failure count for key within the configured window.
	Incr(ctx context.Context, key string) error
	// Reset clears the failure count for key after a successful login.
	Reset(ctx context.Context, key string) error
}

type cacheLimiter struct {
	cacher cache.Cacher[FailureRecord]
	window time.Duration
}

// New returns a cache-backed Limiter. The cache backend (inmemory or redis) is
// selected by config.cache.type.
func New(cfg *config.ConfigurationAuthLoginRateLimit, params cache.Params) (Limiter, error) {
	c, err := cache.New(params, cachePrefix, func(context.Context, string) (FailureRecord, error) {
		return FailureRecord{}, nil
	})
	if err != nil {
		return nil, err
	}
	return &cacheLimiter{cacher: c, window: cfg.Window}, nil
}

// Count returns the current failure count for key.
func (l *cacheLimiter) Count(ctx context.Context, key string) (int64, error) {
	record, err := l.cacher.Get(ctx, key)
	if err != nil {
		recordCacheError()
		return 0, err
	}
	return record.Count, nil
}

// Incr increments the failure count for key and refreshes the window TTL.
func (l *cacheLimiter) Incr(ctx context.Context, key string) error {
	record, err := l.cacher.Get(ctx, key)
	if err != nil {
		recordCacheError()
		return err
	}
	record.Count++
	if err := l.cacher.Set(ctx, key, record, l.window); err != nil {
		recordCacheError()
		return err
	}
	recordFailure()
	return nil
}

// Reset clears the failure count for key.
func (l *cacheLimiter) Reset(ctx context.Context, key string) error {
	if err := l.cacher.Del(ctx, key); err != nil {
		recordCacheError()
		return err
	}
	return nil
}
