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

func newDaemonGcRepositoryRule(db *gorm.DB, opts ...gen.DOOption) daemonGcRepositoryRule {
	_daemonGcRepositoryRule := daemonGcRepositoryRule{}

	_daemonGcRepositoryRule.daemonGcRepositoryRuleDo.UseDB(db, opts...)
	_daemonGcRepositoryRule.daemonGcRepositoryRuleDo.UseModel(&models.DaemonGcRepositoryRule{})

	tableName := _daemonGcRepositoryRule.daemonGcRepositoryRuleDo.TableName()
	_daemonGcRepositoryRule.ALL = field.NewAsterisk(tableName)
	_daemonGcRepositoryRule.CreatedAt = field.NewInt64(tableName, "created_at")
	_daemonGcRepositoryRule.UpdatedAt = field.NewInt64(tableName, "updated_at")
	_daemonGcRepositoryRule.DeletedAt = field.NewUint64(tableName, "deleted_at")
	_daemonGcRepositoryRule.ID = field.NewInt64(tableName, "id")
	_daemonGcRepositoryRule.NamespaceID = field.NewInt64(tableName, "namespace_id")
	_daemonGcRepositoryRule.IsRunning = field.NewBool(tableName, "is_running")
	_daemonGcRepositoryRule.RetentionDay = field.NewInt(tableName, "retention_day")
	_daemonGcRepositoryRule.CronEnabled = field.NewBool(tableName, "cron_enabled")
	_daemonGcRepositoryRule.CronRule = field.NewString(tableName, "cron_rule")
	_daemonGcRepositoryRule.CronNextTrigger = field.NewInt64(tableName, "cron_next_trigger")
	_daemonGcRepositoryRule.Namespace = daemonGcRepositoryRuleBelongsToNamespace{
		db: db.Session(&gorm.Session{}),

		RelationField: field.NewRelation("Namespace", "models.Namespace"),
	}

	_daemonGcRepositoryRule.fillFieldMap()

	return _daemonGcRepositoryRule
}

type daemonGcRepositoryRule struct {
	daemonGcRepositoryRuleDo daemonGcRepositoryRuleDo

	ALL             field.Asterisk
	CreatedAt       field.Int64
	UpdatedAt       field.Int64
	DeletedAt       field.Uint64
	ID              field.Int64
	NamespaceID     field.Int64
	IsRunning       field.Bool
	RetentionDay    field.Int
	CronEnabled     field.Bool
	CronRule        field.String
	CronNextTrigger field.Int64
	Namespace       daemonGcRepositoryRuleBelongsToNamespace

	fieldMap map[string]field.Expr
}

func (d daemonGcRepositoryRule) Table(newTableName string) *daemonGcRepositoryRule {
	d.daemonGcRepositoryRuleDo.UseTable(newTableName)
	return d.updateTableName(newTableName)
}

func (d daemonGcRepositoryRule) As(alias string) *daemonGcRepositoryRule {
	d.daemonGcRepositoryRuleDo.DO = *(d.daemonGcRepositoryRuleDo.As(alias).(*gen.DO))
	return d.updateTableName(alias)
}

func (d *daemonGcRepositoryRule) updateTableName(table string) *daemonGcRepositoryRule {
	d.ALL = field.NewAsterisk(table)
	d.CreatedAt = field.NewInt64(table, "created_at")
	d.UpdatedAt = field.NewInt64(table, "updated_at")
	d.DeletedAt = field.NewUint64(table, "deleted_at")
	d.ID = field.NewInt64(table, "id")
	d.NamespaceID = field.NewInt64(table, "namespace_id")
	d.IsRunning = field.NewBool(table, "is_running")
	d.RetentionDay = field.NewInt(table, "retention_day")
	d.CronEnabled = field.NewBool(table, "cron_enabled")
	d.CronRule = field.NewString(table, "cron_rule")
	d.CronNextTrigger = field.NewInt64(table, "cron_next_trigger")

	d.fillFieldMap()

	return d
}

func (d *daemonGcRepositoryRule) WithContext(ctx context.Context) *daemonGcRepositoryRuleDo {
	return d.daemonGcRepositoryRuleDo.WithContext(ctx)
}

func (d daemonGcRepositoryRule) TableName() string { return d.daemonGcRepositoryRuleDo.TableName() }

