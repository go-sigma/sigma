// Copyright 2025 sigma
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

package query

import (
	"context"
	"database/sql"

	"gorm.io/gorm"
)

//go:generate mockgen -destination=query_mocks.go -package=query github.com/go-sigma/sigma/pkg/dal/query IQuery

type IQuery interface {
	Available() bool
	Begin(opts ...*sql.TxOptions) *QueryTx
	ReadDB() *Query
	ReplaceDB(db *gorm.DB) *Query
	Transaction(fc func(tx *Query) error, opts ...*sql.TxOptions) error
	WithContext(ctx context.Context) *queryCtx
	WriteDB() *Query
	clone(db *gorm.DB) *Query
}
