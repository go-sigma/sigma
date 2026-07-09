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

package graceful

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/go-sigma/sigma/pkg/logger"
)

func TestGetCtx(t *testing.T) {
	resetForTest()
	ctx := GetCtx("test")
	require.NotNil(t, ctx)
	name, ok := ctx.Value(ctxNameKey).(string)
	require.True(t, ok)
	require.Equal(t, "test", name)
}

func TestRunAtShutdown(t *testing.T) {
	resetForTest()
	RunAtShutdown("test", 1, func() {})
	require.Len(t, runAtShutdown, 1)
	RunAtShutdown("test", 2, nil)
	require.Len(t, runAtShutdown, 1)
}

func TestShutdown(t *testing.T) {
	logger.SetLevel("debug")

	t.Run("ordered", func(t *testing.T) {
		resetForTest()
		var orderArray []int
		RunAtShutdown("test1", 1, func() {
			time.Sleep(time.Second)
			orderArray = append(orderArray, 1)
		})
		RunAtShutdown("test2", 2, func() {
			time.Sleep(time.Second)
			orderArray = append(orderArray, 2)
		})
		RunAtShutdown("test5", 5, func() {
			time.Sleep(time.Second)
			orderArray = append(orderArray, 5)
		})
		RunAtShutdown("test4", 4, func() {
			time.Sleep(time.Second)
			orderArray = append(orderArray, 4)
		})
		err := Shutdown(context.Background())
		require.NoError(t, err)
		require.Equal(t, []int{1, 2, 4, 5}, orderArray)
	})

	t.Run("timeout", func(t *testing.T) {
		resetForTest()
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()

		var ran atomic.Int32
		RunAtShutdown("block", 1, func() {
			time.Sleep(time.Second) // exceeds the 500ms budget
			ran.Add(1)
		})
		RunAtShutdown("never", 2, func() {
			ran.Add(1)
		})
		err := Shutdown(ctx)
		require.ErrorIs(t, err, context.DeadlineExceeded)
		// Wait for the abandoned goroutine to drain
		time.Sleep(700 * time.Millisecond)
		require.Equal(t, int32(1), ran.Load(), "only the first hook should have run")
	})

	t.Run("panic_recovery", func(t *testing.T) {
		resetForTest()
		var orderArray []int
		RunAtShutdown("test1", 1, func() {
			panic("test panic")
		})
		RunAtShutdown("test2", 2, func() {
			time.Sleep(time.Second)
			orderArray = append(orderArray, 2)
		})
		RunAtShutdown("test5", 5, func() {
			time.Sleep(time.Second)
			orderArray = append(orderArray, 5)
		})
		RunAtShutdown("test4", 4, func() {
			time.Sleep(time.Second)
			orderArray = append(orderArray, 4)
		})
		err := Shutdown(context.Background())
		require.NoError(t, err)
		require.Equal(t, []int{2, 4, 5}, orderArray)
	})

	t.Run("idempotent", func(t *testing.T) {
		resetForTest()
		var count int
		RunAtShutdown("test", 1, func() {
			count++
		})
		err := Shutdown(context.Background())
		require.NoError(t, err)
		err = Shutdown(context.Background())
		require.NoError(t, err)
		require.Equal(t, 1, count)
	})
}
