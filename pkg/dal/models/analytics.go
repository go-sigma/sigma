// Copyright 2026 sigma
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

// UserActivityHourly stores per-user hourly push increments.
type UserActivityHourly struct {
	CreatedAt int64  `gorm:"autoCreateTime:milli"`
	UpdatedAt int64  `gorm:"autoUpdateTime:milli"`
	ID        string `gorm:"primaryKey;size:36"`

	Hour      int64  `gorm:"uniqueIndex:idx_user_activity_hourly_hour_user;index:idx_user_activity_hourly_user_hour"`
	UserID    string `gorm:"uniqueIndex:idx_user_activity_hourly_hour_user;index:idx_user_activity_hourly_user_hour"`
	PushCount int64  `gorm:"default:0"`
}

// NamespaceActivityHourly stores per-namespace hourly activity increments.
type NamespaceActivityHourly struct {
	CreatedAt int64  `gorm:"autoCreateTime:milli"`
	UpdatedAt int64  `gorm:"autoUpdateTime:milli"`
	ID        string `gorm:"primaryKey;size:36"`

	Hour        int64  `gorm:"uniqueIndex:idx_namespace_activity_hourly_hour_namespace;index:idx_namespace_activity_hourly_namespace_hour"`
	NamespaceID string `gorm:"uniqueIndex:idx_namespace_activity_hourly_hour_namespace;index:idx_namespace_activity_hourly_namespace_hour"`
	PushCount   int64  `gorm:"default:0"`
	PullCount   int64  `gorm:"default:0"`
	SizeDelta   int64  `gorm:"default:0"`
	TagDelta    int64  `gorm:"default:0"`
}
