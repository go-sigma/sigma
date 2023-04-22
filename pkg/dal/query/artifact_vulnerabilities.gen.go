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

func newArtifactVulnerability(db *gorm.DB, opts ...gen.DOOption) artifactVulnerability {
	_artifactVulnerability := artifactVulnerability{}

	_artifactVulnerability.artifactVulnerabilityDo.UseDB(db, opts...)
	_artifactVulnerability.artifactVulnerabilityDo.UseModel(&models.ArtifactVulnerability{})

	tableName := _artifactVulnerability.artifactVulnerabilityDo.TableName()
	_artifactVulnerability.ALL = field.NewAsterisk(tableName)
	_artifactVulnerability.CreatedAt = field.NewTime(tableName, "created_at")
	_artifactVulnerability.UpdatedAt = field.NewTime(tableName, "updated_at")
	_artifactVulnerability.DeletedAt = field.NewUint(tableName, "deleted_at")
	_artifactVulnerability.ID = field.NewUint64(tableName, "id")
	_artifactVulnerability.ArtifactID = field.NewUint64(tableName, "artifact_id")
	_artifactVulnerability.Metadata = field.NewBytes(tableName, "metadata")
	_artifactVulnerability.Raw = field.NewBytes(tableName, "raw")
	_artifactVulnerability.Status = field.NewString(tableName, "status")
	_artifactVulnerability.Stdout = field.NewBytes(tableName, "stdout")
	_artifactVulnerability.Stderr = field.NewBytes(tableName, "stderr")
	_artifactVulnerability.Message = field.NewString(tableName, "message")
	_artifactVulnerability.Artifact = artifactVulnerabilityBelongsToArtifact{
		db: db.Session(&gorm.Session{}),

		RelationField: field.NewRelation("Artifact", "models.Artifact"),
		Repository: struct {
			field.RelationField
			Namespace struct {
				field.RelationField
			}
		}{
			RelationField: field.NewRelation("Artifact.Repository", "models.Repository"),
			Namespace: struct {
				field.RelationField
			}{
				RelationField: field.NewRelation("Artifact.Repository.Namespace", "models.Namespace"),
			},
		},
		Tags: struct {
			field.RelationField
			Repository struct {
				field.RelationField
			}
			Artifact struct {
				field.RelationField
			}
		}{
			RelationField: field.NewRelation("Artifact.Tags", "models.Tag"),
			Repository: struct {
				field.RelationField
			}{
				RelationField: field.NewRelation("Artifact.Tags.Repository", "models.Repository"),
			},
			Artifact: struct {
				field.RelationField
			}{
				RelationField: field.NewRelation("Artifact.Tags.Artifact", "models.Artifact"),
			},
		},
		Blobs: struct {
			field.RelationField
			Artifacts struct {
				field.RelationField
			}
		}{
			RelationField: field.NewRelation("Artifact.Blobs", "models.Blob"),
			Artifacts: struct {
				field.RelationField
			}{
				RelationField: field.NewRelation("Artifact.Blobs.Artifacts", "models.Artifact"),
			},
		},
	}

	_artifactVulnerability.fillFieldMap()

	return _artifactVulnerability
}

type artifactVulnerability struct {
	artifactVulnerabilityDo artifactVulnerabilityDo

	ALL        field.Asterisk
	CreatedAt  field.Time
	UpdatedAt  field.Time
	DeletedAt  field.Uint
	ID         field.Uint64
	ArtifactID field.Uint64
	Metadata   field.Bytes
	Raw        field.Bytes
	Status     field.String
	Stdout     field.Bytes
	Stderr     field.Bytes
	Message    field.String
	Artifact   artifactVulnerabilityBelongsToArtifact

	fieldMap map[string]field.Expr
}

func (a artifactVulnerability) Table(newTableName string) *artifactVulnerability {
	a.artifactVulnerabilityDo.UseTable(newTableName)
	return a.updateTableName(newTableName)
}

func (a artifactVulnerability) As(alias string) *artifactVulnerability {
	a.artifactVulnerabilityDo.DO = *(a.artifactVulnerabilityDo.As(alias).(*gen.DO))
	return a.updateTableName(alias)
}

