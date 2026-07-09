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
	"gorm.io/plugin/soft_delete"

	"github.com/go-sigma/sigma/pkg/api/enums"
)

// DaemonGcTagRule ...
type DaemonGcTagRule struct {
	CreatedAt int64                 `gorm:"autoCreateTime:milli"`
	UpdatedAt int64                 `gorm:"autoUpdateTime:milli"`
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:milli"`
	ID        string                `gorm:"primaryKey;size:36"`

	NamespaceID *string
	Namespace   *Namespace

	IsRunning           bool `gorm:"default:false"`
	CronEnabled         bool `gorm:"default:false"`
	CronRule            *string
	CronNextTrigger     *int64
	RetentionRuleType   enums.RetentionRuleType
	RetentionRuleAmount int64
	RetentionPattern    *string
}

// DaemonGcTagRunner ...
type DaemonGcTagRunner struct {
	CreatedAt int64                 `gorm:"autoCreateTime:milli"`
	UpdatedAt int64                 `gorm:"autoUpdateTime:milli"`
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:milli"`
	ID        string                `gorm:"primaryKey;size:36"`

	RuleID string
	Rule   DaemonGcTagRule

	Message []byte
	Status  enums.TaskCommonStatus

	OperateType   enums.OperateType
	OperateUserID *string
	OperateUser   *User

	StartedAt    *int64
	EndedAt      *int64
	Duration     *int64
	SuccessCount *int64
	FailedCount  *int64
}

// DaemonGcTagRecords ...
type DaemonGcTagRecord struct {
	CreatedAt int64                 `gorm:"autoCreateTime:milli"`
	UpdatedAt int64                 `gorm:"autoUpdateTime:milli"`
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:milli"`
	ID        string                `gorm:"primaryKey;size:36"`

	RunnerID string
	Runner   DaemonGcTagRunner

	Tag     string
	Status  enums.GcRecordStatus `gorm:"default:Success"`
	Message []byte
}

// DaemonGcRepositoryRule ...
type DaemonGcRepositoryRule struct {
	CreatedAt int64                 `gorm:"autoCreateTime:milli"`
	UpdatedAt int64                 `gorm:"autoUpdateTime:milli"`
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:milli"`
	ID        string                `gorm:"primaryKey;size:36"`

	NamespaceID *string
	Namespace   *Namespace

	IsRunning       bool `gorm:"default:false"`
	RetentionDay    int  `gorm:"default:0"`
	CronEnabled     bool `gorm:"default:false"`
	CronRule        *string
	CronNextTrigger *int64
}

// DaemonGcRepositoryRunner ...
type DaemonGcRepositoryRunner struct {
	CreatedAt int64                 `gorm:"autoCreateTime:milli"`
	UpdatedAt int64                 `gorm:"autoUpdateTime:milli"`
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:milli"`
	ID        string                `gorm:"primaryKey;size:36"`

	RuleID string
	Rule   DaemonGcRepositoryRule

	Status  enums.TaskCommonStatus `gorm:"status"`
	Message []byte

	OperateType   enums.OperateType
	OperateUserID *string
	OperateUser   *User

	StartedAt    *int64
	EndedAt      *int64
	Duration     *int64
	SuccessCount *int64
	FailedCount  *int64
}

// DaemonGcRepositoryRecord ...
type DaemonGcRepositoryRecord struct {
	CreatedAt int64                 `gorm:"autoCreateTime:milli"`
	UpdatedAt int64                 `gorm:"autoUpdateTime:milli"`
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:milli"`
	ID        string                `gorm:"primaryKey;size:36"`

	RunnerID string
	Runner   DaemonGcRepositoryRunner

	Repository string
	Status     enums.GcRecordStatus `gorm:"default:Success"`
	Message    []byte
}

