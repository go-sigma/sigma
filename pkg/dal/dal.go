// Copyright 2023 XImager
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

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/logger"
	"github.com/go-sigma/sigma/pkg/types/enums"
)

var (
	// DB is the global database connection
	DB *gorm.DB
	// RedisCli ...
	RedisCli redis.UniversalClient
)

// Initialize initializes the database connection
func Initialize() error {
	err := connectRedis()
	if err != nil {
		return err
	}
	pool := goredis.NewPool(RedisCli)
	rs := redsync.New(pool)
	mutex := rs.NewMutex(consts.LockerMigration, redsync.WithRetryDelay(time.Second*3), redsync.WithTries(10), redsync.WithExpiry(time.Second*30))
	err = mutex.Lock()
	if err != nil {
		return err
	}
	defer func() {
		if ok, err := mutex.Unlock(); !ok || err != nil {
			log.Fatal().Err(err).Msg("Unlock the migration key failed")
		}
	}()
	dbType := enums.MustParseDatabase(viper.GetString("database.type"))
	switch dbType {
	case enums.DatabaseMysql:
		err = connectMysql()
	case enums.DatabasePostgresql:
		err = connectPostgres()
	case enums.DatabaseSqlite3:
		err = connectSqlite3()
	default:
		return fmt.Errorf("unknown database type: %s", dbType)
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

	logLevel := viper.GetString("log.level")
	if logLevel == "debug" {
		query.SetDefault(DB.Debug())
	} else {
		query.SetDefault(DB)
	}
	return nil
}

func connectRedis() error {
	redisOpt, err := redis.ParseURL(viper.GetString("redis.url"))
	if err != nil {
		return err
	}
	RedisCli = redis.NewClient(redisOpt)
	return nil
}

func connectMysql() error {
	host := viper.GetString("database.mysql.host")
	port := viper.GetString("database.mysql.port")
	user := viper.GetString("database.mysql.user")
	password := viper.GetString("database.mysql.password")
	dbname := viper.GetString("database.mysql.dbname")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", user, password, host, port, dbname)
	log.Debug().Str("dsn", dsn).Msg("Connect to mysql database")

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.ZLogger{},
	})
	if err != nil {
		return err
	}
	db = db.WithContext(log.Logger.WithContext(context.Background()))
	DB = db

	err = migrateMysql(dsn)
	if err != nil {
		return err
	}

	return nil
}

func connectPostgres() error {
	host := viper.GetString("database.postgres.host")
	port := viper.GetString("database.postgres.port")
	user := viper.GetString("database.postgres.user")
	password := viper.GetString("database.postgres.password")
	dbname := viper.GetString("database.postgres.dbname")
	sslmode := viper.GetString("database.postgres.sslmode")

	dsn := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=%s", host, port, user, dbname, password, sslmode)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.ZLogger{},
	})
	if err != nil {
		return err
	}
	db = db.WithContext(log.Logger.WithContext(context.Background()))
	DB = db

	migrateDsn := fmt.Sprintf("%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port, dbname)
	err = migratePostgres(migrateDsn)
	if err != nil {
		return err
	}

	return nil
}

func connectSqlite3() error {
	dbname := viper.GetString("database.sqlite3.path")

	db, err := gorm.Open(sqlite.Open(dbname), &gorm.Config{
		Logger: logger.ZLogger{},
	})
	if err != nil {
		return err
	}
	db = db.WithContext(log.Logger.WithContext(context.Background()))
	DB = db

	err = migrateSqlite(dbname)
	if err != nil {
		return err
	}

	return nil
}
