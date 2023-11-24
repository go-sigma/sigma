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

	"github.com/go-sigma/sigma/pkg/dal/models"
)

func newNamespace(db *gorm.DB, opts ...gen.DOOption) namespace {
	_namespace := namespace{}

	_namespace.namespaceDo.UseDB(db, opts...)
	_namespace.namespaceDo.UseModel(&models.Namespace{})

	tableName := _namespace.namespaceDo.TableName()
	_namespace.ALL = field.NewAsterisk(tableName)
	_namespace.CreatedAt = field.NewInt64(tableName, "created_at")
	_namespace.UpdatedAt = field.NewInt64(tableName, "updated_at")
	_namespace.DeletedAt = field.NewUint64(tableName, "deleted_at")
	_namespace.ID = field.NewInt64(tableName, "id")
	_namespace.Name = field.NewString(tableName, "name")
	_namespace.Description = field.NewString(tableName, "description")
	_namespace.Visibility = field.NewField(tableName, "visibility")
	_namespace.TagLimit = field.NewInt64(tableName, "tag_limit")
	_namespace.TagCount = field.NewInt64(tableName, "tag_count")
	_namespace.RepositoryLimit = field.NewInt64(tableName, "repository_limit")
	_namespace.RepositoryCount = field.NewInt64(tableName, "repository_count")
	_namespace.SizeLimit = field.NewInt64(tableName, "size_limit")
	_namespace.Size = field.NewInt64(tableName, "size")

	_namespace.fillFieldMap()

	return _namespace
}

type namespace struct {
	namespaceDo namespaceDo

	ALL             field.Asterisk
	CreatedAt       field.Int64
	UpdatedAt       field.Int64
	DeletedAt       field.Uint64
	ID              field.Int64
	Name            field.String
	Description     field.String
	Visibility      field.Field
	TagLimit        field.Int64
	TagCount        field.Int64
	RepositoryLimit field.Int64
	RepositoryCount field.Int64
	SizeLimit       field.Int64
	Size            field.Int64

	fieldMap map[string]field.Expr
}

func (n namespace) Table(newTableName string) *namespace {
	n.namespaceDo.UseTable(newTableName)
	return n.updateTableName(newTableName)
}

func (n namespace) As(alias string) *namespace {
	n.namespaceDo.DO = *(n.namespaceDo.As(alias).(*gen.DO))
	return n.updateTableName(alias)
}

func (n *namespace) updateTableName(table string) *namespace {
	n.ALL = field.NewAsterisk(table)
	n.CreatedAt = field.NewInt64(table, "created_at")
	n.UpdatedAt = field.NewInt64(table, "updated_at")
	n.DeletedAt = field.NewUint64(table, "deleted_at")
	n.ID = field.NewInt64(table, "id")
	n.Name = field.NewString(table, "name")
	n.Description = field.NewString(table, "description")
	n.Visibility = field.NewField(table, "visibility")
	n.TagLimit = field.NewInt64(table, "tag_limit")
	n.TagCount = field.NewInt64(table, "tag_count")
	n.RepositoryLimit = field.NewInt64(table, "repository_limit")
	n.RepositoryCount = field.NewInt64(table, "repository_count")
	n.SizeLimit = field.NewInt64(table, "size_limit")
	n.Size = field.NewInt64(table, "size")

	n.fillFieldMap()

	return n
}

func (n *namespace) WithContext(ctx context.Context) *namespaceDo {
	return n.namespaceDo.WithContext(ctx)
}

func (n namespace) TableName() string { return n.namespaceDo.TableName() }

func (n namespace) Alias() string { return n.namespaceDo.Alias() }

func (n namespace) Columns(cols ...field.Expr) gen.Columns { return n.namespaceDo.Columns(cols...) }

func (n *namespace) GetFieldByName(fieldName string) (field.OrderExpr, bool) {
	_f, ok := n.fieldMap[fieldName]
	if !ok || _f == nil {
		return nil, false
	}
	_oe, ok := _f.(field.OrderExpr)
	return _oe, ok
}

