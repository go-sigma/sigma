// Copyright 2023 XImager
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
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"github.com/ximager/ximager/pkg/dal"
)

func TestCheckRedis(t *testing.T) {
	viper.SetDefault("redis.url", "redis:///127.0.0.1:6379")
	err := checkRedis()
	assert.Error(t, err)

	miniRedis := miniredis.RunT(t)
	viper.SetDefault("redis.url", "redis://"+miniRedis.Addr())
	err = checkRedis()
	assert.NoError(t, err)

	viper.SetDefault("redis.url", "redis://127.0.0.1:1100")
	err = checkRedis()
	assert.Error(t, err)
}

func TestCheckDatabase(t *testing.T) {
	viper.SetDefault("database.type", "sqlite3")

	err := checkDatabase()
	assert.NoError(t, err)

	viper.SetDefault("database.type", dal.DatabaseMysql.String())
	viper.SetDefault("database.mysql.host", "127.0.0.1")
	viper.SetDefault("database.mysql.port", "3306")
	viper.SetDefault("database.mysql.user", "root")
	viper.SetDefault("database.mysql.password", "ximager")
	viper.SetDefault("database.mysql.database", "ximager")

	err = checkDatabase()
	assert.NoError(t, err)

	viper.SetDefault("database.type", dal.DatabasePostgresql.String())
	viper.SetDefault("database.postgres.host", "localhost")
	viper.SetDefault("database.postgres.port", 5432)
	viper.SetDefault("database.postgres.user", "ximager")
	viper.SetDefault("database.postgres.password", "ximager")
	viper.SetDefault("database.postgres.dbname", "ximager")

	err = checkDatabase()
	assert.NoError(t, err)

	viper.SetDefault("database.type", "fake")

	err = checkDatabase()
	assert.Error(t, err)
}

func TestCheckMysql(t *testing.T) {
	viper.SetDefault("database.type", dal.DatabaseMysql.String())
	viper.SetDefault("database.mysql.host", "127.0.0.1")
	viper.SetDefault("database.mysql.port", "3306")
	viper.SetDefault("database.mysql.user", "root")
	viper.SetDefault("database.mysql.password", "ximager")
	viper.SetDefault("database.mysql.database", "ximager")

	err := checkMysql()
	assert.NoError(t, err)

	viper.SetDefault("database.mysql.port", "3310")

	err = checkMysql()
	assert.Error(t, err)
}

func TestCheckPostgresql(t *testing.T) {
	viper.SetDefault("database.type", dal.DatabasePostgresql.String())
	viper.SetDefault("database.postgres.host", "localhost")
	viper.SetDefault("database.postgres.port", 5432)
	viper.SetDefault("database.postgres.user", "ximager")
	viper.SetDefault("database.postgres.password", "ximager")
	viper.SetDefault("database.postgres.dbname", "ximager")

	err := checkPostgresql()
	assert.NoError(t, err)

	viper.SetDefault("database.postgres.port", 5433)

	err = checkPostgresql()
	assert.Error(t, err)
}

func TestCheckS3(t *testing.T) {
	viper.SetDefault("storage.s3.endpoint", "http://127.0.0.1:9000")
	viper.SetDefault("storage.s3.region", "cn-north-1")
	viper.SetDefault("storage.s3.ak", "ximager")
	viper.SetDefault("storage.s3.sk", "ximager-ximager")
	viper.SetDefault("storage.s3.bucket", "ximager")
	viper.SetDefault("storage.s3.forcePathStyle", true)

	err := checkS3()
	assert.NoError(t, err)

	viper.SetDefault("storage.s3.endpoint", "http://localhost:9011")

	err = checkS3()
	assert.Error(t, err)
}
