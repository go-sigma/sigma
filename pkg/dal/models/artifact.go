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

	"github.com/ximager/ximager/pkg/types/enums"
)

// Artifact represents an artifact
type Artifact struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:milli"`
	ID        int64                 `gorm:"primaryKey"`

	RepositoryID int64
	Digest       string
	Size         int64 `gorm:"default:0"`
	BlobsSize    int64 `gorm:"default:0"`
	ContentType  string
	Raw          []byte

	LastPull  sql.NullTime
	PushedAt  time.Time `gorm:"autoCreateTime"`
	PullTimes int64     `gorm:"default:0"`

	Repository Repository

	ArtifactIndexes []*Artifact `gorm:"many2many:artifact_artifacts;"`
	Blobs           []*Blob     `gorm:"many2many:artifact_blobs;"`
	Tags            []*Tag      `gorm:"foreignKey:ArtifactID;"`
}

// AfterCreate ...
// if something err occurs, the create will be aborted
func (a *Artifact) BeforeCreate(tx *gorm.DB) error {
	if a == nil {
		return nil
	}
	var repositoryObj Repository
	err := tx.Model(&Repository{}).Where(&Repository{ID: a.RepositoryID}).First(&repositoryObj).Error
	if err != nil {
		return err
	}
	var namespaceQuotaObj NamespaceQuota
	err = tx.Model(&NamespaceQuota{}).Where(&NamespaceQuota{NamespaceID: repositoryObj.NamespaceID}).First(&namespaceQuotaObj).Error
	if err != nil {
		return err
	}
	if namespaceQuotaObj.Limit > 0 && namespaceQuotaObj.Usage+a.BlobsSize > namespaceQuotaObj.Limit {
		return errors.New("namespace quota exceeded")
	}
	err = tx.Model(&NamespaceQuota{}).Where(&NamespaceQuota{NamespaceID: repositoryObj.ID}).UpdateColumn("usage", namespaceQuotaObj.Usage+a.BlobsSize).Error
	if err != nil {
		return err
	}
	var repositoryQuotaObj RepositoryQuota
	err = tx.Model(&RepositoryQuota{}).Where(&RepositoryQuota{RepositoryID: repositoryObj.ID}).First(&repositoryQuotaObj).Error
	if err != nil {
		return err
	}
	if repositoryQuotaObj.Limit > 0 && repositoryQuotaObj.Usage+a.BlobsSize > repositoryQuotaObj.Limit {
		return errors.New("repository quota exceeded")
	}
	err = tx.Model(&RepositoryQuota{}).Where(&RepositoryQuota{RepositoryID: repositoryObj.ID}).UpdateColumn("usage", repositoryQuotaObj.Usage+a.BlobsSize).Error
	if err != nil {
		return err
	}
	return nil
}

// BeforeDelete ...
// if something err occurs, the delete will be aborted
func (a *Artifact) BeforeUpdate(tx *gorm.DB) error {
	if a == nil {
		return nil
	}
	if a.DeletedAt != 0 {
		var repositoryObj Repository
		err := tx.Model(&Repository{}).Where("id = ?", a.RepositoryID).First(&repositoryObj).Error
		if err != nil {
			return err
		}
		err = tx.Model(&NamespaceQuota{}).Where("namespace_id = ?", repositoryObj.NamespaceID).Update("usage", gorm.Expr("usage - ?", a.BlobsSize)).Error
		if err != nil {
			return err
		}
		err = tx.Model(&RepositoryQuota{}).Where("repository_id = ?", a.RepositoryID).Update("usage", gorm.Expr("usage + ?", a.BlobsSize)).Error
		if err != nil {
			return err
		}
	}
	return nil
}

// ArtifactSbom represents an artifact sbom
type ArtifactSbom struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:milli"`
	ID        int64                 `gorm:"primaryKey"`

	ArtifactID int64
	Raw        []byte
	Status     enums.TaskCommonStatus
	Stdout     []byte
	Stderr     []byte
	Message    string

	Artifact *Artifact
}

// ArtifactVulnerability represents an artifact vulnerability
type ArtifactVulnerability struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:milli"`
	ID        int64                 `gorm:"primaryKey"`

	ArtifactID int64
	Metadata   []byte // is the trivy db metadata
	Raw        []byte
	Status     enums.TaskCommonStatus
	Stdout     []byte
	Stderr     []byte
	Message    string

	Artifact *Artifact
}