func (a *artifactVulnerability) updateTableName(table string) *artifactVulnerability {
	a.ALL = field.NewAsterisk(table)
	a.CreatedAt = field.NewTime(table, "created_at")
	a.UpdatedAt = field.NewTime(table, "updated_at")
	a.DeletedAt = field.NewUint(table, "deleted_at")
	a.ID = field.NewUint64(table, "id")
	a.ArtifactID = field.NewUint64(table, "artifact_id")
	a.Metadata = field.NewBytes(table, "metadata")
	a.Raw = field.NewBytes(table, "raw")
	a.Status = field.NewString(table, "status")
	a.Stdout = field.NewBytes(table, "stdout")
	a.Stderr = field.NewBytes(table, "stderr")
	a.Message = field.NewString(table, "message")

	a.fillFieldMap()

	return a
}

func (a *artifactVulnerability) WithContext(ctx context.Context) *artifactVulnerabilityDo {
	return a.artifactVulnerabilityDo.WithContext(ctx)
}

func (a artifactVulnerability) TableName() string { return a.artifactVulnerabilityDo.TableName() }

func (a artifactVulnerability) Alias() string { return a.artifactVulnerabilityDo.Alias() }

func (a *artifactVulnerability) GetFieldByName(fieldName string) (field.OrderExpr, bool) {
	_f, ok := a.fieldMap[fieldName]
	if !ok || _f == nil {
		return nil, false
	}
	_oe, ok := _f.(field.OrderExpr)
	return _oe, ok
}

func (a *artifactVulnerability) fillFieldMap() {
	a.fieldMap = make(map[string]field.Expr, 12)
	a.fieldMap["created_at"] = a.CreatedAt
	a.fieldMap["updated_at"] = a.UpdatedAt
	a.fieldMap["deleted_at"] = a.DeletedAt
	a.fieldMap["id"] = a.ID
	a.fieldMap["artifact_id"] = a.ArtifactID
	a.fieldMap["metadata"] = a.Metadata
	a.fieldMap["raw"] = a.Raw
	a.fieldMap["status"] = a.Status
	a.fieldMap["stdout"] = a.Stdout
	a.fieldMap["stderr"] = a.Stderr
	a.fieldMap["message"] = a.Message

}

func (a artifactVulnerability) clone(db *gorm.DB) artifactVulnerability {
	a.artifactVulnerabilityDo.ReplaceConnPool(db.Statement.ConnPool)
	return a
}

func (a artifactVulnerability) replaceDB(db *gorm.DB) artifactVulnerability {
	a.artifactVulnerabilityDo.ReplaceDB(db)
	return a
}

type artifactVulnerabilityBelongsToArtifact struct {
	db *gorm.DB

	field.RelationField

	Repository struct {
		field.RelationField
		Namespace struct {
			field.RelationField
		}
	}
	Tags struct {
		field.RelationField
		Repository struct {
			field.RelationField
		}
		Artifact struct {
			field.RelationField
		}
	}
	Blobs struct {
		field.RelationField
		Artifacts struct {
			field.RelationField
		}
	}
}

func (a artifactVulnerabilityBelongsToArtifact) Where(conds ...field.Expr) *artifactVulnerabilityBelongsToArtifact {
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

func (a artifactVulnerabilityBelongsToArtifact) WithContext(ctx context.Context) *artifactVulnerabilityBelongsToArtifact {
	a.db = a.db.WithContext(ctx)
	return &a
}

func (a artifactVulnerabilityBelongsToArtifact) Model(m *models.ArtifactVulnerability) *artifactVulnerabilityBelongsToArtifactTx {
	return &artifactVulnerabilityBelongsToArtifactTx{a.db.Model(m).Association(a.Name())}
}

type artifactVulnerabilityBelongsToArtifactTx struct{ tx *gorm.Association }

func (a artifactVulnerabilityBelongsToArtifactTx) Find() (result *models.Artifact, err error) {
	return result, a.tx.Find(&result)
}

func (a artifactVulnerabilityBelongsToArtifactTx) Append(values ...*models.Artifact) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Append(targetValues...)
}

func (a artifactVulnerabilityBelongsToArtifactTx) Replace(values ...*models.Artifact) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Replace(targetValues...)
}

func (a artifactVulnerabilityBelongsToArtifactTx) Delete(values ...*models.Artifact) (err error) {
	targetValues := make([]interface{}, len(values))
	for i, v := range values {
		targetValues[i] = v
	}
	return a.tx.Delete(targetValues...)
}

func (a artifactVulnerabilityBelongsToArtifactTx) Clear() error {
	return a.tx.Clear()
}

func (a artifactVulnerabilityBelongsToArtifactTx) Count() int64 {
	return a.tx.Count()
}

type artifactVulnerabilityDo struct{ gen.DO }

func (a artifactVulnerabilityDo) Debug() *artifactVulnerabilityDo {
	return a.withDO(a.DO.Debug())
}

