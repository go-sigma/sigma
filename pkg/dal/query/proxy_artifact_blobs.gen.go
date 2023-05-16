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

func newProxyArtifactBlob(db *gorm.DB, opts ...gen.DOOption) proxyArtifactBlob {
	_proxyArtifactBlob := proxyArtifactBlob{}

	_proxyArtifactBlob.proxyArtifactBlobDo.UseDB(db, opts...)
	_proxyArtifactBlob.proxyArtifactBlobDo.UseModel(&models.ProxyArtifactBlob{})

	tableName := _proxyArtifactBlob.proxyArtifactBlobDo.TableName()
	_proxyArtifactBlob.ALL = field.NewAsterisk(tableName)
	_proxyArtifactBlob.CreatedAt = field.NewTime(tableName, "created_at")
	_proxyArtifactBlob.UpdatedAt = field.NewTime(tableName, "updated_at")
	_proxyArtifactBlob.DeletedAt = field.NewUint(tableName, "deleted_at")
	_proxyArtifactBlob.ID = field.NewUint64(tableName, "id")
	_proxyArtifactBlob.ProxyArtifactTaskID = field.NewUint64(tableName, "proxy_artifact_task_id")
	_proxyArtifactBlob.Blob = field.NewString(tableName, "blob")
	_proxyArtifactBlob.ProxyArtifactTask = proxyArtifactBlobBelongsToProxyArtifactTask{
		db: db.Session(&gorm.Session{}),

		RelationField: field.NewRelation("ProxyArtifactTask", "models.ProxyArtifactTask"),
		Blobs: struct {
			field.RelationField
			ProxyArtifactTask struct {
				field.RelationField
			}
		}{
			RelationField: field.NewRelation("ProxyArtifactTask.Blobs", "models.ProxyArtifactBlob"),
			ProxyArtifactTask: struct {
				field.RelationField
			}{
				RelationField: field.NewRelation("ProxyArtifactTask.Blobs.ProxyArtifactTask", "models.ProxyArtifactTask"),
			},
		},
	}

	_proxyArtifactBlob.fillFieldMap()

	return _proxyArtifactBlob
}

type proxyArtifactBlob struct {
	proxyArtifactBlobDo proxyArtifactBlobDo

	ALL                 field.Asterisk
	CreatedAt           field.Time
	UpdatedAt           field.Time
	DeletedAt           field.Uint
	ID                  field.Uint64
	ProxyArtifactTaskID field.Uint64
	Blob                field.String
	ProxyArtifactTask   proxyArtifactBlobBelongsToProxyArtifactTask

	fieldMap map[string]field.Expr
}

func (p proxyArtifactBlob) Table(newTableName string) *proxyArtifactBlob {
	p.proxyArtifactBlobDo.UseTable(newTableName)
	return p.updateTableName(newTableName)
}

func (p proxyArtifactBlob) As(alias string) *proxyArtifactBlob {
	p.proxyArtifactBlobDo.DO = *(p.proxyArtifactBlobDo.As(alias).(*gen.DO))
	return p.updateTableName(alias)
}

func (p *proxyArtifactBlob) updateTableName(table string) *proxyArtifactBlob {
	p.ALL = field.NewAsterisk(table)
	p.CreatedAt = field.NewTime(table, "created_at")
	p.UpdatedAt = field.NewTime(table, "updated_at")
	p.DeletedAt = field.NewUint(table, "deleted_at")
	p.ID = field.NewUint64(table, "id")
	p.ProxyArtifactTaskID = field.NewUint64(table, "proxy_artifact_task_id")
	p.Blob = field.NewString(table, "blob")

	p.fillFieldMap()

	return p
}

func (p *proxyArtifactBlob) WithContext(ctx context.Context) *proxyArtifactBlobDo {
	return p.proxyArtifactBlobDo.WithContext(ctx)
}

func (p proxyArtifactBlob) TableName() string { return p.proxyArtifactBlobDo.TableName() }

func (p proxyArtifactBlob) Alias() string { return p.proxyArtifactBlobDo.Alias() }

