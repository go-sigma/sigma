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
	"errors"
	"fmt"
	"log/slog"
	"net"
	"strings"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/jackc/pgx/v5/pgconn"
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
	"github.com/go-sigma/sigma/pkg/server/errcode"

	_ "turso.tech/database/tursogo"
)

const (
	// connectRetryBudget bounds how long Initialize waits for a database that
	// is not accepting connections yet. Containers are started before their
	// database is ready (fresh install, database pod restart, namespace
	// re-creation); failing on the first refused connection turns that race
	// into a CrashLoopBackOff. Kubernetes startup probes must allow at least
	// this budget plus app.StartupTimeout before restarting the container.
	connectRetryBudget = 2 * time.Minute
	connectRetryDelay  = 500 * time.Millisecond
	// connectRetryMaxDelay caps the exponential backoff between attempts.
	connectRetryMaxDelay = 5 * time.Second
	// connectDialTimeout bounds a single dial and handshake. Without it a
	// dropped SYN (node gone behind a ClusterIP, DROP policy) blocks until the
	// OS TCP timeout, so the retry budget would stop being an upper bound.
	connectDialTimeout = 10 * time.Second
	// pgerrDatabaseStartingUp is the PostgreSQL SQLSTATE returned while the
	// server is still performing recovery; the port is already accepting TCP.
	pgerrDatabaseStartingUp = "57P03"
)

// Initialize initializes the database connection
func Initialize(digCon *dig.Container) error {
	var cfg *config.Configuration
	if err := digCon.Invoke(func(c *config.Configuration) { cfg = c }); err != nil {
		return err
	}

	db, err := connectDatabaseWithRetry(cfg)
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

// connectDatabaseWithRetry connects to the database, retrying while the
// database is unreachable so that a database that is still booting does not
// kill the process. Errors that retrying cannot fix, such as an unknown
// database type or malformed credentials, are returned immediately.
func connectDatabaseWithRetry(cfg *config.Configuration) (*gorm.DB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), connectRetryBudget)
	defer cancel()

	var db *gorm.DB
	var lastErr error
	err := retryConnect(ctx, func() error {
		connected, err := connectDatabase(cfg)
		if err != nil {
			lastErr = err
			return err
		}
		db = connected
		return nil
	})
	if err != nil {
		if lastErr == nil {
			return nil, err
		}
		// Report the database error rather than a bare deadline, but only claim
		// the budget elapsed when it did: non-retryable failures (bad
		// credentials, unknown database type) return on the first attempt.
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, fmt.Errorf("database unreachable after %s: %w", connectRetryBudget, lastErr)
		}
		return nil, lastErr
	}
	return db, nil
}

// retryConnect runs fn until it succeeds, returns a non-retryable error, or
// ctx is done.
func retryConnect(ctx context.Context, fn func() error) error {
	return retry.Do(fn,
		retry.Context(ctx),
		retry.Attempts(0),
		retry.Delay(connectRetryDelay),
		retry.MaxDelay(connectRetryMaxDelay),
		retry.LastErrorOnly(true),
		retry.RetryIf(isRetryableConnectError),
		retry.OnRetry(func(n uint, err error) {
			slog.Warn("database is not ready, retrying connection", "attempt", n+1, "err", err)
		}),
	)
}

// isRetryableConnectError reports whether err describes a database that is
// temporarily unreachable rather than a configuration mistake.
func isRetryableConnectError(err error) bool {
	// Both pgx (postgres) and go-sql-driver (mysql) wrap a refused or sluggish
	// dial in *net.OpError.
	if _, ok := errcode.AsType[*net.OpError](err); ok {
		return true
	}
	if _, ok := errcode.AsType[*net.DNSError](err); ok {
		return true
	}
	// PostgreSQL accepts TCP connections before it finishes recovery.
	pgErr, ok := errcode.AsType[*pgconn.PgError](err)
	return ok && pgErr.Code == pgerrDatabaseStartingUp
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

// DeInitialize closes the database connection resolved from digCon, releasing
// the connection pool. It returns an error if no database is registered or the
// underlying connection cannot be closed.
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
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&loc=UTC&timeout=%s",
		config.Database.Mysql.Username, config.Database.Mysql.Password,
		config.Database.Mysql.Host, config.Database.Mysql.Port, config.Database.Mysql.Database,
		connectDialTimeout)

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
	dsn := fmt.Sprintf("%s:%s@%s:%d/%s?sslmode=%s&connect_timeout=%d", config.Database.Postgresql.Username,
		config.Database.Postgresql.Password, config.Database.Postgresql.Host,
		config.Database.Postgresql.Port, config.Database.Postgresql.Database,
		config.Database.Postgresql.SslMode, int(connectDialTimeout.Seconds()))
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
		dsn += "&_busy_timeout=30000&_txlock=immediate"
	} else {
		dsn += "?_busy_timeout=30000&_txlock=immediate"
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
