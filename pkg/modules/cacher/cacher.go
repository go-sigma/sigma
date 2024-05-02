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

package cacher

import (
	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/modules/cacher/badger"
	"github.com/go-sigma/sigma/pkg/modules/cacher/definition"
	"github.com/go-sigma/sigma/pkg/modules/cacher/inmemory"
	"github.com/go-sigma/sigma/pkg/modules/cacher/redis"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

// New ...
func New[T any](prefix string, fetcher definition.Fetcher[T]) (definition.Cacher[T], error) {
	var err error
	var cacher definition.Cacher[T]
	config := ptr.To(configs.GetConfiguration())
	switch config.Cache.Type {
	case enums.CacherTypeRedis:
		cacher, err = redis.New[T](config, prefix, fetcher)
	case enums.CacherTypeInmemory:
		cacher, err = inmemory.New[T](config, prefix, fetcher)
	case enums.CacherTypeBadger:
		cacher, err = badger.New[T](config, prefix, fetcher)
	default:
		cacher, err = badger.New[T](config, prefix, fetcher)
	}
	return cacher, err
}
