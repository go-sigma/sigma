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

package tests

import (
	"fmt"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/dal/redis"
	"github.com/go-sigma/sigma/pkg/modules/locker"
	"github.com/go-sigma/sigma/pkg/modules/locker/definition"
	"github.com/go-sigma/sigma/pkg/types/enums"
)

// ciDatabase is the interface for the database in ci tests
type ciDatabase interface {
	// Initialize initializes the database or database file for ci tests
	Initialize(*dig.Container) error
	// DeInitialize remove the database or database file for ci tests
	DeInitialize() error
	// GetName get database name
	GetName() enums.Database
}

type factory interface {
	New() ciDatabase
}

var ciDatabaseFactories = make(map[string]factory)

// registerCIDatabaseFactory registers a storage factory driver by name.
// If registerCIDatabaseFactory is called twice with the same name or if driver is nil, it panics.
func registerCIDatabaseFactory(name string, factory factory) error {
	if _, ok := ciDatabaseFactories[name]; ok {
		return fmt.Errorf("ci database %q already registered", name)
	}
	ciDatabaseFactories[name] = factory
	return nil
}

// Instance ...
type Instance struct {
	database ciDatabase
}

func Initialize(t *testing.T, digCon *dig.Container) (*Instance, error) {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	typ := viper.GetString("ci.database.type")
	if typ == "" {
		typ = enums.DatabaseSqlite3.String()
	}

	err := digCon.Provide(redis.New)
	if err != nil {
		return nil, fmt.Errorf("initialize redis failed: %v", err)
	}

	err = digCon.Provide(func() (definition.Locker, error) {
		return locker.Initialize(digCon)
	})
	if err != nil {
		return nil, fmt.Errorf("initialize locker failed: %v", err)
	}

	factory, ok := ciDatabaseFactories[typ]
	if !ok {
		return nil, fmt.Errorf("ci database %q not registered", typ)
	}

	database := factory.New()

	err = database.Initialize(digCon)
	if err != nil {
		return nil, fmt.Errorf("init ci database %q failed: %w", typ, err)
	}

	return &Instance{database: database}, nil
}

// DeInit ...
func (i *Instance) DeInitialize() error {
	return i.database.DeInitialize()
}
