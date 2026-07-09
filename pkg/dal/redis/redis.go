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

package redis

import (
	"context"
	"fmt"
	"sync"

	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"

	"github.com/go-sigma/sigma/pkg/config"
)

// Client ...
var Client redis.UniversalClient

// ClientFactory lazily creates a Redis client.
type ClientFactory func() (redis.UniversalClient, error)

// NewClientFactory returns a memoized Redis client factory.
func NewClientFactory(config *config.Configuration) ClientFactory {
	var once sync.Once
	var cli redis.UniversalClient
	var err error
	return func() (redis.UniversalClient, error) {
		once.Do(func() {
			cli, err = New(config)
		})
		return cli, err
	}
}

// New new redis instance
func New(config *config.Configuration) (redis.UniversalClient, error) {
	if !config.Redis.Enabled() {
		return nil, nil
	}
	opt := &redis.UniversalOptions{
		Addrs:        append([]string(nil), config.Redis.Addrs...),
		Username:     config.Redis.Username,
		Password:     config.Redis.Password,
		DB:           config.Redis.DB,
		MasterName:   config.Redis.MasterName,
		PoolSize:     config.Redis.PoolSize,
		MinIdleConns: config.Redis.MinIdleConns,
		MaxRetries:   config.Redis.MaxRetries,
	}
	if config.Redis.URL != "" {
		parsed, err := redis.ParseURL(config.Redis.URL)
		if err != nil {
			return nil, fmt.Errorf("redis parse url failed: %v", err)
		}
		opt.Addrs = []string{parsed.Addr}
		opt.Username = parsed.Username
		opt.Password = parsed.Password
		opt.DB = parsed.DB
	}
	if len(opt.Addrs) == 0 && opt.MasterName == "" {
		return nil, fmt.Errorf("redis: no addrs or url provided")
	}
	redisCli := redis.NewUniversalClient(opt)
	if err := redisotel.InstrumentTracing(redisCli); err != nil {
		return nil, fmt.Errorf("instrument redis with otel: %w", err)
	}
	res, err := redisCli.Ping(context.Background()).Result()
	if err != nil {
		return nil, fmt.Errorf("redis ping failed: %v", err)
	}
	if res != "PONG" {
		return nil, fmt.Errorf("redis ping should got PONG, real: %s", res)
	}
	return redisCli, nil
}