func (d daemonGcRepositoryRule) Alias() string { return d.daemonGcRepositoryRuleDo.Alias() }

func (d daemonGcRepositoryRule) Columns(cols ...field.Expr) gen.Columns {
	return d.daemonGcRepositoryRuleDo.Columns(cols...)
}

func (d *daemonGcRepositoryRule) GetFieldByName(fieldName string) (field.OrderExpr, bool) {
	_f, ok := d.fieldMap[fieldName]
	if !ok || _f == nil {
		return nil, false
	}
	_oe, ok := _f.(field.OrderExpr)
	return _oe, ok
}

func (d *daemonGcRepositoryRule) fillFieldMap() {
	d.fieldMap = make(map[string]field.Expr, 11)
	d.fieldMap["created_at"] = d.CreatedAt
	d.fieldMap["updated_at"] = d.UpdatedAt
	d.fieldMap["deleted_at"] = d.DeletedAt
	d.fieldMap["id"] = d.ID
	d.fieldMap["namespace_id"] = d.NamespaceID
	d.fieldMap["is_running"] = d.IsRunning
	d.fieldMap["retention_day"] = d.RetentionDay
	d.fieldMap["cron_enabled"] = d.CronEnabled
	d.fieldMap["cron_rule"] = d.CronRule
	d.fieldMap["cron_next_trigger"] = d.CronNextTrigger

}

func (d daemonGcRepositoryRule) clone(db *gorm.DB) daemonGcRepositoryRule {
	d.daemonGcRepositoryRuleDo.ReplaceConnPool(db.Statement.ConnPool)
	return d
}

func (d daemonGcRepositoryRule) replaceDB(db *gorm.DB) daemonGcRepositoryRule {
	d.daemonGcRepositoryRuleDo.ReplaceDB(db)
	return d
}

type daemonGcRepositoryRuleBelongsToNamespace struct {
	db *gorm.DB

	field.RelationField
}

func (a daemonGcRepositoryRuleBelongsToNamespace) Where(conds ...field.Expr) *daemonGcRepositoryRuleBelongsToNamespace {
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

func (a daemonGcRepositoryRuleBelongsToNamespace) WithContext(ctx context.Context) *daemonGcRepositoryRuleBelongsToNamespace {
	a.db = a.db.WithContext(ctx)
	return &a
}

func (a daemonGcRepositoryRuleBelongsToNamespace) Session(session *gorm.Session) *daemonGcRepositoryRuleBelongsToNamespace {
	a.db = a.db.Session(session)
	return &a
}

func (a daemonGcRepositoryRuleBelongsToNamespace) Model(m *models.DaemonGcRepositoryRule) *daemonGcRepositoryRuleBelongsToNamespaceTx {
	return &daemonGcRepositoryRuleBelongsToNamespaceTx{a.db.Model(m).Association(a.Name())}
}

type daemonGcRepositoryRuleBelongsToNamespaceTx struct{ tx *gorm.Association }

func (a daemonGcRepositoryRuleBelongsToNamespaceTx) Find() (result *models.Namespace, err error) {
	return result, a.tx.Find(&result)
}

func (a daemonGcRepositoryRuleBelongsToNamespaceTx) Append(values ...*models.Namespace) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Append(targetValues...)
}

func (a daemonGcRepositoryRuleBelongsToNamespaceTx) Replace(values ...*models.Namespace) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Replace(targetValues...)
}

func (a daemonGcRepositoryRuleBelongsToNamespaceTx) Delete(values ...*models.Namespace) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Delete(targetValues...)
}

func (a daemonGcRepositoryRuleBelongsToNamespaceTx) Clear() error {
	return a.tx.Clear()
}

func (a daemonGcRepositoryRuleBelongsToNamespaceTx) Count() int64 {
	return a.tx.Count()
}

type daemonGcRepositoryRuleDo struct{ gen.DO }

func (d daemonGcRepositoryRuleDo) Debug() *daemonGcRepositoryRuleDo {
	return d.withDO(d.DO.Debug())
}

func (d daemonGcRepositoryRuleDo) WithContext(ctx context.Context) *daemonGcRepositoryRuleDo {
	return d.withDO(d.DO.WithContext(ctx))
}

func (d daemonGcRepositoryRuleDo) ReadDB() *daemonGcRepositoryRuleDo {
	return d.Clauses(dbresolver.Read)
}

