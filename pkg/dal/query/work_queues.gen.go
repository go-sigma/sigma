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

func newWorkQueue(db *gorm.DB, opts ...gen.DOOption) workQueue {
	_workQueue := workQueue{}

	_workQueue.workQueueDo.UseDB(db, opts...)
	_workQueue.workQueueDo.UseModel(&models.WorkQueue{})

	tableName := _workQueue.workQueueDo.TableName()
	_workQueue.ALL = field.NewAsterisk(tableName)
	_workQueue.CreatedAt = field.NewTime(tableName, "created_at")
	_workQueue.UpdatedAt = field.NewTime(tableName, "updated_at")
	_workQueue.DeletedAt = field.NewUint(tableName, "deleted_at")
	_workQueue.ID = field.NewInt64(tableName, "id")
	_workQueue.Topic = field.NewString(tableName, "topic")
	_workQueue.Payload = field.NewBytes(tableName, "payload")
	_workQueue.Times = field.NewInt(tableName, "times")
	_workQueue.Version = field.NewString(tableName, "version")
	_workQueue.Status = field.NewField(tableName, "status")

	_workQueue.fillFieldMap()

	return _workQueue
}

type workQueue struct {
	workQueueDo workQueueDo

	ALL       field.Asterisk
	CreatedAt field.Time
	UpdatedAt field.Time
	DeletedAt field.Uint
	ID        field.Int64
	Topic     field.String
	Payload   field.Bytes
	Times     field.Int
	Version   field.String
	Status    field.Field

	fieldMap map[string]field.Expr
}

func (w workQueue) Table(newTableName string) *workQueue {
	w.workQueueDo.UseTable(newTableName)
	return w.updateTableName(newTableName)
}

func (w workQueue) As(alias string) *workQueue {
	w.workQueueDo.DO = *(w.workQueueDo.As(alias).(*gen.DO))
	return w.updateTableName(alias)
}

func (w *workQueue) updateTableName(table string) *workQueue {
	w.ALL = field.NewAsterisk(table)
	w.CreatedAt = field.NewTime(table, "created_at")
	w.UpdatedAt = field.NewTime(table, "updated_at")
	w.DeletedAt = field.NewUint(table, "deleted_at")
	w.ID = field.NewInt64(table, "id")
	w.Topic = field.NewString(table, "topic")
	w.Payload = field.NewBytes(table, "payload")
	w.Times = field.NewInt(table, "times")
	w.Version = field.NewString(table, "version")
	w.Status = field.NewField(table, "status")

	w.fillFieldMap()

	return w
}

func (w *workQueue) WithContext(ctx context.Context) *workQueueDo {
	return w.workQueueDo.WithContext(ctx)
}

func (w workQueue) TableName() string { return w.workQueueDo.TableName() }

func (w workQueue) Alias() string { return w.workQueueDo.Alias() }

func (w workQueue) Columns(cols ...field.Expr) gen.Columns { return w.workQueueDo.Columns(cols...) }

func (w *workQueue) GetFieldByName(fieldName string) (field.OrderExpr, bool) {
	_f, ok := w.fieldMap[fieldName]
	if !ok || _f == nil {
		return nil, false
	}
	_oe, ok := _f.(field.OrderExpr)
	return _oe, ok
}

func (w *workQueue) fillFieldMap() {
	w.fieldMap = make(map[string]field.Expr, 9)
	w.fieldMap["created_at"] = w.CreatedAt
	w.fieldMap["updated_at"] = w.UpdatedAt
	w.fieldMap["deleted_at"] = w.DeletedAt
	w.fieldMap["id"] = w.ID
	w.fieldMap["topic"] = w.Topic
	w.fieldMap["payload"] = w.Payload
	w.fieldMap["times"] = w.Times
	w.fieldMap["version"] = w.Version
	w.fieldMap["status"] = w.Status
}

func (w workQueue) clone(db *gorm.DB) workQueue {
	w.workQueueDo.ReplaceConnPool(db.Statement.ConnPool)
	return w
}

