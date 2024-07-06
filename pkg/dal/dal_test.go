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

	"github.com/jackc/pgx/v4"
	gonanoid "github.com/matoous/go-nanoid"
	"github.com/stretchr/testify/assert"

	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/dal"
	"github.com/go-sigma/sigma/pkg/logger"
	"github.com/go-sigma/sigma/pkg/types/enums"
)

func TestInitialize(t *testing.T) {
	logger.SetLevel("debug")

	dbPath := fmt.Sprintf("%s.db", gonanoid.MustGenerate("abcdefghijklmnopqrstuvwxyz", 6))

	err := dal.Initialize(configs.Configuration{
		Database: configs.ConfigurationDatabase{
			Type: enums.DatabaseSqlite3,
			Sqlite3: configs.ConfigurationDatabaseSqlite3{
				Path: dbPath,
			},
		},
	})
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

	err = dal.Initialize(configs.Configuration{
		Database: configs.ConfigurationDatabase{
			Type: enums.DatabaseMysql,
			Mysql: configs.ConfigurationDatabaseMysql{
				Host:     "127.0.0.1",
				Port:     3306,
				User:     "root",
				Password: "sigma",
				DBName:   dbname,
			},
		},
	})
	assert.NoError(t, err)

	db, err = sql.Open("mysql", "root:sigma@tcp(127.0.0.1:3306)/")
	assert.NoError(t, err)
	_, err = db.Exec(fmt.Sprintf("DROP DATABASE %s", dbname))
	assert.NoError(t, err)
	assert.NoError(t, db.Close())

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, "postgres://sigma:sigma@localhost:5432/?sslmode=disable")
	assert.NoError(t, err)

	_, err = conn.Exec(ctx, fmt.Sprintf("CREATE DATABASE \"%s\"", dbname))
	assert.NoError(t, err)
	assert.NoError(t, db.Close())

	err = dal.Initialize(configs.Configuration{
		Database: configs.ConfigurationDatabase{
			Type: enums.DatabasePostgresql,
			Postgresql: configs.ConfigurationDatabasePostgresql{
				Host:     "localhost",
				Port:     5432,
				User:     "sigma",
				Password: "sigma",
				DBName:   dbname,
				SslMode:  "disable",
			},
		},
	})
	assert.NoError(t, err)
}

func TestInitialize1(t *testing.T) {
	logger.SetLevel("debug")

	dbPath := fmt.Sprintf("%s.db", gonanoid.MustGenerate("abcdefghijklmnopqrstuvwxyz", 6))

	err := dal.Initialize(configs.Configuration{
		Database: configs.ConfigurationDatabase{
			Type: enums.DatabaseSqlite3,
			Sqlite3: configs.ConfigurationDatabaseSqlite3{
				Path: dbPath,
			},
		},
	})
	assert.NoError(t, err)

	db, err := dal.DB.DB()
	assert.NoError(t, err)
	assert.NoError(t, db.Close())

	err = os.Remove(dbPath)
	assert.NoError(t, err)
}

func TestInitializeDatabaseUnknown(t *testing.T) {
	assert.Error(t, dal.Initialize(configs.Configuration{}))
}
