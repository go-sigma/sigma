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

	"github.com/go-sigma/sigma/pkg/api/enums"
)

// Namespace represents a namespace
type Namespace struct {
	CreatedAt int64                 `gorm:"autoCreateTime:milli"`
	UpdatedAt int64                 `gorm:"autoUpdateTime:milli"`
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:milli"`
	ID        string                `gorm:"primaryKey;size:36"`

	Name            string `gorm:"uniqueIndex"`
	Description     *string
	Overview        []byte
	Visibility      enums.Visibility `gorm:"default:public"`
	TagLimit        int64            `gorm:"default:0"`
	TagCount        int64            `gorm:"default:0"`
	RepositoryLimit int64            `gorm:"default:0"`
	RepositoryCount int64            `gorm:"default:0"`
	SizeLimit       int64            `gorm:"default:0"`
	Size            int64            `gorm:"default:0"`
	SizeDirty       bool             `gorm:"default:false;index"`
}
