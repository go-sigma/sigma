// Copyright 2023 sigma
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package models

import (
	"gorm.io/plugin/soft_delete"
)

// Cache cache
type Cache struct {
	CreatedAt int64                 `gorm:"autoCreateTime:milli"`
	UpdatedAt int64                 `gorm:"autoUpdateTime:milli"`
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
