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

	"github.com/ximager/ximager/pkg/types"
)

// ArtifactProxyTask represents an artifact proxy task
type ProxyArtifactTask struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:milli"`
	ID        uint64                `gorm:"primaryKey"`

	Status  types.TaskCommonStatus
	Message string

	Blobs []ProxyArtifactBlob `gorm:"foreignKey:ProxyArtifactTaskID;references:ID"`
}

// ProxyArtifactBlob represents an artifact proxy task
type ProxyArtifactBlob struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:milli"`
	ID        uint64                `gorm:"primaryKey"`

	ProxyArtifactTaskID uint64
	Blob                string

	ProxyArtifactTask ProxyArtifactTask `gorm:"foreignKey:ProxyArtifactTaskID;references:ID"`
}

// ProxyTagTask represents an artifact proxy task
type ProxyTagTask struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:milli"`
	ID        uint64                `gorm:"primaryKey"`

	Manifest string
	Status   types.TaskCommonStatus
	Message  string
}
