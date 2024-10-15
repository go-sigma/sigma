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
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/require"
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/dal/redis"
	"github.com/go-sigma/sigma/pkg/logger"
	"github.com/go-sigma/sigma/pkg/types/enums"
)

func TestNew(t *testing.T) {
	logger.SetLevel("debug")

	tests := []struct {
		name      string
		newDigCon func(*testing.T) *dig.Container
		wantErr   bool
	}{
		{
			name: "normal",
			newDigCon: func(t *testing.T) *dig.Container {
				digCon := dig.New()
				err := digCon.Provide(func() configs.Configuration {
					return configs.Configuration{
						Locker: configs.ConfigurationLocker{
							Type:   enums.LockerTypeRedis,
							Prefix: "sigma-locker",
						},
						Redis: configs.ConfigurationRedis{
							Type: enums.RedisTypeExternal,
							Url:  "redis://" + miniredis.RunT(t).Addr(),
						},
					}
				})
				require.NoError(t, err)

				err = digCon.Provide(redis.New)
				require.NoError(t, err)

				return digCon
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			locker, err := New(tt.newDigCon(t))
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			ctx := context.Background()

			{
				const key = "test-redis-lock"
				var concurrency uint64 = 500
				var wg sync.WaitGroup
				var cnt uint64 = 0
				for i := uint64(0); i < concurrency; i++ {
					wg.Add(1)
					go func() {
						defer wg.Done()
						l, err := locker.Acquire(ctx, key, time.Second*1, time.Second*3)
						require.Equal(t, true, err == nil || errors.Is(err, context.DeadlineExceeded))
						if !(err == nil || errors.Is(err, context.DeadlineExceeded)) {
							require.NoError(t, fmt.Errorf("acquire lock failed"))
						}
						if l != nil {
							<-time.After(time.Millisecond * 100)
							err = l.Unlock(ctx)
							require.NoError(t, err)
							atomic.AddUint64(&cnt, 1)
						}
					}()
				}
				wg.Wait()
				require.True(t, true, concurrency > cnt && cnt > 1)
			}
			{
				const key = "test-redis-lock"
				var concurrency uint64 = 500
				var wg sync.WaitGroup
				var cnt uint64 = 0
				for i := uint64(0); i < concurrency; i++ {
					wg.Add(1)
					go func() {
						defer wg.Done()
						err := locker.AcquireWithRenew(ctx, key, time.Second*1, time.Second*3)
						if errors.Is(err, context.DeadlineExceeded) {
							atomic.AddUint64(&cnt, 1)
						}
					}()
				}
				wg.Wait()
				require.Equal(t, true, cnt >= 1)
			}
		})
	}
}
