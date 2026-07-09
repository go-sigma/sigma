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
	"gorm.io/plugin/soft_delete"

	"github.com/go-sigma/sigma/pkg/api/enums"
)

// Artifact represents an artifact
type Artifact struct {
	CreatedAt int64                 `gorm:"autoCreateTime:milli"`
	UpdatedAt int64                 `gorm:"autoUpdateTime:milli"`
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:milli"`
	ID        string                `gorm:"primaryKey;size:36"`

	NamespaceID string
	Namespace   Namespace

	RepositoryID string
	Repository   Repository

	Digest          string
	Size            int64 `gorm:"default:0"`
	BlobsSize       int64 `gorm:"default:0"`
	ContentType     string
	ConfigRaw       []byte
	ConfigMediaType *string
	Type            enums.ArtifactType `gorm:"default:Unknown"`

	LastPull  int64
	PushedAt  int64 `gorm:"autoCreateTime:milli"`
	PullTimes int64 `gorm:"default:0"`

	Vulnerability ArtifactVulnerability `gorm:"foreignKey:ArtifactID;"`
	Sbom          ArtifactSbom          `gorm:"foreignKey:ArtifactID;"`

	ReferrerID *string
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
	ArtifactSizeByNamespace(namespaceID string) (gen.T, error)
	// SELECT sum(blobs_size) as size FROM @@table WHERE repository_id = @repositoryID
	ArtifactSizeByRepository(repositoryID string) (gen.T, error)
}

// ArtifactAssociated ...
type ArtifactAssociated interface {
	// SELECT COUNT(artifact_id) as count FROM artifact_artifacts LEFT JOIN artifacts ON artifacts.id = artifact_artifacts.artifact_id WHERE artifacts.deleted_at = 0 AND artifact_sub_id=@artifactID
	ArtifactAssociated(artifactID string) (gen.M, error)
}

// ArtifactSbom represents an artifact sbom
type ArtifactSbom struct {
	CreatedAt int64                 `gorm:"autoUpdateTime:milli"`
	UpdatedAt int64                 `gorm:"autoUpdateTime:milli"`
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:milli"`
	ID        string                `gorm:"primaryKey;size:36"`

	ArtifactID string
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
	ID        string                `gorm:"primaryKey;size:36"`

	ArtifactID string
	Version    int64 // is the vuln db built time
	Raw        []byte
	Result     []byte
	Status     enums.TaskCommonStatus
	Stdout     []byte
	Stderr     []byte
	Message    string

	Artifact *Artifact
}
