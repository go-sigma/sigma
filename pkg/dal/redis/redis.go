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
	"fmt"

	"github.com/redis/go-redis/v9"

	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/types/enums"
)

// Client ...
var Client redis.UniversalClient

// Initialize init redis
func Initialize(ctx context.Context, config configs.Configuration) error {
	if config.Redis.Type == enums.RedisTypeNone {
		return nil
	}
	redisOpt, err := redis.ParseURL(config.Redis.Url)
	if err != nil {
		return fmt.Errorf("redis.ParseURL error: %v", err)
	}
	redisCli := redis.NewClient(redisOpt)
	res, err := redisCli.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("redis ping error: %v", err)
	}
	if res != "PONG" {
		return fmt.Errorf("redis ping should got PONG, real: %s", res)
	}
	Client = redisCli
	return nil
}
