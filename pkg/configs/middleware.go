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

package configs

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/jackc/pgx/v4"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"

	"github.com/go-sigma/sigma/pkg/types/enums"

	_ "github.com/go-sigma/sigma/pkg/dal"
)

func init() {
	checkers = append(checkers, checkRedis, checkDatabase, checkStorage)
}

func checkRedis() error {
	redisOpt, err := redis.ParseURL(viper.GetString("redis.url"))
	if err != nil {
		return fmt.Errorf("redis.ParseURL error: %v", err)
	}
	redisCli := redis.NewClient(redisOpt)
	err = redisCli.Ping(context.Background()).Err()
	if err != nil {
		return fmt.Errorf("redis.Ping error: %v", err)
	}
	err = redisCli.Close()
	if err != nil {
		return fmt.Errorf("redis.Close error: %v", err)
	}
	return nil
}

func checkDatabase() error {
	dbType := viper.GetString("database.type")

	typ, err := enums.ParseDatabase(dbType)
	if err != nil {
		return fmt.Errorf("database type is invalid, just support: %s, %s, %s", enums.DatabasePostgresql, enums.DatabaseMysql, enums.DatabaseSqlite3)
	}

	switch typ {
	case enums.DatabaseMysql:
		return checkMysql()
	case enums.DatabasePostgresql:
		return checkPostgresql()
	case enums.DatabaseSqlite3:
		return nil
	default:
		return fmt.Errorf("unknown database type: %s", dbType)
	}
}

func checkMysql() error {
	host := viper.GetString("database.mysql.host")
	port := viper.GetString("database.mysql.port")
	user := viper.GetString("database.mysql.user")
	password := viper.GetString("database.mysql.password")
	dbname := viper.GetString("database.mysql.database")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", user, password, host, port, dbname)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("sql.Open error: %v", err)
	}
	err = db.Ping()
	if err != nil {
		return fmt.Errorf("db.Ping error: %v", err)
	}
	err = db.Close()
	if err != nil {
		return fmt.Errorf("db.Close error: %v", err)
	}
	return nil
}

func checkPostgresql() error {
	host := viper.GetString("database.postgres.host")
	port := viper.GetString("database.postgres.port")
	user := viper.GetString("database.postgres.user")
	password := viper.GetString("database.postgres.password")
	dbname := viper.GetString("database.postgres.dbname")

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port, dbname))
	if err != nil {
		return fmt.Errorf("pgx.Connect error: %v", err)
	}
	err = conn.Close(ctx)
	if err != nil {
		return fmt.Errorf("conn.Close error: %v", err)
	}
	return nil
}

func checkStorage() error {
	switch viper.GetString("storage.type") {
	case "filesystem":
		return nil
	case "s3":
		return checkStorageS3()
	default:
		return fmt.Errorf("Not support storage type")
	}
}

func checkStorageS3() error {
	endpoint := viper.GetString("storage.s3.endpoint")
	region := viper.GetString("storage.s3.region")
	ak := viper.GetString("storage.s3.ak")
	sk := viper.GetString("storage.s3.sk")
	bucket := viper.GetString("storage.s3.bucket")
	forcePathStyle := viper.GetBool("storage.s3.forcePathStyle")

	sess, err := session.NewSession(&aws.Config{
		Endpoint:         aws.String(endpoint),
		Region:           aws.String(region),
		S3ForcePathStyle: aws.Bool(forcePathStyle),
		Credentials:      credentials.NewStaticCredentials(ak, sk, ""),
	})
	if err != nil {
		return fmt.Errorf("failed to create new session with aws config: %v", err)
	}
	s3Cli := s3.New(sess)
	_, err = s3Cli.HeadBucket(&s3.HeadBucketInput{Bucket: aws.String(bucket)})
	if err != nil {
		return fmt.Errorf("s3.HeadBucket error: %v", err)
	}
	return nil
}
