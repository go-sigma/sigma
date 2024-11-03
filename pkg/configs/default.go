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

package configs

import (
	"strings"
	"time"

	"github.com/go-sigma/sigma/pkg/types/enums"
)

func defaultSettings() {
	if configuration.HTTP.Endpoint == "" {
		configuration.HTTP.Endpoint = "http://127.0.0.1:3000"
	}
	if configuration.HTTP.InternalEndpoint == "" {
		configuration.HTTP.InternalEndpoint = "http://127.0.0.1:3000"
	}
	if configuration.Auth.Jwt.Ttl == 0 {
		configuration.Auth.Jwt.Ttl = time.Hour
	}
	if configuration.Auth.Jwt.RefreshTtl == 0 {
		configuration.Auth.Jwt.RefreshTtl = time.Hour * 24
	}
	if configuration.Namespace.Visibility.String() == "" {
		configuration.Namespace.Visibility = enums.VisibilityPrivate
	}
	if configuration.Daemon.Builder.Kubernetes.Namespace == "" {
		configuration.Daemon.Builder.Kubernetes.Namespace = "default"
	}
	if configuration.Daemon.Builder.Podman.URI == "" {
		configuration.Daemon.Builder.Podman.URI = "unix:///run/podman/podman.sock"
	}
	if configuration.WorkQueue.Inmemory.Concurrency == 0 {
		configuration.WorkQueue.Inmemory.Concurrency = 1024
	}

	// for cache
	if configuration.Cache.Type == enums.CacherTypeInmemory && configuration.Cache.Inmemory.Size == 0 {
		configuration.Cache.Inmemory.Size = 10240
	}
	if len(strings.TrimSpace(configuration.Cache.Prefix)) == 0 {
		configuration.Cache.Prefix = "sigma-cache"
	}
	if configuration.Cache.Type == enums.CacherTypeRedis && configuration.Cache.Redis.Ttl == 0 {
		configuration.Cache.Redis.Ttl = time.Hour * 72
	}
	if configuration.Cache.Type == enums.CacherTypeBadger && configuration.Cache.Badger.Ttl == 0 {
		configuration.Cache.Badger.Ttl = time.Hour * 72
	}

	// for badger
	if (configuration.Cache.Type == enums.CacherTypeBadger || configuration.Locker.Type == enums.LockerTypeBadger) && !configuration.Badger.Enabled {
		configuration.Badger.Enabled = true
	}
	if configuration.Badger.Enabled && len(strings.TrimSpace(configuration.Badger.Path)) == 0 {
		configuration.Badger.Path = "/var/lib/sigma/badger/"
	}

	// for locker
	if strings.TrimSpace(configuration.Locker.Prefix) == "" {
		configuration.Locker.Prefix = "sigma-locker"
	}
}
