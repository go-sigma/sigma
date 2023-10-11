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

package database

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	jsoniter "github.com/json-iterator/go"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/dal/dao"
	"github.com/go-sigma/sigma/pkg/modules/cacher/definition"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

// ValueWithTtl ...
type ValueWithTtl struct {
	Value json.RawMessage
	Ttl   *time.Time
}

type cacher[T any] struct {
	cacheService dao.CacheService
	prefix       string
	fetcher      definition.Fetcher[T]
	config       configs.Configuration
}

// New returns a new Cacher.
func New[T any](config configs.Configuration, prefix string, fetcher definition.Fetcher[T]) (definition.Cacher[T], error) {
	return &cacher[T]{
		cacheService: dao.NewCacheServiceFactory().New(),
		prefix:       prefix,
		fetcher:      fetcher,
		config:       config,
	}, nil
}

// Set sets the value of given key if it is new to the cache.
// Param val should not be nil.
func (c *cacher[T]) Set(ctx context.Context, key string, val T, ttls ...time.Duration) error {
	content, err := jsoniter.Marshal(val)
	if err != nil {
		return fmt.Errorf("marshal value failed: %w", err)
	}
	value := ValueWithTtl{
		Value: content,
		Ttl:   ptr.Of(time.Now().Add(c.config.Cache.Ttl)),
	}
	if len(ttls) > 0 {
		value.Ttl = ptr.Of(time.Now().Add(ttls[0]))
	}
	return c.cacheService.Save(ctx, c.key(key), utils.MustMarshal(value), c.config.Cache.Database.Size, c.config.Cache.Database.Threshold)
}

// Get tries to fetch a value corresponding to the given key from the cache.
// If error occurs during the first time fetching, it will be cached until the
// sequential fetching triggered by the refresh goroutine succeed.
func (c *cacher[T]) Get(ctx context.Context, key string) (T, error) {
	var result T
	content, err := c.cacheService.Get(ctx, c.key(key))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if c.fetcher == nil {
				return result, definition.ErrNotFound
			}
			result, err = c.fetcher(key)
			if err != nil {
				return result, err
			}
			err = c.Set(ctx, key, result, c.config.Cache.Ttl)
			if err != nil {
				return result, fmt.Errorf("set value failed: %w", err)
			}
			return result, nil
		}
		return result, fmt.Errorf("get value failed: %w", err)
	}
	var val ValueWithTtl
	err = jsoniter.Unmarshal(content.Val, &val)
	if err != nil {
		return result, fmt.Errorf("unmarshal value failed: %w", err)
	}
	if val.Ttl != nil && val.Ttl.After(time.Now()) {
		err = jsoniter.Unmarshal(val.Value, &result)
		if err != nil {
			return result, fmt.Errorf("unmarshal value failed: %w", err)
		}
		return result, nil
	}
	err = c.Del(ctx, key)
	if err != nil {
		return result, err
	}
	return result, definition.ErrNotFound
}

// Del deletes the value corresponding to the given key from the cache.
func (c *cacher[T]) Del(ctx context.Context, key string) error {
	return c.cacheService.Delete(ctx, c.key(key))
}

func (c *cacher[T]) key(key string) string {
	return fmt.Sprintf("%s:%s", c.prefix, key)
}
