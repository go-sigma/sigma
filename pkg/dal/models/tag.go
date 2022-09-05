package models

import (
	"database/sql"
	"time"

	"gorm.io/gen"
	"gorm.io/gorm"
)

// Tag represents a tag
type Tag struct {
	gorm.Model
	ID           uint `gorm:"primaryKey"`
	RepositoryID uint
	ArtifactID   uint
	Name         string
	Digest       string
	Size         int64

	LastPull  sql.NullTime
	PushedAt  time.Time `gorm:"not null"`
	PullTimes uint      `gorm:"default:0"`

	Repository Repository
	Artifact   Artifact
}

type TagQuerier interface {
	// DeleteByName query data by name and age and return it as map
	//
	// UPDATE `tags` LEFT JOIN `repositories` ON `tags`.`repository_id` = `repositories`.`id`
	// SET `tags`.`deleted_at` = NOW() WHERE
	// `repositories`.`name` = @repository AND `tags`.`name` = @tag AND `tags`.`deleted_at` IS NULL
	DeleteByName(repository, tag string) (gen.RowsAffected, error)
}
