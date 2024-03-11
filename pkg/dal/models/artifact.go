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
	"gorm.io/gen"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/xerrors"
)

// Artifact represents an artifact
type Artifact struct {
	CreatedAt int64                 `gorm:"autoCreateTime:milli"`
	UpdatedAt int64                 `gorm:"autoUpdateTime:milli"`
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:milli"`
	ID        int64                 `gorm:"primaryKey"`

	RepositoryID int64
	Repository   Repository

	Digest          string
	Size            int64 `gorm:"default:0"`
	BlobsSize       int64 `gorm:"default:0"`
	ContentType     string
	Raw             []byte
	ConfigRaw       []byte
	ConfigMediaType *string
	Type            enums.ArtifactType `gorm:"default:Unknown"`

	LastPull  int64
	PushedAt  int64 `gorm:"autoCreateTime:milli"`
	PullTimes int64 `gorm:"default:0"`

	Vulnerability ArtifactVulnerability `gorm:"foreignKey:ArtifactID;"`
	Sbom          ArtifactSbom          `gorm:"foreignKey:ArtifactID;"`

	ReferrerID *int64
	Referrer   *Artifact

	// ArtifactSubs In the artifact_artifacts table, artifact_id refers to the upper-level artifact index,
	// and artifact_sub_id refers to the lower-level artifact.
	ArtifactSubs []*Artifact `gorm:"many2many:artifact_artifacts;"`
	Blobs        []*Blob     `gorm:"many2many:artifact_blobs;"`
	Tags         []*Tag      `gorm:"foreignKey:ArtifactID;"`
}

// ArtifactSizeByNamespaceOrRepository ...
type ArtifactSizeByNamespaceOrRepository interface {
	// SELECT sum(blobs_size) as size FROM @@table WHERE repository_id in (
	// SELECT id from repositories where namespace_id = @namespaceID)
	ArtifactSizeByNamespace(namespaceID int64) (gen.T, error)
	// SELECT sum(blobs_size) as size FROM @@table WHERE repository_id = @repositoryID
	ArtifactSizeByRepository(repositoryID int64) (gen.T, error)
}

// ArtifactAssociated ...
type ArtifactAssociated interface {
	// SELECT COUNT(artifact_id) as count FROM artifact_artifacts LEFT JOIN artifacts ON artifacts.id = artifact_artifacts.artifact_id WHERE artifacts.deleted_at = 0 AND artifact_sub_id=@artifactID
	ArtifactAssociated(artifactID int64) (gen.M, error)
}

// AfterCreate ...
func (a *Artifact) BeforeCreate(tx *gorm.DB) error {
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
	if namespaceObj.SizeLimit > 0 && namespaceObj.Size+a.BlobsSize > namespaceObj.SizeLimit {
		return xerrors.GenDSErrCodeResourceSizeQuotaExceedNamespace(namespaceObj.Name, namespaceObj.Size, namespaceObj.SizeLimit, a.BlobsSize)
	}
	if repositoryObj.SizeLimit > 0 && repositoryObj.Size+a.BlobsSize > repositoryObj.SizeLimit {
		return xerrors.GenDSErrCodeResourceSizeQuotaExceedRepository(repositoryObj.Name, repositoryObj.Size, repositoryObj.SizeLimit, a.BlobsSize)
	}

	// we should check all the checker here, and update the size and tag count
	err = tx.Model(&Namespace{}).Where(&Namespace{ID: repositoryObj.NamespaceID}).UpdateColumns(
		map[string]any{
			"size": namespaceObj.Size + a.BlobsSize,
		}).Error
	if err != nil {
		return err
	}
	err = tx.Model(&Repository{}).Where(&Repository{ID: repositoryObj.ID}).UpdateColumns(map[string]any{
		"size": repositoryObj.Size + a.BlobsSize,
	}).Error
	if err != nil {
		return err
	}
	return nil
}

// BeforeUpdate ...
func (a *Artifact) BeforeUpdate(tx *gorm.DB) error {
	if a == nil {
		return nil
	}
	var repositoryObj Repository
	err := tx.Model(&Repository{}).Where("id = ?", a.RepositoryID).First(&repositoryObj).Error
	if err != nil {
		return err
	}

	err = tx.Exec(`UPDATE
  namespaces
SET
  size = (
    SELECT
      SUM(artifacts.blobs_size)
    FROM
      repositories
      INNER JOIN artifacts ON repositories.id = artifacts.repository_id
    WHERE
      repositories.namespace_id = ?)
WHERE
  id = ?`, repositoryObj.NamespaceID, repositoryObj.NamespaceID).Error
	if err != nil {
		return err
	}
	err = tx.Exec(`UPDATE
  repositories
SET
  size = (
    SELECT
      SUM(size)
    FROM
      artifacts
    WHERE
		  artifacts.repository_id = ?)
WHERE
  id = ?`, repositoryObj.ID, repositoryObj.ID).Error
	if err != nil {
		return err
	}
	return nil
}

// ArtifactSbom represents an artifact sbom
type ArtifactSbom struct {
	CreatedAt int64                 `gorm:"autoUpdateTime:milli"`
	UpdatedAt int64                 `gorm:"autoUpdateTime:milli"`
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:milli"`
	ID        int64                 `gorm:"primaryKey"`

	ArtifactID int64
	Raw        []byte
	Result     []byte
	Status     enums.TaskCommonStatus
	Stdout     []byte
	Stderr     []byte
	Message    string

	Artifact *Artifact
}

// ArtifactVulnerability represents an artifact vulnerability
type ArtifactVulnerability struct {
	CreatedAt int64                 `gorm:"autoCreateTime:milli"`
	UpdatedAt int64                 `gorm:"autoUpdateTime:milli"`
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:milli"`
	ID        int64                 `gorm:"primaryKey"`

	ArtifactID int64
	Metadata   []byte // is the trivy db metadata
	Raw        []byte
	Result     []byte
	Status     enums.TaskCommonStatus
	Stdout     []byte
	Stderr     []byte
	Message    string

	Artifact *Artifact
}
