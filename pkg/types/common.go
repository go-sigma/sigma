// Copyright 2023 XImager
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

// CommonList is the common list struct
type CommonList struct {
	Total int64 `json:"total" example:"1"`
	Items []any `json:"items"`
}

// Pagination is the pagination struct
type Pagination struct {
	Page  *int `json:"page" query:"page" example:"1"`
	Limit *int `json:"limit" query:"limit" example:"10"`
}

// Sortable is the sort struct
type Sortable struct {
	Sort   *string           `json:"sort" query:"sort" example:"created_at"`
	Method *enums.SortMethod `json:"method" query:"method" example:"desc"`
}
