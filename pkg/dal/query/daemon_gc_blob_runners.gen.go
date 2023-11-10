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

func newDaemonGcBlobRunner(db *gorm.DB, opts ...gen.DOOption) daemonGcBlobRunner {
	_daemonGcBlobRunner := daemonGcBlobRunner{}

	_daemonGcBlobRunner.daemonGcBlobRunnerDo.UseDB(db, opts...)
	_daemonGcBlobRunner.daemonGcBlobRunnerDo.UseModel(&models.DaemonGcBlobRunner{})

	tableName := _daemonGcBlobRunner.daemonGcBlobRunnerDo.TableName()
	_daemonGcBlobRunner.ALL = field.NewAsterisk(tableName)
	_daemonGcBlobRunner.CreatedAt = field.NewTime(tableName, "created_at")
	_daemonGcBlobRunner.UpdatedAt = field.NewTime(tableName, "updated_at")
	_daemonGcBlobRunner.DeletedAt = field.NewUint(tableName, "deleted_at")
	_daemonGcBlobRunner.ID = field.NewInt64(tableName, "id")
	_daemonGcBlobRunner.RuleID = field.NewInt64(tableName, "rule_id")
	_daemonGcBlobRunner.Status = field.NewField(tableName, "status")
	_daemonGcBlobRunner.Message = field.NewBytes(tableName, "message")
	_daemonGcBlobRunner.StartedAt = field.NewTime(tableName, "started_at")
	_daemonGcBlobRunner.EndedAt = field.NewTime(tableName, "ended_at")
	_daemonGcBlobRunner.Duration = field.NewInt64(tableName, "duration")
	_daemonGcBlobRunner.Rule = daemonGcBlobRunnerBelongsToRule{
		db: db.Session(&gorm.Session{}),

		RelationField: field.NewRelation("Rule", "models.DaemonGcBlobRule"),
	}

	_daemonGcBlobRunner.fillFieldMap()

	return _daemonGcBlobRunner
}

type daemonGcBlobRunner struct {
	daemonGcBlobRunnerDo daemonGcBlobRunnerDo

	ALL       field.Asterisk
	CreatedAt field.Time
	UpdatedAt field.Time
	DeletedAt field.Uint
	ID        field.Int64
	RuleID    field.Int64
	Status    field.Field
	Message   field.Bytes
	StartedAt field.Time
	EndedAt   field.Time
	Duration  field.Int64
	Rule      daemonGcBlobRunnerBelongsToRule

	fieldMap map[string]field.Expr
}

func (d daemonGcBlobRunner) Table(newTableName string) *daemonGcBlobRunner {
	d.daemonGcBlobRunnerDo.UseTable(newTableName)
	return d.updateTableName(newTableName)
}

func (d daemonGcBlobRunner) As(alias string) *daemonGcBlobRunner {
	d.daemonGcBlobRunnerDo.DO = *(d.daemonGcBlobRunnerDo.As(alias).(*gen.DO))
	return d.updateTableName(alias)
}

func (d *daemonGcBlobRunner) updateTableName(table string) *daemonGcBlobRunner {
	d.ALL = field.NewAsterisk(table)
	d.CreatedAt = field.NewTime(table, "created_at")
	d.UpdatedAt = field.NewTime(table, "updated_at")
	d.DeletedAt = field.NewUint(table, "deleted_at")
	d.ID = field.NewInt64(table, "id")
	d.RuleID = field.NewInt64(table, "rule_id")
	d.Status = field.NewField(table, "status")
	d.Message = field.NewBytes(table, "message")
	d.StartedAt = field.NewTime(table, "started_at")
	d.EndedAt = field.NewTime(table, "ended_at")
	d.Duration = field.NewInt64(table, "duration")

	d.fillFieldMap()

	return d
}

func (d *daemonGcBlobRunner) WithContext(ctx context.Context) *daemonGcBlobRunnerDo {
	return d.daemonGcBlobRunnerDo.WithContext(ctx)
}

func (d daemonGcBlobRunner) TableName() string { return d.daemonGcBlobRunnerDo.TableName() }

func (d daemonGcBlobRunner) Alias() string { return d.daemonGcBlobRunnerDo.Alias() }

