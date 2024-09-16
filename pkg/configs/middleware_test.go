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
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"

	"github.com/go-sigma/sigma/pkg/types/enums"
)

func TestCheckRedis(t *testing.T) {
	err := checkRedis(Configuration{Redis: ConfigurationRedis{Type: enums.RedisTypeNone}})
	assert.NoError(t, err)

	err = checkRedis(Configuration{Redis: ConfigurationRedis{Type: enums.RedisType("invalid")}})
	assert.Error(t, err)

	miniRedis := miniredis.RunT(t)
	err = checkRedis(Configuration{Redis: ConfigurationRedis{Type: enums.RedisTypeExternal, Url: "redis://" + miniRedis.Addr()}})
	assert.NoError(t, err)

	err = checkRedis(Configuration{Redis: ConfigurationRedis{Type: enums.RedisTypeExternal, Url: "redis://127.0.0.1:1100"}})
	assert.Error(t, err)
}

func TestCheckDatabase(t *testing.T) {
	err := checkDatabase(Configuration{Database: ConfigurationDatabase{Type: enums.DatabaseSqlite3}})
	assert.NoError(t, err)

	err = checkDatabase(Configuration{Database: ConfigurationDatabase{
		Type: enums.DatabaseMysql,
		Mysql: ConfigurationDatabaseMysql{
			Host:     "127.0.0.1",
			Port:     3306,
			Username: "sigma",
			Password: "sigma",
			Database: "sigma",
		},
	}})
	assert.NoError(t, err)

	err = checkDatabase(Configuration{Database: ConfigurationDatabase{Type: enums.Database("invalid")}})
	assert.Error(t, err)
}

func TestCheckMysql(t *testing.T) {
	var config = Configuration{
		Database: ConfigurationDatabase{
			Type: enums.DatabaseMysql,
			Mysql: ConfigurationDatabaseMysql{
				Host:     "127.0.0.1",
				Port:     3306,
				Username: "sigma",
				Password: "sigma",
				Database: "sigma",
			},
		},
	}

	err := checkMysql(config)
	assert.NoError(t, err)

	config.Database.Mysql.Port = 3310

	err = checkMysql(config)
	assert.Error(t, err)
}

func TestCheckPostgresql(t *testing.T) {
	var config = Configuration{
		Database: ConfigurationDatabase{
			Type: enums.DatabasePostgresql,
			Postgresql: ConfigurationDatabasePostgresql{
				Host:     "localhost",
				Port:     5432,
				Username: "sigma",
				Password: "sigma",
				Database: "sigma",
			},
		},
	}

	err := checkPostgresql(config)
	assert.NoError(t, err)

	config.Database.Postgresql.Port = 5433

	err = checkPostgresql(config)
	assert.Error(t, err)
}

func TestCheckS3(t *testing.T) {
	config := Configuration{
		Storage: ConfigurationStorage{
			Type: "s3",
			S3: ConfigurationStorageS3{
				Endpoint:       "http://127.0.0.1:9000",
				Region:         "cn-north-1",
				Ak:             "sigma",
				Sk:             "sigma-sigma",
				Bucket:         "sigma",
				ForcePathStyle: true,
			},
		},
	}
	err := checkStorage(config)
	assert.NoError(t, err)

	config.Storage.S3.Endpoint = "http://localhost:9011"
	err = checkStorage(config)
	assert.Error(t, err)
}