func (p *proxyArtifactBlob) GetFieldByName(fieldName string) (field.OrderExpr, bool) {
	_f, ok := p.fieldMap[fieldName]
	if !ok || _f == nil {
		return nil, false
	}
	_oe, ok := _f.(field.OrderExpr)
	return _oe, ok
}

func (p *proxyArtifactBlob) fillFieldMap() {
	p.fieldMap = make(map[string]field.Expr, 7)
	p.fieldMap["created_at"] = p.CreatedAt
	p.fieldMap["updated_at"] = p.UpdatedAt
	p.fieldMap["deleted_at"] = p.DeletedAt
	p.fieldMap["id"] = p.ID
	p.fieldMap["proxy_artifact_task_id"] = p.ProxyArtifactTaskID
	p.fieldMap["blob"] = p.Blob

}

func (p proxyArtifactBlob) clone(db *gorm.DB) proxyArtifactBlob {
	p.proxyArtifactBlobDo.ReplaceConnPool(db.Statement.ConnPool)
	return p
}

func (p proxyArtifactBlob) replaceDB(db *gorm.DB) proxyArtifactBlob {
	p.proxyArtifactBlobDo.ReplaceDB(db)
	return p
}

type proxyArtifactBlobBelongsToProxyArtifactTask struct {
	db *gorm.DB

	field.RelationField

	Blobs struct {
		field.RelationField
		ProxyArtifactTask struct {
			field.RelationField
		}
	}
}

func (a proxyArtifactBlobBelongsToProxyArtifactTask) Where(conds ...field.Expr) *proxyArtifactBlobBelongsToProxyArtifactTask {
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

func (a proxyArtifactBlobBelongsToProxyArtifactTask) WithContext(ctx context.Context) *proxyArtifactBlobBelongsToProxyArtifactTask {
	a.db = a.db.WithContext(ctx)
	return &a
}

func (a proxyArtifactBlobBelongsToProxyArtifactTask) Session(session *gorm.Session) *proxyArtifactBlobBelongsToProxyArtifactTask {
	a.db = a.db.Session(session)
	return &a
}

func (a proxyArtifactBlobBelongsToProxyArtifactTask) Model(m *models.ProxyArtifactBlob) *proxyArtifactBlobBelongsToProxyArtifactTaskTx {
	return &proxyArtifactBlobBelongsToProxyArtifactTaskTx{a.db.Model(m).Association(a.Name())}
}

type proxyArtifactBlobBelongsToProxyArtifactTaskTx struct{ tx *gorm.Association }

func (a proxyArtifactBlobBelongsToProxyArtifactTaskTx) Find() (result *models.ProxyArtifactTask, err error) {
	return result, a.tx.Find(&result)
}

func (a proxyArtifactBlobBelongsToProxyArtifactTaskTx) Append(values ...*models.ProxyArtifactTask) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Append(targetValues...)
}

func (a proxyArtifactBlobBelongsToProxyArtifactTaskTx) Replace(values ...*models.ProxyArtifactTask) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Replace(targetValues...)
}

func (a proxyArtifactBlobBelongsToProxyArtifactTaskTx) Delete(values ...*models.ProxyArtifactTask) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Delete(targetValues...)
}

func (a proxyArtifactBlobBelongsToProxyArtifactTaskTx) Clear() error {
	return a.tx.Clear()
}

func (a proxyArtifactBlobBelongsToProxyArtifactTaskTx) Count() int64 {
	return a.tx.Count()
}

type proxyArtifactBlobDo struct{ gen.DO }

func (p proxyArtifactBlobDo) Debug() *proxyArtifactBlobDo {
	return p.withDO(p.DO.Debug())
}

func (p proxyArtifactBlobDo) WithContext(ctx context.Context) *proxyArtifactBlobDo {
	return p.withDO(p.DO.WithContext(ctx))
}

func (p proxyArtifactBlobDo) ReadDB() *proxyArtifactBlobDo {
	return p.Clauses(dbresolver.Read)
}

func (p proxyArtifactBlobDo) WriteDB() *proxyArtifactBlobDo {
	return p.Clauses(dbresolver.Write)
}

func (p proxyArtifactBlobDo) Session(config *gorm.Session) *proxyArtifactBlobDo {
	return p.withDO(p.DO.Session(config))
}

