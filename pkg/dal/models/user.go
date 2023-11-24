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

// User is the model for the user table.
type User struct {
	CreatedAt int64                 `gorm:"autoUpdateTime:milli"`
	UpdatedAt int64                 `gorm:"autoUpdateTime:milli"`
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:milli"`
	ID        int64                 `gorm:"primaryKey"`

	Username       string
	Password       *string
	Email          *string
	LastLogin      time.Time        `gorm:"autoCreateTime"`
	Status         enums.UserStatus `gorm:"default:Active"`
	Role           enums.UserRole   `gorm:"default:User"`
	NamespaceLimit int64            `gorm:"default:0"`
	NamespaceCount int64            `gorm:"default:0"`
}

// User3rdParty ...
type User3rdParty struct {
	CreatedAt int64                 `gorm:"autoUpdateTime:milli"`
	UpdatedAt int64                 `gorm:"autoUpdateTime:milli"`
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:milli"`
	ID        int64                 `gorm:"primaryKey"`

	UserID       int64
	Provider     enums.Provider
	AccountID    *string
	Token        *string
	RefreshToken *string

	CrLastUpdateTimestamp time.Time              `gorm:"autoCreateTime"`
	CrLastUpdateStatus    enums.TaskCommonStatus `gorm:"default:Doing"`
	CrLastUpdateMessage   *string

	User User
}

// TableName ...
func (User3rdParty) TableName() string {
	return "user_3rdparty"
}

type UserRecoverCode struct {
	CreatedAt int64                 `gorm:"autoUpdateTime:milli"`
	UpdatedAt int64                 `gorm:"autoUpdateTime:milli"`
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:milli"`
	ID        int64                 `gorm:"primaryKey"`

	UserID int64
	Code   string

	User User
}