// DaemonGcArtifactRule ...
type DaemonGcArtifactRule struct {
	CreatedAt int64                 `gorm:"autoCreateTime:milli"`
	UpdatedAt int64                 `gorm:"autoUpdateTime:milli"`
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:milli"`
	ID        string                `gorm:"primaryKey;size:36"`

	NamespaceID *string
	Namespace   *Namespace

	IsRunning       bool `gorm:"default:false"`
	RetentionDay    int  `gorm:"default:0"`
	CronEnabled     bool `gorm:"default:false"`
	CronRule        *string
	CronNextTrigger *int64
}

type DaemonGcArtifactRunner struct {
	CreatedAt int64                 `gorm:"autoCreateTime:milli"`
	UpdatedAt int64                 `gorm:"autoUpdateTime:milli"`
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:milli"`
	ID        string                `gorm:"primaryKey;size:36"`

	RuleID string
	Rule   DaemonGcArtifactRule

	Status  enums.TaskCommonStatus `gorm:"status"`
	Message []byte

	OperateType   enums.OperateType
	OperateUserID *string
	OperateUser   *User

	StartedAt    *int64
	EndedAt      *int64
	Duration     *int64
	SuccessCount *int64
	FailedCount  *int64
}

// DaemonGcArtifactRecord ...
type DaemonGcArtifactRecord struct {
	CreatedAt int64                 `gorm:"autoCreateTime:milli"`
	UpdatedAt int64                 `gorm:"autoUpdateTime:milli"`
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:milli"`
	ID        string                `gorm:"primaryKey;size:36"`

	RunnerID string
	Runner   DaemonGcArtifactRunner

	Digest  string
	Status  enums.GcRecordStatus `gorm:"default:Success"`
	Message []byte
}

// DaemonGcBlobRule ...
type DaemonGcBlobRule struct {
	CreatedAt int64                 `gorm:"autoCreateTime:milli"`
	UpdatedAt int64                 `gorm:"autoUpdateTime:milli"`
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:milli"`
	ID        string                `gorm:"primaryKey;size:36"`

	IsRunning       bool `gorm:"default:false"`
	RetentionDay    int  `gorm:"default:0"`
	CronEnabled     bool `gorm:"default:false"`
	CronRule        *string
	CronNextTrigger *int64
}

// DaemonGcBlobRunner ...
type DaemonGcBlobRunner struct {
	CreatedAt int64                 `gorm:"autoCreateTime:milli"`
	UpdatedAt int64                 `gorm:"autoUpdateTime:milli"`
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:milli"`
	ID        string                `gorm:"primaryKey;size:36"`

	RuleID string
	Rule   DaemonGcBlobRule

	Status  enums.TaskCommonStatus `gorm:"status"`
	Message []byte

	OperateType   enums.OperateType
	OperateUserID *string
	OperateUser   *User

	StartedAt    *int64
	EndedAt      *int64
	Duration     *int64
	SuccessCount *int64
	FailedCount  *int64
}

type DaemonGcBlobRecord struct {
	CreatedAt int64                 `gorm:"autoCreateTime:milli"`
	UpdatedAt int64                 `gorm:"autoUpdateTime:milli"`
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:milli"`
	ID        string                `gorm:"primaryKey;size:36"`

	RunnerID string
	Runner   DaemonGcBlobRunner

	Digest  string
	Status  enums.GcRecordStatus `gorm:"default:Success"`
	Message []byte
}

// GcStorageDeletionTask records object storage deletion intent for retryable GC.
type GcStorageDeletionTask struct {
	CreatedAt int64                 `gorm:"autoCreateTime:milli"`
	UpdatedAt int64                 `gorm:"autoUpdateTime:milli"`
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:milli"`
	ID        string                `gorm:"primaryKey;size:36"`

	Daemon          enums.Daemon
	RunnerID        string
	ResourceType    string
	ResourceID      string
	StoragePath     string
	StoragePathHash string
	Status          enums.TaskCommonStatus `gorm:"default:Pending"`
	Attempts        int
	Message         []byte
}
