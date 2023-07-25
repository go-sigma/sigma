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
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func testFetcher1(key string) (string, error) {
	return "test", nil
}

func TestNew(t *testing.T) {
	miniRedis := miniredis.RunT(t)
	viper.SetDefault("redis.url", "redis://"+miniRedis.Addr())
	redisCli, err := redis.ParseURL("redis://" + miniRedis.Addr())
	assert.NoError(t, err)
	type args struct {
		redisCli redis.UniversalClient
		prefix   string
	}
	tests := []struct {
		name      string
		args      args
		afterFunc func(t *testing.T, acache Cacher[string])
	}{
		{
			name: "normal",
			args: args{
				prefix:   "normal",
				redisCli: redis.NewClient(redisCli),
			},
			afterFunc: func(t *testing.T, acache Cacher[string]) {
				if acache == nil {
					t.Errorf("New() got nil")
					return
				}
				ctx := context.Background()
				value, err := acache.Get(ctx, "test")
				assert.NoError(t, err)
				assert.Equal(t, "test", value)
				value, err = acache.Get(ctx, "test")
				assert.NoError(t, err)
				assert.Equal(t, "test", value)
				err = acache.Del(ctx, "test")
				assert.NoError(t, err)
				err = acache.Set(ctx, "test", "test", time.Second*3)
				assert.NoError(t, err)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.afterFunc(t, New(tt.args.redisCli, tt.args.prefix, testFetcher1))
		})
	}
}
