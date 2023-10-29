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
	"time"

	"gorm.io/plugin/soft_delete"

	"github.com/go-sigma/sigma/pkg/types/enums"
)

// DaemonLog represents an artifact
type DaemonLog struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:milli"`
	ID        int64                 `gorm:"primaryKey"`

	NamespaceID *int64
	Type        enums.Daemon
	Action      enums.AuditAction
	Resource    string
	Status      enums.TaskCommonStatus
	Message     []byte
}

// DaemonGcRepositoryRunner ...
type DaemonGcRepositoryRunner struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:milli"`
	ID        int64                 `gorm:"primaryKey"`

	Status      enums.TaskCommonStatus `gorm:"status"`
	Message     []byte
	NamespaceID *int64

	StartedAt *time.Time
	EndedAt   *time.Time
	Duration  *int64

	Namespace *Namespace
}

// DaemonGcRepositoryRecord ...
type DaemonGcRepositoryRecord struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:milli"`
	ID        int64                 `gorm:"primaryKey"`

	RunnerID   int64
	Repository string

	Runner DaemonGcRepositoryRunner
}

type DaemonGcArtifactRunner struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:milli"`
	ID        int64                 `gorm:"primaryKey"`

	Status      enums.TaskCommonStatus `gorm:"status"`
	Message     []byte
	NamespaceID *int64

	StartedAt *time.Time
	EndedAt   *time.Time
	Duration  *int64

	Namespace *Namespace
}

type DaemonGcArtifactRecord struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:milli"`
	ID        int64                 `gorm:"primaryKey"`

	RunnerID int64
	Digest   string

	Runner DaemonGcArtifactRunner
}

type DaemonGcBlobRunner struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:milli"`
	ID        int64                 `gorm:"primaryKey"`

	Status  enums.TaskCommonStatus `gorm:"status"`
	Message []byte

	StartedAt *time.Time
	EndedAt   *time.Time
	Duration  *int64
}

type DaemonGcBlobRecord struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:milli"`
	ID        int64                 `gorm:"primaryKey"`

	RunnerID int64
	Digest   string

	Runner DaemonGcBlobRunner
}
