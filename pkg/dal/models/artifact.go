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
	"time"

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
	Size         int64
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
