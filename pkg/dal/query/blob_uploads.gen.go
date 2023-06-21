// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.

package query

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"

	"gorm.io/gen"
	"gorm.io/gen/field"

	"gorm.io/plugin/dbresolver"

	"github.com/ximager/ximager/pkg/dal/models"
)

func newBlobUpload(db *gorm.DB, opts ...gen.DOOption) blobUpload {
	_blobUpload := blobUpload{}

	_blobUpload.blobUploadDo.UseDB(db, opts...)
	_blobUpload.blobUploadDo.UseModel(&models.BlobUpload{})

	tableName := _blobUpload.blobUploadDo.TableName()
	_blobUpload.ALL = field.NewAsterisk(tableName)
	_blobUpload.CreatedAt = field.NewTime(tableName, "created_at")
	_blobUpload.UpdatedAt = field.NewTime(tableName, "updated_at")
	_blobUpload.DeletedAt = field.NewUint(tableName, "deleted_at")
	_blobUpload.ID = field.NewInt64(tableName, "id")
	_blobUpload.PartNumber = field.NewInt(tableName, "part_number")
	_blobUpload.UploadID = field.NewString(tableName, "upload_id")
	_blobUpload.Etag = field.NewString(tableName, "etag")
	_blobUpload.Repository = field.NewString(tableName, "repository")
	_blobUpload.FileID = field.NewString(tableName, "file_id")
	_blobUpload.Size = field.NewInt64(tableName, "size")

	_blobUpload.fillFieldMap()

	return _blobUpload
}

type blobUpload struct {
	blobUploadDo blobUploadDo

	ALL        field.Asterisk
	CreatedAt  field.Time
	UpdatedAt  field.Time
	DeletedAt  field.Uint
	ID         field.Int64
	PartNumber field.Int
	UploadID   field.String
	Etag       field.String
	Repository field.String
	FileID     field.String
	Size       field.Int64

	fieldMap map[string]field.Expr
}

func (b blobUpload) Table(newTableName string) *blobUpload {
	b.blobUploadDo.UseTable(newTableName)
	return b.updateTableName(newTableName)
}

func (b blobUpload) As(alias string) *blobUpload {
	b.blobUploadDo.DO = *(b.blobUploadDo.As(alias).(*gen.DO))
	return b.updateTableName(alias)
}

func (b *blobUpload) updateTableName(table string) *blobUpload {
	b.ALL = field.NewAsterisk(table)
	b.CreatedAt = field.NewTime(table, "created_at")
	b.UpdatedAt = field.NewTime(table, "updated_at")
	b.DeletedAt = field.NewUint(table, "deleted_at")
	b.ID = field.NewInt64(table, "id")
	b.PartNumber = field.NewInt(table, "part_number")
	b.UploadID = field.NewString(table, "upload_id")
	b.Etag = field.NewString(table, "etag")
	b.Repository = field.NewString(table, "repository")
	b.FileID = field.NewString(table, "file_id")
	b.Size = field.NewInt64(table, "size")

	b.fillFieldMap()

	return b
}

func (b *blobUpload) WithContext(ctx context.Context) *blobUploadDo {
	return b.blobUploadDo.WithContext(ctx)
}

func (b blobUpload) TableName() string { return b.blobUploadDo.TableName() }

func (b blobUpload) Alias() string { return b.blobUploadDo.Alias() }

func (b *blobUpload) GetFieldByName(fieldName string) (field.OrderExpr, bool) {
	_f, ok := b.fieldMap[fieldName]
	if !ok || _f == nil {
		return nil, false
	}
	_oe, ok := _f.(field.OrderExpr)
	return _oe, ok
}

func (b *blobUpload) fillFieldMap() {
	b.fieldMap = make(map[string]field.Expr, 10)
	b.fieldMap["created_at"] = b.CreatedAt
	b.fieldMap["updated_at"] = b.UpdatedAt
	b.fieldMap["deleted_at"] = b.DeletedAt
	b.fieldMap["id"] = b.ID
	b.fieldMap["part_number"] = b.PartNumber
	b.fieldMap["upload_id"] = b.UploadID
	b.fieldMap["etag"] = b.Etag
	b.fieldMap["repository"] = b.Repository
	b.fieldMap["file_id"] = b.FileID
	b.fieldMap["size"] = b.Size
}

