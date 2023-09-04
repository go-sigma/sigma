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

// Builder represents a builder
type Builder struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:milli"`
	ID        int64                 `gorm:"primaryKey"`

	RepositoryID int64
	Active       bool

	Source enums.BuilderSource

	// source CodeRepository
	CodeRepositoryID *int64
	// source Dockerfile
	Dockerfile []byte
	// source SelfCodeRepository
	ScmRepository     string
	ScmCredentialType enums.ScmCredentialType
	ScmToken          string
	ScmSshKey         string
	ScmUsername       string
	ScmPassword       string

	// common settings
	ScmDepth     int
	ScmSubmodule bool

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
	BuildkitContext            string `gorm:"default:."`
	BuildkitDockerfile         string `gorm:"default:Dockerfile"`
	BuildkitPlatforms          string `gorm:"default:linux/amd64"`

	Repository *Repository
}

// BuilderRunner represents a builder runner
type BuilderRunner struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:milli"`
	ID        int64                 `gorm:"primaryKey"`

	BuilderID int64
	Log       []byte
	Status    enums.BuildStatus

	Tag               string
	ScmBranch         string
	BuildkitPlatforms string `gorm:"default:linux/amd64"`

	Builder Builder
}
