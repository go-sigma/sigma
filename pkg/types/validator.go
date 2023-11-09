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

// GetValidatorReferenceRequest ...
type GetValidatorReferenceRequest struct {
	Reference string `json:"reference" query:"reference" validate:"required"`
}

// GetValidatorTagRequest ...
type GetValidatorTagRequest struct {
	Tag string `json:"tag" query:"tag" validate:"required"`
}

// ValidatePasswordRequest ...
type ValidatePasswordRequest struct {
	Password string `json:"password" validate:"required" example:"Admin@123"`
}

// ValidateCronRequest ...
type ValidateCronRequest struct {
	Cron string `json:"cron" validate:"required" example:"0 0 * * 6"`
}