func (b blobUpload) clone(db *gorm.DB) blobUpload {
	b.blobUploadDo.ReplaceConnPool(db.Statement.ConnPool)
	return b
}

func (b blobUpload) replaceDB(db *gorm.DB) blobUpload {
	b.blobUploadDo.ReplaceDB(db)
	return b
}

type blobUploadDo struct{ gen.DO }

func (b blobUploadDo) Debug() *blobUploadDo {
	return b.withDO(b.DO.Debug())
}

func (b blobUploadDo) WithContext(ctx context.Context) *blobUploadDo {
	return b.withDO(b.DO.WithContext(ctx))
}

func (b blobUploadDo) ReadDB() *blobUploadDo {
	return b.Clauses(dbresolver.Read)
}

func (b blobUploadDo) WriteDB() *blobUploadDo {
	return b.Clauses(dbresolver.Write)
}

func (b blobUploadDo) Session(config *gorm.Session) *blobUploadDo {
	return b.withDO(b.DO.Session(config))
}

func (b blobUploadDo) Clauses(conds ...clause.Expression) *blobUploadDo {
	return b.withDO(b.DO.Clauses(conds...))
}

func (b blobUploadDo) Returning(value interface{}, columns ...string) *blobUploadDo {
	return b.withDO(b.DO.Returning(value, columns...))
}

func (b blobUploadDo) Not(conds ...gen.Condition) *blobUploadDo {
	return b.withDO(b.DO.Not(conds...))
}

func (b blobUploadDo) Or(conds ...gen.Condition) *blobUploadDo {
	return b.withDO(b.DO.Or(conds...))
}

func (b blobUploadDo) Select(conds ...field.Expr) *blobUploadDo {
	return b.withDO(b.DO.Select(conds...))
}

func (b blobUploadDo) Where(conds ...gen.Condition) *blobUploadDo {
	return b.withDO(b.DO.Where(conds...))
}

func (b blobUploadDo) Exists(subquery interface{ UnderlyingDB() *gorm.DB }) *blobUploadDo {
	return b.Where(field.CompareSubQuery(field.ExistsOp, nil, subquery.UnderlyingDB()))
}

func (b blobUploadDo) Order(conds ...field.Expr) *blobUploadDo {
	return b.withDO(b.DO.Order(conds...))
}

func (b blobUploadDo) Distinct(cols ...field.Expr) *blobUploadDo {
	return b.withDO(b.DO.Distinct(cols...))
}

func (b blobUploadDo) Omit(cols ...field.Expr) *blobUploadDo {
	return b.withDO(b.DO.Omit(cols...))
}

func (b blobUploadDo) Join(table schema.Tabler, on ...field.Expr) *blobUploadDo {
	return b.withDO(b.DO.Join(table, on...))
}

func (b blobUploadDo) LeftJoin(table schema.Tabler, on ...field.Expr) *blobUploadDo {
	return b.withDO(b.DO.LeftJoin(table, on...))
}

func (b blobUploadDo) RightJoin(table schema.Tabler, on ...field.Expr) *blobUploadDo {
	return b.withDO(b.DO.RightJoin(table, on...))
}

func (b blobUploadDo) Group(cols ...field.Expr) *blobUploadDo {
	return b.withDO(b.DO.Group(cols...))
}

func (b blobUploadDo) Having(conds ...gen.Condition) *blobUploadDo {
	return b.withDO(b.DO.Having(conds...))
}

func (b blobUploadDo) Limit(limit int) *blobUploadDo {
	return b.withDO(b.DO.Limit(limit))
}

func (b blobUploadDo) Offset(offset int) *blobUploadDo {
	return b.withDO(b.DO.Offset(offset))
}

func (b blobUploadDo) Scopes(funcs ...func(gen.Dao) gen.Dao) *blobUploadDo {
	return b.withDO(b.DO.Scopes(funcs...))
}

func (b blobUploadDo) Unscoped() *blobUploadDo {
	return b.withDO(b.DO.Unscoped())
}

func (b blobUploadDo) Create(values ...*models.BlobUpload) error {
	if len(values) == 0 {
		return nil
	}
	return b.DO.Create(values)
}

