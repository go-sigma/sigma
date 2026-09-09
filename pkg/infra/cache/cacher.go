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

package cacher

import (
	"context"
	"fmt"
	"time"

	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/config"
	dalredis "github.com/go-sigma/sigma/pkg/dal/redis"
)

//go:generate mockgen -destination=cacher_mocks.go -package=cacher github.com/go-sigma/sigma/pkg/infra/cache Cacher

// Fetcher loads a value when the cache misses.
type Fetcher[T any] func(ctx context.Context, key string) (T, error)

// Cacher ...
type Cacher[T any] interface {
	// Set sets the value of given key if it is new to the cache.
	// Param val should not be nil.
	Set(ctx context.Context, key string, val T, ttls ...time.Duration) error
	// Get tries to fetch a value corresponding to the given key from the cache.
	// If error occurs during the first time fetching, it will be cached until the
	// sequential fetching triggered by the refresh goroutine succeed.
	Get(ctx context.Context, key string) (T, error)
	// Del deletes the value corresponding to the given key from the cache.
	Del(ctx context.Context, key string) error
}

var (
	ErrValNil   = fmt.Errorf("val should not be nil")
	ErrNotFound = fmt.Errorf("key not found")
)

// Params declares dependencies needed to construct a cacher.
type Params struct {
	Config             *config.Configuration
	RedisClientFactory dalredis.ClientFactory
}

func genKey(config *config.Configuration, prefix, key string) string {
	return fmt.Sprintf("%s:%s:%s", config.Cache.Prefix, prefix, key)
}

func genNegativeKey(config *config.Configuration, prefix, key string) string {
	return fmt.Sprintf("%s:missing", genKey(config, prefix, key))
}

// New creates a read-through cache with default options.
//
// It is a convenience wrapper around NewWithOptions that uses Options{} so the
// TTL, NegativeTTL and IsNotFound defaults are resolved from the config (see
// Options.withDefaults). Pass a custom Options to NewWithOptions when you need
// to override those defaults.
//
// On a miss, Get invokes fetcher to load the value; when fetcher returns an
// error matching IsNotFound, a negative entry is cached so subsequent lookups
// short-circuit without hitting the backend until NegativeTTL expires.
func New[T any](params Params, prefix string, fetcher Fetcher[T]) (Cacher[T], error) {
	return NewWithOptions(params, prefix, fetcher, Options{})
}

// NewWithOptions creates a cache with custom read-through options.
func NewWithOptions[T any](params Params, prefix string, fetcher Fetcher[T], options Options) (Cacher[T], error) {
	var err error
	var cacher Cacher[T]
	options = options.withDefaults(params)
	switch params.Config.Cache.Type {
	case enums.CacherTypeRedis:
		cacher, err = newRedis(params, prefix, fetcher, options)
	case enums.CacherTypeInmemory:
		cacher, err = newMemory(params.Config, prefix, fetcher, options)
	default:
		cacher, err = newMemory(params.Config, prefix, fetcher, options)
	}
	return cacher, err
}
