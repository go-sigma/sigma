// Copyright 2024 sigma
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

package inits

import (
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/dal/badger"
	"github.com/go-sigma/sigma/pkg/dal/redis"
	"github.com/go-sigma/sigma/pkg/modules/locker"
	"github.com/go-sigma/sigma/pkg/modules/locker/definition"
	"github.com/go-sigma/sigma/pkg/utils/password"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
	"github.com/go-sigma/sigma/pkg/utils/token"
)

// DigCon is the dependency injection container
var DigCon = dig.New()

// NewDigContainer ...
func NewDigContainer() error {
	for _, e := range []error{
		// init config
		DigCon.Provide(func() configs.Configuration {
			return ptr.To(configs.GetConfiguration())
		}),
		// init redis
		DigCon.Provide(redis.New),
		// init badger
		DigCon.Provide(badger.New),
		// init password
		DigCon.Provide(password.New),
		// init token
		DigCon.Provide(func() (token.Service, error) {
			return token.New(DigCon)
		}),
		// init locker
		DigCon.Provide(func() (definition.Locker, error) {
			return locker.Initialize(DigCon)
		}),
	} {
		if e != nil {
			return e
		}
	}
	return nil
}
