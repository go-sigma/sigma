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

package distribution

import (
	"cmp"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/consts"
)

var (
	// ErrNext continue next router
	ErrNext = fmt.Errorf("continue next router")
)

// All handles the all request. The factories must have been prepared with Prepare first, so no
// dig container is touched while serving a request.
func All(c *gin.Context) error {
	c.Writer.Header().Set(consts.APIVersionKey, consts.APIVersionValue)

	for index, factory := range routerFactories {
		err := factory.Value.Initialize(c)
		if err != nil {
			if errors.Is(err, ErrNext) {
				continue
			}
			return fmt.Errorf("failed to initialize router factory index(%d): %v", index, err)
		}
		return nil
	}

	slog.Error("uri cannot match any route", "Uri", c.Request.RequestURI, "Method", c.Request.Method)
	c.AbortWithStatus(http.StatusMethodNotAllowed)
	return nil
}

// Prepare resolves every registered factory from the container. It runs once during startup:
// dig containers are not safe for concurrent use, so request handling must not invoke them.
func Prepare(digCon *dig.Container) error {
	for index, factory := range routerFactories {
		if err := factory.Value.Prepare(digCon); err != nil {
			return fmt.Errorf("failed to prepare router factory index(%d): %v", index, err)
		}
	}
	return nil
}

// Factory is the interface for the storage router factory
//
// Prepare resolves the handler dependencies once at startup and Initialize dispatches a request
// to that prepared handler, so the dig container is never touched from a request goroutine.
type Factory interface {
	// Prepare resolves the factory handler from the container.
	Prepare(digCon *dig.Container) error
	// Initialize dispatches the request to the prepared handler.
	Initialize(c *gin.Context) error
}

type Item struct {
	Key   int
	Value Factory
}

var routerFactories = make([]Item, 0, 10)

// RegisterRouterFactory registers a new router factory
func RegisterRouterFactory(factory Factory, index int) error {
	for _, router := range routerFactories {
		if router.Key == index {
			return fmt.Errorf("router %d already registered", index)
		}
	}
	routerFactories = append(routerFactories, Item{Key: index, Value: factory})
	slices.SortStableFunc(routerFactories, func(a, b Item) int {
		return cmp.Compare(a.Key, b.Key)
	})
	return nil
}
