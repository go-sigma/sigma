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

package counter

import (
	"context"
	"fmt"

	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/config"
	dalredis "github.com/go-sigma/sigma/pkg/dal/redis"
	"github.com/go-sigma/sigma/pkg/infra/registry"
)

//go:generate go tool mockgen -destination=counter_mocks.go -package=counter github.com/go-sigma/sigma/pkg/infra/counter Counter,Factory

// Counter provides hash and set operations backed by in-memory or Redis storage.
type Counter interface {
	// HIncrBy increments hash fields by the given deltas.
	HIncrBy(ctx context.Context, key string, fields map[string]int64) error
	// HGetAll returns all fields and values of a hash.
	HGetAll(ctx context.Context, key string) (map[string]int64, error)
	// Del removes one or more keys.
	Del(ctx context.Context, keys ...string) error
	// SAdd adds members to a set.
	SAdd(ctx context.Context, key string, members ...string) error
	// SMembers returns all members of a set.
	SMembers(ctx context.Context, key string) ([]string, error)
	// SRem removes members from a set.
	SRem(ctx context.Context, key string, members ...string) error
	// Rename atomically renames a key. Returns nil if the target already exists.
	Rename(ctx context.Context, oldKey, newKey string) error
	// Exists returns the number of keys that exist among the given keys.
	Exists(ctx context.Context, keys ...string) (int64, error)
	// Shutdown gracefully shuts down the counter.
	Shutdown(ctx context.Context) error
}

// Factory creates a Counter instance.
type Factory interface {
	New(params Params) (Counter, error)
}

// Params declares dependencies needed to construct a counter.
type Params struct {
	dig.In

	Config             *config.Configuration
	RedisClientFactory dalredis.ClientFactory `optional:"true"`
}

// Factories holds registered counter factories keyed by backend name.
var Factories = make(registry.Factories[string, Factory])

// Initialize creates a Counter based on the configured analytics counter backend.
func Initialize(params Params) (Counter, error) {
	backend := params.Config.Analytics.CounterBackend
	if backend == "" {
		backend = "inmemory"
	}
	factory, ok := Factories[backend]
	if !ok {
		return nil, fmt.Errorf("counter backend %q not registered", backend)
	}
	return factory.New(params)
}
