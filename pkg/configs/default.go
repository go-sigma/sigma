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
	"time"

	"github.com/spf13/viper"

	"github.com/go-sigma/sigma/pkg/types/enums"
)

func defaultSettings() {
	viper.SetDefault("auth.jwt.type", "RS256")            // the jwt token type
	viper.SetDefault("auth.jwt.ttl", time.Hour)           // the jwt token ttl
	viper.SetDefault("auth.jwt.refreshTtl", time.Hour*24) // the refresh token ttl

	viper.SetDefault("storage.rootDirectory", "/var/lib/sigma") // the root directory for filesystem storage

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
}
