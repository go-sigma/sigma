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
	"time"

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/redis/go-redis/v9"

	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/modules/locker/definition"
)

type lockerRedis struct {
	redsync *redsync.Redsync
}

func New(config configs.Configuration) (definition.Locker, error) {
	redisOpt, err := redis.ParseURL(config.Redis.Url)
	if err != nil {
		return nil, err
	}
	return &lockerRedis{
		redsync: redsync.New(goredis.NewPool(redis.NewClient(redisOpt))),
	}, nil
}

type lock struct {
	mutex *redsync.Mutex
}

func (l lockerRedis) Lock(ctx context.Context, name string, expire time.Duration) (definition.Lock, error) {
	var opts = []redsync.Option{redsync.WithRetryDelay(consts.LockerRetryDelay), redsync.WithTries(consts.LockerRetryMaxTimes)}
	if expire != 0 {
		opts = append(opts, redsync.WithExpiry(expire))
	}
	mutex := l.redsync.NewMutex(consts.LockerMigration, opts...)
	return &lock{mutex: mutex}, nil
}

func (l lock) Unlock() error {
	_, err := l.mutex.Unlock()
	return err
}