func (w workQueue) replaceDB(db *gorm.DB) workQueue {
	w.workQueueDo.ReplaceDB(db)
	return w
}

type workQueueDo struct{ gen.DO }

func (w workQueueDo) Debug() *workQueueDo {
	return w.withDO(w.DO.Debug())
}

func (w workQueueDo) WithContext(ctx context.Context) *workQueueDo {
	return w.withDO(w.DO.WithContext(ctx))
}

func (w workQueueDo) ReadDB() *workQueueDo {
	return w.Clauses(dbresolver.Read)
}

func (w workQueueDo) WriteDB() *workQueueDo {
	return w.Clauses(dbresolver.Write)
}

func (w workQueueDo) Session(config *gorm.Session) *workQueueDo {
	return w.withDO(w.DO.Session(config))
}

func (w workQueueDo) Clauses(conds ...clause.Expression) *workQueueDo {
	return w.withDO(w.DO.Clauses(conds...))
}

func (w workQueueDo) Returning(value interface{}, columns ...string) *workQueueDo {
	return w.withDO(w.DO.Returning(value, columns...))
}

func (w workQueueDo) Not(conds ...gen.Condition) *workQueueDo {
	return w.withDO(w.DO.Not(conds...))
}

func (w workQueueDo) Or(conds ...gen.Condition) *workQueueDo {
	return w.withDO(w.DO.Or(conds...))
}

func (w workQueueDo) Select(conds ...field.Expr) *workQueueDo {
	return w.withDO(w.DO.Select(conds...))
}

func (w workQueueDo) Where(conds ...gen.Condition) *workQueueDo {
	return w.withDO(w.DO.Where(conds...))
}

func (w workQueueDo) Order(conds ...field.Expr) *workQueueDo {
	return w.withDO(w.DO.Order(conds...))
}

func (w workQueueDo) Distinct(cols ...field.Expr) *workQueueDo {
	return w.withDO(w.DO.Distinct(cols...))
}

func (w workQueueDo) Omit(cols ...field.Expr) *workQueueDo {
	return w.withDO(w.DO.Omit(cols...))
}

func (w workQueueDo) Join(table schema.Tabler, on ...field.Expr) *workQueueDo {
	return w.withDO(w.DO.Join(table, on...))
}

func (w workQueueDo) LeftJoin(table schema.Tabler, on ...field.Expr) *workQueueDo {
	return w.withDO(w.DO.LeftJoin(table, on...))
}

func (w workQueueDo) RightJoin(table schema.Tabler, on ...field.Expr) *workQueueDo {
	return w.withDO(w.DO.RightJoin(table, on...))
}

func (w workQueueDo) Group(cols ...field.Expr) *workQueueDo {
	return w.withDO(w.DO.Group(cols...))
}

func (w workQueueDo) Having(conds ...gen.Condition) *workQueueDo {
	return w.withDO(w.DO.Having(conds...))
}

func (w workQueueDo) Limit(limit int) *workQueueDo {
	return w.withDO(w.DO.Limit(limit))
}

func (w workQueueDo) Offset(offset int) *workQueueDo {
	return w.withDO(w.DO.Offset(offset))
}

func (w workQueueDo) Scopes(funcs ...func(gen.Dao) gen.Dao) *workQueueDo {
	return w.withDO(w.DO.Scopes(funcs...))
}

func (w workQueueDo) Unscoped() *workQueueDo {
	return w.withDO(w.DO.Unscoped())
}

func (w workQueueDo) Create(values ...*models.WorkQueue) error {
	if len(values) == 0 {
		return nil
	}
	return w.DO.Create(values)
}

func (w workQueueDo) CreateInBatches(values []*models.WorkQueue, batchSize int) error {
	return w.DO.CreateInBatches(values, batchSize)
}

