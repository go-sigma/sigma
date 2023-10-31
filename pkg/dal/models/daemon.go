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

// DaemonGcTagRule ...
type DaemonGcTagRule struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:milli"`
	ID        int64                 `gorm:"primaryKey"`

	NamespaceID *int64
	Namespace   *Namespace

	IsRunning           bool `gorm:"default:false"`
	CronEnabled         bool `gorm:"default:false"`
	CronRule            *string
	CronNextTrigger     *time.Time
	RetentionRuleType   *enums.RetentionRuleType
	RetentionRuleAmount *int64
	RetentionPattern    []byte
}

// DaemonGcTagRunner ...
type DaemonGcTagRunner struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:milli"`
	ID        int64                 `gorm:"primaryKey"`

	RuleID int64
	Rule   DaemonGcTagRule

	Message []byte
	Status  enums.TaskCommonStatus

	StartedAt *time.Time
	EndedAt   *time.Time
	Duration  *int64
}

// DaemonGcTagRecords ...
type DaemonGcTagRecord struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:milli"`
	ID        int64                 `gorm:"primaryKey"`

	RunnerID int64
	Runner   DaemonGcTagRunner

	Tag string
}

// DaemonGcRepositoryRule ...
type DaemonGcRepositoryRule struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:milli"`
	ID        int64                 `gorm:"primaryKey"`

	NamespaceID *int64
	Namespace   *Namespace

	IsRunning       bool `gorm:"default:false"`
	CronEnabled     bool `gorm:"default:false"`
	CronRule        *string
	CronNextTrigger *time.Time
}

// DaemonGcRepositoryRunner ...
type DaemonGcRepositoryRunner struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:milli"`
	ID        int64                 `gorm:"primaryKey"`

	RuleID int64
	Rule   DaemonGcTagRule

	Status  enums.TaskCommonStatus `gorm:"status"`
	Message []byte

	StartedAt *time.Time
	EndedAt   *time.Time
	Duration  *int64
}

// DaemonGcRepositoryRecord ...
type DaemonGcRepositoryRecord struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:milli"`
	ID        int64                 `gorm:"primaryKey"`

	RunnerID int64
	Runner   DaemonGcRepositoryRunner

	Repository string
}

// DaemonGcArtifactRule ...
type DaemonGcArtifactRule struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:milli"`
	ID        int64                 `gorm:"primaryKey"`

	NamespaceID *int64
	Namespace   *Namespace

	IsRunning       bool `gorm:"default:false"`
	CronEnabled     bool `gorm:"default:false"`
	CronRule        *string
	CronNextTrigger *time.Time
}

type DaemonGcArtifactRunner struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:milli"`
	ID        int64                 `gorm:"primaryKey"`

	RuleID int64
	Rule   DaemonGcArtifactRule

	Status  enums.TaskCommonStatus `gorm:"status"`
	Message []byte

	StartedAt *time.Time
	EndedAt   *time.Time
	Duration  *int64
}

// DaemonGcArtifactRecord ...
type DaemonGcArtifactRecord struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:milli"`
	ID        int64                 `gorm:"primaryKey"`

	RunnerID int64
	Digest   string

	Runner DaemonGcArtifactRunner
}

// DaemonGcBlobRule ...
type DaemonGcBlobRule struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:milli"`
	ID        int64                 `gorm:"primaryKey"`

	IsRunning       bool `gorm:"default:false"`
	CronEnabled     bool `gorm:"default:false"`
	CronRule        *string
	CronNextTrigger *time.Time
}

// DaemonGcBlobRunner ...
type DaemonGcBlobRunner struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:milli"`
	ID        int64                 `gorm:"primaryKey"`

	RuleID int64
	Rule   DaemonGcBlobRule

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
	Runner   DaemonGcBlobRunner

	Digest string
}
