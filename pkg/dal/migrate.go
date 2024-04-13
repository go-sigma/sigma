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
	"embed"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/rs/zerolog/log"
)

//go:embed migrations/mysql/*.sql
var mysqlFS embed.FS

//go:embed migrations/postgresql/*.sql
var postgresqlFS embed.FS

//go:embed migrations/sqlite3/*.sql
var sqliteFS embed.FS

func migrateMysql(database string) error {
	d, err := iofs.New(mysqlFS, "migrations/mysql")
	if err != nil {
		return err
	}
	rawDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("get raw db instance failed")
	}
	migrateDriver, err := mysql.WithInstance(rawDB, &mysql.Config{})
	if err != nil {
		return fmt.Errorf("get migrate driver failed")
	}
	m, err := migrate.NewWithInstance("iofs", d, database, migrateDriver)
	if err != nil {
		return err
	}
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return err
	}
	version, dirty, err := m.Version()
	if err != nil {
		return err
	}
	log.Info().Uint("version", version).Bool("dirty", dirty).Msg("Migrate database")
	return nil
}

func migratePostgres(database string) error {
	d, err := iofs.New(postgresqlFS, "migrations/postgresql")
	if err != nil {
		return err
	}
	rawDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("get raw db instance failed")
	}
	migrateDriver, err := postgres.WithInstance(rawDB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("get migrate driver failed")
	}
	m, err := migrate.NewWithInstance("iofs", d, database, migrateDriver)
	if err != nil {
		return err
	}
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return err
	}
	version, dirty, err := m.Version()
	if err != nil {
		return err
	}
	log.Info().Uint("version", version).Bool("dirty", dirty).Msg("Migrate database")
	return nil
}

func migrateSqlite() error {
	d, err := iofs.New(sqliteFS, "migrations/sqlite3")
	if err != nil {
		return err
	}
	rawDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("get raw db instance failed")
	}
	migrateDriver, err := sqlite3.WithInstance(rawDB, &sqlite3.Config{})
	if err != nil {
		return fmt.Errorf("get migrate driver failed")
	}
	m, err := migrate.NewWithInstance("iofs", d, "", migrateDriver)
	if err != nil {
		return err
	}
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return err
	}
	version, dirty, err := m.Version()
	if err != nil {
		return err
	}
	log.Info().Uint("version", version).Bool("dirty", dirty).Msg("Migrate database")
	return nil
}
