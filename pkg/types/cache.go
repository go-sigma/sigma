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

package types

// CreateCacheRequest ...
type CreateCacheRequest struct {
	BuilderID int64 `json:"builder_id" query:"builder_id" validate:"required,number" example:"10"`
	RunnerID  int64 `json:"runner_id" query:"runner_id" validate:"required,number" example:"10"`
}

// DeleteCacheRequest ...
type DeleteCacheRequest struct {
	BuilderID int64 `json:"builder_id" query:"builder_id" validate:"required,number" example:"10"`
	RunnerID  int64 `json:"runner_id" query:"runner_id" validate:"required,number" example:"10"`
}

// GetCacheRequest ...
type GetCacheRequest struct {
	BuilderID int64 `json:"builder_id" query:"builder_id" validate:"required,number" example:"10"`
	RunnerID  int64 `json:"runner_id" query:"runner_id" validate:"required,number" example:"10"`
}
