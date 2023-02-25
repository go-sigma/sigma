// The MIT License (MIT)
//
// Copyright © 2023 Tosone <i@tosone.cn>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

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
	case "postgres":
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
	dbname := viper.GetString("database.sqlite.dbname")

	db, err := gorm.Open(sqlite.Open(dbname), &gorm.Config{})
	if err != nil {
		return err
	}
	DB = db

	return nil
}
