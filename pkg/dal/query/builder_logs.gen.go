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

func newBuilderLog(db *gorm.DB, opts ...gen.DOOption) builderLog {
	_builderLog := builderLog{}

	_builderLog.builderLogDo.UseDB(db, opts...)
	_builderLog.builderLogDo.UseModel(&models.BuilderRunner{})

	tableName := _builderLog.builderLogDo.TableName()
	_builderLog.ALL = field.NewAsterisk(tableName)
	_builderLog.CreatedAt = field.NewTime(tableName, "created_at")
	_builderLog.UpdatedAt = field.NewTime(tableName, "updated_at")
	_builderLog.DeletedAt = field.NewUint(tableName, "deleted_at")
	_builderLog.ID = field.NewInt64(tableName, "id")
	_builderLog.BuilderID = field.NewInt64(tableName, "builder_id")
	_builderLog.Log = field.NewBytes(tableName, "log")
	_builderLog.Status = field.NewField(tableName, "status")
	_builderLog.Builder = builderLogBelongsToBuilder{
		db: db.Session(&gorm.Session{}),

		RelationField: field.NewRelation("Builder", "models.Builder"),
		Repository: struct {
			field.RelationField
			Namespace struct {
				field.RelationField
			}
		}{
			RelationField: field.NewRelation("Builder.Repository", "models.Repository"),
			Namespace: struct {
				field.RelationField
			}{
				RelationField: field.NewRelation("Builder.Repository.Namespace", "models.Namespace"),
			},
		},
	}

	_builderLog.fillFieldMap()

	return _builderLog
}

type builderLog struct {
	builderLogDo builderLogDo

	ALL       field.Asterisk
	CreatedAt field.Time
	UpdatedAt field.Time
	DeletedAt field.Uint
	ID        field.Int64
	BuilderID field.Int64
	Log       field.Bytes
	Status    field.Field
	Builder   builderLogBelongsToBuilder

	fieldMap map[string]field.Expr
}

func (b builderLog) Table(newTableName string) *builderLog {
	b.builderLogDo.UseTable(newTableName)
	return b.updateTableName(newTableName)
}

func (b builderLog) As(alias string) *builderLog {
	b.builderLogDo.DO = *(b.builderLogDo.As(alias).(*gen.DO))
	return b.updateTableName(alias)
}

func (b *builderLog) updateTableName(table string) *builderLog {
	b.ALL = field.NewAsterisk(table)
	b.CreatedAt = field.NewTime(table, "created_at")
	b.UpdatedAt = field.NewTime(table, "updated_at")
	b.DeletedAt = field.NewUint(table, "deleted_at")
	b.ID = field.NewInt64(table, "id")
	b.BuilderID = field.NewInt64(table, "builder_id")
	b.Log = field.NewBytes(table, "log")
	b.Status = field.NewField(table, "status")

	b.fillFieldMap()

	return b
}

func (b *builderLog) WithContext(ctx context.Context) *builderLogDo {
	return b.builderLogDo.WithContext(ctx)
}

func (b builderLog) TableName() string { return b.builderLogDo.TableName() }

func (b builderLog) Alias() string { return b.builderLogDo.Alias() }

func (b builderLog) Columns(cols ...field.Expr) gen.Columns { return b.builderLogDo.Columns(cols...) }

func (b *builderLog) GetFieldByName(fieldName string) (field.OrderExpr, bool) {
	_f, ok := b.fieldMap[fieldName]
	if !ok || _f == nil {
		return nil, false
	}
	_oe, ok := _f.(field.OrderExpr)
	return _oe, ok
}

func (b *builderLog) fillFieldMap() {
	b.fieldMap = make(map[string]field.Expr, 8)
	b.fieldMap["created_at"] = b.CreatedAt
	b.fieldMap["updated_at"] = b.UpdatedAt
	b.fieldMap["deleted_at"] = b.DeletedAt
	b.fieldMap["id"] = b.ID
	b.fieldMap["builder_id"] = b.BuilderID
	b.fieldMap["log"] = b.Log
	b.fieldMap["status"] = b.Status

}

func (b builderLog) clone(db *gorm.DB) builderLog {
	b.builderLogDo.ReplaceConnPool(db.Statement.ConnPool)
	return b
}

func (b builderLog) replaceDB(db *gorm.DB) builderLog {
	b.builderLogDo.ReplaceDB(db)
	return b
}

type builderLogBelongsToBuilder struct {
	db *gorm.DB

	field.RelationField

	Repository struct {
		field.RelationField
		Namespace struct {
			field.RelationField
		}
	}
}