func (n *namespace) fillFieldMap() {
	n.fieldMap = make(map[string]field.Expr, 13)
	n.fieldMap["created_at"] = n.CreatedAt
	n.fieldMap["updated_at"] = n.UpdatedAt
	n.fieldMap["deleted_at"] = n.DeletedAt
	n.fieldMap["id"] = n.ID
	n.fieldMap["name"] = n.Name
	n.fieldMap["description"] = n.Description
	n.fieldMap["visibility"] = n.Visibility
	n.fieldMap["tag_limit"] = n.TagLimit
	n.fieldMap["tag_count"] = n.TagCount
	n.fieldMap["repository_limit"] = n.RepositoryLimit
	n.fieldMap["repository_count"] = n.RepositoryCount
	n.fieldMap["size_limit"] = n.SizeLimit
	n.fieldMap["size"] = n.Size
}

func (n namespace) clone(db *gorm.DB) namespace {
	n.namespaceDo.ReplaceConnPool(db.Statement.ConnPool)
	return n
}

func (n namespace) replaceDB(db *gorm.DB) namespace {
	n.namespaceDo.ReplaceDB(db)
	return n
}

type namespaceDo struct{ gen.DO }

func (n namespaceDo) Debug() *namespaceDo {
	return n.withDO(n.DO.Debug())
}

func (n namespaceDo) WithContext(ctx context.Context) *namespaceDo {
	return n.withDO(n.DO.WithContext(ctx))
}

func (n namespaceDo) ReadDB() *namespaceDo {
	return n.Clauses(dbresolver.Read)
}

func (n namespaceDo) WriteDB() *namespaceDo {
	return n.Clauses(dbresolver.Write)
}

func (n namespaceDo) Session(config *gorm.Session) *namespaceDo {
	return n.withDO(n.DO.Session(config))
}

func (n namespaceDo) Clauses(conds ...clause.Expression) *namespaceDo {
	return n.withDO(n.DO.Clauses(conds...))
}

func (n namespaceDo) Returning(value interface{}, columns ...string) *namespaceDo {
	return n.withDO(n.DO.Returning(value, columns...))
}

func (n namespaceDo) Not(conds ...gen.Condition) *namespaceDo {
	return n.withDO(n.DO.Not(conds...))
}

func (n namespaceDo) Or(conds ...gen.Condition) *namespaceDo {
	return n.withDO(n.DO.Or(conds...))
}

func (n namespaceDo) Select(conds ...field.Expr) *namespaceDo {
	return n.withDO(n.DO.Select(conds...))
}

func (n namespaceDo) Where(conds ...gen.Condition) *namespaceDo {
	return n.withDO(n.DO.Where(conds...))
}

func (n namespaceDo) Order(conds ...field.Expr) *namespaceDo {
	return n.withDO(n.DO.Order(conds...))
}

func (n namespaceDo) Distinct(cols ...field.Expr) *namespaceDo {
	return n.withDO(n.DO.Distinct(cols...))
}

func (n namespaceDo) Omit(cols ...field.Expr) *namespaceDo {
	return n.withDO(n.DO.Omit(cols...))
}

func (n namespaceDo) Join(table schema.Tabler, on ...field.Expr) *namespaceDo {
	return n.withDO(n.DO.Join(table, on...))
}

func (n namespaceDo) LeftJoin(table schema.Tabler, on ...field.Expr) *namespaceDo {
	return n.withDO(n.DO.LeftJoin(table, on...))
}

func (n namespaceDo) RightJoin(table schema.Tabler, on ...field.Expr) *namespaceDo {
	return n.withDO(n.DO.RightJoin(table, on...))
}

func (n namespaceDo) Group(cols ...field.Expr) *namespaceDo {
	return n.withDO(n.DO.Group(cols...))
}

func (n namespaceDo) Having(conds ...gen.Condition) *namespaceDo {
	return n.withDO(n.DO.Having(conds...))
}

func (n namespaceDo) Limit(limit int) *namespaceDo {
	return n.withDO(n.DO.Limit(limit))
}

func (n namespaceDo) Offset(offset int) *namespaceDo {
	return n.withDO(n.DO.Offset(offset))
}

func (n namespaceDo) Scopes(funcs ...func(gen.Dao) gen.Dao) *namespaceDo {
	return n.withDO(n.DO.Scopes(funcs...))
}

func (n namespaceDo) Unscoped() *namespaceDo {
	return n.withDO(n.DO.Unscoped())
}

