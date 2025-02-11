// Copyright 2025 sigma
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

// CasbinRule represents a casbin rule.
type CasbinRule struct {
	ID    int64   `gorm:"primaryKey"`
	PType *string `gorm:"column:ptype"`
	V0    *string `gorm:"column:v0"`
	V1    *string `gorm:"column:v1"`
	V2    *string `gorm:"column:v2"`
	V3    *string `gorm:"column:v3"`
	V4    *string `gorm:"column:v4"`
	V5    *string `gorm:"column:v5"`
}
