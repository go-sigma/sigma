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

package dal_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/jackc/pgx/v4"
	gonanoid "github.com/matoous/go-nanoid"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"github.com/go-sigma/sigma/pkg/dal"
	"github.com/go-sigma/sigma/pkg/logger"
	"github.com/go-sigma/sigma/pkg/types/enums"
)

func TestInitialize(t *testing.T) {
	logger.SetLevel("debug")

	miniRedis := miniredis.RunT(t)
	viper.SetDefault("redis.url", "redis://"+miniRedis.Addr())

	dbPath := fmt.Sprintf("%s.db", gonanoid.MustGenerate("abcdefghijklmnopqrstuvwxyz", 6))
	viper.SetDefault("database.type", "sqlite3")
	viper.SetDefault("database.sqlite3.path", dbPath)

	err := dal.Initialize()
	assert.NoError(t, err)

	db, err := dal.DB.DB()
	assert.NoError(t, err)
	assert.NoError(t, db.Close())

	err = os.Remove(dbPath)
	assert.NoError(t, err)

	db, err = sql.Open("mysql", "root:sigma@tcp(127.0.0.1:3306)/")
	assert.NoError(t, err)

	dbname := gonanoid.MustGenerate("abcdefghijklmnopqrstuvwxyz", 6)
	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", dbname))
	assert.NoError(t, err)
	assert.NoError(t, db.Close())

	viper.SetDefault("database.type", enums.DatabaseMysql.String())
	viper.SetDefault("database.mysql.host", "127.0.0.1")
	viper.SetDefault("database.mysql.port", "3306")
	viper.SetDefault("database.mysql.user", "root")
	viper.SetDefault("database.mysql.password", "sigma")
	viper.SetDefault("database.mysql.database", "sigma")
	viper.SetDefault("database.mysql.dbname", dbname)

	err = dal.Initialize()
	assert.NoError(t, err)

	db, err = sql.Open("mysql", "root:sigma@tcp(127.0.0.1:3306)/")
	assert.NoError(t, err)
	_, err = db.Exec(fmt.Sprintf("DROP DATABASE %s", dbname))
	assert.NoError(t, err)
	assert.NoError(t, db.Close())

	viper.SetDefault("database.type", enums.DatabasePostgresql.String())
	viper.SetDefault("database.postgres.host", "localhost")
	viper.SetDefault("database.postgres.port", 5432)
	viper.SetDefault("database.postgres.user", "sigma")
	viper.SetDefault("database.postgres.password", "sigma")
	viper.SetDefault("database.postgres.dbname", "sigma")
	viper.SetDefault("database.postgres.dbname", dbname)

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, "postgres://sigma:sigma@localhost:5432/?sslmode=disable")
	assert.NoError(t, err)

	_, err = conn.Exec(ctx, fmt.Sprintf("CREATE DATABASE \"%s\"", dbname))
	assert.NoError(t, err)
	assert.NoError(t, db.Close())

	err = dal.Initialize()
	assert.NoError(t, err)
}

func TestInitialize1(t *testing.T) {
	viper.SetDefault("log.level", "debug")
	logger.SetLevel("debug")

	dbPath := fmt.Sprintf("%s.db", gonanoid.MustGenerate("abcdefghijklmnopqrstuvwxyz", 6))
	viper.SetDefault("database.type", "sqlite3")
	viper.SetDefault("database.sqlite3.path", dbPath)

	miniRedis := miniredis.RunT(t)
	viper.SetDefault("redis.url", "redis://"+miniRedis.Addr())

	err := dal.Initialize()
	assert.NoError(t, err)

	db, err := dal.DB.DB()
	assert.NoError(t, err)
	assert.NoError(t, db.Close())

	err = os.Remove(dbPath)
	assert.NoError(t, err)
}

func TestInitialize2(t *testing.T) {
	defer func() {
		err := recover()
		assert.NotNil(t, err)
	}()

	miniRedis := miniredis.RunT(t)
	viper.SetDefault("redis.url", "redis://"+miniRedis.Addr())

	viper.SetDefault("database.type", "unknown")
	err := dal.Initialize()
	assert.Error(t, err)
}