func (a artifactVulnerabilityDo) WithContext(ctx context.Context) *artifactVulnerabilityDo {
	return a.withDO(a.DO.WithContext(ctx))
}

func (a artifactVulnerabilityDo) ReadDB() *artifactVulnerabilityDo {
	return a.Clauses(dbresolver.Read)
}

func (a artifactVulnerabilityDo) WriteDB() *artifactVulnerabilityDo {
	return a.Clauses(dbresolver.Write)
}

func (a artifactVulnerabilityDo) Session(config *gorm.Session) *artifactVulnerabilityDo {
	return a.withDO(a.DO.Session(config))
}

func (a artifactVulnerabilityDo) Clauses(conds ...clause.Expression) *artifactVulnerabilityDo {
	return a.withDO(a.DO.Clauses(conds...))
}

func (a artifactVulnerabilityDo) Returning(value interface{}, columns ...string) *artifactVulnerabilityDo {
	return a.withDO(a.DO.Returning(value, columns...))
}

func (a artifactVulnerabilityDo) Not(conds ...gen.Condition) *artifactVulnerabilityDo {
	return a.withDO(a.DO.Not(conds...))
}

func (a artifactVulnerabilityDo) Or(conds ...gen.Condition) *artifactVulnerabilityDo {
	return a.withDO(a.DO.Or(conds...))
}

func (a artifactVulnerabilityDo) Select(conds ...field.Expr) *artifactVulnerabilityDo {
	return a.withDO(a.DO.Select(conds...))
}

func (a artifactVulnerabilityDo) Where(conds ...gen.Condition) *artifactVulnerabilityDo {
	return a.withDO(a.DO.Where(conds...))
}

func (a artifactVulnerabilityDo) Exists(subquery interface{ UnderlyingDB() *gorm.DB }) *artifactVulnerabilityDo {
	return a.Where(field.CompareSubQuery(field.ExistsOp, nil, subquery.UnderlyingDB()))
}

func (a artifactVulnerabilityDo) Order(conds ...field.Expr) *artifactVulnerabilityDo {
	return a.withDO(a.DO.Order(conds...))
}

func (a artifactVulnerabilityDo) Distinct(cols ...field.Expr) *artifactVulnerabilityDo {
	return a.withDO(a.DO.Distinct(cols...))
}

func (a artifactVulnerabilityDo) Omit(cols ...field.Expr) *artifactVulnerabilityDo {
	return a.withDO(a.DO.Omit(cols...))
}

func (a artifactVulnerabilityDo) Join(table schema.Tabler, on ...field.Expr) *artifactVulnerabilityDo {
	return a.withDO(a.DO.Join(table, on...))
}

func (a artifactVulnerabilityDo) LeftJoin(table schema.Tabler, on ...field.Expr) *artifactVulnerabilityDo {
	return a.withDO(a.DO.LeftJoin(table, on...))
}

func (a artifactVulnerabilityDo) RightJoin(table schema.Tabler, on ...field.Expr) *artifactVulnerabilityDo {
	return a.withDO(a.DO.RightJoin(table, on...))
}

func (a artifactVulnerabilityDo) Group(cols ...field.Expr) *artifactVulnerabilityDo {
	return a.withDO(a.DO.Group(cols...))
}

func (a artifactVulnerabilityDo) Having(conds ...gen.Condition) *artifactVulnerabilityDo {
	return a.withDO(a.DO.Having(conds...))
}

func (a artifactVulnerabilityDo) Limit(limit int) *artifactVulnerabilityDo {
	return a.withDO(a.DO.Limit(limit))
}

func (a artifactVulnerabilityDo) Offset(offset int) *artifactVulnerabilityDo {
	return a.withDO(a.DO.Offset(offset))
}

func (a artifactVulnerabilityDo) Scopes(funcs ...func(gen.Dao) gen.Dao) *artifactVulnerabilityDo {
	return a.withDO(a.DO.Scopes(funcs...))
}

func (a artifactVulnerabilityDo) Unscoped() *artifactVulnerabilityDo {
	return a.withDO(a.DO.Unscoped())
}

func (a artifactVulnerabilityDo) Create(values ...*models.ArtifactVulnerability) error {
	if len(values) == 0 {
		return nil
	}
	return a.DO.Create(values)
}

func (a artifactVulnerabilityDo) CreateInBatches(values []*models.ArtifactVulnerability, batchSize int) error {
	return a.DO.CreateInBatches(values, batchSize)
}

