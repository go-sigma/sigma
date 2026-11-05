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

// QuotaUsage is the aggregated usage per quota scope. The global physical
// usage is stored in the single row with scope_type = "global" and an empty
// scope_id; namespace/repository counters currently stay on their own tables.
type QuotaUsage struct {
	UpdatedAt int64 `gorm:"autoUpdateTime:milli"`

	ScopeType string `gorm:"primaryKey;size:16"`
	ScopeID   string `gorm:"primaryKey;size:36"`
	Size      int64  `gorm:"default:0"`
}
