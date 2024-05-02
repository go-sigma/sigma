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

package inmemory

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/go-sigma/sigma/pkg/configs"
)

func fetcher1(key string) (string, error) {
	return "new-val", nil
}

func TestNew(t *testing.T) {
	config := configs.Configuration{
		Cache: configs.ConfigurationCache{
			Inmemory: configs.ConfigurationCacheInmemory{
				Prefix: "sigma-cache",
				Size:   1000,
			},
		},
	}
	cacher, err := New(config, "test", fetcher1)
	assert.NoError(t, err)
	assert.NotNil(t, cacher)

	ctx := context.Background()
	err = cacher.Set(ctx, "test", "test")
	assert.NoError(t, err)

	res, err := cacher.Get(ctx, "test")
	assert.NoError(t, err)
	assert.Equal(t, "test", res)

	err = cacher.Del(ctx, "test")
	assert.NoError(t, err)

	res, err = cacher.Get(ctx, "test")
	assert.NoError(t, err)
	assert.Equal(t, "new-val", res)

	err = cacher.Set(ctx, "m-test", "m-val")
	assert.NoError(t, err)

	for i := 0; i < 1024; i++ {
		err = cacher.Set(ctx, fmt.Sprintf("key-%d", i), "val")
		assert.NoError(t, err)
	}

	res, err = cacher.Get(ctx, "m-test")
	assert.NoError(t, err)
	assert.Equal(t, "new-val", res)
}
