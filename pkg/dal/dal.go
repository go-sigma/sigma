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
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"go.uber.org/dig"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/plugin/opentelemetry/tracing"

	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/infra/lock"
	"github.com/go-sigma/sigma/pkg/logger"

	_ "turso.tech/database/tursogo"
)

// Initialize initializes the database connection
func Initialize(digCon *dig.Container) error {
	var cfg *config.Configuration
	if err := digCon.Invoke(func(c *config.Configuration) { cfg = c }); err != nil {
		return err
	}

	db, err := connectDatabase(cfg)
	if err != nil {
		return err
	}
	initialized := false
	defer func() {
		if !initialized {
			_ = closeDatabase(db)
		}
	}()

	err = digCon.Provide(func() *gorm.DB { return db })
	if err != nil {
		return err
	}

	if cfg.Log.Level == enums.LogLevelDebug || cfg.Log.Level == enums.LogLevelTrace {
		query.SetDefault(db.Debug())
	} else {
		query.SetDefault(db)
	}

	ctx, ctxCancel := context.WithCancel(context.Background())
	defer ctxCancel()

	var lockerInst lock.Locker
	err = digCon.Invoke(func(l lock.Locker) { lockerInst = l })
	if err != nil {
		return err
	}
	err = lockerInst.AcquireWithRenew(ctx, consts.LockerMigration, time.Second*3, time.Second*5)
	if err != nil {
		return err
	}

	rawDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("get raw db instance failed: %v", err)
	}

	err = MigrateDatabase(cfg, rawDB)
	if err != nil {
		return err
	}

	err = initDigContainer(digCon)
	if err != nil {
		return err
	}

	initialized = true
	return nil
}

func connectDatabase(cfg *config.Configuration) (*gorm.DB, error) {
	switch cfg.Database.Type {
	case enums.DatabaseMysql:
		return connectMysql(cfg)
	case enums.DatabasePostgresql:
		return connectPostgres(cfg)
	case enums.DatabaseSqlite3:
		return connectSqlite3(cfg)
	case enums.DatabaseTurso:
		return connectTurso(cfg)
	default:
		return nil, fmt.Errorf("unknown database type: %s", cfg.Database.Type)
	}
}

// DeInitialize ...
func DeInitialize(digCon *dig.Container) error {
	var db *gorm.DB
	if err := digCon.Invoke(func(database *gorm.DB) { db = database }); err != nil {
		return err
	}
	return closeDatabase(db)
}

func closeDatabase(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("database is not initialized")
	}
	conn, err := db.DB()
	if err != nil {
		return fmt.Errorf("get raw db instance failed: %v", err)
	}
	return conn.Close()
}

func connectMysql(config *config.Configuration) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&loc=UTC",
		config.Database.Mysql.Username, config.Database.Mysql.Password,
		config.Database.Mysql.Host, config.Database.Mysql.Port, config.Database.Mysql.Database)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
		Logger: logger.ZLogger{},
	})
	if err != nil {
		return nil, err
	}
	if err := db.Use(tracing.NewPlugin()); err != nil {
		_ = closeDatabase(db)
		return nil, fmt.Errorf("attach otelgorm plugin: %w", err)
	}
	db = db.WithContext(context.Background())

	if err := applyConnPool(db, config, "mysql", ""); err != nil {
		_ = closeDatabase(db)
		return nil, err
	}

	return db, nil
}

func connectPostgres(config *config.Configuration) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@%s:%d/%s?sslmode=%s", config.Database.Postgresql.Username,
		config.Database.Postgresql.Password, config.Database.Postgresql.Host,
		config.Database.Postgresql.Port, config.Database.Postgresql.Database,
		config.Database.Postgresql.SslMode)
	db, err := gorm.Open(postgres.Open("postgresql://"+dsn), &gorm.Config{
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
		Logger: logger.ZLogger{},
	})
	if err != nil {
		return nil, err
	}
	if err := db.Use(tracing.NewPlugin()); err != nil {
		_ = closeDatabase(db)
		return nil, fmt.Errorf("attach otelgorm plugin: %w", err)
	}
	db = db.WithContext(context.Background())

	if err := applyConnPool(db, config, "postgresql", ""); err != nil {
		_ = closeDatabase(db)
		return nil, err
	}

	return db, nil
}

func connectSqlite3(config *config.Configuration) (*gorm.DB, error) {
	dbname := config.Database.Sqlite3.Path
	dsn := "file:" + dbname
	if strings.Contains(dbname, "?") {
		dsn += "&_busy_timeout=30000"
	} else {
		dsn += "?_busy_timeout=30000"
	}

	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
		Logger: logger.ZLogger{},
	})
	if err != nil {
		return nil, err
	}
	if err := db.Use(tracing.NewPlugin()); err != nil {
		_ = closeDatabase(db)
		return nil, fmt.Errorf("attach otelgorm plugin: %w", err)
	}
	db = db.WithContext(context.Background())

	if err := applyConnPool(db, config, "sqlite3", dbname); err != nil {
		_ = closeDatabase(db)
		return nil, err
	}

	return db, nil
}

func connectTurso(config *config.Configuration) (*gorm.DB, error) {
	rawDB, err := sql.Open("turso", config.Database.Turso.DSN)
	if err != nil {
		return nil, fmt.Errorf("open turso database failed: %w", err)
	}

	db, err := gorm.Open(sqlite.Dialector{Conn: rawDB}, &gorm.Config{
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
		Logger: logger.ZLogger{},
	})
	if err != nil {
		_ = rawDB.Close()
		return nil, err
	}
	if err := db.Use(tracing.NewPlugin()); err != nil {
		_ = closeDatabase(db)
		return nil, fmt.Errorf("attach otelgorm plugin: %w", err)
	}
	db = db.WithContext(context.Background())

	if err := applyConnPool(db, config, "turso", config.Database.Turso.DSN); err != nil {
		_ = closeDatabase(db)
		return nil, err
	}

	return db, nil
}

// applyConnPool configures the connection pool for all database drivers.
// For sqlite-like drivers, sqliteDBName is used to detect in-memory mode (single connection).
// For mysql/postgresql, sqliteDBName is ignored and config values are used directly.
func applyConnPool(db *gorm.DB, config *config.Configuration, driver, sqliteDBName string) error {
	rawDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("get raw db instance failed: %v", err)
	}
	dbCfg := config.Database
	maxOpen, maxIdle := dbCfg.MaxOpenConns, dbCfg.MaxIdleConns
	lifetime, idleTime := dbCfg.ConnMaxLifetime, dbCfg.ConnMaxIdleTime

	if driver == "sqlite3" || driver == "turso" {
		if strings.Contains(sqliteDBName, "mode=memory") || sqliteDBName == ":memory:" {
			maxOpen = 1
		} else {
			maxOpen = 10
		}
		maxIdle = 3
		lifetime = time.Hour
		idleTime = time.Hour
	}

	rawDB.SetMaxOpenConns(maxOpen)
	rawDB.SetMaxIdleConns(maxIdle)
	rawDB.SetConnMaxLifetime(lifetime)
	rawDB.SetConnMaxIdleTime(idleTime)
	slog.Info("database connection pool configured",
		"driver", driver,
		"maxOpenConns", maxOpen,
		"maxIdleConns", maxIdle,
		"connMaxLifetime", lifetime,
		"connMaxIdleTime", idleTime,
	)
	return nil
}
