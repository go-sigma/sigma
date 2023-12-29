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

	"github.com/go-sigma/sigma/pkg/types/enums"

	_ "github.com/go-sql-driver/mysql"
)

func init() {
	checkers = append(checkers, checkRedis, checkDatabase, checkStorage)
}

func checkRedis(config Configuration) error {
	if config.Redis.Type == enums.RedisTypeNone {
		return nil
	}
	if config.Redis.Type != enums.RedisTypeExternal {
		return fmt.Errorf("Unknown redis type: %s", config.Redis.Type)
	}
	redisOpt, err := redis.ParseURL(config.Redis.Url)
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

func checkDatabase(config Configuration) error {
	dbType := config.Database.Type

	switch dbType {
	case enums.DatabaseMysql:
		return checkMysql(config)
	case enums.DatabasePostgresql:
		return checkPostgresql(config)
	case enums.DatabaseSqlite3:
		return nil
	default:
		return fmt.Errorf("unknown database type: %s", dbType)
	}
}

func checkMysql(config Configuration) error {
	host := config.Database.Mysql.Host
	port := config.Database.Mysql.Port
	user := config.Database.Mysql.User
	password := config.Database.Mysql.Password
	dbname := config.Database.Mysql.DBName

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", user, password, host, port, dbname) // TODO: query values
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

func checkPostgresql(config Configuration) error {
	host := config.Database.Postgresql.Host
	port := config.Database.Postgresql.Port
	user := config.Database.Postgresql.User
	password := config.Database.Postgresql.Password
	dbname := config.Database.Postgresql.DBName

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", user, password, host, port, dbname))
	if err != nil {
		return fmt.Errorf("pgx.Connect error: %v", err)
	}
	err = conn.Close(ctx)
	if err != nil {
		return fmt.Errorf("conn.Close error: %v", err)
	}
	return nil
}

func checkStorage(config Configuration) error {
	storageType := config.Storage.Type
	switch storageType {
	case enums.StorageTypeFilesystem:
		return nil
	case enums.StorageTypeS3:
		return checkStorageS3(config)
	default:
		return fmt.Errorf("Not support storage type")
	}
}

func checkStorageS3(config Configuration) error {
	endpoint := config.Storage.S3.Endpoint
	region := config.Storage.S3.Region
	ak := config.Storage.S3.Ak
	sk := config.Storage.S3.Sk
	bucket := config.Storage.S3.Bucket
	forcePathStyle := config.Storage.S3.ForcePathStyle

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
