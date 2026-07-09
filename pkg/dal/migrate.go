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
	"log/slog"

	"github.com/pressly/goose/v3"

	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/config"
)

//go:embed migrations/mysql/*.sql
var mysqlFS embed.FS

//go:embed migrations/postgresql/*.sql
var postgresqlFS embed.FS

//go:embed migrations/sqlite3/*.sql
var sqliteFS embed.FS

//go:embed migrations/turso/*.sql
var tursoFS embed.FS

// MigrateDatabase migrates the database to the latest version
func MigrateDatabase(config *config.Configuration, rawDB *sql.DB) error {
	var fs embed.FS
	var dir string
	var dialect string
	switch config.Database.Type {
	case enums.DatabaseMysql:
		fs = mysqlFS
		dir = "migrations/mysql"
		dialect = "mysql"
	case enums.DatabasePostgresql:
		fs = postgresqlFS
		dir = "migrations/postgresql"
		dialect = "postgres"
	case enums.DatabaseSqlite3:
		fs = sqliteFS
		dir = "migrations/sqlite3"
		dialect = "sqlite3"
	case enums.DatabaseTurso:
		fs = tursoFS
		dir = "migrations/turso"
		dialect = "sqlite3"
	}

	goose.SetBaseFS(fs)
	if err := goose.SetDialect(dialect); err != nil {
		return fmt.Errorf("set migrate dialect failed: %v", err)
	}

	if err := goose.Up(rawDB, dir); err != nil {
		return fmt.Errorf("migrate up failed: %v", err)
	}

	version, err := goose.GetDBVersion(rawDB)
	if err != nil {
		return fmt.Errorf("get migrate version failed: %v", err)
	}
	slog.Info("migrate database succeed", "version", version)
	return nil
}