// Save : !!! underlying implementation is different with GORM
// The method is equivalent to executing the statement: db.Clauses(clause.OnConflict{UpdateAll: true}).Create(values)
func (a artifactVulnerabilityDo) Save(values ...*models.ArtifactVulnerability) error {
	if len(values) == 0 {
		return nil
	}
	return a.DO.Save(values)
}

func (a artifactVulnerabilityDo) First() (*models.ArtifactVulnerability, error) {
	if result, err := a.DO.First(); err != nil {
		return nil, err
	} else {
		return result.(*models.ArtifactVulnerability), nil
	}
}

func (a artifactVulnerabilityDo) Take() (*models.ArtifactVulnerability, error) {
	if result, err := a.DO.Take(); err != nil {
		return nil, err
	} else {
		return result.(*models.ArtifactVulnerability), nil
	}
}

func (a artifactVulnerabilityDo) Last() (*models.ArtifactVulnerability, error) {
	if result, err := a.DO.Last(); err != nil {
		return nil, err
	} else {
		return result.(*models.ArtifactVulnerability), nil
	}
}

func (a artifactVulnerabilityDo) Find() ([]*models.ArtifactVulnerability, error) {
	result, err := a.DO.Find()
	return result.([]*models.ArtifactVulnerability), err
}

func (a artifactVulnerabilityDo) FindInBatch(batchSize int, fc func(tx gen.Dao, batch int) error) (results []*models.ArtifactVulnerability, err error) {
	buf := make([]*models.ArtifactVulnerability, 0, batchSize)
	err = a.DO.FindInBatches(&buf, batchSize, func(tx gen.Dao, batch int) error {
		defer func() { results = append(results, buf...) }()
		return fc(tx, batch)
	})
	return results, err
}

func (a artifactVulnerabilityDo) FindInBatches(result *[]*models.ArtifactVulnerability, batchSize int, fc func(tx gen.Dao, batch int) error) error {
	return a.DO.FindInBatches(result, batchSize, fc)
}

func (a artifactVulnerabilityDo) Attrs(attrs ...field.AssignExpr) *artifactVulnerabilityDo {
	return a.withDO(a.DO.Attrs(attrs...))
}

func (a artifactVulnerabilityDo) Assign(attrs ...field.AssignExpr) *artifactVulnerabilityDo {
	return a.withDO(a.DO.Assign(attrs...))
}

func (a artifactVulnerabilityDo) Joins(fields ...field.RelationField) *artifactVulnerabilityDo {
	for _, _f := range fields {
		a = *a.withDO(a.DO.Joins(_f))
	}
	return &a
}

func (a artifactVulnerabilityDo) Preload(fields ...field.RelationField) *artifactVulnerabilityDo {
	for _, _f := range fields {
		a = *a.withDO(a.DO.Preload(_f))
	}
	return &a
}

func (a artifactVulnerabilityDo) FirstOrInit() (*models.ArtifactVulnerability, error) {
	if result, err := a.DO.FirstOrInit(); err != nil {
		return nil, err
	} else {
		return result.(*models.ArtifactVulnerability), nil
	}
}

func (a artifactVulnerabilityDo) FirstOrCreate() (*models.ArtifactVulnerability, error) {
	if result, err := a.DO.FirstOrCreate(); err != nil {
		return nil, err
	} else {
		return result.(*models.ArtifactVulnerability), nil
	}
}

func (a artifactVulnerabilityDo) FindByPage(offset int, limit int) (result []*models.ArtifactVulnerability, count int64, err error) {
	result, err = a.Offset(offset).Limit(limit).Find()
	if err != nil {
		return
	}

	if size := len(result); 0 < limit && 0 < size && size < limit {
		count = int64(size + offset)
		return
	}

	count, err = a.Offset(-1).Limit(-1).Count()
	return
}

func (a artifactVulnerabilityDo) ScanByPage(result interface{}, offset int, limit int) (count int64, err error) {
	count, err = a.Count()
	if err != nil {
		return
	}

	err = a.Offset(offset).Limit(limit).Scan(result)
	return
}

func (a artifactVulnerabilityDo) Scan(result interface{}) (err error) {
	return a.DO.Scan(result)
}

func (a artifactVulnerabilityDo) Delete(models ...*models.ArtifactVulnerability) (result gen.ResultInfo, err error) {
	return a.DO.Delete(models)
}

func (a *artifactVulnerabilityDo) withDO(do gen.Dao) *artifactVulnerabilityDo {
	a.DO = *do.(*gen.DO)
	return a
}
