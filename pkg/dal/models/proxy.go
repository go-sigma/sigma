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
)

// ProxyTaskArtifact represents an artifact proxy task
type ProxyTaskArtifact struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:milli"`
	ID        int64                 `gorm:"primaryKey"`

	Repository  string
	Digest      string
	Size        int64
	ContentType string
	Raw         []byte

	Blobs []ProxyTaskArtifactBlob `gorm:"foreignKey:ProxyTaskArtifactID;references:ID"`
}

// ProxyTaskArtifactBlob represents an proxy task artifact blobs
type ProxyTaskArtifactBlob struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:milli"`
	ID        int64                 `gorm:"primaryKey"`

	ProxyTaskArtifactID int64
	Blob                string
}

// ProxyTaskTag represents a proxy task tag
type ProxyTaskTag struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:milli"`
	ID        int64                 `gorm:"primaryKey"`

	Repository  string
	Reference   string
	Size        int64
	ContentType string
	Raw         []byte

	Manifests []ProxyTaskTagManifest `gorm:"foreignKey:ProxyTaskTagID;references:ID"`
}

// ProxyTaskTagManifest represents proxy task tag manifests
type ProxyTaskTagManifest struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:milli"`
	ID        int64                 `gorm:"primaryKey"`

	ProxyTaskTagID int64
	Digest         string
}
