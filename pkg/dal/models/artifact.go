package models

import (
	"database/sql"
	"time"

	"gorm.io/gorm"
)

// Artifact represents an artifact
type Artifact struct {
	gorm.Model
	ID               uint `gorm:"primaryKey"`
	RepositoryID     uint
	Digest           string
	Size             int64
	ContentType      string
	Raw              string
	HistoryCreatedBy *string

	LastPull  sql.NullTime
	PushedAt  time.Time `gorm:"not null"`
	PullTimes uint      `gorm:"default:0"`

	Repository Repository
	Blobs      []*Blob `gorm:"many2many:artifact_blobs;"`
	Tags       []*Tag  `gorm:"foreignKey:ArtifactID;references:ID;"`
}
