// Copyright 2024 sigma
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

package dao_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/xo/dburl"
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/dal"
	"github.com/go-sigma/sigma/pkg/dal/badger"
	"github.com/go-sigma/sigma/pkg/modules/locker"
	"github.com/go-sigma/sigma/pkg/modules/locker/definition"
	"github.com/go-sigma/sigma/pkg/tests"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils/ptr"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func initDal(t *testing.T) *dig.Container {
	config, err := tests.GetConfig()
	require.NoError(t, err)

	digCon := dig.New()
	require.NoError(t, digCon.Provide(func() configs.Configuration {
		database := strings.ReplaceAll(uuid.Must(uuid.NewV7()).String(), "-", "")
		config.Database.Sqlite3.Path = fmt.Sprintf("%s.db", database)
		if config.Database.Type == enums.DatabaseMysql {
			config.Database.Mysql.Database = database
			initMysqlDatabase(t, database)
		} else if config.Database.Type == enums.DatabasePostgresql {
			config.Database.Postgresql.Database = database
			config.Database.Postgresql.SslMode = "disable"
			initPostgresqlDatabase(t, database)
		}
		return ptr.To(config)
	}))
	require.NoError(t, digCon.Provide(func() (definition.Locker, error) { return locker.Initialize(digCon) }))
	require.NoError(t, digCon.Provide(badger.New))
	require.NoError(t, dal.Initialize(digCon))

	return digCon
}

func initMysqlDatabase(t *testing.T, database string) {
	db, err := dburl.Open("mysql://root:sigma@127.0.0.1:3306")
	require.NoError(t, err)
	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE `%s` DEFAULT CHARACTER SET utf8mb4 DEFAULT COLLATE utf8mb4_unicode_ci", database))
	require.NoError(t, err)
}

func initPostgresqlDatabase(t *testing.T, database string) {
	db, err := dburl.Open("pgx://sigma:sigma@localhost:5432?sslmode=disable")
	require.NoError(t, err)
	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE \"%s\";", database))
	require.NoError(t, err)
}
