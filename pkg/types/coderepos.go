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

import "github.com/go-sigma/sigma/pkg/types/enums"

type CodeRepositoryItem struct {
	ID       int64  `json:"id" example:"1"`
	Name     string `json:"name" example:"sigma"`
	Owner    string `json:"owner" example:"go-sigma"`
	CloneUrl string `json:"clone_url" example:"https://github.com/go-sigma/sigma.git"`
	SshUrl   string `json:"ssh_url" example:"git@github.com:go-sigma/sigma.git"`
}

// ListCodeRepositoryRequest represents the request to list code repository.
type ListCodeRepositoryRequest struct {
	Pagination
	Sortable

	Provider enums.Provider `json:"provider" param:"provider" validate:"required,is_valid_provider"`
	Owner    string         `json:"owner"`
}
