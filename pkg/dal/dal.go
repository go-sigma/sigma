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

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/logger"
	"github.com/go-sigma/sigma/pkg/modules/locker"
	"github.com/go-sigma/sigma/pkg/types/enums"
)

var (
	// DB is the global database connection
	DB *gorm.DB
)

// Initialize initializes the database connection
func Initialize(config configs.Configuration) error {
	var err error
	var dsn string
	// dbType := enums.MustParseDatabase(viper.GetString("database.type"))
	switch config.Database.Type {
	case enums.DatabaseMysql:
		dsn, err = connectMysql(config)
	case enums.DatabasePostgresql:
		dsn, err = connectPostgres(config)
	case enums.DatabaseSqlite3:
		dsn, err = connectSqlite3(config)
	default:
		return fmt.Errorf("unknown database type: %s", config.Database.Type)
	}
	if err != nil {
		return err
	}
	logLevel := viper.GetString("log.level")
	if logLevel == "debug" {
		query.SetDefault(DB.Debug())
	} else {
		query.SetDefault(DB)
	}

	err = DB.AutoMigrate(&models.Locker{})
	if err != nil {
		return err
	}

	locker, err := locker.New(config)
	if err != nil {
		return err
	}
	lock, err := locker.Lock(context.Background(), consts.LockerMigration, time.Second*30)
	if err != nil {
		return err
	}
	defer func() {
		err := lock.Unlock()
		if err != nil {
			log.Error().Err(err).Msg("Migrate locker release failed")
		}
	}()

	switch config.Database.Type {
	case enums.DatabaseMysql:
		err = migrateMysql(dsn)
	case enums.DatabasePostgresql:
		err = migratePostgres(dsn)
	case enums.DatabaseSqlite3:
		err = migrateSqlite(dsn)
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

	return nil
}

func connectMysql(config configs.Configuration) (string, error) {
	// host := viper.GetString("database.mysql.host")
	// port := viper.GetString("database.mysql.port")
	// user := viper.GetString("database.mysql.user")
	// password := viper.GetString("database.mysql.password")
	// dbname := viper.GetString("database.mysql.dbname")
	host := config.Database.Mysql.Host
	port := config.Database.Mysql.Port
	user := config.Database.Mysql.User
	password := config.Database.Mysql.Password
	dbname := config.Database.Mysql.DBName

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=UTC", user, password, host, port, dbname)
	log.Debug().Str("dsn", dsn).Msg("Connect to mysql database")

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
		Logger: logger.ZLogger{},
	})
	if err != nil {
		return "", err
	}
	db = db.WithContext(log.Logger.WithContext(context.Background()))
	DB = db

	return dsn, nil
}

func connectPostgres(config configs.Configuration) (string, error) {
	// host := viper.GetString("database.postgres.host")
	// port := viper.GetString("database.postgres.port")
	// user := viper.GetString("database.postgres.user")
	// password := viper.GetString("database.postgres.password")
	// dbname := viper.GetString("database.postgres.dbname")
	// sslmode := viper.GetString("database.postgres.sslmode")

	host := config.Database.Postgresql.Host
	port := config.Database.Postgresql.Port
	user := config.Database.Postgresql.User
	password := config.Database.Postgresql.Password
	dbname := config.Database.Postgresql.DBName
	sslmode := config.Database.Postgresql.SslMode

	dsn := fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s sslmode=%s", host, port, user, dbname, password, sslmode)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
		Logger: logger.ZLogger{},
	})
	if err != nil {
		return "", err
	}
	db = db.WithContext(log.Logger.WithContext(context.Background()))
	DB = db

	migrateDsn := fmt.Sprintf("%s:%s@%s:%d/%s?sslmode=disable", user, password, host, port, dbname)

	return migrateDsn, nil
}

func connectSqlite3(config configs.Configuration) (string, error) {
	// dbname := viper.GetString("database.sqlite3.path")
	dbname := config.Database.Sqlite3.Path

	db, err := gorm.Open(sqlite.Open(dbname), &gorm.Config{
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
		Logger: logger.ZLogger{},
	})
	if err != nil {
		return "", err
	}
	db = db.WithContext(log.Logger.WithContext(context.Background()))
	DB = db

	return dbname, nil
}
