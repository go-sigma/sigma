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
