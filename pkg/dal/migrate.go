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

package dal

import (
	"database/sql"
	"embed"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/rs/zerolog/log"

	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/types/enums"
)

//go:embed migrations/mysql/*.sql
var mysqlFS embed.FS

//go:embed migrations/postgresql/*.sql
var postgresqlFS embed.FS

//go:embed migrations/sqlite3/*.sql
var sqliteFS embed.FS

// MigrateDatabase migrates the database to the latest version
func MigrateDatabase(config configs.Configuration, rawDB *sql.DB) error {
	var err error
	var sourceDriver source.Driver
	switch config.Database.Type {
	case enums.DatabaseMysql:
		sourceDriver, err = iofs.New(mysqlFS, "migrations/mysql")
	case enums.DatabasePostgresql:
		sourceDriver, err = iofs.New(postgresqlFS, "migrations/postgresql")
	case enums.DatabaseSqlite3:
		sourceDriver, err = iofs.New(sqliteFS, "migrations/sqlite3")
	}
	if err != nil {
		return fmt.Errorf("new iofs instance failed: %v", err)
	}

	var databaseDriver database.Driver
	switch config.Database.Type {
	case enums.DatabaseMysql:
		databaseDriver, err = mysql.WithInstance(rawDB, &mysql.Config{})
	case enums.DatabasePostgresql:
		databaseDriver, err = postgres.WithInstance(rawDB, &postgres.Config{})
	case enums.DatabaseSqlite3:
		databaseDriver, err = sqlite3.WithInstance(rawDB, &sqlite3.Config{})
	}
	if err != nil {
		return fmt.Errorf("get migrate driver failed: %v", err)
	}
	m, err := migrate.NewWithInstance("iofs", sourceDriver, "", databaseDriver)
	if err != nil {
		return fmt.Errorf("new migrate instance failed: %v", err)
	}
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migrate up failed: %v", err)
	}
	version, dirty, err := m.Version()
	if err != nil {
		return fmt.Errorf("get migrate version failed: %v", err)
	}
	log.Info().Uint("version", version).Bool("dirty", dirty).Msg("migrate database succeed")
	return nil
}
