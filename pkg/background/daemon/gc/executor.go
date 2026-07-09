// Copyright 2026 sigma
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

package gc

import (
	"context"
	"sync"
)

func runWorkers[T any](ctx context.Context, workers int, items []T, fn func(context.Context, T) error) error {
	if workers <= 0 {
		workers = 1
	}
	if workers > len(items) {
		workers = len(items)
	}
	if workers == 0 {
		return nil
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	jobs := make(chan T)
	var wait sync.WaitGroup
	var errOnce sync.Once
	var firstErr error

	for range workers {
		wait.Go(func() {
			for item := range jobs {
				if ctx.Err() != nil {
					return
				}
				if err := fn(ctx, item); err != nil {
					errOnce.Do(func() {
						firstErr = err
						cancel()
					})
					return
				}
			}
		})
	}

	for _, item := range items {
		select {
		case <-ctx.Done():
			close(jobs)
			wait.Wait()
			return firstErr
		case jobs <- item:
		}
	}
	close(jobs)
	wait.Wait()
	return firstErr
}