func (p proxyArtifactBlobDo) Clauses(conds ...clause.Expression) *proxyArtifactBlobDo {
	return p.withDO(p.DO.Clauses(conds...))
}

func (p proxyArtifactBlobDo) Returning(value interface{}, columns ...string) *proxyArtifactBlobDo {
	return p.withDO(p.DO.Returning(value, columns...))
}

func (p proxyArtifactBlobDo) Not(conds ...gen.Condition) *proxyArtifactBlobDo {
	return p.withDO(p.DO.Not(conds...))
}

func (p proxyArtifactBlobDo) Or(conds ...gen.Condition) *proxyArtifactBlobDo {
	return p.withDO(p.DO.Or(conds...))
}

func (p proxyArtifactBlobDo) Select(conds ...field.Expr) *proxyArtifactBlobDo {
	return p.withDO(p.DO.Select(conds...))
}

func (p proxyArtifactBlobDo) Where(conds ...gen.Condition) *proxyArtifactBlobDo {
	return p.withDO(p.DO.Where(conds...))
}

func (p proxyArtifactBlobDo) Exists(subquery interface{ UnderlyingDB() *gorm.DB }) *proxyArtifactBlobDo {
	return p.Where(field.CompareSubQuery(field.ExistsOp, nil, subquery.UnderlyingDB()))
}

func (p proxyArtifactBlobDo) Order(conds ...field.Expr) *proxyArtifactBlobDo {
	return p.withDO(p.DO.Order(conds...))
}

func (p proxyArtifactBlobDo) Distinct(cols ...field.Expr) *proxyArtifactBlobDo {
	return p.withDO(p.DO.Distinct(cols...))
}

func (p proxyArtifactBlobDo) Omit(cols ...field.Expr) *proxyArtifactBlobDo {
	return p.withDO(p.DO.Omit(cols...))
}

func (p proxyArtifactBlobDo) Join(table schema.Tabler, on ...field.Expr) *proxyArtifactBlobDo {
	return p.withDO(p.DO.Join(table, on...))
}

func (p proxyArtifactBlobDo) LeftJoin(table schema.Tabler, on ...field.Expr) *proxyArtifactBlobDo {
	return p.withDO(p.DO.LeftJoin(table, on...))
}

func (p proxyArtifactBlobDo) RightJoin(table schema.Tabler, on ...field.Expr) *proxyArtifactBlobDo {
	return p.withDO(p.DO.RightJoin(table, on...))
}

func (p proxyArtifactBlobDo) Group(cols ...field.Expr) *proxyArtifactBlobDo {
	return p.withDO(p.DO.Group(cols...))
}

func (p proxyArtifactBlobDo) Having(conds ...gen.Condition) *proxyArtifactBlobDo {
	return p.withDO(p.DO.Having(conds...))
}

func (p proxyArtifactBlobDo) Limit(limit int) *proxyArtifactBlobDo {
	return p.withDO(p.DO.Limit(limit))
}

func (p proxyArtifactBlobDo) Offset(offset int) *proxyArtifactBlobDo {
	return p.withDO(p.DO.Offset(offset))
}

func (p proxyArtifactBlobDo) Scopes(funcs ...func(gen.Dao) gen.Dao) *proxyArtifactBlobDo {
	return p.withDO(p.DO.Scopes(funcs...))
}

func (p proxyArtifactBlobDo) Unscoped() *proxyArtifactBlobDo {
	return p.withDO(p.DO.Unscoped())
}

func (p proxyArtifactBlobDo) Create(values ...*models.ProxyArtifactBlob) error {
	if len(values) == 0 {
		return nil
	}
	return p.DO.Create(values)
}

func (p proxyArtifactBlobDo) CreateInBatches(values []*models.ProxyArtifactBlob, batchSize int) error {
	return p.DO.CreateInBatches(values, batchSize)
}

// Save : !!! underlying implementation is different with GORM
// The method is equivalent to executing the statement: db.Clauses(clause.OnConflict{UpdateAll: true}).Create(values)
func (p proxyArtifactBlobDo) Save(values ...*models.ProxyArtifactBlob) error {
	if len(values) == 0 {
		return nil
	}
	return p.DO.Save(values)
}