func (a builderLogBelongsToBuilder) Where(conds ...field.Expr) *builderLogBelongsToBuilder {
	if len(conds) == 0 {
		return &a
	}

	exprs := make([]clause.Expression, 0, len(conds))
	for _, cond := range conds {
		exprs = append(exprs, cond.BeCond().(clause.Expression))
	}
	a.db = a.db.Clauses(clause.Where{Exprs: exprs})
	return &a
}

func (a builderLogBelongsToBuilder) WithContext(ctx context.Context) *builderLogBelongsToBuilder {
	a.db = a.db.WithContext(ctx)
	return &a
}

func (a builderLogBelongsToBuilder) Session(session *gorm.Session) *builderLogBelongsToBuilder {
	a.db = a.db.Session(session)
	return &a
}

func (a builderLogBelongsToBuilder) Model(m *models.BuilderRunner) *builderLogBelongsToBuilderTx {
	return &builderLogBelongsToBuilderTx{a.db.Model(m).Association(a.Name())}
}

type builderLogBelongsToBuilderTx struct{ tx *gorm.Association }

func (a builderLogBelongsToBuilderTx) Find() (result *models.Builder, err error) {
	return result, a.tx.Find(&result)
}

func (a builderLogBelongsToBuilderTx) Append(values ...*models.Builder) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Append(targetValues...)
}

func (a builderLogBelongsToBuilderTx) Replace(values ...*models.Builder) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Replace(targetValues...)
}

func (a builderLogBelongsToBuilderTx) Delete(values ...*models.Builder) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Delete(targetValues...)
}

func (a builderLogBelongsToBuilderTx) Clear() error {
	return a.tx.Clear()
}

func (a builderLogBelongsToBuilderTx) Count() int64 {
	return a.tx.Count()
}

type builderLogDo struct{ gen.DO }

func (b builderLogDo) Debug() *builderLogDo {
	return b.withDO(b.DO.Debug())
}

func (b builderLogDo) WithContext(ctx context.Context) *builderLogDo {
	return b.withDO(b.DO.WithContext(ctx))
}

func (b builderLogDo) ReadDB() *builderLogDo {
	return b.Clauses(dbresolver.Read)
}

func (b builderLogDo) WriteDB() *builderLogDo {
	return b.Clauses(dbresolver.Write)
}

func (b builderLogDo) Session(config *gorm.Session) *builderLogDo {
	return b.withDO(b.DO.Session(config))
}

func (b builderLogDo) Clauses(conds ...clause.Expression) *builderLogDo {
	return b.withDO(b.DO.Clauses(conds...))
}

func (b builderLogDo) Returning(value interface{}, columns ...string) *builderLogDo {
	return b.withDO(b.DO.Returning(value, columns...))
}

func (b builderLogDo) Not(conds ...gen.Condition) *builderLogDo {
	return b.withDO(b.DO.Not(conds...))
}

func (b builderLogDo) Or(conds ...gen.Condition) *builderLogDo {
	return b.withDO(b.DO.Or(conds...))
}

func (b builderLogDo) Select(conds ...field.Expr) *builderLogDo {
	return b.withDO(b.DO.Select(conds...))
}

func (b builderLogDo) Where(conds ...gen.Condition) *builderLogDo {
	return b.withDO(b.DO.Where(conds...))
}

func (b builderLogDo) Order(conds ...field.Expr) *builderLogDo {
	return b.withDO(b.DO.Order(conds...))
}

func (b builderLogDo) Distinct(cols ...field.Expr) *builderLogDo {
	return b.withDO(b.DO.Distinct(cols...))
}

func (b builderLogDo) Omit(cols ...field.Expr) *builderLogDo {
	return b.withDO(b.DO.Omit(cols...))
}

func (b builderLogDo) Join(table schema.Tabler, on ...field.Expr) *builderLogDo {
	return b.withDO(b.DO.Join(table, on...))
}

func (b builderLogDo) LeftJoin(table schema.Tabler, on ...field.Expr) *builderLogDo {
	return b.withDO(b.DO.LeftJoin(table, on...))
}

func (b builderLogDo) RightJoin(table schema.Tabler, on ...field.Expr) *builderLogDo {
	return b.withDO(b.DO.RightJoin(table, on...))
}

func (b builderLogDo) Group(cols ...field.Expr) *builderLogDo {
	return b.withDO(b.DO.Group(cols...))
}

func (b builderLogDo) Having(conds ...gen.Condition) *builderLogDo {
	return b.withDO(b.DO.Having(conds...))
}

