// The MIT License (MIT)
//
// Copyright © 2023 Tosone <i@tosone.cn>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package models

import (
	"database/sql"
	"time"

	"gorm.io/gen"
	"gorm.io/plugin/soft_delete"
)

// Tag represents a tag
type Tag struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:milli"`
	ID        uint                  `gorm:"primaryKey"`

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
