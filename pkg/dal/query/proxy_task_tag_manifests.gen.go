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

func newProxyTaskTagManifest(db *gorm.DB, opts ...gen.DOOption) proxyTaskTagManifest {
	_proxyTaskTagManifest := proxyTaskTagManifest{}

	_proxyTaskTagManifest.proxyTaskTagManifestDo.UseDB(db, opts...)
	_proxyTaskTagManifest.proxyTaskTagManifestDo.UseModel(&models.ProxyTaskTagManifest{})

	tableName := _proxyTaskTagManifest.proxyTaskTagManifestDo.TableName()
	_proxyTaskTagManifest.ALL = field.NewAsterisk(tableName)
	_proxyTaskTagManifest.CreatedAt = field.NewTime(tableName, "created_at")
	_proxyTaskTagManifest.UpdatedAt = field.NewTime(tableName, "updated_at")
	_proxyTaskTagManifest.DeletedAt = field.NewUint(tableName, "deleted_at")
	_proxyTaskTagManifest.ID = field.NewUint64(tableName, "id")
	_proxyTaskTagManifest.ProxyTaskTagID = field.NewUint64(tableName, "proxy_task_tag_id")
	_proxyTaskTagManifest.Digest = field.NewString(tableName, "digest")

	_proxyTaskTagManifest.fillFieldMap()

	return _proxyTaskTagManifest
}

type proxyTaskTagManifest struct {
	proxyTaskTagManifestDo proxyTaskTagManifestDo

	ALL            field.Asterisk
	CreatedAt      field.Time
	UpdatedAt      field.Time
	DeletedAt      field.Uint
	ID             field.Uint64
	ProxyTaskTagID field.Uint64
	Digest         field.String

	fieldMap map[string]field.Expr
}

func (p proxyTaskTagManifest) Table(newTableName string) *proxyTaskTagManifest {
	p.proxyTaskTagManifestDo.UseTable(newTableName)
	return p.updateTableName(newTableName)
}

func (p proxyTaskTagManifest) As(alias string) *proxyTaskTagManifest {
	p.proxyTaskTagManifestDo.DO = *(p.proxyTaskTagManifestDo.As(alias).(*gen.DO))
	return p.updateTableName(alias)
}

func (p *proxyTaskTagManifest) updateTableName(table string) *proxyTaskTagManifest {
	p.ALL = field.NewAsterisk(table)
	p.CreatedAt = field.NewTime(table, "created_at")
	p.UpdatedAt = field.NewTime(table, "updated_at")
	p.DeletedAt = field.NewUint(table, "deleted_at")
	p.ID = field.NewUint64(table, "id")
	p.ProxyTaskTagID = field.NewUint64(table, "proxy_task_tag_id")
	p.Digest = field.NewString(table, "digest")

	p.fillFieldMap()

	return p
}

func (p *proxyTaskTagManifest) WithContext(ctx context.Context) *proxyTaskTagManifestDo {
	return p.proxyTaskTagManifestDo.WithContext(ctx)
}

func (p proxyTaskTagManifest) TableName() string { return p.proxyTaskTagManifestDo.TableName() }

func (p proxyTaskTagManifest) Alias() string { return p.proxyTaskTagManifestDo.Alias() }

func (p *proxyTaskTagManifest) GetFieldByName(fieldName string) (field.OrderExpr, bool) {
	_f, ok := p.fieldMap[fieldName]
	if !ok || _f == nil {
		return nil, false
	}
	_oe, ok := _f.(field.OrderExpr)
	return _oe, ok
}

func (p *proxyTaskTagManifest) fillFieldMap() {
	p.fieldMap = make(map[string]field.Expr, 6)
	p.fieldMap["created_at"] = p.CreatedAt
	p.fieldMap["updated_at"] = p.UpdatedAt
	p.fieldMap["deleted_at"] = p.DeletedAt
	p.fieldMap["id"] = p.ID
	p.fieldMap["proxy_task_tag_id"] = p.ProxyTaskTagID
	p.fieldMap["digest"] = p.Digest
}

func (p proxyTaskTagManifest) clone(db *gorm.DB) proxyTaskTagManifest {
	p.proxyTaskTagManifestDo.ReplaceConnPool(db.Statement.ConnPool)
	return p
}

func (p proxyTaskTagManifest) replaceDB(db *gorm.DB) proxyTaskTagManifest {
	p.proxyTaskTagManifestDo.ReplaceDB(db)
	return p
}

type proxyTaskTagManifestDo struct{ gen.DO }