func (d daemonGcBlobRunner) Columns(cols ...field.Expr) gen.Columns {
	return d.daemonGcBlobRunnerDo.Columns(cols...)
}

func (d *daemonGcBlobRunner) GetFieldByName(fieldName string) (field.OrderExpr, bool) {
	_f, ok := d.fieldMap[fieldName]
	if !ok || _f == nil {
		return nil, false
	}
	_oe, ok := _f.(field.OrderExpr)
	return _oe, ok
}

func (d *daemonGcBlobRunner) fillFieldMap() {
	d.fieldMap = make(map[string]field.Expr, 11)
	d.fieldMap["created_at"] = d.CreatedAt
	d.fieldMap["updated_at"] = d.UpdatedAt
	d.fieldMap["deleted_at"] = d.DeletedAt
	d.fieldMap["id"] = d.ID
	d.fieldMap["rule_id"] = d.RuleID
	d.fieldMap["status"] = d.Status
	d.fieldMap["message"] = d.Message
	d.fieldMap["started_at"] = d.StartedAt
	d.fieldMap["ended_at"] = d.EndedAt
	d.fieldMap["duration"] = d.Duration

}

func (d daemonGcBlobRunner) clone(db *gorm.DB) daemonGcBlobRunner {
	d.daemonGcBlobRunnerDo.ReplaceConnPool(db.Statement.ConnPool)
	return d
}

func (d daemonGcBlobRunner) replaceDB(db *gorm.DB) daemonGcBlobRunner {
	d.daemonGcBlobRunnerDo.ReplaceDB(db)
	return d
}

type daemonGcBlobRunnerBelongsToRule struct {
	db *gorm.DB

	field.RelationField
}

func (a daemonGcBlobRunnerBelongsToRule) Where(conds ...field.Expr) *daemonGcBlobRunnerBelongsToRule {
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

func (a daemonGcBlobRunnerBelongsToRule) WithContext(ctx context.Context) *daemonGcBlobRunnerBelongsToRule {
	a.db = a.db.WithContext(ctx)
	return &a
}

func (a daemonGcBlobRunnerBelongsToRule) Session(session *gorm.Session) *daemonGcBlobRunnerBelongsToRule {
	a.db = a.db.Session(session)
	return &a
}

func (a daemonGcBlobRunnerBelongsToRule) Model(m *models.DaemonGcBlobRunner) *daemonGcBlobRunnerBelongsToRuleTx {
	return &daemonGcBlobRunnerBelongsToRuleTx{a.db.Model(m).Association(a.Name())}
}

type daemonGcBlobRunnerBelongsToRuleTx struct{ tx *gorm.Association }

func (a daemonGcBlobRunnerBelongsToRuleTx) Find() (result *models.DaemonGcBlobRule, err error) {
	return result, a.tx.Find(&result)
}

func (a daemonGcBlobRunnerBelongsToRuleTx) Append(values ...*models.DaemonGcBlobRule) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Append(targetValues...)
}

func (a daemonGcBlobRunnerBelongsToRuleTx) Replace(values ...*models.DaemonGcBlobRule) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Replace(targetValues...)
}

func (a daemonGcBlobRunnerBelongsToRuleTx) Delete(values ...*models.DaemonGcBlobRule) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Delete(targetValues...)
}

func (a daemonGcBlobRunnerBelongsToRuleTx) Clear() error {
	return a.tx.Clear()
}

func (a daemonGcBlobRunnerBelongsToRuleTx) Count() int64 {
	return a.tx.Count()
}

type daemonGcBlobRunnerDo struct{ gen.DO }

func (d daemonGcBlobRunnerDo) Debug() *daemonGcBlobRunnerDo {
	return d.withDO(d.DO.Debug())
}

func (d daemonGcBlobRunnerDo) WithContext(ctx context.Context) *daemonGcBlobRunnerDo {
	return d.withDO(d.DO.WithContext(ctx))
}

func (d daemonGcBlobRunnerDo) ReadDB() *daemonGcBlobRunnerDo {
	return d.Clauses(dbresolver.Read)
}

func (d daemonGcBlobRunnerDo) WriteDB() *daemonGcBlobRunnerDo {
	return d.Clauses(dbresolver.Write)
}

func (d daemonGcBlobRunnerDo) Session(config *gorm.Session) *daemonGcBlobRunnerDo {
	return d.withDO(d.DO.Session(config))
}

