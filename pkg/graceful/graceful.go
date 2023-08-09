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
	"sort"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

const (
	ctxNameKey = "name"
)

var (
	gracefulShutdownTimeout = time.Second * 30
	ctx, ctxCancel          = context.WithCancel(context.Background())
	runAtShutdown           []item
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
	runAtShutdown = append(runAtShutdown, item{
		name:  name,
		index: index,
		f:     f,
	})
}

// Shutdown ...
func Shutdown() {
	ctxCancel()

	sort.SliceStable(runAtShutdown, func(i, j int) bool {
		return runAtShutdown[i].index < runAtShutdown[j].index
	})

	ctx, ctxCancel := context.WithTimeout(context.Background(), gracefulShutdownTimeout)

	waitGroupDone := make(chan struct{})
	go func() {
		for _, item := range runAtShutdown {
			select {
			case <-ctx.Done():
				return
			default:
			}
			waitGroup := &sync.WaitGroup{}
			waitGroup.Add(1)
			go func(name string, f func()) {
				defer func() {
					waitGroup.Done()
					err := recover()
					if err != nil {
						log.Error().Msgf("Panic during shuting down %s", name)
					}
				}()
				log.Info().Str("name", name).Msg("Shuting down")
				f()
			}(item.name, item.f)
			waitGroup.Wait()
		}
		waitGroupDone <- struct{}{}
	}()

	select {
	case <-ctx.Done():
		log.Error().Msg("Timeout shuting down")
	case <-waitGroupDone:
	}
	ctxCancel()
}
