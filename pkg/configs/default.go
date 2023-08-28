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
)

func defaultSettings() {
	viper.SetDefault("auth.jwt.type", "RS256")            // the jwt token type
	viper.SetDefault("auth.jwt.ttl", time.Hour)           // the jwt token ttl
	viper.SetDefault("auth.jwt.refreshTtl", time.Hour*24) // the refresh token ttl

	viper.SetDefault("storage.rootDirectory", "/var/lib/sigma") // the root directory for filesystem storage

	viper.SetDefault("server.endpoint", "http://127.0.0.1:3000")
	viper.SetDefault("server.internalEndpoint", "http://127.0.0.1:3000")

	configuration.HTTP.Endpoint = "http://127.0.0.1:3000"
	configuration.HTTP.InternalEndpoint = "http://127.0.0.1:3000"
}