func (p proxyTaskTagManifestDo) Debug() *proxyTaskTagManifestDo {
	return p.withDO(p.DO.Debug())
}

func (p proxyTaskTagManifestDo) WithContext(ctx context.Context) *proxyTaskTagManifestDo {
	return p.withDO(p.DO.WithContext(ctx))
}

func (p proxyTaskTagManifestDo) ReadDB() *proxyTaskTagManifestDo {
	return p.Clauses(dbresolver.Read)
}

func (p proxyTaskTagManifestDo) WriteDB() *proxyTaskTagManifestDo {
	return p.Clauses(dbresolver.Write)
}

func (p proxyTaskTagManifestDo) Session(config *gorm.Session) *proxyTaskTagManifestDo {
	return p.withDO(p.DO.Session(config))
}

func (p proxyTaskTagManifestDo) Clauses(conds ...clause.Expression) *proxyTaskTagManifestDo {
	return p.withDO(p.DO.Clauses(conds...))
}

func (p proxyTaskTagManifestDo) Returning(value interface{}, columns ...string) *proxyTaskTagManifestDo {
	return p.withDO(p.DO.Returning(value, columns...))
}

func (p proxyTaskTagManifestDo) Not(conds ...gen.Condition) *proxyTaskTagManifestDo {
	return p.withDO(p.DO.Not(conds...))
}

func (p proxyTaskTagManifestDo) Or(conds ...gen.Condition) *proxyTaskTagManifestDo {
	return p.withDO(p.DO.Or(conds...))
}

func (p proxyTaskTagManifestDo) Select(conds ...field.Expr) *proxyTaskTagManifestDo {
	return p.withDO(p.DO.Select(conds...))
}

func (p proxyTaskTagManifestDo) Where(conds ...gen.Condition) *proxyTaskTagManifestDo {
	return p.withDO(p.DO.Where(conds...))
}

func (p proxyTaskTagManifestDo) Exists(subquery interface{ UnderlyingDB() *gorm.DB }) *proxyTaskTagManifestDo {
	return p.Where(field.CompareSubQuery(field.ExistsOp, nil, subquery.UnderlyingDB()))
}

func (p proxyTaskTagManifestDo) Order(conds ...field.Expr) *proxyTaskTagManifestDo {
	return p.withDO(p.DO.Order(conds...))
}

func (p proxyTaskTagManifestDo) Distinct(cols ...field.Expr) *proxyTaskTagManifestDo {
	return p.withDO(p.DO.Distinct(cols...))
}

func (p proxyTaskTagManifestDo) Omit(cols ...field.Expr) *proxyTaskTagManifestDo {
	return p.withDO(p.DO.Omit(cols...))
}

func (p proxyTaskTagManifestDo) Join(table schema.Tabler, on ...field.Expr) *proxyTaskTagManifestDo {
	return p.withDO(p.DO.Join(table, on...))
}

func (p proxyTaskTagManifestDo) LeftJoin(table schema.Tabler, on ...field.Expr) *proxyTaskTagManifestDo {
	return p.withDO(p.DO.LeftJoin(table, on...))
}

func (p proxyTaskTagManifestDo) RightJoin(table schema.Tabler, on ...field.Expr) *proxyTaskTagManifestDo {
	return p.withDO(p.DO.RightJoin(table, on...))
}

func (p proxyTaskTagManifestDo) Group(cols ...field.Expr) *proxyTaskTagManifestDo {
	return p.withDO(p.DO.Group(cols...))
}

func (p proxyTaskTagManifestDo) Having(conds ...gen.Condition) *proxyTaskTagManifestDo {
	return p.withDO(p.DO.Having(conds...))
}

func (p proxyTaskTagManifestDo) Limit(limit int) *proxyTaskTagManifestDo {
	return p.withDO(p.DO.Limit(limit))
}

func (p proxyTaskTagManifestDo) Offset(offset int) *proxyTaskTagManifestDo {
	return p.withDO(p.DO.Offset(offset))
}

func (p proxyTaskTagManifestDo) Scopes(funcs ...func(gen.Dao) gen.Dao) *proxyTaskTagManifestDo {
	return p.withDO(p.DO.Scopes(funcs...))
}

func (p proxyTaskTagManifestDo) Unscoped() *proxyTaskTagManifestDo {
	return p.withDO(p.DO.Unscoped())
}

func (p proxyTaskTagManifestDo) Create(values ...*models.ProxyTaskTagManifest) error {
	if len(values) == 0 {
		return nil
	}
	return p.DO.Create(values)
}