func (n namespaceDo) Create(values ...*models.Namespace) error {
	if len(values) == 0 {
		return nil
	}
	return n.DO.Create(values)
}

func (n namespaceDo) CreateInBatches(values []*models.Namespace, batchSize int) error {
	return n.DO.CreateInBatches(values, batchSize)
}

// Save : !!! underlying implementation is different with GORM
// The method is equivalent to executing the statement: db.Clauses(clause.OnConflict{UpdateAll: true}).Create(values)
func (n namespaceDo) Save(values ...*models.Namespace) error {
	if len(values) == 0 {
		return nil
	}
	return n.DO.Save(values)
}

func (n namespaceDo) First() (*models.Namespace, error) {
	if result, err := n.DO.First(); err != nil {
		return nil, err
	} else {
		return result.(*models.Namespace), nil
	}
}

func (n namespaceDo) Take() (*models.Namespace, error) {
	if result, err := n.DO.Take(); err != nil {
		return nil, err
	} else {
		return result.(*models.Namespace), nil
	}
}

func (n namespaceDo) Last() (*models.Namespace, error) {
	if result, err := n.DO.Last(); err != nil {
		return nil, err
	} else {
		return result.(*models.Namespace), nil
	}
}

func (n namespaceDo) Find() ([]*models.Namespace, error) {
	result, err := n.DO.Find()
	return result.([]*models.Namespace), err
}

func (n namespaceDo) FindInBatch(batchSize int, fc func(tx gen.Dao, batch int) error) (results []*models.Namespace, err error) {
	buf := make([]*models.Namespace, 0, batchSize)
	err = n.DO.FindInBatches(&buf, batchSize, func(tx gen.Dao, batch int) error {
		defer func() { results = append(results, buf...) }()
		return fc(tx, batch)
	})
	return results, err
}

func (n namespaceDo) FindInBatches(result *[]*models.Namespace, batchSize int, fc func(tx gen.Dao, batch int) error) error {
	return n.DO.FindInBatches(result, batchSize, fc)
}

func (n namespaceDo) Attrs(attrs ...field.AssignExpr) *namespaceDo {
	return n.withDO(n.DO.Attrs(attrs...))
}

func (n namespaceDo) Assign(attrs ...field.AssignExpr) *namespaceDo {
	return n.withDO(n.DO.Assign(attrs...))
}

func (n namespaceDo) Joins(fields ...field.RelationField) *namespaceDo {
	for _, _f := range fields {
		n = *n.withDO(n.DO.Joins(_f))
	}
	return &n
}

func (n namespaceDo) Preload(fields ...field.RelationField) *namespaceDo {
	for _, _f := range fields {
		n = *n.withDO(n.DO.Preload(_f))
	}
	return &n
}

func (n namespaceDo) FirstOrInit() (*models.Namespace, error) {
	if result, err := n.DO.FirstOrInit(); err != nil {
		return nil, err
	} else {
		return result.(*models.Namespace), nil
	}
}

func (n namespaceDo) FirstOrCreate() (*models.Namespace, error) {
	if result, err := n.DO.FirstOrCreate(); err != nil {
		return nil, err
	} else {
		return result.(*models.Namespace), nil
	}
}

func (n namespaceDo) FindByPage(offset int, limit int) (result []*models.Namespace, count int64, err error) {
	result, err = n.Offset(offset).Limit(limit).Find()
	if err != nil {
		return
	}

	if size := len(result); 0 < limit && 0 < size && size < limit {
		count = int64(size + offset)
		return
	}

	count, err = n.Offset(-1).Limit(-1).Count()
	return
}

func (n namespaceDo) ScanByPage(result interface{}, offset int, limit int) (count int64, err error) {
	count, err = n.Count()
	if err != nil {
		return
	}

	err = n.Offset(offset).Limit(limit).Scan(result)
	return
}

func (n namespaceDo) Scan(result interface{}) (err error) {
	return n.DO.Scan(result)
}

func (n namespaceDo) Delete(models ...*models.Namespace) (result gen.ResultInfo, err error) {
	return n.DO.Delete(models)
}

func (n *namespaceDo) withDO(do gen.Dao) *namespaceDo {
	n.DO = *do.(*gen.DO)
	return n
}
