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

// CommonList is the common list struct
type CommonList struct {
	Total int64 `json:"total"`
	Items []any `json:"items"`
}

// Pagination is the pagination struct
type Pagination struct {
	PageSize int `json:"page_size" query:"page_size" validate:"required,gte=10,lte=100"`
	PageNum  int `json:"page_num" query:"page_num" validate:"required,gte=1"`
}