// Save : !!! underlying implementation is different with GORM
// The method is equivalent to executing the statement: db.Clauses(clause.OnConflict{UpdateAll: true}).Create(values)
func (w workQueueDo) Save(values ...*models.WorkQueue) error {
	if len(values) == 0 {
		return nil
	}
	return w.DO.Save(values)
}

func (w workQueueDo) First() (*models.WorkQueue, error) {
	if result, err := w.DO.First(); err != nil {
		return nil, err
	} else {
		return result.(*models.WorkQueue), nil
	}
}

func (w workQueueDo) Take() (*models.WorkQueue, error) {
	if result, err := w.DO.Take(); err != nil {
		return nil, err
	} else {
		return result.(*models.WorkQueue), nil
	}
}

func (w workQueueDo) Last() (*models.WorkQueue, error) {
	if result, err := w.DO.Last(); err != nil {
		return nil, err
	} else {
		return result.(*models.WorkQueue), nil
	}
}

func (w workQueueDo) Find() ([]*models.WorkQueue, error) {
	result, err := w.DO.Find()
	return result.([]*models.WorkQueue), err
}

func (w workQueueDo) FindInBatch(batchSize int, fc func(tx gen.Dao, batch int) error) (results []*models.WorkQueue, err error) {
	buf := make([]*models.WorkQueue, 0, batchSize)
	err = w.DO.FindInBatches(&buf, batchSize, func(tx gen.Dao, batch int) error {
		defer func() { results = append(results, buf...) }()
		return fc(tx, batch)
	})
	return results, err
}

func (w workQueueDo) FindInBatches(result *[]*models.WorkQueue, batchSize int, fc func(tx gen.Dao, batch int) error) error {
	return w.DO.FindInBatches(result, batchSize, fc)
}

func (w workQueueDo) Attrs(attrs ...field.AssignExpr) *workQueueDo {
	return w.withDO(w.DO.Attrs(attrs...))
}

func (w workQueueDo) Assign(attrs ...field.AssignExpr) *workQueueDo {
	return w.withDO(w.DO.Assign(attrs...))
}

func (w workQueueDo) Joins(fields ...field.RelationField) *workQueueDo {
	for _, _f := range fields {
		w = *w.withDO(w.DO.Joins(_f))
	}
	return &w
}

func (w workQueueDo) Preload(fields ...field.RelationField) *workQueueDo {
	for _, _f := range fields {
		w = *w.withDO(w.DO.Preload(_f))
	}
	return &w
}

func (w workQueueDo) FirstOrInit() (*models.WorkQueue, error) {
	if result, err := w.DO.FirstOrInit(); err != nil {
		return nil, err
	} else {
		return result.(*models.WorkQueue), nil
	}
}

func (w workQueueDo) FirstOrCreate() (*models.WorkQueue, error) {
	if result, err := w.DO.FirstOrCreate(); err != nil {
		return nil, err
	} else {
		return result.(*models.WorkQueue), nil
	}
}

func (w workQueueDo) FindByPage(offset int, limit int) (result []*models.WorkQueue, count int64, err error) {
	result, err = w.Offset(offset).Limit(limit).Find()
	if err != nil {
		return
	}

	if size := len(result); 0 < limit && 0 < size && size < limit {
		count = int64(size + offset)
		return
	}

	count, err = w.Offset(-1).Limit(-1).Count()
	return
}

func (w workQueueDo) ScanByPage(result interface{}, offset int, limit int) (count int64, err error) {
	count, err = w.Count()
	if err != nil {
		return
	}

	err = w.Offset(offset).Limit(limit).Scan(result)
	return
}

func (w workQueueDo) Scan(result interface{}) (err error) {
	return w.DO.Scan(result)
}

func (w workQueueDo) Delete(models ...*models.WorkQueue) (result gen.ResultInfo, err error) {
	return w.DO.Delete(models)
}

func (w *workQueueDo) withDO(do gen.Dao) *workQueueDo {
	w.DO = *do.(*gen.DO)
	return w
}
