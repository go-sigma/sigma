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
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/rs/zerolog/log"

	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
)

//go:embed migrations/mysql/*.sql
var mysqlFS embed.FS

//go:embed migrations/postgresql/*.sql
var postgresqlFS embed.FS

//go:embed migrations/sqlite3/*.sql
var sqlite3FS embed.FS

func migrateMysql(dsn string) error {
	d, err := iofs.New(mysqlFS, "migrations/mysql")
	if err != nil {
		return err
	}
	m, err := migrate.NewWithSourceInstance("iofs", d, fmt.Sprintf("mysql://%s", dsn))
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

func migratePostgres(dsn string) error {
	d, err := iofs.New(postgresqlFS, "migrations/postgresql")
	if err != nil {
		return err
	}
	m, err := migrate.NewWithSourceInstance("iofs", d, fmt.Sprintf("postgres://%s", dsn))
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

func migrateSqlite(dsn string) error {
	d, err := iofs.New(sqlite3FS, "migrations/sqlite3")
	if err != nil {
		return err
	}
	m, err := migrate.NewWithSourceInstance("iofs", d, fmt.Sprintf("sqlite3://%s", dsn))
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
