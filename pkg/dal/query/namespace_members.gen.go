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

func newNamespaceMember(db *gorm.DB, opts ...gen.DOOption) namespaceMember {
	_namespaceMember := namespaceMember{}

	_namespaceMember.namespaceMemberDo.UseDB(db, opts...)
	_namespaceMember.namespaceMemberDo.UseModel(&models.NamespaceMember{})

	tableName := _namespaceMember.namespaceMemberDo.TableName()
	_namespaceMember.ALL = field.NewAsterisk(tableName)
	_namespaceMember.CreatedAt = field.NewTime(tableName, "created_at")
	_namespaceMember.UpdatedAt = field.NewTime(tableName, "updated_at")
	_namespaceMember.DeletedAt = field.NewUint(tableName, "deleted_at")
	_namespaceMember.ID = field.NewInt64(tableName, "id")
	_namespaceMember.UserID = field.NewInt64(tableName, "user_id")
	_namespaceMember.NamespaceID = field.NewInt64(tableName, "namespace_id")
	_namespaceMember.Role = field.NewField(tableName, "role")
	_namespaceMember.User = namespaceMemberBelongsToUser{
		db: db.Session(&gorm.Session{}),

		RelationField: field.NewRelation("User", "models.User"),
	}

	_namespaceMember.Namespace = namespaceMemberBelongsToNamespace{
		db: db.Session(&gorm.Session{}),

		RelationField: field.NewRelation("Namespace", "models.Namespace"),
	}

	_namespaceMember.fillFieldMap()

	return _namespaceMember
}

type namespaceMember struct {
	namespaceMemberDo namespaceMemberDo

	ALL         field.Asterisk
	CreatedAt   field.Time
	UpdatedAt   field.Time
	DeletedAt   field.Uint
	ID          field.Int64
	UserID      field.Int64
	NamespaceID field.Int64
	Role        field.Field
	User        namespaceMemberBelongsToUser

	Namespace namespaceMemberBelongsToNamespace

	fieldMap map[string]field.Expr
}

func (n namespaceMember) Table(newTableName string) *namespaceMember {
	n.namespaceMemberDo.UseTable(newTableName)
	return n.updateTableName(newTableName)
}

func (n namespaceMember) As(alias string) *namespaceMember {
	n.namespaceMemberDo.DO = *(n.namespaceMemberDo.As(alias).(*gen.DO))
	return n.updateTableName(alias)
}

func (n *namespaceMember) updateTableName(table string) *namespaceMember {
	n.ALL = field.NewAsterisk(table)
	n.CreatedAt = field.NewTime(table, "created_at")
	n.UpdatedAt = field.NewTime(table, "updated_at")
	n.DeletedAt = field.NewUint(table, "deleted_at")
	n.ID = field.NewInt64(table, "id")
	n.UserID = field.NewInt64(table, "user_id")
	n.NamespaceID = field.NewInt64(table, "namespace_id")
	n.Role = field.NewField(table, "role")

	n.fillFieldMap()

	return n
}

func (n *namespaceMember) WithContext(ctx context.Context) *namespaceMemberDo {
	return n.namespaceMemberDo.WithContext(ctx)
}

func (n namespaceMember) TableName() string { return n.namespaceMemberDo.TableName() }

func (n namespaceMember) Alias() string { return n.namespaceMemberDo.Alias() }

func (n namespaceMember) Columns(cols ...field.Expr) gen.Columns {
	return n.namespaceMemberDo.Columns(cols...)
}

func (n *namespaceMember) GetFieldByName(fieldName string) (field.OrderExpr, bool) {
	_f, ok := n.fieldMap[fieldName]
	if !ok || _f == nil {
		return nil, false
	}
	_oe, ok := _f.(field.OrderExpr)
	return _oe, ok
}

