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

package locker

import (
	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/modules/locker/badger"
	"github.com/go-sigma/sigma/pkg/modules/locker/definition"
	"github.com/go-sigma/sigma/pkg/modules/locker/redis"
	"github.com/go-sigma/sigma/pkg/types/enums"
)

// Locker ...
var Locker definition.Locker

// New ...
func Initialize(config configs.Configuration) error {
	var err error
	switch config.Locker.Type {
	case enums.LockerTypeBadger:
		Locker, err = badger.New(config)
	case enums.LockerTypeRedis:
		Locker, err = redis.New(config)
	default:
		Locker, err = badger.New(config)
	}
	return err
}
