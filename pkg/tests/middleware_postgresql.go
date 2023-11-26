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

package tests

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4"
	gonanoid "github.com/matoous/go-nanoid"
	"github.com/rs/zerolog/log"

	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/dal"
	"github.com/go-sigma/sigma/pkg/types/enums"
)

func init() {
	err := RegisterCIDatabaseFactory("postgresql", &postgresqlFactory{})
	if err != nil {
		panic(err)
	}
}

type postgresqlFactory struct{}

var _ Factory = &postgresqlFactory{}

func (postgresqlFactory) New() CIDatabase {
	return &postgresqlCIDatabase{}
}

type postgresqlCIDatabase struct {
	dbname string
}

var _ CIDatabase = &postgresqlCIDatabase{}

// Init sets the default values for the database configuration in ci tests
func (d *postgresqlCIDatabase) Init() error {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, "postgres://sigma:sigma@localhost:5432/?sslmode=disable")
	if err != nil {
		return err
	}
	d.dbname = gonanoid.MustGenerate("abcdefghijklmnopqrstuvwxyz", 6)

	_, err = conn.Exec(ctx, fmt.Sprintf("CREATE DATABASE \"%s\"", d.dbname))
	if err != nil {
		return err
	}
	err = conn.Close(ctx)
	if err != nil {
		return err
	}

	err = dal.Initialize(configs.Configuration{
		Database: configs.ConfigurationDatabase{
			Type: enums.DatabasePostgresql,
			Postgresql: configs.ConfigurationDatabasePostgresql{
				Host:     "127.0.0.1",
				Port:     5432,
				User:     "sigma",
				Password: "sigma",
				DBName:   d.dbname,
				SslMode:  "disable",
			},
		},
	})
	if err != nil {
		return err
	}
	return nil
}

// DeInit remove the database or database file for ci tests
func (d *postgresqlCIDatabase) DeInit() error {
	// For unknown reason, postgresql does not allow to drop the database
	log.Debug().Str("database", d.dbname).Msg("postgresql does not allow to drop the database, skipping")

	// ctx := context.Background()
	// conn, err := pgx.Connect(ctx, "postgres://sigma:sigma@localhost:5432/?sslmode=disable")
	// if err != nil {
	// 	return err
	// }
	// _, err = conn.Exec(ctx, fmt.Sprintf("DROP DATABASE \"%s\"", d.database))
	// if err != nil {
	// 	return err
	// }
	// err = conn.Close(ctx)
	// if err != nil {
	// 	return err
	// }
	return nil
}
