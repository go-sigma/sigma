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
	"github.com/go-sigma/sigma/pkg/modules/locker/database"
	"github.com/go-sigma/sigma/pkg/modules/locker/definition"
	"github.com/go-sigma/sigma/pkg/modules/locker/redis"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

// New ...
func New() (definition.Locker, error) {
	var err error
	var lock definition.Locker
	config := ptr.To(configs.GetConfiguration())
	switch config.Locker.Type {
	case enums.LockerTypeDatabase:
		lock, err = database.New(config)
	case enums.LockerTypeRedis:
		lock, err = redis.New(config)
	default:
		lock, err = database.New(config)
		// return nil, fmt.Errorf("Locker %s not support", config.Cache.Type)
	}
	return lock, err
}
