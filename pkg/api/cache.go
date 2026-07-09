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

package api

// CreateCacheRequest is the request for creating a builder cache.
type CreateCacheRequest struct {
	BuilderID string `json:"builder_id" param:"builder_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// DeleteCacheRequest is the request for deleting a builder cache.
type DeleteCacheRequest struct {
	BuilderID string `json:"builder_id" param:"builder_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// GetCacheRequest is the request for getting a builder cache.
type GetCacheRequest struct {
	BuilderID string `json:"builder_id" param:"builder_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
}
