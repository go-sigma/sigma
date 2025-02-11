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
	"errors"
	"fmt"
	"net/http"
	"sort"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/consts"
)

var (
	// ErrNext continue next router
	ErrNext = fmt.Errorf("continue next router")
)

// All handles the all request
func All(c echo.Context, digCon *dig.Container) error {
	c.Response().Header().Set(consts.APIVersionKey, consts.APIVersionValue)

	for index, factory := range routerFactories {
		err := factory.Value.Initialize(c, digCon)
		if err != nil {
			if errors.Is(err, ErrNext) {
				continue
			}
			return fmt.Errorf("failed to initialize router factory index(%d): %v", index, err)
		}
		return nil
	}

	log.Error().Str("Uri", c.Request().RequestURI).Str("Method", c.Request().Method).Msg("Uri cannot match any route")
	return c.NoContent(http.StatusMethodNotAllowed)
}

// Factory is the interface for the storage router factory
type Factory interface {
	Initialize(echo.Context, *dig.Container) error
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
	sort.SliceStable(routerFactories, func(i, j int) bool {
		return routerFactories[i].Key < routerFactories[j].Key
	})
	return nil
}
