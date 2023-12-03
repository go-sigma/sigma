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

	"github.com/go-sigma/sigma/pkg/types/enums"
)

// CodeRepository ...
type CodeRepository struct {
	CreatedAt int64                 `gorm:"autoUpdateTime:milli"`
	UpdatedAt int64                 `gorm:"autoUpdateTime:milli"`
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:milli"`
	ID        int64                 `gorm:"primaryKey"`

	RepositoryID string

	User3rdPartyID int64        `gorm:"column:user_3rdparty_id"`
	User3rdParty   User3rdParty `gorm:"foreignKey:User3rdPartyID"`

	OwnerID  string
	Owner    string // in github named owner.name
	IsOrg    bool
	Name     string // in github named full_name
	SshUrl   string // in github named ssh_url
	CloneUrl string // in github named clone_url

	OciRepoCount int64

	Branches []CodeRepositoryBranch
}

// CodeRepositoryBranch ...
type CodeRepositoryBranch struct {
	CreatedAt int64                 `gorm:"autoUpdateTime:milli"`
	UpdatedAt int64                 `gorm:"autoUpdateTime:milli"`
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:milli"`
	ID        int64                 `gorm:"primaryKey"`

	CodeRepositoryID int64
	Name             string
}

// CodeRepositoryOwner ...
type CodeRepositoryOwner struct {
	CreatedAt int64                 `gorm:"autoUpdateTime:milli"`
	UpdatedAt int64                 `gorm:"autoUpdateTime:milli"`
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:milli"`
	ID        int64                 `gorm:"primaryKey"`

	User3rdPartyID int64 `gorm:"column:user_3rdparty_id"`
	OwnerID        string
	Owner          string // in github named owner.name
	IsOrg          bool

	User3rdParty User3rdParty `gorm:"foreignKey:User3rdPartyID"`
}

// CodeRepositoryCloneCredential ...
type CodeRepositoryCloneCredential struct {
	CreatedAt int64                 `gorm:"autoUpdateTime:milli"`
	UpdatedAt int64                 `gorm:"autoUpdateTime:milli"`
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:milli"`
	ID        int64                 `gorm:"primaryKey"`

	User3rdPartyID int64 `gorm:"column:user_3rdparty_id"`
	Type           enums.ScmCredentialType
	SshKey         *string
	Username       *string
	Password       *string
	Token          *string

	User3rdParty User3rdParty `gorm:"foreignKey:User3rdPartyID"`
}
