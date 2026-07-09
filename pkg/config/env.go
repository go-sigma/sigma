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

package config

import "github.com/caarlos0/env/v9"

type configurationEnv struct {
	Database configurationDatabaseEnv `envPrefix:"DATABASE_"`
	Redis    configurationRedisEnv    `envPrefix:"REDIS_"`
	Storage  configurationStorageEnv  `envPrefix:"STORAGE_"`
}

type configurationDatabaseEnv struct {
	Mysql      configurationDatabasePasswordEnv `envPrefix:"MYSQL_"`
	PostgreSQL configurationDatabasePasswordEnv `envPrefix:"POSTGRESQL_"`
}

type configurationDatabasePasswordEnv struct {
	Password *string `env:"PASSWORD"`
}

type configurationRedisEnv struct {
	URL      *string `env:"URL"`
	Username *string `env:"USERNAME"`
	Password *string `env:"PASSWORD"`
}

type configurationStorageEnv struct {
	S3  configurationStorageAccessKeyEnv `envPrefix:"S3_"`
	Cos configurationStorageAccessKeyEnv `envPrefix:"COS_"`
	Oss configurationStorageAccessKeyEnv `envPrefix:"OSS_"`
}

type configurationStorageAccessKeyEnv struct {
	AK *string `env:"AK"`
	SK *string `env:"SK"`
}

// ApplyEnvOverrides applies environment variable overrides after YAML loading.
func (c *Configuration) ApplyEnvOverrides() error {
	var cfg configurationEnv
	if err := env.ParseWithOptions(&cfg, env.Options{Prefix: "SIGMA_"}); err != nil {
		return err
	}
	if cfg.Database.Mysql.Password != nil {
		c.Database.Mysql.Password = *cfg.Database.Mysql.Password
	}
	if cfg.Database.PostgreSQL.Password != nil {
		c.Database.Postgresql.Password = *cfg.Database.PostgreSQL.Password
	}
	if cfg.Redis.URL != nil {
		c.Redis.URL = *cfg.Redis.URL
	}
	if cfg.Redis.Username != nil {
		c.Redis.Username = *cfg.Redis.Username
	}
	if cfg.Redis.Password != nil {
		c.Redis.Password = *cfg.Redis.Password
	}
	if cfg.Storage.S3.AK != nil {
		c.Storage.S3.Ak = *cfg.Storage.S3.AK
	}
	if cfg.Storage.S3.SK != nil {
		c.Storage.S3.Sk = *cfg.Storage.S3.SK
	}
	if cfg.Storage.Cos.AK != nil {
		c.Storage.Cos.Ak = *cfg.Storage.Cos.AK
	}
	if cfg.Storage.Cos.SK != nil {
		c.Storage.Cos.Sk = *cfg.Storage.Cos.SK
	}
	if cfg.Storage.Oss.AK != nil {
		c.Storage.Oss.Ak = *cfg.Storage.Oss.AK
	}
	if cfg.Storage.Oss.SK != nil {
		c.Storage.Oss.Sk = *cfg.Storage.Oss.SK
	}
	return nil
}