func (n *namespaceMember) fillFieldMap() {
	n.fieldMap = make(map[string]field.Expr, 9)
	n.fieldMap["created_at"] = n.CreatedAt
	n.fieldMap["updated_at"] = n.UpdatedAt
	n.fieldMap["deleted_at"] = n.DeletedAt
	n.fieldMap["id"] = n.ID
	n.fieldMap["user_id"] = n.UserID
	n.fieldMap["namespace_id"] = n.NamespaceID
	n.fieldMap["role"] = n.Role

}

func (n namespaceMember) clone(db *gorm.DB) namespaceMember {
	n.namespaceMemberDo.ReplaceConnPool(db.Statement.ConnPool)
	return n
}

func (n namespaceMember) replaceDB(db *gorm.DB) namespaceMember {
	n.namespaceMemberDo.ReplaceDB(db)
	return n
}

type namespaceMemberBelongsToUser struct {
	db *gorm.DB

	field.RelationField
}

func (a namespaceMemberBelongsToUser) Where(conds ...field.Expr) *namespaceMemberBelongsToUser {
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

func (a namespaceMemberBelongsToUser) WithContext(ctx context.Context) *namespaceMemberBelongsToUser {
	a.db = a.db.WithContext(ctx)
	return &a
}

func (a namespaceMemberBelongsToUser) Session(session *gorm.Session) *namespaceMemberBelongsToUser {
	a.db = a.db.Session(session)
	return &a
}

func (a namespaceMemberBelongsToUser) Model(m *models.NamespaceMember) *namespaceMemberBelongsToUserTx {
	return &namespaceMemberBelongsToUserTx{a.db.Model(m).Association(a.Name())}
}

type namespaceMemberBelongsToUserTx struct{ tx *gorm.Association }

func (a namespaceMemberBelongsToUserTx) Find() (result *models.User, err error) {
	return result, a.tx.Find(&result)
}

func (a namespaceMemberBelongsToUserTx) Append(values ...*models.User) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Append(targetValues...)
}

func (a namespaceMemberBelongsToUserTx) Replace(values ...*models.User) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Replace(targetValues...)
}

func (a namespaceMemberBelongsToUserTx) Delete(values ...*models.User) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Delete(targetValues...)
}

func (a namespaceMemberBelongsToUserTx) Clear() error {
	return a.tx.Clear()
}

func (a namespaceMemberBelongsToUserTx) Count() int64 {
	return a.tx.Count()
}

type namespaceMemberBelongsToNamespace struct {
	db *gorm.DB

	field.RelationField
}

func (a namespaceMemberBelongsToNamespace) Where(conds ...field.Expr) *namespaceMemberBelongsToNamespace {
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

func (a namespaceMemberBelongsToNamespace) WithContext(ctx context.Context) *namespaceMemberBelongsToNamespace {
	a.db = a.db.WithContext(ctx)
	return &a
}

func (a namespaceMemberBelongsToNamespace) Session(session *gorm.Session) *namespaceMemberBelongsToNamespace {
	a.db = a.db.Session(session)
	return &a
}

func (a namespaceMemberBelongsToNamespace) Model(m *models.NamespaceMember) *namespaceMemberBelongsToNamespaceTx {
	return &namespaceMemberBelongsToNamespaceTx{a.db.Model(m).Association(a.Name())}
}

type namespaceMemberBelongsToNamespaceTx struct{ tx *gorm.Association }

func (a namespaceMemberBelongsToNamespaceTx) Find() (result *models.Namespace, err error) {
	return result, a.tx.Find(&result)
}

func (a namespaceMemberBelongsToNamespaceTx) Append(values ...*models.Namespace) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Append(targetValues...)
}

func (a namespaceMemberBelongsToNamespaceTx) Replace(values ...*models.Namespace) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Replace(targetValues...)
}

func (a namespaceMemberBelongsToNamespaceTx) Delete(values ...*models.Namespace) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Delete(targetValues...)
}

func (a namespaceMemberBelongsToNamespaceTx) Clear() error {
	return a.tx.Clear()
}

func (a namespaceMemberBelongsToNamespaceTx) Count() int64 {
	return a.tx.Count()
}

