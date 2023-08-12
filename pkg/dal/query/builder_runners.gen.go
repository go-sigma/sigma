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

func newBuilderRunner(db *gorm.DB, opts ...gen.DOOption) builderRunner {
	_builderRunner := builderRunner{}

	_builderRunner.builderRunnerDo.UseDB(db, opts...)
	_builderRunner.builderRunnerDo.UseModel(&models.BuilderRunner{})

	tableName := _builderRunner.builderRunnerDo.TableName()
	_builderRunner.ALL = field.NewAsterisk(tableName)
	_builderRunner.CreatedAt = field.NewTime(tableName, "created_at")
	_builderRunner.UpdatedAt = field.NewTime(tableName, "updated_at")
	_builderRunner.DeletedAt = field.NewUint(tableName, "deleted_at")
	_builderRunner.ID = field.NewInt64(tableName, "id")
	_builderRunner.BuilderID = field.NewInt64(tableName, "builder_id")
	_builderRunner.Log = field.NewBytes(tableName, "log")
	_builderRunner.Status = field.NewField(tableName, "status")
	_builderRunner.Builder = builderRunnerBelongsToBuilder{
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

	_builderRunner.fillFieldMap()

	return _builderRunner
}

type builderRunner struct {
	builderRunnerDo builderRunnerDo

	ALL       field.Asterisk
	CreatedAt field.Time
	UpdatedAt field.Time
	DeletedAt field.Uint
	ID        field.Int64
	BuilderID field.Int64
	Log       field.Bytes
	Status    field.Field
	Builder   builderRunnerBelongsToBuilder

	fieldMap map[string]field.Expr
}

func (b builderRunner) Table(newTableName string) *builderRunner {
	b.builderRunnerDo.UseTable(newTableName)
	return b.updateTableName(newTableName)
}

func (b builderRunner) As(alias string) *builderRunner {
	b.builderRunnerDo.DO = *(b.builderRunnerDo.As(alias).(*gen.DO))
	return b.updateTableName(alias)
}

func (b *builderRunner) updateTableName(table string) *builderRunner {
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

func (b *builderRunner) WithContext(ctx context.Context) *builderRunnerDo {
	return b.builderRunnerDo.WithContext(ctx)
}

func (b builderRunner) TableName() string { return b.builderRunnerDo.TableName() }

func (b builderRunner) Alias() string { return b.builderRunnerDo.Alias() }

func (b builderRunner) Columns(cols ...field.Expr) gen.Columns {
	return b.builderRunnerDo.Columns(cols...)
}

func (b *builderRunner) GetFieldByName(fieldName string) (field.OrderExpr, bool) {
	_f, ok := b.fieldMap[fieldName]
	if !ok || _f == nil {
		return nil, false
	}
	_oe, ok := _f.(field.OrderExpr)
	return _oe, ok
}

func (b *builderRunner) fillFieldMap() {
	b.fieldMap = make(map[string]field.Expr, 8)
	b.fieldMap["created_at"] = b.CreatedAt
	b.fieldMap["updated_at"] = b.UpdatedAt
	b.fieldMap["deleted_at"] = b.DeletedAt
	b.fieldMap["id"] = b.ID
	b.fieldMap["builder_id"] = b.BuilderID
	b.fieldMap["log"] = b.Log
	b.fieldMap["status"] = b.Status

}

func (b builderRunner) clone(db *gorm.DB) builderRunner {
	b.builderRunnerDo.ReplaceConnPool(db.Statement.ConnPool)
	return b
}

func (b builderRunner) replaceDB(db *gorm.DB) builderRunner {
	b.builderRunnerDo.ReplaceDB(db)
	return b
}

type builderRunnerBelongsToBuilder struct {
	db *gorm.DB

	field.RelationField

	Repository struct {
		field.RelationField
		Namespace struct {
			field.RelationField
		}
	}
}

func (a builderRunnerBelongsToBuilder) Where(conds ...field.Expr) *builderRunnerBelongsToBuilder {
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

func (a builderRunnerBelongsToBuilder) WithContext(ctx context.Context) *builderRunnerBelongsToBuilder {
	a.db = a.db.WithContext(ctx)
	return &a
}

func (a builderRunnerBelongsToBuilder) Session(session *gorm.Session) *builderRunnerBelongsToBuilder {
	a.db = a.db.Session(session)
	return &a
}

func (a builderRunnerBelongsToBuilder) Model(m *models.BuilderRunner) *builderRunnerBelongsToBuilderTx {
	return &builderRunnerBelongsToBuilderTx{a.db.Model(m).Association(a.Name())}
}

type builderRunnerBelongsToBuilderTx struct{ tx *gorm.Association }

func (a builderRunnerBelongsToBuilderTx) Find() (result *models.Builder, err error) {
	return result, a.tx.Find(&result)
}

func (a builderRunnerBelongsToBuilderTx) Append(values ...*models.Builder) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Append(targetValues...)
}

func (a builderRunnerBelongsToBuilderTx) Replace(values ...*models.Builder) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Replace(targetValues...)
}

func (a builderRunnerBelongsToBuilderTx) Delete(values ...*models.Builder) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Delete(targetValues...)
}