func (d daemonGcRepositoryRuleDo) WriteDB() *daemonGcRepositoryRuleDo {
	return d.Clauses(dbresolver.Write)
}

func (d daemonGcRepositoryRuleDo) Session(config *gorm.Session) *daemonGcRepositoryRuleDo {
	return d.withDO(d.DO.Session(config))
}

func (d daemonGcRepositoryRuleDo) Clauses(conds ...clause.Expression) *daemonGcRepositoryRuleDo {
	return d.withDO(d.DO.Clauses(conds...))
}

func (d daemonGcRepositoryRuleDo) Returning(value interface{}, columns ...string) *daemonGcRepositoryRuleDo {
	return d.withDO(d.DO.Returning(value, columns...))
}

func (d daemonGcRepositoryRuleDo) Not(conds ...gen.Condition) *daemonGcRepositoryRuleDo {
	return d.withDO(d.DO.Not(conds...))
}

func (d daemonGcRepositoryRuleDo) Or(conds ...gen.Condition) *daemonGcRepositoryRuleDo {
	return d.withDO(d.DO.Or(conds...))
}

func (d daemonGcRepositoryRuleDo) Select(conds ...field.Expr) *daemonGcRepositoryRuleDo {
	return d.withDO(d.DO.Select(conds...))
}

func (d daemonGcRepositoryRuleDo) Where(conds ...gen.Condition) *daemonGcRepositoryRuleDo {
	return d.withDO(d.DO.Where(conds...))
}

func (d daemonGcRepositoryRuleDo) Order(conds ...field.Expr) *daemonGcRepositoryRuleDo {
	return d.withDO(d.DO.Order(conds...))
}

func (d daemonGcRepositoryRuleDo) Distinct(cols ...field.Expr) *daemonGcRepositoryRuleDo {
	return d.withDO(d.DO.Distinct(cols...))
}

func (d daemonGcRepositoryRuleDo) Omit(cols ...field.Expr) *daemonGcRepositoryRuleDo {
	return d.withDO(d.DO.Omit(cols...))
}

func (d daemonGcRepositoryRuleDo) Join(table schema.Tabler, on ...field.Expr) *daemonGcRepositoryRuleDo {
	return d.withDO(d.DO.Join(table, on...))
}

func (d daemonGcRepositoryRuleDo) LeftJoin(table schema.Tabler, on ...field.Expr) *daemonGcRepositoryRuleDo {
	return d.withDO(d.DO.LeftJoin(table, on...))
}

func (d daemonGcRepositoryRuleDo) RightJoin(table schema.Tabler, on ...field.Expr) *daemonGcRepositoryRuleDo {
	return d.withDO(d.DO.RightJoin(table, on...))
}

func (d daemonGcRepositoryRuleDo) Group(cols ...field.Expr) *daemonGcRepositoryRuleDo {
	return d.withDO(d.DO.Group(cols...))
}

func (d daemonGcRepositoryRuleDo) Having(conds ...gen.Condition) *daemonGcRepositoryRuleDo {
	return d.withDO(d.DO.Having(conds...))
}

func (d daemonGcRepositoryRuleDo) Limit(limit int) *daemonGcRepositoryRuleDo {
	return d.withDO(d.DO.Limit(limit))
}

func (d daemonGcRepositoryRuleDo) Offset(offset int) *daemonGcRepositoryRuleDo {
	return d.withDO(d.DO.Offset(offset))
}

func (d daemonGcRepositoryRuleDo) Scopes(funcs ...func(gen.Dao) gen.Dao) *daemonGcRepositoryRuleDo {
	return d.withDO(d.DO.Scopes(funcs...))
}

func (d daemonGcRepositoryRuleDo) Unscoped() *daemonGcRepositoryRuleDo {
	return d.withDO(d.DO.Unscoped())
}

func (d daemonGcRepositoryRuleDo) Create(values ...*models.DaemonGcRepositoryRule) error {
	if len(values) == 0 {
		return nil
	}
	return d.DO.Create(values)
}

func (d daemonGcRepositoryRuleDo) CreateInBatches(values []*models.DaemonGcRepositoryRule, batchSize int) error {
	return d.DO.CreateInBatches(values, batchSize)
}

// Save : !!! underlying implementation is different with GORM
// The method is equivalent to executing the statement: db.Clauses(clause.OnConflict{UpdateAll: true}).Create(values)
func (d daemonGcRepositoryRuleDo) Save(values ...*models.DaemonGcRepositoryRule) error {
	if len(values) == 0 {
		return nil
	}
	return d.DO.Save(values)
}