func (d daemonGcBlobRunnerDo) Clauses(conds ...clause.Expression) *daemonGcBlobRunnerDo {
	return d.withDO(d.DO.Clauses(conds...))
}

func (d daemonGcBlobRunnerDo) Returning(value interface{}, columns ...string) *daemonGcBlobRunnerDo {
	return d.withDO(d.DO.Returning(value, columns...))
}

func (d daemonGcBlobRunnerDo) Not(conds ...gen.Condition) *daemonGcBlobRunnerDo {
	return d.withDO(d.DO.Not(conds...))
}

func (d daemonGcBlobRunnerDo) Or(conds ...gen.Condition) *daemonGcBlobRunnerDo {
	return d.withDO(d.DO.Or(conds...))
}

func (d daemonGcBlobRunnerDo) Select(conds ...field.Expr) *daemonGcBlobRunnerDo {
	return d.withDO(d.DO.Select(conds...))
}

func (d daemonGcBlobRunnerDo) Where(conds ...gen.Condition) *daemonGcBlobRunnerDo {
	return d.withDO(d.DO.Where(conds...))
}

func (d daemonGcBlobRunnerDo) Order(conds ...field.Expr) *daemonGcBlobRunnerDo {
	return d.withDO(d.DO.Order(conds...))
}

func (d daemonGcBlobRunnerDo) Distinct(cols ...field.Expr) *daemonGcBlobRunnerDo {
	return d.withDO(d.DO.Distinct(cols...))
}

func (d daemonGcBlobRunnerDo) Omit(cols ...field.Expr) *daemonGcBlobRunnerDo {
	return d.withDO(d.DO.Omit(cols...))
}

func (d daemonGcBlobRunnerDo) Join(table schema.Tabler, on ...field.Expr) *daemonGcBlobRunnerDo {
	return d.withDO(d.DO.Join(table, on...))
}

func (d daemonGcBlobRunnerDo) LeftJoin(table schema.Tabler, on ...field.Expr) *daemonGcBlobRunnerDo {
	return d.withDO(d.DO.LeftJoin(table, on...))
}

func (d daemonGcBlobRunnerDo) RightJoin(table schema.Tabler, on ...field.Expr) *daemonGcBlobRunnerDo {
	return d.withDO(d.DO.RightJoin(table, on...))
}

func (d daemonGcBlobRunnerDo) Group(cols ...field.Expr) *daemonGcBlobRunnerDo {
	return d.withDO(d.DO.Group(cols...))
}

func (d daemonGcBlobRunnerDo) Having(conds ...gen.Condition) *daemonGcBlobRunnerDo {
	return d.withDO(d.DO.Having(conds...))
}

func (d daemonGcBlobRunnerDo) Limit(limit int) *daemonGcBlobRunnerDo {
	return d.withDO(d.DO.Limit(limit))
}

func (d daemonGcBlobRunnerDo) Offset(offset int) *daemonGcBlobRunnerDo {
	return d.withDO(d.DO.Offset(offset))
}

func (d daemonGcBlobRunnerDo) Scopes(funcs ...func(gen.Dao) gen.Dao) *daemonGcBlobRunnerDo {
	return d.withDO(d.DO.Scopes(funcs...))
}

func (d daemonGcBlobRunnerDo) Unscoped() *daemonGcBlobRunnerDo {
	return d.withDO(d.DO.Unscoped())
}

func (d daemonGcBlobRunnerDo) Create(values ...*models.DaemonGcBlobRunner) error {
	if len(values) == 0 {
		return nil
	}
	return d.DO.Create(values)
}

func (d daemonGcBlobRunnerDo) CreateInBatches(values []*models.DaemonGcBlobRunner, batchSize int) error {
	return d.DO.CreateInBatches(values, batchSize)
}

// Save : !!! underlying implementation is different with GORM
// The method is equivalent to executing the statement: db.Clauses(clause.OnConflict{UpdateAll: true}).Create(values)
func (d daemonGcBlobRunnerDo) Save(values ...*models.DaemonGcBlobRunner) error {
	if len(values) == 0 {
		return nil
	}
	return d.DO.Save(values)
}