func (a builderRunnerBelongsToBuilderTx) Clear() error {
	return a.tx.Clear()
}

func (a builderRunnerBelongsToBuilderTx) Count() int64 {
	return a.tx.Count()
}

type builderRunnerDo struct{ gen.DO }

func (b builderRunnerDo) Debug() *builderRunnerDo {
	return b.withDO(b.DO.Debug())
}

func (b builderRunnerDo) WithContext(ctx context.Context) *builderRunnerDo {
	return b.withDO(b.DO.WithContext(ctx))
}

func (b builderRunnerDo) ReadDB() *builderRunnerDo {
	return b.Clauses(dbresolver.Read)
}

func (b builderRunnerDo) WriteDB() *builderRunnerDo {
	return b.Clauses(dbresolver.Write)
}

func (b builderRunnerDo) Session(config *gorm.Session) *builderRunnerDo {
	return b.withDO(b.DO.Session(config))
}

func (b builderRunnerDo) Clauses(conds ...clause.Expression) *builderRunnerDo {
	return b.withDO(b.DO.Clauses(conds...))
}

func (b builderRunnerDo) Returning(value interface{}, columns ...string) *builderRunnerDo {
	return b.withDO(b.DO.Returning(value, columns...))
}

func (b builderRunnerDo) Not(conds ...gen.Condition) *builderRunnerDo {
	return b.withDO(b.DO.Not(conds...))
}

func (b builderRunnerDo) Or(conds ...gen.Condition) *builderRunnerDo {
	return b.withDO(b.DO.Or(conds...))
}

func (b builderRunnerDo) Select(conds ...field.Expr) *builderRunnerDo {
	return b.withDO(b.DO.Select(conds...))
}

func (b builderRunnerDo) Where(conds ...gen.Condition) *builderRunnerDo {
	return b.withDO(b.DO.Where(conds...))
}

func (b builderRunnerDo) Order(conds ...field.Expr) *builderRunnerDo {
	return b.withDO(b.DO.Order(conds...))
}

func (b builderRunnerDo) Distinct(cols ...field.Expr) *builderRunnerDo {
	return b.withDO(b.DO.Distinct(cols...))
}

func (b builderRunnerDo) Omit(cols ...field.Expr) *builderRunnerDo {
	return b.withDO(b.DO.Omit(cols...))
}

func (b builderRunnerDo) Join(table schema.Tabler, on ...field.Expr) *builderRunnerDo {
	return b.withDO(b.DO.Join(table, on...))
}

func (b builderRunnerDo) LeftJoin(table schema.Tabler, on ...field.Expr) *builderRunnerDo {
	return b.withDO(b.DO.LeftJoin(table, on...))
}

func (b builderRunnerDo) RightJoin(table schema.Tabler, on ...field.Expr) *builderRunnerDo {
	return b.withDO(b.DO.RightJoin(table, on...))
}

func (b builderRunnerDo) Group(cols ...field.Expr) *builderRunnerDo {
	return b.withDO(b.DO.Group(cols...))
}

func (b builderRunnerDo) Having(conds ...gen.Condition) *builderRunnerDo {
	return b.withDO(b.DO.Having(conds...))
}

func (b builderRunnerDo) Limit(limit int) *builderRunnerDo {
	return b.withDO(b.DO.Limit(limit))
}

func (b builderRunnerDo) Offset(offset int) *builderRunnerDo {
	return b.withDO(b.DO.Offset(offset))
}

func (b builderRunnerDo) Scopes(funcs ...func(gen.Dao) gen.Dao) *builderRunnerDo {
	return b.withDO(b.DO.Scopes(funcs...))
}

