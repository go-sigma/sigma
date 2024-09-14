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
	"math/rand/v2" // nolint: gosec
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"

	"github.com/go-sigma/sigma/pkg/configs"
	rds "github.com/go-sigma/sigma/pkg/dal/redis"
	"github.com/go-sigma/sigma/pkg/modules/locker/definition"
	"github.com/go-sigma/sigma/pkg/types/enums"
)

var (
	luaRelease = redis.NewScript(`if redis.call("get", KEYS[1]) == ARGV[1] then return redis.call("del", KEYS[1]) else return 0 end`)
	luaRenew   = redis.NewScript(`if redis.call('get', KEYS[1]) == ARGV[1] then return redis.call('expire', KEYS[1], ARGV[2]) else return 0 end`)
)

type lockerRedis struct {
	redisCli redis.UniversalClient
}

func New(config configs.Configuration) (definition.Locker, error) {
	if config.Redis.Type != enums.RedisTypeExternal {
		return nil, fmt.Errorf("redislock: please check redis configuration, it should be external")
	}
	return &lockerRedis{
		redisCli: rds.Client,
	}, nil
}

type lock struct {
	redisCli   redis.UniversalClient
	key, value string
	expire     time.Duration
}

func (l lockerRedis) Acquire(ctx context.Context, key string, expire, waitTimeout time.Duration) (definition.Lock, error) {
	if expire < definition.MinLockExpire {
		return nil, definition.ErrLockTooShort
	}
	ddlCtx, cancel := context.WithTimeout(ctx, waitTimeout)
	defer cancel()
	val := fmt.Sprintf("%d-%d", rand.Int(), time.Now().Nanosecond()) // nolint: gosec
	ticker := time.NewTicker(time.Duration(100) * time.Millisecond)
	defer func() {
		ticker.Stop()
	}()
	for {
		select {
		case <-ddlCtx.Done():
			return nil, ddlCtx.Err()
		case <-ticker.C:
			ok, err := l.redisCli.SetNX(ddlCtx, key, val, expire).Result()
			if err != nil {
				return nil, err
			}
			if ok {
				return &lock{
					redisCli: l.redisCli,
					key:      key,
					value:    val,
					expire:   expire,
				}, nil
			}
		}
	}
}

// AcquireWithRenew acquire lock with renew the lock
func (l lockerRedis) AcquireWithRenew(ctx context.Context, key string, expire, waitTimeout time.Duration) error {
	lock, err := l.Acquire(ctx, key, expire, waitTimeout)
	if err != nil {
		return err
	}

	var tick = time.Duration(100) * time.Millisecond

	go func() {
		ticker := time.NewTicker(tick)
		defer func() {
			ticker.Stop()
		}()
		for {
			select {
			case <-ctx.Done():
				err := lock.Unlock(context.Background()) // should always release the locker
				if err != nil {
					log.Error().Err(err).Msg("release lock failed")
				}
				return
			case <-ticker.C:
			}
			if err := lock.Renew(ctx, expire); err != nil {
				return
			}
		}
	}()
	return nil
}

// Renew renew the lock
func (l lock) Renew(ctx context.Context, ttls ...time.Duration) error {
	var expire time.Duration
	if len(ttls) == 0 {
		expire = l.expire
	} else {
		expire = ttls[0]
	}
	if expire < definition.MinLockExpire {
		return definition.ErrLockTooShort
	}
	res, err := luaRenew.Run(ctx, l.redisCli, []string{l.key}, l.value, int64(expire/time.Second)).Result()
	if err == redis.Nil {
		return definition.ErrLockNotHeld
	} else if err != nil {
		return err
	}

	if i, ok := res.(int64); !ok || i != 1 {
		return definition.ErrLockNotHeld
	}
	return nil
}

// Unlock unlock the lock
func (l lock) Unlock(ctx context.Context) error {
	res, err := luaRelease.Run(ctx, l.redisCli, []string{l.key}, l.value).Result()
	if err == redis.Nil {
		return definition.ErrLockNotHeld
	} else if err != nil {
		return err
	}

	if i, ok := res.(int64); !ok || i != 1 {
		return definition.ErrLockNotHeld
	}
	return nil
}