func (d daemonGcRepositoryRuleDo) First() (*models.DaemonGcRepositoryRule, error) {
	if result, err := d.DO.First(); err != nil {
		return nil, err
	} else {
		return result.(*models.DaemonGcRepositoryRule), nil
	}
}

func (d daemonGcRepositoryRuleDo) Take() (*models.DaemonGcRepositoryRule, error) {
	if result, err := d.DO.Take(); err != nil {
		return nil, err
	} else {
		return result.(*models.DaemonGcRepositoryRule), nil
	}
}

func (d daemonGcRepositoryRuleDo) Last() (*models.DaemonGcRepositoryRule, error) {
	if result, err := d.DO.Last(); err != nil {
		return nil, err
	} else {
		return result.(*models.DaemonGcRepositoryRule), nil
	}
}

func (d daemonGcRepositoryRuleDo) Find() ([]*models.DaemonGcRepositoryRule, error) {
	result, err := d.DO.Find()
	return result.([]*models.DaemonGcRepositoryRule), err
}

func (d daemonGcRepositoryRuleDo) FindInBatch(batchSize int, fc func(tx gen.Dao, batch int) error) (results []*models.DaemonGcRepositoryRule, err error) {
	buf := make([]*models.DaemonGcRepositoryRule, 0, batchSize)
	err = d.DO.FindInBatches(&buf, batchSize, func(tx gen.Dao, batch int) error {
		defer func() { results = append(results, buf...) }()
		return fc(tx, batch)
	})
	return results, err
}

func (d daemonGcRepositoryRuleDo) FindInBatches(result *[]*models.DaemonGcRepositoryRule, batchSize int, fc func(tx gen.Dao, batch int) error) error {
	return d.DO.FindInBatches(result, batchSize, fc)
}

func (d daemonGcRepositoryRuleDo) Attrs(attrs ...field.AssignExpr) *daemonGcRepositoryRuleDo {
	return d.withDO(d.DO.Attrs(attrs...))
}

func (d daemonGcRepositoryRuleDo) Assign(attrs ...field.AssignExpr) *daemonGcRepositoryRuleDo {
	return d.withDO(d.DO.Assign(attrs...))
}

func (d daemonGcRepositoryRuleDo) Joins(fields ...field.RelationField) *daemonGcRepositoryRuleDo {
	for _, _f := range fields {
		d = *d.withDO(d.DO.Joins(_f))
	}
	return &d
}

func (d daemonGcRepositoryRuleDo) Preload(fields ...field.RelationField) *daemonGcRepositoryRuleDo {
	for _, _f := range fields {
		d = *d.withDO(d.DO.Preload(_f))
	}
	return &d
}

func (d daemonGcRepositoryRuleDo) FirstOrInit() (*models.DaemonGcRepositoryRule, error) {
	if result, err := d.DO.FirstOrInit(); err != nil {
		return nil, err
	} else {
		return result.(*models.DaemonGcRepositoryRule), nil
	}
}

func (d daemonGcRepositoryRuleDo) FirstOrCreate() (*models.DaemonGcRepositoryRule, error) {
	if result, err := d.DO.FirstOrCreate(); err != nil {
		return nil, err
	} else {
		return result.(*models.DaemonGcRepositoryRule), nil
	}
}

func (d daemonGcRepositoryRuleDo) FindByPage(offset int, limit int) (result []*models.DaemonGcRepositoryRule, count int64, err error) {
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

func (d daemonGcRepositoryRuleDo) ScanByPage(result interface{}, offset int, limit int) (count int64, err error) {
	count, err = d.Count()
	if err != nil {
		return
	}

	err = d.Offset(offset).Limit(limit).Scan(result)
	return
}

func (d daemonGcRepositoryRuleDo) Scan(result interface{}) (err error) {
	return d.DO.Scan(result)
}

func (d daemonGcRepositoryRuleDo) Delete(models ...*models.DaemonGcRepositoryRule) (result gen.ResultInfo, err error) {
	return d.DO.Delete(models)
}

func (d *daemonGcRepositoryRuleDo) withDO(do gen.Dao) *daemonGcRepositoryRuleDo {
	d.DO = *do.(*gen.DO)
	return d
}
