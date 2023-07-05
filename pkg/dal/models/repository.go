// Copyright 2023 XImager
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
	"time"

	"gorm.io/plugin/soft_delete"

	"github.com/ximager/ximager/pkg/types/enums"
)

// Repository represents a repository
type Repository struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:milli"`
	ID        int64                 `gorm:"primaryKey"`

	NamespaceID int64
	Name        string
	Description *string
	Overview    []byte
	Visibility  enums.Visibility `gorm:"default:public"`
	Limit       int64            `gorm:"default:0"`
	Usage       int64            `gorm:"default:0"`

	Namespace Namespace
	Tags      []*RepositoryTag
}

// RepositoryTag represents repository tags
type RepositoryTag struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:milli"`
	ID        int64                 `gorm:"primaryKey"`

	Name         string
	RepositoryID int64
}
