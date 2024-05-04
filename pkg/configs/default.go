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

	"github.com/spf13/viper"

	"github.com/go-sigma/sigma/pkg/types/enums"
)

func defaultSettings() {
	viper.SetDefault("auth.jwt.type", "RS256")            // the jwt token type
	viper.SetDefault("auth.jwt.ttl", time.Hour)           // the jwt token ttl
	viper.SetDefault("auth.jwt.refreshTtl", time.Hour*24) // the refresh token ttl

	viper.SetDefault("server.endpoint", "http://127.0.0.1:3000")
	viper.SetDefault("server.internalEndpoint", "http://127.0.0.1:3000")

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
	if configuration.Cache.Type == enums.CacherTypeInmemory && len(strings.TrimSpace(configuration.Cache.Inmemory.Prefix)) == 0 {
		configuration.Cache.Inmemory.Prefix = "sigma-cache"
	}
	if configuration.Cache.Type == enums.CacherTypeRedis && configuration.Cache.Redis.Ttl == 0 {
		configuration.Cache.Redis.Ttl = time.Hour * 72
	}
	if configuration.Cache.Type == enums.CacherTypeRedis && len(strings.TrimSpace(configuration.Cache.Redis.Prefix)) == 0 {
		configuration.Cache.Redis.Prefix = "sigma-cache"
	}
	if configuration.Cache.Type == enums.CacherTypeBadger && configuration.Cache.Badger.Ttl == 0 {
		configuration.Cache.Badger.Ttl = time.Hour * 72
	}
	if configuration.Cache.Type == enums.CacherTypeBadger && len(strings.TrimSpace(configuration.Cache.Badger.Prefix)) == 0 {
		configuration.Cache.Badger.Prefix = "sigma-cache"
	}

	// for badger
	if configuration.Badger.Enabled && len(strings.TrimSpace(configuration.Badger.Path)) == 0 {
		configuration.Badger.Path = "/var/lib/sigma/badger/"
	}

	// for locker
	if configuration.Locker.Type == enums.LockerTypeBadger && strings.TrimSpace(configuration.Locker.Badger.Prefix) == "" {
		configuration.Locker.Badger.Prefix = "sigma-locker"
	}
	if configuration.Locker.Type == enums.LockerTypeRedis && strings.TrimSpace(configuration.Locker.Redis.Prefix) == "" {
		configuration.Locker.Redis.Prefix = "sigma-locker"
	}
}
