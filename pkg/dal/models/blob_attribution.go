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

// BlobAttribution records which blob is charged to which quota scope.
//
// A blob can be shared by multiple namespaces and repositories, so the
// ownership cannot live on blobs itself. The primary key makes the write
// idempotent: concurrent manifest pushes only insert the row once.
type BlobAttribution struct {
	CreatedAt int64 `gorm:"autoCreateTime:milli"`

	ScopeType string `gorm:"primaryKey;size:16"` // namespace / repository
	ScopeID   string `gorm:"primaryKey;size:36"`
	BlobID    string `gorm:"primaryKey;size:36"`
	Size      int64  `gorm:"default:0"`
	BatchID   string `gorm:"size:36;default:''"`
}
