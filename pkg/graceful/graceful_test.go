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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetCtx(t *testing.T) {
	ctx := GetCtx("test")
	assert.NotEqual(t, ctx, nil)
	name, ok := ctx.Value(ctxNameKey).(string)
	assert.True(t, ok)
	assert.Equal(t, name, "test")
}

func TestRunAtShutdown(t *testing.T) {
	runAtShutdown = []item{}

	RunAtShutdown("test", 1, func() {})
	assert.Equal(t, len(runAtShutdown), 1)
	RunAtShutdown("test", 2, nil)
	assert.Equal(t, len(runAtShutdown), 1)
}

func TestShutdown(t *testing.T) {
	runAtShutdown = []item{}

	orderArray := []int{}
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
	Shutdown()

	assert.Equal(t, []int{1, 2, 4, 5}, orderArray)

	/* TestShutdownTimeout*/
	runAtShutdown = []item{}
	orderArray = []int{}
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

	gracefulShutdownTimeout = time.Second * 3
	Shutdown()

	assert.Equal(t, []int{1, 2}, orderArray)

	/* TestShutdownPanic*/
	time.Sleep(3 * time.Second)
	runAtShutdown = []item{}
	orderArray = []int{}
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
	gracefulShutdownTimeout = time.Second * 30
	Shutdown()

	assert.Equal(t, []int{2, 4, 5}, orderArray)
}
