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

import "gorm.io/plugin/soft_delete"

// Door ...
type Door struct {
	CreatedAt int64                 `gorm:"autoCreateTime:milli"`
	UpdatedAt int64                 `gorm:"autoUpdateTime:milli"`
	DeletedAt soft_delete.DeletedAt `gorm:"softDelete:milli"`
	ID        int64                 `gorm:"primaryKey"`
	ULID      string                `gorm:"column:ulid,maxsize:36"`

	Role     string `gorm:"column:role"`
	Resource string `gorm:"column:resource"`
	Effect   string `gorm:"column:effect"`
}

type AuthSource struct {
	Name     string `json:"name"`
	Position string `json:"position"`
}
