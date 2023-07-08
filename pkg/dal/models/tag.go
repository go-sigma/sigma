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
	"database/sql"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

// Tag represents a tag
type Tag struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:milli"`
	ID        int64                 `gorm:"primaryKey"`

	RepositoryID int64
	ArtifactID   int64
	Name         string

	LastPull  sql.NullTime
	PushedAt  time.Time `gorm:"autoCreateTime"`
	PullTimes int64     `gorm:"default:0"`

	Repository *Repository
	Artifact   *Artifact
}

// AfterCreate ...
// if something err occurs, the create will be aborted
func (a *Tag) BeforeCreate(tx *gorm.DB) error {
	if a == nil {
		return nil
	}
	var repositoryObj Repository
	err := tx.Model(&Repository{}).Where(&Repository{ID: a.RepositoryID}).First(&repositoryObj).Error
	if err != nil {
		return err
	}
	var namespaceObj Namespace
	err = tx.Model(&Namespace{}).Where(&Namespace{ID: repositoryObj.NamespaceID}).First(&namespaceObj).Error
	if err != nil {
		return err
	}

	if namespaceObj.TagLimit > 0 && namespaceObj.TagCount+1 > namespaceObj.TagLimit {
		return errors.New("namespace's tag quota exceeded")
	}
	if repositoryObj.TagLimit > 0 && repositoryObj.TagCount+1 > repositoryObj.TagLimit {
		return errors.New("repository's tag quota exceeded")
	}

	err = tx.Model(&Namespace{}).Where(&Namespace{ID: repositoryObj.NamespaceID}).UpdateColumns(
		map[string]any{
			"tag_count": namespaceObj.TagCount + 1,
		}).Error
	if err != nil {
		return err
	}
	err = tx.Model(&Repository{}).Where(&Repository{ID: repositoryObj.ID}).UpdateColumns(map[string]any{
		"tag_count": repositoryObj.TagCount + 1,
	}).Error
	if err != nil {
		return err
	}
	return nil
}

// BeforeUpdate ...
func (a *Tag) BeforeUpdate(tx *gorm.DB) error {
	if a == nil {
		return nil
	}
	var repositoryObj Repository
	err := tx.Model(&Repository{}).Where(&Repository{ID: a.RepositoryID}).First(&repositoryObj).Error
	if err != nil {
		return err
	}

	err = tx.Exec(`UPDATE
  namespaces
SET
  tag_count = (
    SELECT
      COUNT(tags.id)
    FROM
      repositories
      INNER JOIN tags ON repositories.id = tags.repository_id
    WHERE
      repositories.namespace_id = ?)
WHERE
  id = ?`, repositoryObj.NamespaceID, repositoryObj.NamespaceID).Error
	if err != nil {
		return err
	}

	tx.Exec(`UPDATE
  repositories
SET
  tag_count = (
    SELECT
      count(tags.name)
    FROM
      tags
    WHERE
      tags.repository_id = ?)
WHERE
  id = ?`, repositoryObj.ID, repositoryObj.ID)
	if err != nil {
		return err
	}

	return nil
}
