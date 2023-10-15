package models

import (
	"time"

	"gorm.io/plugin/soft_delete"
)

// Setting setting
type Setting struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:milli"`
	ID        int64                 `gorm:"primaryKey"`

	Key string `gorm:"uniqueIndex,size:256"`
	Val []byte
}
