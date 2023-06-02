// Copyright 2023 XImager
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
	"fmt"
	"net/http"
	"sort"

	"github.com/labstack/echo/v4"

	"github.com/ximager/ximager/pkg/consts"
)

// All handles the all request
func All(c echo.Context) error {
	c.Response().Header().Set(consts.APIVersionKey, consts.APIVersionValue)

	sort.SliceStable(routerFactories, func(i, j int) bool {
		return routerFactories[i].Key < routerFactories[j].Key
	})

	for index, factory := range routerFactories {
		if err := factory.Value.Initialize(c); err != nil {
			return fmt.Errorf("failed to initialize router factory index(%d): %v", index, err)
		}
	}

	return c.NoContent(http.StatusOK)
}

// Factory is the interface for the storage router factory
type Factory interface {
	Initialize(ctx echo.Context) error
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
	return nil
}