func (b blobUploadDo) CreateInBatches(values []*models.BlobUpload, batchSize int) error {
	return b.DO.CreateInBatches(values, batchSize)
}

// Save : !!! underlying implementation is different with GORM
// The method is equivalent to executing the statement: db.Clauses(clause.OnConflict{UpdateAll: true}).Create(values)
func (b blobUploadDo) Save(values ...*models.BlobUpload) error {
	if len(values) == 0 {
		return nil
	}
	return b.DO.Save(values)
}

func (b blobUploadDo) First() (*models.BlobUpload, error) {
	if result, err := b.DO.First(); err != nil {
		return nil, err
	} else {
		return result.(*models.BlobUpload), nil
	}
}

func (b blobUploadDo) Take() (*models.BlobUpload, error) {
	if result, err := b.DO.Take(); err != nil {
		return nil, err
	} else {
		return result.(*models.BlobUpload), nil
	}
}

func (b blobUploadDo) Last() (*models.BlobUpload, error) {
	if result, err := b.DO.Last(); err != nil {
		return nil, err
	} else {
		return result.(*models.BlobUpload), nil
	}
}

func (b blobUploadDo) Find() ([]*models.BlobUpload, error) {
	result, err := b.DO.Find()
	return result.([]*models.BlobUpload), err
}

func (b blobUploadDo) FindInBatch(batchSize int, fc func(tx gen.Dao, batch int) error) (results []*models.BlobUpload, err error) {
	buf := make([]*models.BlobUpload, 0, batchSize)
	err = b.DO.FindInBatches(&buf, batchSize, func(tx gen.Dao, batch int) error {
		defer func() { results = append(results, buf...) }()
		return fc(tx, batch)
	})
	return results, err
}

func (b blobUploadDo) FindInBatches(result *[]*models.BlobUpload, batchSize int, fc func(tx gen.Dao, batch int) error) error {
	return b.DO.FindInBatches(result, batchSize, fc)
}

func (b blobUploadDo) Attrs(attrs ...field.AssignExpr) *blobUploadDo {
	return b.withDO(b.DO.Attrs(attrs...))
}

func (b blobUploadDo) Assign(attrs ...field.AssignExpr) *blobUploadDo {
	return b.withDO(b.DO.Assign(attrs...))
}

func (b blobUploadDo) Joins(fields ...field.RelationField) *blobUploadDo {
	for _, _f := range fields {
		b = *b.withDO(b.DO.Joins(_f))
	}
	return &b
}

func (b blobUploadDo) Preload(fields ...field.RelationField) *blobUploadDo {
	for _, _f := range fields {
		b = *b.withDO(b.DO.Preload(_f))
	}
	return &b
}

func (b blobUploadDo) FirstOrInit() (*models.BlobUpload, error) {
	if result, err := b.DO.FirstOrInit(); err != nil {
		return nil, err
	} else {
		return result.(*models.BlobUpload), nil
	}
}

func (b blobUploadDo) FirstOrCreate() (*models.BlobUpload, error) {
	if result, err := b.DO.FirstOrCreate(); err != nil {
		return nil, err
	} else {
		return result.(*models.BlobUpload), nil
	}
}

func (b blobUploadDo) FindByPage(offset int, limit int) (result []*models.BlobUpload, count int64, err error) {
	result, err = b.Offset(offset).Limit(limit).Find()
	if err != nil {
		return
	}

	if size := len(result); 0 < limit && 0 < size && size < limit {
		count = int64(size + offset)
		return
	}

	count, err = b.Offset(-1).Limit(-1).Count()
	return
}

func (b blobUploadDo) ScanByPage(result interface{}, offset int, limit int) (count int64, err error) {
	count, err = b.Count()
	if err != nil {
		return
	}

	err = b.Offset(offset).Limit(limit).Scan(result)
	return
}

func (b blobUploadDo) Scan(result interface{}) (err error) {
	return b.DO.Scan(result)
}

func (b blobUploadDo) Delete(models ...*models.BlobUpload) (result gen.ResultInfo, err error) {
	return b.DO.Delete(models)
}

func (b *blobUploadDo) withDO(do gen.Dao) *blobUploadDo {
	b.DO = *do.(*gen.DO)
	return b
}
