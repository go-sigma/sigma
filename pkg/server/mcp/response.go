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

package mcpserver

import "github.com/go-sigma/sigma/pkg/api"

func listResponse(items any, total int64, pagination api.Pagination) map[string]any {
	return map[string]any{
		"items":      items,
		"total":      total,
		"pagination": pagination,
	}
}

func okResponse() map[string]bool {
	return map[string]bool{"ok": true}
}
