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

package timewheel

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewTimeWheel(t *testing.T) {
	var myNum int32 = 1
	tw := NewTimeWheel(context.Background())
	tw.AddRunner(func(_ context.Context, _ TimeWheel) {
		atomic.AddInt32(&myNum, 1)
	})
	tw.TickNext(time.Second)
	time.Sleep(time.Second * 2)
	tw.Stop()
	assert.Equal(t, int32(2), atomic.LoadInt32(&myNum))
}

func TestNewTimeWheelWithMaxTicker(t *testing.T) {
	var myNum int32 = 1
	tw := NewTimeWheel(context.Background(), time.Second*2)
	tw.AddRunner(func(_ context.Context, _ TimeWheel) {
		atomic.AddInt32(&myNum, 1)
	})
	tw.TickNext(time.Second)
	time.Sleep(time.Second * 3)
	tw.Stop()
	assert.Equal(t, int32(3), atomic.LoadInt32(&myNum))
}