func (d daemonGcBlobRunnerDo) First() (*models.DaemonGcBlobRunner, error) {
	if result, err := d.DO.First(); err != nil {
		return nil, err
	} else {
		return result.(*models.DaemonGcBlobRunner), nil
	}
}

func (d daemonGcBlobRunnerDo) Take() (*models.DaemonGcBlobRunner, error) {
	if result, err := d.DO.Take(); err != nil {
		return nil, err
	} else {
		return result.(*models.DaemonGcBlobRunner), nil
	}
}

func (d daemonGcBlobRunnerDo) Last() (*models.DaemonGcBlobRunner, error) {
	if result, err := d.DO.Last(); err != nil {
		return nil, err
	} else {
		return result.(*models.DaemonGcBlobRunner), nil
	}
}

func (d daemonGcBlobRunnerDo) Find() ([]*models.DaemonGcBlobRunner, error) {
	result, err := d.DO.Find()
	return result.([]*models.DaemonGcBlobRunner), err
}

func (d daemonGcBlobRunnerDo) FindInBatch(batchSize int, fc func(tx gen.Dao, batch int) error) (results []*models.DaemonGcBlobRunner, err error) {
	buf := make([]*models.DaemonGcBlobRunner, 0, batchSize)
	err = d.DO.FindInBatches(&buf, batchSize, func(tx gen.Dao, batch int) error {
		defer func() { results = append(results, buf...) }()
		return fc(tx, batch)
	})
	return results, err
}

func (d daemonGcBlobRunnerDo) FindInBatches(result *[]*models.DaemonGcBlobRunner, batchSize int, fc func(tx gen.Dao, batch int) error) error {
	return d.DO.FindInBatches(result, batchSize, fc)
}

func (d daemonGcBlobRunnerDo) Attrs(attrs ...field.AssignExpr) *daemonGcBlobRunnerDo {
	return d.withDO(d.DO.Attrs(attrs...))
}

func (d daemonGcBlobRunnerDo) Assign(attrs ...field.AssignExpr) *daemonGcBlobRunnerDo {
	return d.withDO(d.DO.Assign(attrs...))
}

func (d daemonGcBlobRunnerDo) Joins(fields ...field.RelationField) *daemonGcBlobRunnerDo {
	for _, _f := range fields {
		d = *d.withDO(d.DO.Joins(_f))
	}
	return &d
}

func (d daemonGcBlobRunnerDo) Preload(fields ...field.RelationField) *daemonGcBlobRunnerDo {
	for _, _f := range fields {
		d = *d.withDO(d.DO.Preload(_f))
	}
	return &d
}

func (d daemonGcBlobRunnerDo) FirstOrInit() (*models.DaemonGcBlobRunner, error) {
	if result, err := d.DO.FirstOrInit(); err != nil {
		return nil, err
	} else {
		return result.(*models.DaemonGcBlobRunner), nil
	}
}

func (d daemonGcBlobRunnerDo) FirstOrCreate() (*models.DaemonGcBlobRunner, error) {
	if result, err := d.DO.FirstOrCreate(); err != nil {
		return nil, err
	} else {
		return result.(*models.DaemonGcBlobRunner), nil
	}
}

func (d daemonGcBlobRunnerDo) FindByPage(offset int, limit int) (result []*models.DaemonGcBlobRunner, count int64, err error) {
	result, err = d.Offset(offset).Limit(limit).Find()
	if err != nil {
		return
	}

	if size := len(result); 0 < limit && 0 < size && size < limit {
		count = int64(size + offset)
		return
	}

	count, err = d.Offset(-1).Limit(-1).Count()
	return
}

func (d daemonGcBlobRunnerDo) ScanByPage(result interface{}, offset int, limit int) (count int64, err error) {
	count, err = d.Count()
	if err != nil {
		return
	}

	err = d.Offset(offset).Limit(limit).Scan(result)
	return
}

func (d daemonGcBlobRunnerDo) Scan(result interface{}) (err error) {
	return d.DO.Scan(result)
}

func (d daemonGcBlobRunnerDo) Delete(models ...*models.DaemonGcBlobRunner) (result gen.ResultInfo, err error) {
	return d.DO.Delete(models)
}

func (d *daemonGcBlobRunnerDo) withDO(do gen.Dao) *daemonGcBlobRunnerDo {
	d.DO = *do.(*gen.DO)
	return d
}