package models

import (
	"time"

	"gorm.io/plugin/soft_delete"
)

// Cache cache
type Cache struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:milli"`
	ID        int64                 `gorm:"primaryKey"`

	Key string `gorm:"uniqueIndex,size:256"`
	Val []byte
}

// CacheQuery ...
type CacheQuery interface {
	// DELETE FROM @@table WHERE id in (
	// SELECT id from @@table ORDER BY created_at ASC LIMIT (
	// (SELECT COUNT(id) FROM @@table) - (@size * (1 - @threshold))))
	DeleteOutsideThreshold(size int64, threshold float64) error
}