type namespaceMemberDo struct{ gen.DO }

func (n namespaceMemberDo) Debug() *namespaceMemberDo {
	return n.withDO(n.DO.Debug())
}

func (n namespaceMemberDo) WithContext(ctx context.Context) *namespaceMemberDo {
	return n.withDO(n.DO.WithContext(ctx))
}

func (n namespaceMemberDo) ReadDB() *namespaceMemberDo {
	return n.Clauses(dbresolver.Read)
}

func (n namespaceMemberDo) WriteDB() *namespaceMemberDo {
	return n.Clauses(dbresolver.Write)
}

func (n namespaceMemberDo) Session(config *gorm.Session) *namespaceMemberDo {
	return n.withDO(n.DO.Session(config))
}

func (n namespaceMemberDo) Clauses(conds ...clause.Expression) *namespaceMemberDo {
	return n.withDO(n.DO.Clauses(conds...))
}

func (n namespaceMemberDo) Returning(value interface{}, columns ...string) *namespaceMemberDo {
	return n.withDO(n.DO.Returning(value, columns...))
}

func (n namespaceMemberDo) Not(conds ...gen.Condition) *namespaceMemberDo {
	return n.withDO(n.DO.Not(conds...))
}

func (n namespaceMemberDo) Or(conds ...gen.Condition) *namespaceMemberDo {
	return n.withDO(n.DO.Or(conds...))
}

func (n namespaceMemberDo) Select(conds ...field.Expr) *namespaceMemberDo {
	return n.withDO(n.DO.Select(conds...))
}

func (n namespaceMemberDo) Where(conds ...gen.Condition) *namespaceMemberDo {
	return n.withDO(n.DO.Where(conds...))
}

func (n namespaceMemberDo) Order(conds ...field.Expr) *namespaceMemberDo {
	return n.withDO(n.DO.Order(conds...))
}

func (n namespaceMemberDo) Distinct(cols ...field.Expr) *namespaceMemberDo {
	return n.withDO(n.DO.Distinct(cols...))
}

func (n namespaceMemberDo) Omit(cols ...field.Expr) *namespaceMemberDo {
	return n.withDO(n.DO.Omit(cols...))
}

func (n namespaceMemberDo) Join(table schema.Tabler, on ...field.Expr) *namespaceMemberDo {
	return n.withDO(n.DO.Join(table, on...))
}

func (n namespaceMemberDo) LeftJoin(table schema.Tabler, on ...field.Expr) *namespaceMemberDo {
	return n.withDO(n.DO.LeftJoin(table, on...))
}

func (n namespaceMemberDo) RightJoin(table schema.Tabler, on ...field.Expr) *namespaceMemberDo {
	return n.withDO(n.DO.RightJoin(table, on...))
}

func (n namespaceMemberDo) Group(cols ...field.Expr) *namespaceMemberDo {
	return n.withDO(n.DO.Group(cols...))
}

func (n namespaceMemberDo) Having(conds ...gen.Condition) *namespaceMemberDo {
	return n.withDO(n.DO.Having(conds...))
}

func (n namespaceMemberDo) Limit(limit int) *namespaceMemberDo {
	return n.withDO(n.DO.Limit(limit))
}

func (n namespaceMemberDo) Offset(offset int) *namespaceMemberDo {
	return n.withDO(n.DO.Offset(offset))
}

func (n namespaceMemberDo) Scopes(funcs ...func(gen.Dao) gen.Dao) *namespaceMemberDo {
	return n.withDO(n.DO.Scopes(funcs...))
}

func (n namespaceMemberDo) Unscoped() *namespaceMemberDo {
	return n.withDO(n.DO.Unscoped())
}

func (n namespaceMemberDo) Create(values ...*models.NamespaceMember) error {
	if len(values) == 0 {
		return nil
	}
	return n.DO.Create(values)
}

func (n namespaceMemberDo) CreateInBatches(values []*models.NamespaceMember, batchSize int) error {
	return n.DO.CreateInBatches(values, batchSize)
}

