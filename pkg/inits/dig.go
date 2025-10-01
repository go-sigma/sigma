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
	"sync"

	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/dal/badger"
	"github.com/go-sigma/sigma/pkg/dal/redis"
	"github.com/go-sigma/sigma/pkg/modules/locker"
	"github.com/go-sigma/sigma/pkg/utils/password"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
	"github.com/go-sigma/sigma/pkg/utils/token"
)

var digCon = dig.New()
var digConOnce = sync.Once{}

// NewDigContainer ...
func NewDigContainer() (*dig.Container, error) {
	var err error
	digConOnce.Do(func() {
		for _, e := range []error{
			digCon.Provide(func() configs.Configuration { return ptr.To(configs.GetConfig()) }), // init config
			digCon.Provide(redis.New),    // init redis
			digCon.Provide(badger.New),   // init badger
			digCon.Provide(password.New), // init password
			digCon.Provide(func() (token.Service, error) { return token.New(digCon) }),         // init token
			digCon.Provide(func() (locker.Locker, error) { return locker.Initialize(digCon) }), // init locker
		} {
			if e != nil {
				err = e
				return
			}
		}
	})
	return digCon, err
}