func (b builderRunnerDo) Unscoped() *builderRunnerDo {
	return b.withDO(b.DO.Unscoped())
}

func (b builderRunnerDo) Create(values ...*models.BuilderRunner) error {
	if len(values) == 0 {
		return nil
	}
	return b.DO.Create(values)
}

func (b builderRunnerDo) CreateInBatches(values []*models.BuilderRunner, batchSize int) error {
	return b.DO.CreateInBatches(values, batchSize)
}

// Save : !!! underlying implementation is different with GORM
// The method is equivalent to executing the statement: db.Clauses(clause.OnConflict{UpdateAll: true}).Create(values)
func (b builderRunnerDo) Save(values ...*models.BuilderRunner) error {
	if len(values) == 0 {
		return nil
	}
	return b.DO.Save(values)
}

func (b builderRunnerDo) First() (*models.BuilderRunner, error) {
	if result, err := b.DO.First(); err != nil {
		return nil, err
	} else {
		return result.(*models.BuilderRunner), nil
	}
}

func (b builderRunnerDo) Take() (*models.BuilderRunner, error) {
	if result, err := b.DO.Take(); err != nil {
		return nil, err
	} else {
		return result.(*models.BuilderRunner), nil
	}
}

func (b builderRunnerDo) Last() (*models.BuilderRunner, error) {
	if result, err := b.DO.Last(); err != nil {
		return nil, err
	} else {
		return result.(*models.BuilderRunner), nil
	}
}

func (b builderRunnerDo) Find() ([]*models.BuilderRunner, error) {
	result, err := b.DO.Find()
	return result.([]*models.BuilderRunner), err
}

func (b builderRunnerDo) FindInBatch(batchSize int, fc func(tx gen.Dao, batch int) error) (results []*models.BuilderRunner, err error) {
	buf := make([]*models.BuilderRunner, 0, batchSize)
	err = b.DO.FindInBatches(&buf, batchSize, func(tx gen.Dao, batch int) error {
		defer func() { results = append(results, buf...) }()
		return fc(tx, batch)
	})
	return results, err
}

func (b builderRunnerDo) FindInBatches(result *[]*models.BuilderRunner, batchSize int, fc func(tx gen.Dao, batch int) error) error {
	return b.DO.FindInBatches(result, batchSize, fc)
}

func (b builderRunnerDo) Attrs(attrs ...field.AssignExpr) *builderRunnerDo {
	return b.withDO(b.DO.Attrs(attrs...))
}

func (b builderRunnerDo) Assign(attrs ...field.AssignExpr) *builderRunnerDo {
	return b.withDO(b.DO.Assign(attrs...))
}

func (b builderRunnerDo) Joins(fields ...field.RelationField) *builderRunnerDo {
	for _, _f := range fields {
		b = *b.withDO(b.DO.Joins(_f))
	}
	return &b
}

func (b builderRunnerDo) Preload(fields ...field.RelationField) *builderRunnerDo {
	for _, _f := range fields {
		b = *b.withDO(b.DO.Preload(_f))
	}
	return &b
}

func (b builderRunnerDo) FirstOrInit() (*models.BuilderRunner, error) {
	if result, err := b.DO.FirstOrInit(); err != nil {
		return nil, err
	} else {
		return result.(*models.BuilderRunner), nil
	}
}

func (b builderRunnerDo) FirstOrCreate() (*models.BuilderRunner, error) {
	if result, err := b.DO.FirstOrCreate(); err != nil {
		return nil, err
	} else {
		return result.(*models.BuilderRunner), nil
	}
}

func (b builderRunnerDo) FindByPage(offset int, limit int) (result []*models.BuilderRunner, count int64, err error) {
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

func (b builderRunnerDo) ScanByPage(result interface{}, offset int, limit int) (count int64, err error) {
	count, err = b.Count()
	if err != nil {
		return
	}

	err = b.Offset(offset).Limit(limit).Scan(result)
	return
}

func (b builderRunnerDo) Scan(result interface{}) (err error) {
	return b.DO.Scan(result)
}

func (b builderRunnerDo) Delete(models ...*models.BuilderRunner) (result gen.ResultInfo, err error) {
	return b.DO.Delete(models)
}

func (b *builderRunnerDo) withDO(do gen.Dao) *builderRunnerDo {
	b.DO = *do.(*gen.DO)
	return b
}