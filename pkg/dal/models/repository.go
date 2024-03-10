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
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/go-sigma/sigma/pkg/xerrors"
)

// Repository represents a repository
type Repository struct {
	CreatedAt int64                 `gorm:"autoCreateTime:milli"`
	UpdatedAt int64                 `gorm:"autoUpdateTime:milli"`
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:milli"`
	ID        int64                 `gorm:"primaryKey"`

	NamespaceID int64
	Name        string
	Description *string
	Overview    []byte
	TagLimit    int64 `gorm:"default:0"`
	TagCount    int64 `gorm:"default:0"`
	SizeLimit   int64 `gorm:"default:0"`
	Size        int64 `gorm:"default:0"`

	Namespace Namespace
	Builder   *Builder
}

// BeforeCreate ...
func (a *Repository) BeforeCreate(tx *gorm.DB) error {
	if a == nil {
		return nil
	}
	var namespaceObj Namespace
	err := tx.Model(&Namespace{}).Where(&Namespace{ID: a.NamespaceID}).First(&namespaceObj).Error
	if err != nil {
		return err
	}
	if namespaceObj.RepositoryLimit > 0 && namespaceObj.RepositoryCount+1 > namespaceObj.RepositoryLimit {
		return xerrors.GenDSErrCodeResourceCountQuotaExceedNamespaceRepository(namespaceObj.Name, namespaceObj.RepositoryLimit)
	}
	err = tx.Model(&Namespace{}).Where(&Namespace{ID: a.NamespaceID}).UpdateColumns(
		map[string]any{
			"repository_count": namespaceObj.RepositoryCount + 1,
		}).Error
	if err != nil {
		return err
	}
	return nil
}

// AfterUpdate ...
func (a *Repository) AfterUpdate(tx *gorm.DB) error {
	if a == nil {
		return nil
	}
	err := tx.Exec(`UPDATE
  namespaces
SET
  repository_count = (
    SELECT
      count(repositories.id)
    FROM
      repositories
    WHERE
      namespace_id = ?)
WHERE
  id = ?`, a.NamespaceID, a.NamespaceID).Error
	if err != nil {
		return err
	}
	return nil
}