// Save : !!! underlying implementation is different with GORM
// The method is equivalent to executing the statement: db.Clauses(clause.OnConflict{UpdateAll: true}).Create(values)
func (n namespaceMemberDo) Save(values ...*models.NamespaceMember) error {
	if len(values) == 0 {
		return nil
	}
	return n.DO.Save(values)
}

func (n namespaceMemberDo) First() (*models.NamespaceMember, error) {
	if result, err := n.DO.First(); err != nil {
		return nil, err
	} else {
		return result.(*models.NamespaceMember), nil
	}
}

func (n namespaceMemberDo) Take() (*models.NamespaceMember, error) {
	if result, err := n.DO.Take(); err != nil {
		return nil, err
	} else {
		return result.(*models.NamespaceMember), nil
	}
}

func (n namespaceMemberDo) Last() (*models.NamespaceMember, error) {
	if result, err := n.DO.Last(); err != nil {
		return nil, err
	} else {
		return result.(*models.NamespaceMember), nil
	}
}

func (n namespaceMemberDo) Find() ([]*models.NamespaceMember, error) {
	result, err := n.DO.Find()
	return result.([]*models.NamespaceMember), err
}

func (n namespaceMemberDo) FindInBatch(batchSize int, fc func(tx gen.Dao, batch int) error) (results []*models.NamespaceMember, err error) {
	buf := make([]*models.NamespaceMember, 0, batchSize)
	err = n.DO.FindInBatches(&buf, batchSize, func(tx gen.Dao, batch int) error {
		defer func() { results = append(results, buf...) }()
		return fc(tx, batch)
	})
	return results, err
}

func (n namespaceMemberDo) FindInBatches(result *[]*models.NamespaceMember, batchSize int, fc func(tx gen.Dao, batch int) error) error {
	return n.DO.FindInBatches(result, batchSize, fc)
}

func (n namespaceMemberDo) Attrs(attrs ...field.AssignExpr) *namespaceMemberDo {
	return n.withDO(n.DO.Attrs(attrs...))
}

func (n namespaceMemberDo) Assign(attrs ...field.AssignExpr) *namespaceMemberDo {
	return n.withDO(n.DO.Assign(attrs...))
}

func (n namespaceMemberDo) Joins(fields ...field.RelationField) *namespaceMemberDo {
	for _, _f := range fields {
		n = *n.withDO(n.DO.Joins(_f))
	}
	return &n
}

func (n namespaceMemberDo) Preload(fields ...field.RelationField) *namespaceMemberDo {
	for _, _f := range fields {
		n = *n.withDO(n.DO.Preload(_f))
	}
	return &n
}

func (n namespaceMemberDo) FirstOrInit() (*models.NamespaceMember, error) {
	if result, err := n.DO.FirstOrInit(); err != nil {
		return nil, err
	} else {
		return result.(*models.NamespaceMember), nil
	}
}

func (n namespaceMemberDo) FirstOrCreate() (*models.NamespaceMember, error) {
	if result, err := n.DO.FirstOrCreate(); err != nil {
		return nil, err
	} else {
		return result.(*models.NamespaceMember), nil
	}
}

func (n namespaceMemberDo) FindByPage(offset int, limit int) (result []*models.NamespaceMember, count int64, err error) {
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

func (n namespaceMemberDo) ScanByPage(result interface{}, offset int, limit int) (count int64, err error) {
	count, err = n.Count()
	if err != nil {
		return
	}

	err = n.Offset(offset).Limit(limit).Scan(result)
	return
}

func (n namespaceMemberDo) Scan(result interface{}) (err error) {
	return n.DO.Scan(result)
}

func (n namespaceMemberDo) Delete(models ...*models.NamespaceMember) (result gen.ResultInfo, err error) {
	return n.DO.Delete(models)
}

func (n *namespaceMemberDo) withDO(do gen.Dao) *namespaceMemberDo {
	n.DO = *do.(*gen.DO)
	return n
}