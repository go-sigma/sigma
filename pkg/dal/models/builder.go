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

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

// Builder represents a builder
type Builder struct {
	CreatedAt int64                 `gorm:"autoUpdateTime:milli"`
	UpdatedAt int64                 `gorm:"autoUpdateTime:milli"`
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:milli"`
	ID        int64                 `gorm:"primaryKey"`

	RepositoryID int64

	Source enums.BuilderSource

	// source CodeRepository
	CodeRepositoryID *int64
	// source Dockerfile
	Dockerfile []byte
	// source SelfCodeRepository
	ScmRepository     *string
	ScmCredentialType *enums.ScmCredentialType
	ScmToken          *string
	ScmSshKey         *string
	ScmUsername       *string
	ScmPassword       *string

	// common settings
	ScmBranch    *string // for SelfCodeRepository and CodeRepository
	ScmDepth     *int
	ScmSubmodule *bool

	// cron settings
	CronRule        *string
	CronBranch      *string
	CronTagTemplate *string
	CronNextTrigger *time.Time

	// webhook settings
	WebhookBranchName        *string
	WebhookBranchTagTemplate *string
	WebhookTagTagTemplate    *string

	// buildkit settings
	BuildkitInsecureRegistries string
	BuildkitContext            *string
	BuildkitDockerfile         *string
	BuildkitPlatforms          string `gorm:"default:linux/amd64"`
	BuildkitBuildArgs          *string

	Repository *Repository
}

// BuilderRunner represents a builder runner
type BuilderRunner struct {
	CreatedAt int64                 `gorm:"autoUpdateTime:milli"`
	UpdatedAt int64                 `gorm:"autoUpdateTime:milli"`
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:milli"`
	ID        int64                 `gorm:"primaryKey"`

	BuilderID int64
	Log       []byte
	Status    enums.BuildStatus `gorm:"default:Pending"`

	Tag         *string
	RawTag      string
	Description *string
	ScmBranch   *string

	StartedAt *time.Time
	EndedAt   *time.Time
	Duration  *int64

	Builder Builder
}

// AfterUpdate ...
func (b *BuilderRunner) AfterUpdate(tx *gorm.DB) error {
	if b == nil {
		return nil
	}

	var runnerObj BuilderRunner
	err := tx.Model(&BuilderRunner{}).Where("id = ?", b.ID).First(&runnerObj).Error
	if err != nil {
		return err
	}

	if runnerObj.Duration != nil {
		return nil
	}

	if runnerObj.StartedAt != nil && runnerObj.EndedAt != nil {
		var duration = runnerObj.EndedAt.Sub(ptr.To(runnerObj.StartedAt))
		err = tx.Model(&BuilderRunner{}).Where("id = ?", b.ID).Updates(
			map[string]any{
				"duration": duration.Milliseconds(),
				"id":       b.ID,
			}).Error
		if err != nil {
			return err
		}
	}

	return nil
}
