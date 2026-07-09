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

package bootstrap

import (
	"sync"

	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/dal/redis"
	"github.com/go-sigma/sigma/pkg/infra/counter"
	"github.com/go-sigma/sigma/pkg/infra/lock"
	"github.com/go-sigma/sigma/pkg/infra/workq"
	"github.com/go-sigma/sigma/pkg/service/password"
	"github.com/go-sigma/sigma/pkg/service/token"
	"github.com/go-sigma/sigma/pkg/storage"
)

var digCon = dig.New()
var digConOnce = sync.Once{}

// NewDigContainer ...
func NewDigContainer() (*dig.Container, error) {
	var err error
	digConOnce.Do(func() {
		for _, e := range []error{
			digCon.Provide(config.GetConfig),       // init config
			digCon.Provide(redis.NewClientFactory), // init redis
			digCon.Provide(password.New),           // init password
			digCon.Provide(token.New),              // init token
			digCon.Provide(lock.Initialize),        // init locker
			digCon.Provide(counter.Initialize),     // init counter
			digCon.Provide(workq.InitProducer),
			digCon.Provide(workq.NewHandlerRegistry),
			digCon.Provide(storage.Initialize),
		} {
			if e != nil {
				err = e
				return
			}
		}
	})
	return digCon, err
}
