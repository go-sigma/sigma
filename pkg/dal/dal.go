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
	"fmt"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/rs/zerolog/log"
	"go.uber.org/dig"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/logger"
	"github.com/go-sigma/sigma/pkg/modules/locker/definition"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils"
)

var (
	// DB is the global database connection
	DB *gorm.DB
)

// Initialize initializes the database connection
func Initialize(digCon *dig.Container) error {
	var err error
	var dsn string

	config := utils.MustGetObjFromDigCon[configs.Configuration](digCon)

	switch config.Database.Type {
	case enums.DatabaseMysql:
		dsn, err = connectMysql(config)
	case enums.DatabasePostgresql:
		dsn, err = connectPostgres(config)
	case enums.DatabaseSqlite3:
		err = connectSqlite3(config)
	default:
		return fmt.Errorf("unknown database type: %s", config.Database.Type)
	}

	if err != nil {
		return err
	}

	if config.Log.Level == enums.LogLevelDebug || config.Log.Level == enums.LogLevelTrace {
		query.SetDefault(DB.Debug())
	} else {
		query.SetDefault(DB)
	}

	ctx, ctxCancel := context.WithCancel(context.Background())
	defer ctxCancel()

	locker := utils.MustGetObjFromDigCon[definition.Locker](digCon)
	err = locker.AcquireWithRenew(ctx, consts.LockerMigration, time.Second*3, time.Second*5)
	if err != nil {
		return err
	}

	switch config.Database.Type {
	case enums.DatabaseMysql:
		err = migrateMysql(dsn)
	case enums.DatabasePostgresql:
		err = migratePostgres(dsn)
	case enums.DatabaseSqlite3:
		err = migrateSqlite()
	default:
		return fmt.Errorf("unknown database type: %s", config.Database.Type)
	}
	if err != nil {
		return err
	}

	err = setAuthModel(DB)
	if err != nil {
		return err
	}

	err = AuthEnforcer.LoadPolicy()
	if err != nil {
		return err
	}

	err = initDigContainer(digCon)
	if err != nil {
		return err
	}

	return nil
}

// DeInitialize ...
func DeInitialize() error {
	conn, err := DB.DB()
	if err != nil {
		return fmt.Errorf("get raw db instance failed: %v", err)
	}
	return conn.Close()
}

func connectMysql(config configs.Configuration) (string, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=UTC&multiStatements=true",
		config.Database.Mysql.Username, config.Database.Mysql.Password,
		config.Database.Mysql.Host, config.Database.Mysql.Port, config.Database.Mysql.Database)

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
		Logger: logger.ZLogger{},
	})
	if err != nil {
		return "", err
	}
	DB = DB.WithContext(log.Logger.WithContext(context.Background()))

	return dsn, nil
}

func connectPostgres(config configs.Configuration) (string, error) {
	dsn := fmt.Sprintf("%s:%s@%s:%d/%s?sslmode=%s", config.Database.Postgresql.Username,
		config.Database.Postgresql.Password, config.Database.Postgresql.Host,
		config.Database.Postgresql.Port, config.Database.Postgresql.Database,
		config.Database.Postgresql.SslMode)
	var err error
	DB, err = gorm.Open(postgres.Open("postgresql://"+dsn), &gorm.Config{
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
		Logger: logger.ZLogger{},
	})
	if err != nil {
		return "", err
	}
	DB = DB.WithContext(log.Logger.WithContext(context.Background()))
	return dsn, nil
}

func connectSqlite3(config configs.Configuration) error {
	dbname := config.Database.Sqlite3.Path

	var err error
	DB, err = gorm.Open(sqlite.Open("file:"+dbname+"?_busy_timeout=30000"), &gorm.Config{
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
		Logger: logger.ZLogger{},
	})
	if err != nil {
		return err
	}
	DB = DB.WithContext(log.Logger.WithContext(context.Background()))

	rawDB, err := DB.DB()
	if err != nil {
		return err
	}
	rawDB.SetMaxOpenConns(10)
	rawDB.SetMaxIdleConns(3)
	rawDB.SetConnMaxIdleTime(time.Hour)
	rawDB.SetConnMaxLifetime(time.Hour)

	return nil
}
