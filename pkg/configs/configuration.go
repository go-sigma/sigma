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

import "github.com/go-sigma/sigma/pkg/types/enums"

// ConfigurationLog ...
type ConfigurationLog struct {
	Level      enums.LogLevel
	ProxyLevel enums.LogLevel
}

// ConfigurationDatabaseSqlite3 ...
type ConfigurationDatabaseSqlite3 struct {
	Path string
}

// ConfigurationDatabaseMysql ...
type ConfigurationDatabaseMysql struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

// ConfigurationDatabase ...
type ConfigurationDatabasePostgresql struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SslMode  bool
}

// ConfigurationDatabase ...
type ConfigurationDatabase struct {
	Type       enums.Database
	Sqlite3    ConfigurationDatabaseSqlite3
	Mysql      ConfigurationDatabaseMysql
	Postgresql ConfigurationDatabasePostgresql
}

// ConfigurationRedis ...
type ConfigurationRedis struct {
	Type enums.RedisType
	Url  string
}

// ConfigurationCache ...
type ConfigurationCache struct {
	Type enums.CacheType
}

// ConfigurationWorkQueue ...
type ConfigurationWorkQueue struct {
	Type enums.WorkQueueType
}

// ConfigurationNamespace ...
type ConfigurationNamespace struct {
	AutoCreate bool
	Visibility enums.Visibility
}

// Configuration ...
type Configuration struct {
	Log       ConfigurationLog
	Database  ConfigurationDatabase
	Deploy    enums.Deploy
	Redis     ConfigurationRedis
	Cache     ConfigurationCache
	WorkQueue ConfigurationWorkQueue
	Namespace ConfigurationNamespace
}
