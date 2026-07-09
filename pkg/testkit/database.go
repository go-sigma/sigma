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

package testkit

import (
	"crypto/rand"
	"database/sql"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync/atomic"
	"testing"

	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/dal"
	"github.com/go-sigma/sigma/pkg/infra/lock"

	_ "github.com/go-sigma/sigma/pkg/infra/lock/inmemory"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	// EnvDatabaseType selects the database used by database-backed tests
	EnvDatabaseType = "CI_DATABASE_TYPE"

	databaseNamePrefix = "sigma_test"
	mySQLAdminDSN      = "root:sigma@tcp(127.0.0.1:3306)/"                                //nolint:gosec // CI fixture credentials
	postgreSQLAdminDSN = "postgres://sigma:sigma@localhost:5432/postgres?sslmode=disable" //nolint:gosec // CI fixture credentials
)

var (
	databaseNamePattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,62}$`)
	databaseNameSeq     atomic.Uint64
)

// DatabaseType returns the configured test database type or defaultType when
// EnvDatabaseType is not set
func DatabaseType(defaultType enums.Database) (enums.Database, error) {
	value, ok := os.LookupEnv(EnvDatabaseType)
	if !ok {
		if !defaultType.IsValid() {
			return "", fmt.Errorf("invalid default database type %q", defaultType)
		}
		return defaultType, nil
	}

	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("%s must not be empty", EnvDatabaseType)
	}
	databaseType, err := enums.ParseDatabase(value)
	if err != nil {
		return "", fmt.Errorf("parse %s %q: %w", EnvDatabaseType, value, err)
	}
	return databaseType, nil
}

// NewDatabaseName returns a random identifier-safe database name for tests
func NewDatabaseName() string {
	var bs [8]byte
	if _, err := rand.Read(bs[:]); err != nil {
		panic(fmt.Sprintf("generate database name: %v", err))
	}
	return fmt.Sprintf("%s_%020d_%010d", databaseNamePrefix, binary.BigEndian.Uint64(bs[:]), databaseNameSeq.Add(1))
}

// PrepareDatabase creates an isolated database configuration for a test
func PrepareDatabase(t testing.TB, databaseType enums.Database) (config.ConfigurationDatabase, error) {
	t.Helper()

	databaseName := NewDatabaseName()
	databaseConfig := config.ConfigurationDatabase{Type: databaseType}

	switch databaseType {
	case enums.DatabaseSqlite3:
		databaseConfig.Sqlite3.Path = databaseName + "?mode=memory&cache=shared"
	case enums.DatabaseTurso:
		databaseConfig.Turso.DSN = filepath.Join(t.TempDir(), databaseName+".db")
	case enums.DatabaseMysql:
		databaseConfig.Mysql = config.ConfigurationDatabaseMysql{
			Host:     "127.0.0.1",
			Port:     3306,
			Username: "root",
			Password: "sigma",
			Database: databaseName,
		}
		if err := createMySQLDatabase(databaseName); err != nil {
			return config.ConfigurationDatabase{}, err
		}
		t.Cleanup(func() {
			if err := dropMySQLDatabase(databaseName); err != nil {
				t.Errorf("drop mysql test database %q: %v", databaseName, err)
			}
		})
	case enums.DatabasePostgresql:
		databaseConfig.Postgresql = config.ConfigurationDatabasePostgresql{
			Host:     "localhost",
			Port:     5432,
			Username: "sigma",
			Password: "sigma",
			Database: databaseName,
			SslMode:  "disable",
		}
		if err := createPostgreSQLDatabase(databaseName); err != nil {
			return config.ConfigurationDatabase{}, err
		}
		t.Cleanup(func() {
			if err := dropPostgreSQLDatabase(databaseName); err != nil {
				t.Errorf("drop postgresql test database %q: %v", databaseName, err)
			}
		})
	default:
		return config.ConfigurationDatabase{}, fmt.Errorf("unsupported test database type %q", databaseType)
	}

	databaseConfig.WithDefaults()
	return databaseConfig, nil
}

// Initialize configures and initializes the DAL for an existing dig container
func Initialize(t testing.TB, digCon *dig.Container, databaseType enums.Database) error {
	t.Helper()

	databaseConfig, err := PrepareDatabase(t, databaseType)
	if err != nil {
		return fmt.Errorf("prepare %s test database: %w", databaseType, err)
	}
	if err = digCon.Decorate(func(cfg *config.Configuration) *config.Configuration {
		cfg.Database = databaseConfig
		cfg.Locker.Type = enums.LockerTypeInmemory
		return cfg
	}); err != nil {
		return fmt.Errorf("configure %s test database: %w", databaseType, err)
	}
	if err = digCon.Provide(lock.Initialize); err != nil {
		return fmt.Errorf("provide test locker: %w", err)
	}
	if err = dal.Initialize(digCon); err != nil {
		return fmt.Errorf("initialize %s test database: %w", databaseType, err)
	}

	t.Cleanup(func() {
		if err := dal.DeInitialize(digCon); err != nil {
			t.Errorf("close %s test database: %v", databaseType, err)
		}
	})
	return nil
}

// InitializeIntegration initializes an integration test DAL, defaulting to
// PostgreSQL while allowing EnvDatabaseType to select another database
func InitializeIntegration(t testing.TB, digCon *dig.Container) error {
	t.Helper()

	databaseType, err := DatabaseType(enums.DatabasePostgresql)
	if err != nil {
		return fmt.Errorf("select integration test database: %w", err)
	}
	return Initialize(t, digCon, databaseType)
}

// InitDal creates a fresh DAL container using the selected test database
func InitDal(t testing.TB, defaultType enums.Database) *dig.Container {
	t.Helper()

	databaseType, err := DatabaseType(defaultType)
	if err != nil {
		t.Fatalf("select test database: %v", err)
	}
	digCon := dig.New()
	if err = digCon.Provide(func() *config.Configuration { return &config.Configuration{} }); err != nil {
		t.Fatalf("provide test configuration: %v", err)
	}
	if err = Initialize(t, digCon, databaseType); err != nil {
		t.Fatalf("initialize test DAL: %v", err)
	}
	return digCon
}

// InitRepository creates a fresh DAL container for repository tests
func InitRepository(t testing.TB) *dig.Container {
	t.Helper()
	return InitDal(t, enums.DatabaseSqlite3)
}

// RequireSQLiteDialect skips the test unless the selected database uses the
// SQLite SQL dialect
func RequireSQLiteDialect(t testing.TB) {
	t.Helper()

	databaseType, err := DatabaseType(enums.DatabaseSqlite3)
	if err != nil {
		t.Fatalf("select test database: %v", err)
	}
	if databaseType != enums.DatabaseSqlite3 && databaseType != enums.DatabaseTurso {
		t.Skipf("requires sqlite dialect, got %s", databaseType)
	}
}

func createMySQLDatabase(database string) error {
	if err := validateDatabaseName(database); err != nil {
		return err
	}
	return withAdminDatabase("mysql", mySQLAdminDSN, func(db *sql.DB) error {
		_, err := db.Exec(fmt.Sprintf(
			"CREATE DATABASE `%s` DEFAULT CHARACTER SET utf8mb4 DEFAULT COLLATE utf8mb4_unicode_ci",
			database,
		))
		return err
	})
}

func dropMySQLDatabase(database string) error {
	if err := validateDatabaseName(database); err != nil {
		return err
	}
	return withAdminDatabase("mysql", mySQLAdminDSN, func(db *sql.DB) error {
		_, err := db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS `%s`", database))
		return err
	})
}

func createPostgreSQLDatabase(database string) error {
	if err := validateDatabaseName(database); err != nil {
		return err
	}
	return withAdminDatabase("pgx", postgreSQLAdminDSN, func(db *sql.DB) error {
		_, err := db.Exec(fmt.Sprintf(`CREATE DATABASE "%s"`, database))
		return err
	})
}

func dropPostgreSQLDatabase(database string) error {
	if err := validateDatabaseName(database); err != nil {
		return err
	}
	return withAdminDatabase("pgx", postgreSQLAdminDSN, func(db *sql.DB) error {
		if _, err := db.Exec(
			"SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = $1 AND pid <> pg_backend_pid()",
			database,
		); err != nil {
			return err
		}
		_, err := db.Exec(fmt.Sprintf(`DROP DATABASE IF EXISTS "%s"`, database))
		return err
	})
}

func validateDatabaseName(database string) error {
	if !databaseNamePattern.MatchString(database) {
		return fmt.Errorf("invalid test database name %q", database)
	}
	return nil
}

func withAdminDatabase(driver, dsn string, fn func(*sql.DB) error) (err error) {
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := db.Close(); err == nil {
			err = closeErr
		}
	}()
	return fn(db)
}
