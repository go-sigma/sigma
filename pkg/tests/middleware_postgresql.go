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
	"github.com/rs/zerolog/log"
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/dal"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils"
)

func init() {
	err := registerCIDatabaseFactory("postgresql", &postgresqlFactory{})
	if err != nil {
		panic(err)
	}
}

type postgresqlFactory struct{}

var _ factory = &postgresqlFactory{}

func (postgresqlFactory) New() ciDatabase {
	return &postgresqlCIDatabase{}
}

type postgresqlCIDatabase struct {
	database string
}

var _ ciDatabase = &postgresqlCIDatabase{}

// Initialize sets the default values for the database configuration in ci tests
func (d *postgresqlCIDatabase) Initialize(digCon *dig.Container) error {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, "postgres://sigma:sigma@localhost:5432/?sslmode=disable")
	if err != nil {
		return err
	}
	d.database = utils.MustGetObjFromDigCon[configs.Configuration](digCon).Database.Postgresql.Database
	_, err = conn.Exec(ctx, fmt.Sprintf("CREATE DATABASE \"%s\"", d.database))
	if err != nil {
		return err
	}
	err = conn.Close(ctx)
	if err != nil {
		return err
	}
	return dal.Initialize(digCon)
}

// DeInitialize remove the database or database file for ci tests
func (d *postgresqlCIDatabase) DeInitialize() error {
	// For unknown reason, postgresql does not allow to drop the database
	log.Debug().Str("database", d.database).Msg("postgresql does not allow to drop the database, skipping")

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

// GetName get database name
func (d *postgresqlCIDatabase) GetName() enums.Database {
	return enums.DatabasePostgresql
}
