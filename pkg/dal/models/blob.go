package models

import (
	"database/sql"
	"time"

	"gorm.io/gorm"
)

// Blob represents a blob
type Blob struct {
	gorm.Model
	ID          uint `gorm:"primaryKey"`
	Digest      string
	Size        int64
	ContentType string

	LastPull  sql.NullTime
	PushedAt  time.Time `gorm:"not null"`
	PullTimes uint      `gorm:"default:0"`

	Artifacts []*Artifact `gorm:"many2many:artifact_blobs;"`
}
