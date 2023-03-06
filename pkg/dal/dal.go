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
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/ximager/ximager/pkg/dal/query"
)

var (
	// DB is the global database connection
	DB *gorm.DB
)

// Initialize initializes the database connection
func Initialize() error {
	var err error
	dbType := viper.GetString("database.type")
	switch dbType {
	case "mysql":
		err = connectMysql()
	case "postgresql":
		err = connectPostgres()
	case "sqlite":
		err = connectSqlite()
	default:
		return fmt.Errorf("unknown database type: %s", dbType)
	}
	if err != nil {
		return err
	}
	logLevel := viper.GetInt("log.level")
	if logLevel == 0 {
		query.SetDefault(DB.Debug())
	} else {
		query.SetDefault(DB)
	}
	return nil
}

func connectMysql() error {
	host := viper.GetString("database.mysql.host")
	port := viper.GetString("database.mysql.port")
	user := viper.GetString("database.mysql.user")
	password := viper.GetString("database.mysql.password")
	dbname := viper.GetString("database.mysql.database")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", user, password, host, port, dbname)
	log.Debug().Str("dsn", dsn).Msg("Connect to mysql database")

	err := migrateMysql(dsn)
	if err != nil {
		return err
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}
	DB = db

	return nil
}

func connectPostgres() error {
	host := viper.GetString("database.postgres.host")
	port := viper.GetString("database.postgres.port")
	user := viper.GetString("database.postgres.user")
	password := viper.GetString("database.postgres.password")
	dbname := viper.GetString("database.postgres.dbname")

	migrateDsn := fmt.Sprintf("%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port, dbname)
	err := migratePostgres(migrateDsn)
	if err != nil {
		return err
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=disable", host, port, user, dbname, password)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}
	DB = db

	return nil
}

func connectSqlite() error {
	dbname := viper.GetString("database.sqlite.path")

	db, err := gorm.Open(sqlite.Open(dbname), &gorm.Config{})
	if err != nil {
		return err
	}
	DB = db

	err = migrateSqlite(dbname)
	if err != nil {
		return err
	}

	return nil
}