func (p proxyArtifactBlobDo) First() (*models.ProxyArtifactBlob, error) {
	if result, err := p.DO.First(); err != nil {
		return nil, err
	} else {
		return result.(*models.ProxyArtifactBlob), nil
	}
}

func (p proxyArtifactBlobDo) Take() (*models.ProxyArtifactBlob, error) {
	if result, err := p.DO.Take(); err != nil {
		return nil, err
	} else {
		return result.(*models.ProxyArtifactBlob), nil
	}
}

func (p proxyArtifactBlobDo) Last() (*models.ProxyArtifactBlob, error) {
	if result, err := p.DO.Last(); err != nil {
		return nil, err
	} else {
		return result.(*models.ProxyArtifactBlob), nil
	}
}

func (p proxyArtifactBlobDo) Find() ([]*models.ProxyArtifactBlob, error) {
	result, err := p.DO.Find()
	return result.([]*models.ProxyArtifactBlob), err
}

func (p proxyArtifactBlobDo) FindInBatch(batchSize int, fc func(tx gen.Dao, batch int) error) (results []*models.ProxyArtifactBlob, err error) {
	buf := make([]*models.ProxyArtifactBlob, 0, batchSize)
	err = p.DO.FindInBatches(&buf, batchSize, func(tx gen.Dao, batch int) error {
		defer func() { results = append(results, buf...) }()
		return fc(tx, batch)
	})
	return results, err
}

func (p proxyArtifactBlobDo) FindInBatches(result *[]*models.ProxyArtifactBlob, batchSize int, fc func(tx gen.Dao, batch int) error) error {
	return p.DO.FindInBatches(result, batchSize, fc)
}

func (p proxyArtifactBlobDo) Attrs(attrs ...field.AssignExpr) *proxyArtifactBlobDo {
	return p.withDO(p.DO.Attrs(attrs...))
}

func (p proxyArtifactBlobDo) Assign(attrs ...field.AssignExpr) *proxyArtifactBlobDo {
	return p.withDO(p.DO.Assign(attrs...))
}

func (p proxyArtifactBlobDo) Joins(fields ...field.RelationField) *proxyArtifactBlobDo {
	for _, _f := range fields {
		p = *p.withDO(p.DO.Joins(_f))
	}
	return &p
}

func (p proxyArtifactBlobDo) Preload(fields ...field.RelationField) *proxyArtifactBlobDo {
	for _, _f := range fields {
		p = *p.withDO(p.DO.Preload(_f))
	}
	return &p
}

func (p proxyArtifactBlobDo) FirstOrInit() (*models.ProxyArtifactBlob, error) {
	if result, err := p.DO.FirstOrInit(); err != nil {
		return nil, err
	} else {
		return result.(*models.ProxyArtifactBlob), nil
	}
}

func (p proxyArtifactBlobDo) FirstOrCreate() (*models.ProxyArtifactBlob, error) {
	if result, err := p.DO.FirstOrCreate(); err != nil {
		return nil, err
	} else {
		return result.(*models.ProxyArtifactBlob), nil
	}
}

func (p proxyArtifactBlobDo) FindByPage(offset int, limit int) (result []*models.ProxyArtifactBlob, count int64, err error) {
	result, err = p.Offset(offset).Limit(limit).Find()
	if err != nil {
		return
	}

	if size := len(result); 0 < limit && 0 < size && size < limit {
		count = int64(size + offset)
		return
	}

	count, err = p.Offset(-1).Limit(-1).Count()
	return
}

func (p proxyArtifactBlobDo) ScanByPage(result interface{}, offset int, limit int) (count int64, err error) {
	count, err = p.Count()
	if err != nil {
		return
	}

	err = p.Offset(offset).Limit(limit).Scan(result)
	return
}

func (p proxyArtifactBlobDo) Scan(result interface{}) (err error) {
	return p.DO.Scan(result)
}

func (p proxyArtifactBlobDo) Delete(models ...*models.ProxyArtifactBlob) (result gen.ResultInfo, err error) {
	return p.DO.Delete(models)
}

func (p *proxyArtifactBlobDo) withDO(do gen.Dao) *proxyArtifactBlobDo {
	p.DO = *do.(*gen.DO)
	return p
}