func (b builderLogDo) Limit(limit int) *builderLogDo {
	return b.withDO(b.DO.Limit(limit))
}

func (b builderLogDo) Offset(offset int) *builderLogDo {
	return b.withDO(b.DO.Offset(offset))
}

func (b builderLogDo) Scopes(funcs ...func(gen.Dao) gen.Dao) *builderLogDo {
	return b.withDO(b.DO.Scopes(funcs...))
}

func (b builderLogDo) Unscoped() *builderLogDo {
	return b.withDO(b.DO.Unscoped())
}

func (b builderLogDo) Create(values ...*models.BuilderRunner) error {
	if len(values) == 0 {
		return nil
	}
	return b.DO.Create(values)
}

func (b builderLogDo) CreateInBatches(values []*models.BuilderRunner, batchSize int) error {
	return b.DO.CreateInBatches(values, batchSize)
}

// Save : !!! underlying implementation is different with GORM
// The method is equivalent to executing the statement: db.Clauses(clause.OnConflict{UpdateAll: true}).Create(values)
func (b builderLogDo) Save(values ...*models.BuilderRunner) error {
	if len(values) == 0 {
		return nil
	}
	return b.DO.Save(values)
}

func (b builderLogDo) First() (*models.BuilderRunner, error) {
	if result, err := b.DO.First(); err != nil {
		return nil, err
	} else {
		return result.(*models.BuilderRunner), nil
	}
}

func (b builderLogDo) Take() (*models.BuilderRunner, error) {
	if result, err := b.DO.Take(); err != nil {
		return nil, err
	} else {
		return result.(*models.BuilderRunner), nil
	}
}

func (b builderLogDo) Last() (*models.BuilderRunner, error) {
	if result, err := b.DO.Last(); err != nil {
		return nil, err
	} else {
		return result.(*models.BuilderRunner), nil
	}
}

func (b builderLogDo) Find() ([]*models.BuilderRunner, error) {
	result, err := b.DO.Find()
	return result.([]*models.BuilderRunner), err
}

func (b builderLogDo) FindInBatch(batchSize int, fc func(tx gen.Dao, batch int) error) (results []*models.BuilderRunner, err error) {
	buf := make([]*models.BuilderRunner, 0, batchSize)
	err = b.DO.FindInBatches(&buf, batchSize, func(tx gen.Dao, batch int) error {
		defer func() { results = append(results, buf...) }()
		return fc(tx, batch)
	})
	return results, err
}

func (b builderLogDo) FindInBatches(result *[]*models.BuilderRunner, batchSize int, fc func(tx gen.Dao, batch int) error) error {
	return b.DO.FindInBatches(result, batchSize, fc)
}

func (b builderLogDo) Attrs(attrs ...field.AssignExpr) *builderLogDo {
	return b.withDO(b.DO.Attrs(attrs...))
}

func (b builderLogDo) Assign(attrs ...field.AssignExpr) *builderLogDo {
	return b.withDO(b.DO.Assign(attrs...))
}

func (b builderLogDo) Joins(fields ...field.RelationField) *builderLogDo {
	for _, _f := range fields {
		b = *b.withDO(b.DO.Joins(_f))
	}
	return &b
}

func (b builderLogDo) Preload(fields ...field.RelationField) *builderLogDo {
	for _, _f := range fields {
		b = *b.withDO(b.DO.Preload(_f))
	}
	return &b
}

func (b builderLogDo) FirstOrInit() (*models.BuilderRunner, error) {
	if result, err := b.DO.FirstOrInit(); err != nil {
		return nil, err
	} else {
		return result.(*models.BuilderRunner), nil
	}
}

func (b builderLogDo) FirstOrCreate() (*models.BuilderRunner, error) {
	if result, err := b.DO.FirstOrCreate(); err != nil {
		return nil, err
	} else {
		return result.(*models.BuilderRunner), nil
	}
}

func (b builderLogDo) FindByPage(offset int, limit int) (result []*models.BuilderRunner, count int64, err error) {
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

func (b builderLogDo) ScanByPage(result interface{}, offset int, limit int) (count int64, err error) {
	count, err = b.Count()
	if err != nil {
		return
	}

	err = b.Offset(offset).Limit(limit).Scan(result)
	return
}

func (b builderLogDo) Scan(result interface{}) (err error) {
	return b.DO.Scan(result)
}

func (b builderLogDo) Delete(models ...*models.BuilderRunner) (result gen.ResultInfo, err error) {
	return b.DO.Delete(models)
}

func (b *builderLogDo) withDO(do gen.Dao) *builderLogDo {
	b.DO = *do.(*gen.DO)
	return b
}
