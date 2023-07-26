// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.

package query

import (
	"context"
	"database/sql"

	"gorm.io/gorm"

	"gorm.io/gen"

	"gorm.io/plugin/dbresolver"
)

var (
	Q                     = new(Query)
	Artifact              *artifact
	ArtifactSbom          *artifactSbom
	ArtifactVulnerability *artifactVulnerability
	Audit                 *audit
	Blob                  *blob
	BlobUpload            *blobUpload
	CasbinRule            *casbinRule
	DaemonLog             *daemonLog
	Namespace             *namespace
	Repository            *repository
	Tag                   *tag
	User                  *user
	UserRecoverCode       *userRecoverCode
)

func SetDefault(db *gorm.DB, opts ...gen.DOOption) {
	*Q = *Use(db, opts...)
	Artifact = &Q.Artifact
	ArtifactSbom = &Q.ArtifactSbom
	ArtifactVulnerability = &Q.ArtifactVulnerability
	Audit = &Q.Audit
	Blob = &Q.Blob
	BlobUpload = &Q.BlobUpload
	CasbinRule = &Q.CasbinRule
	DaemonLog = &Q.DaemonLog
	Namespace = &Q.Namespace
	Repository = &Q.Repository
	Tag = &Q.Tag
	User = &Q.User
	UserRecoverCode = &Q.UserRecoverCode
}

func Use(db *gorm.DB, opts ...gen.DOOption) *Query {
	return &Query{
		db:                    db,
		Artifact:              newArtifact(db, opts...),
		ArtifactSbom:          newArtifactSbom(db, opts...),
		ArtifactVulnerability: newArtifactVulnerability(db, opts...),
		Audit:                 newAudit(db, opts...),
		Blob:                  newBlob(db, opts...),
		BlobUpload:            newBlobUpload(db, opts...),
		CasbinRule:            newCasbinRule(db, opts...),
		DaemonLog:             newDaemonLog(db, opts...),
		Namespace:             newNamespace(db, opts...),
		Repository:            newRepository(db, opts...),
		Tag:                   newTag(db, opts...),
		User:                  newUser(db, opts...),
		UserRecoverCode:       newUserRecoverCode(db, opts...),
	}
}

type Query struct {
	db *gorm.DB

	Artifact              artifact
	ArtifactSbom          artifactSbom
	ArtifactVulnerability artifactVulnerability
	Audit                 audit
	Blob                  blob
	BlobUpload            blobUpload
	CasbinRule            casbinRule
	DaemonLog             daemonLog
	Namespace             namespace
	Repository            repository
	Tag                   tag
	User                  user
	UserRecoverCode       userRecoverCode
}

func (q *Query) Available() bool { return q.db != nil }

func (q *Query) clone(db *gorm.DB) *Query {
	return &Query{
		db:                    db,
		Artifact:              q.Artifact.clone(db),
		ArtifactSbom:          q.ArtifactSbom.clone(db),
		ArtifactVulnerability: q.ArtifactVulnerability.clone(db),
		Audit:                 q.Audit.clone(db),
		Blob:                  q.Blob.clone(db),
		BlobUpload:            q.BlobUpload.clone(db),
		CasbinRule:            q.CasbinRule.clone(db),
		DaemonLog:             q.DaemonLog.clone(db),
		Namespace:             q.Namespace.clone(db),
		Repository:            q.Repository.clone(db),
		Tag:                   q.Tag.clone(db),
		User:                  q.User.clone(db),
		UserRecoverCode:       q.UserRecoverCode.clone(db),
	}
}

func (q *Query) ReadDB() *Query {
	return q.ReplaceDB(q.db.Clauses(dbresolver.Read))
}

func (q *Query) WriteDB() *Query {
	return q.ReplaceDB(q.db.Clauses(dbresolver.Write))
}

func (q *Query) ReplaceDB(db *gorm.DB) *Query {
	return &Query{
		db:                    db,
		Artifact:              q.Artifact.replaceDB(db),
		ArtifactSbom:          q.ArtifactSbom.replaceDB(db),
		ArtifactVulnerability: q.ArtifactVulnerability.replaceDB(db),
		Audit:                 q.Audit.replaceDB(db),
		Blob:                  q.Blob.replaceDB(db),
		BlobUpload:            q.BlobUpload.replaceDB(db),
		CasbinRule:            q.CasbinRule.replaceDB(db),
		DaemonLog:             q.DaemonLog.replaceDB(db),
		Namespace:             q.Namespace.replaceDB(db),
		Repository:            q.Repository.replaceDB(db),
		Tag:                   q.Tag.replaceDB(db),
		User:                  q.User.replaceDB(db),
		UserRecoverCode:       q.UserRecoverCode.replaceDB(db),
	}
}

type queryCtx struct {
	Artifact              *artifactDo
	ArtifactSbom          *artifactSbomDo
	ArtifactVulnerability *artifactVulnerabilityDo
	Audit                 *auditDo
	Blob                  *blobDo
	BlobUpload            *blobUploadDo
	CasbinRule            *casbinRuleDo
	DaemonLog             *daemonLogDo
	Namespace             *namespaceDo
	Repository            *repositoryDo
	Tag                   *tagDo
	User                  *userDo
	UserRecoverCode       *userRecoverCodeDo
}

func (q *Query) WithContext(ctx context.Context) *queryCtx {
	return &queryCtx{
		Artifact:              q.Artifact.WithContext(ctx),
		ArtifactSbom:          q.ArtifactSbom.WithContext(ctx),
		ArtifactVulnerability: q.ArtifactVulnerability.WithContext(ctx),
		Audit:                 q.Audit.WithContext(ctx),
		Blob:                  q.Blob.WithContext(ctx),
		BlobUpload:            q.BlobUpload.WithContext(ctx),
		CasbinRule:            q.CasbinRule.WithContext(ctx),
		DaemonLog:             q.DaemonLog.WithContext(ctx),
		Namespace:             q.Namespace.WithContext(ctx),
		Repository:            q.Repository.WithContext(ctx),
		Tag:                   q.Tag.WithContext(ctx),
		User:                  q.User.WithContext(ctx),
		UserRecoverCode:       q.UserRecoverCode.WithContext(ctx),
	}
}

func (q *Query) Transaction(fc func(tx *Query) error, opts ...*sql.TxOptions) error {
	return q.db.Transaction(func(tx *gorm.DB) error { return fc(q.clone(tx)) }, opts...)
}

func (q *Query) Begin(opts ...*sql.TxOptions) *QueryTx {
	tx := q.db.Begin(opts...)
	return &QueryTx{Query: q.clone(tx), Error: tx.Error}
}

type QueryTx struct {
	*Query
	Error error
}

func (q *QueryTx) Commit() error {
	return q.db.Commit().Error
}

func (q *QueryTx) Rollback() error {
	return q.db.Rollback().Error
}

func (q *QueryTx) SavePoint(name string) error {
	return q.db.SavePoint(name).Error
}

func (q *QueryTx) RollbackTo(name string) error {
	return q.db.RollbackTo(name).Error
}
