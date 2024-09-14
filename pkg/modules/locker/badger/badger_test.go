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

package badger_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/go-sigma/sigma/pkg/configs"
	rBadger "github.com/go-sigma/sigma/pkg/dal/badger"
	"github.com/go-sigma/sigma/pkg/logger"
	"github.com/go-sigma/sigma/pkg/modules/locker/badger"
)

func TestDatabaseAcquire(t *testing.T) {
	logger.SetLevel("debug")

	p, _ := os.MkdirTemp("", "badger")
	config := configs.Configuration{
		Badger: configs.ConfigurationBadger{
			Path: p,
		},
	}
	defer os.RemoveAll(p) // nolint: errcheck

	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	assert.NoError(t, rBadger.Initialize(ctx, config))

	c, err := badger.New(config)
	assert.NoError(t, err)

	const key = "test-redis-lock"
	var concurrency uint64 = 500

	var wg sync.WaitGroup
	var cnt uint64 = 0
	for i := uint64(0); i < concurrency; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()
			l, err := c.Acquire(ctx, key, time.Second*1, time.Second*3)
			assert.Equal(t, true, err == nil || errors.Is(err, context.DeadlineExceeded))
			if !(err == nil || errors.Is(err, context.DeadlineExceeded)) {
				fmt.Println(err)
			}
			if l != nil {
				<-time.After(time.Millisecond * 100)
				defer l.Unlock(ctx) // nolint: errcheck
				atomic.AddUint64(&cnt, 1)
			}
		}()
	}
	wg.Wait()
	assert.True(t, true, concurrency > cnt && cnt > 1)
}

func TestDatabaseAcquireWithRenew(t *testing.T) {
	logger.SetLevel("debug")

	p, _ := os.MkdirTemp("", "badger")
	config := configs.Configuration{
		Badger: configs.ConfigurationBadger{
			Path: p,
		},
	}
	defer os.RemoveAll(p) // nolint: errcheck

	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	assert.NoError(t, rBadger.Initialize(ctx, config))

	c, err := badger.New(config)
	assert.NoError(t, err)

	const key = "test-redis-lock"
	var concurrency uint64 = 500

	var wg sync.WaitGroup
	var cnt uint64 = 0
	for i := uint64(0); i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := c.AcquireWithRenew(ctx, key, time.Second*1, time.Second*3)
			if errors.Is(err, context.DeadlineExceeded) {
				atomic.AddUint64(&cnt, 1)
			}
		}()
	}
	wg.Wait()
	assert.Equal(t, true, cnt >= 1)
}
