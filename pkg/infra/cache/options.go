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

package cacher

import (
	"errors"
	"time"

	"github.com/go-sigma/sigma/pkg/api/enums"
)

// Options controls cache read-through behavior.
type Options struct {
	TTL         time.Duration
	NegativeTTL time.Duration
	IsNotFound  func(error) bool
}

func (o Options) withDefaults(params Params) Options {
	if o.TTL == 0 {
		o.TTL = params.Config.Cache.TTL
		if params.Config.Cache.Type == enums.CacherTypeRedis && params.Config.Cache.Redis.Ttl != 0 {
			o.TTL = params.Config.Cache.Redis.Ttl
		}
	}
	if o.NegativeTTL == 0 {
		o.NegativeTTL = params.Config.Cache.NegativeTTL
	}
	if o.IsNotFound == nil {
		o.IsNotFound = func(err error) bool {
			return errors.Is(err, ErrNotFound)
		}
	}
	return o
}
