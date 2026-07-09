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
	"cmp"
	"context"
	"log/slog"
	"slices"
	"sync"
	"time"
)

const (
	ctxNameKey = "name"
)

var (
	mu             sync.Mutex
	runAtShutdown  []item
	defaultTimeout = 30 * time.Second
	ctx, ctxCancel = context.WithCancel(context.Background())
	shutdownOnce   sync.Once
	shutdownDone   = make(chan struct{})
)

type item struct {
	name  string
	index int
	f     func()
}

// GetCtx ...
func GetCtx(name string) context.Context {
	return context.WithValue(ctx, ctxNameKey, name) // nolint: staticcheck
}

// RunAtShutdown ...
func RunAtShutdown(name string, index int, f func()) {
	if f == nil {
		return
	}
	mu.Lock()
	runAtShutdown = append(runAtShutdown, item{
		name:  name,
		index: index,
		f:     f,
	})
	mu.Unlock()
}

// Shutdown runs all registered shutdown hooks in ascending index order.
// It is idempotent: repeated calls are safe no-ops. Returns
// context.DeadlineExceeded when the shutdown budget is exceeded before all
// hooks finish.
//
// The caller-supplied ctx is the single shutdown budget: when it has a
// deadline, that deadline governs all hooks; when it does not, a default
// 30s budget is derived so passing context.Background() still works.
func Shutdown(parent context.Context) error {
	var timedOut bool
	shutdownOnce.Do(func() {
		ctxCancel()

		mu.Lock()
		snapshot := slices.Clone(runAtShutdown)
		mu.Unlock()

		slices.SortStableFunc(snapshot, func(a, b item) int {
			return cmp.Compare(a.index, b.index)
		})

		gCtx := parent
		if _, ok := parent.Deadline(); !ok {
			var gCtxCancel context.CancelFunc
			gCtx, gCtxCancel = context.WithTimeout(parent, defaultTimeout)
			defer gCtxCancel()
		}

		waitGroupDone := make(chan struct{})
		go func() {
			for _, it := range snapshot {
				select {
				case <-gCtx.Done():
					return
				default:
				}
				func(name string, f func()) {
					defer func() {
						if r := recover(); r != nil {
							slog.Error("Panic during shutting down", "name", name, "err", r)
						}
					}()
					slog.Info("shutting down", "name", name)
					f()
				}(it.name, it.f)
			}
			waitGroupDone <- struct{}{}
		}()

		select {
		case <-gCtx.Done():
			timedOut = true
			slog.Error("timeout shutting down")
		case <-waitGroupDone:
		}
		close(shutdownDone)
	})
	if timedOut {
		return context.DeadlineExceeded
	}
	return nil
}

// resetForTest resets all package-level state. Only for use in tests.
func resetForTest() {
	mu.Lock()
	defer mu.Unlock()
	runAtShutdown = nil
	shutdownOnce = sync.Once{}
	shutdownDone = make(chan struct{})
	ctx, ctxCancel = context.WithCancel(context.Background())
}
