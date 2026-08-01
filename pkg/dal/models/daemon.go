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

// DaemonGcRule defines retention configuration for one garbage collection type.
type DaemonGcRule struct {
	CreatedAt int64                 `gorm:"autoCreateTime:milli"`
	UpdatedAt int64                 `gorm:"autoUpdateTime:milli"`
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:milli"`
	ID        string                `gorm:"primaryKey;size:36"`

	Type        enums.Daemon
	NamespaceID *string
	Namespace   *Namespace

	IsRunning           bool `gorm:"default:false"`
	RetentionDay        int  `gorm:"default:0"`
	CronEnabled         bool `gorm:"default:false"`
	CronRule            *string
	CronNextTrigger     *int64
	RetentionRuleType   enums.RetentionRuleType
	RetentionRuleAmount int64
	RetentionPattern    *string
}

// DaemonGcRunner records one garbage collection execution.
type DaemonGcRunner struct {
	CreatedAt int64                 `gorm:"autoCreateTime:milli"`
	UpdatedAt int64                 `gorm:"autoUpdateTime:milli"`
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:milli"`
	ID        string                `gorm:"primaryKey;size:36"`

	RuleID string
	Rule   DaemonGcRule

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

// DaemonGcRecord records one resource processed by a garbage collection execution.
type DaemonGcRecord struct {
	CreatedAt int64                 `gorm:"autoCreateTime:milli"`
	UpdatedAt int64                 `gorm:"autoUpdateTime:milli"`
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:milli"`
	ID        string                `gorm:"primaryKey;size:36"`

	RunnerID string
	Runner   DaemonGcRunner

	Resource string
	Status   enums.GcRecordStatus `gorm:"default:Success"`
	Message  []byte
}

// Legacy type aliases preserve the public service and handler contracts while
// the persistence layer uses the shared GC tables.
type (
	DaemonGcTagRule          = DaemonGcRule
	DaemonGcRepositoryRule   = DaemonGcRule
	DaemonGcArtifactRule     = DaemonGcRule
	DaemonGcBlobRule         = DaemonGcRule
	DaemonGcTagRunner        = DaemonGcRunner
	DaemonGcRepositoryRunner = DaemonGcRunner
	DaemonGcArtifactRunner   = DaemonGcRunner
	DaemonGcBlobRunner       = DaemonGcRunner
	DaemonGcTagRecord        = DaemonGcRecord
	DaemonGcRepositoryRecord = DaemonGcRecord
	DaemonGcArtifactRecord   = DaemonGcRecord
	DaemonGcBlobRecord       = DaemonGcRecord
)

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
