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
	"os"

	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/dal"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils"
)

func init() {
	err := registerCIDatabaseFactory("sqlite3", &sqlite3Factory{})
	if err != nil {
		panic(err)
	}
}

type sqlite3Factory struct{}

var _ factory = &sqlite3Factory{}

func (sqlite3Factory) New() ciDatabase {
	return &sqlite3CIDatabase{}
}

type sqlite3CIDatabase struct {
	path string
}

var _ ciDatabase = &sqlite3CIDatabase{}

// Initialize sets the default values for the database configuration in ci tests
func (d *sqlite3CIDatabase) Initialize(digCon *dig.Container) error {
	d.path = utils.MustGetObjFromDigCon[configs.Configuration](digCon).Database.Sqlite3.Path
	return dal.Initialize(digCon)
}

// DeInitialize remove the database or database file for ci tests
func (d *sqlite3CIDatabase) DeInitialize() error {
	err := os.Remove(d.path)
	if err != nil {
		return err
	}
	return nil
}

// GetName get database name
func (d *sqlite3CIDatabase) GetName() enums.Database {
	return enums.DatabaseSqlite3
}