func (p proxyTaskTagManifestDo) CreateInBatches(values []*models.ProxyTaskTagManifest, batchSize int) error {
	return p.DO.CreateInBatches(values, batchSize)
}

// Save : !!! underlying implementation is different with GORM
// The method is equivalent to executing the statement: db.Clauses(clause.OnConflict{UpdateAll: true}).Create(values)
func (p proxyTaskTagManifestDo) Save(values ...*models.ProxyTaskTagManifest) error {
	if len(values) == 0 {
		return nil
	}
	return p.DO.Save(values)
}

func (p proxyTaskTagManifestDo) First() (*models.ProxyTaskTagManifest, error) {
	if result, err := p.DO.First(); err != nil {
		return nil, err
	} else {
		return result.(*models.ProxyTaskTagManifest), nil
	}
}

func (p proxyTaskTagManifestDo) Take() (*models.ProxyTaskTagManifest, error) {
	if result, err := p.DO.Take(); err != nil {
		return nil, err
	} else {
		return result.(*models.ProxyTaskTagManifest), nil
	}
}

func (p proxyTaskTagManifestDo) Last() (*models.ProxyTaskTagManifest, error) {
	if result, err := p.DO.Last(); err != nil {
		return nil, err
	} else {
		return result.(*models.ProxyTaskTagManifest), nil
	}
}

func (p proxyTaskTagManifestDo) Find() ([]*models.ProxyTaskTagManifest, error) {
	result, err := p.DO.Find()
	return result.([]*models.ProxyTaskTagManifest), err
}

func (p proxyTaskTagManifestDo) FindInBatch(batchSize int, fc func(tx gen.Dao, batch int) error) (results []*models.ProxyTaskTagManifest, err error) {
	buf := make([]*models.ProxyTaskTagManifest, 0, batchSize)
	err = p.DO.FindInBatches(&buf, batchSize, func(tx gen.Dao, batch int) error {
		defer func() { results = append(results, buf...) }()
		return fc(tx, batch)
	})
	return results, err
}

func (p proxyTaskTagManifestDo) FindInBatches(result *[]*models.ProxyTaskTagManifest, batchSize int, fc func(tx gen.Dao, batch int) error) error {
	return p.DO.FindInBatches(result, batchSize, fc)
}

func (p proxyTaskTagManifestDo) Attrs(attrs ...field.AssignExpr) *proxyTaskTagManifestDo {
	return p.withDO(p.DO.Attrs(attrs...))
}

func (p proxyTaskTagManifestDo) Assign(attrs ...field.AssignExpr) *proxyTaskTagManifestDo {
	return p.withDO(p.DO.Assign(attrs...))
}

func (p proxyTaskTagManifestDo) Joins(fields ...field.RelationField) *proxyTaskTagManifestDo {
	for _, _f := range fields {
		p = *p.withDO(p.DO.Joins(_f))
	}
	return &p
}

func (p proxyTaskTagManifestDo) Preload(fields ...field.RelationField) *proxyTaskTagManifestDo {
	for _, _f := range fields {
		p = *p.withDO(p.DO.Preload(_f))
	}
	return &p
}

func (p proxyTaskTagManifestDo) FirstOrInit() (*models.ProxyTaskTagManifest, error) {
	if result, err := p.DO.FirstOrInit(); err != nil {
		return nil, err
	} else {
		return result.(*models.ProxyTaskTagManifest), nil
	}
}

func (p proxyTaskTagManifestDo) FirstOrCreate() (*models.ProxyTaskTagManifest, error) {
	if result, err := p.DO.FirstOrCreate(); err != nil {
		return nil, err
	} else {
		return result.(*models.ProxyTaskTagManifest), nil
	}
}

func (p proxyTaskTagManifestDo) FindByPage(offset int, limit int) (result []*models.ProxyTaskTagManifest, count int64, err error) {
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

func (p proxyTaskTagManifestDo) ScanByPage(result interface{}, offset int, limit int) (count int64, err error) {
	count, err = p.Count()
	if err != nil {
		return
	}

	err = p.Offset(offset).Limit(limit).Scan(result)
	return
}

func (p proxyTaskTagManifestDo) Scan(result interface{}) (err error) {
	return p.DO.Scan(result)
}

func (p proxyTaskTagManifestDo) Delete(models ...*models.ProxyTaskTagManifest) (result gen.ResultInfo, err error) {
	return p.DO.Delete(models)
}

func (p *proxyTaskTagManifestDo) withDO(do gen.Dao) *proxyTaskTagManifestDo {
	p.DO = *do.(*gen.DO)
	return p
}
