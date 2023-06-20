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

package upload

import (
	"strings"

	"github.com/labstack/echo/v4"
)

func (h *handler) getRepository(c echo.Context) string {
	uri := c.Request().URL.Path
	return strings.TrimPrefix(strings.TrimSuffix(uri[:strings.LastIndex(uri, "/")], "/blobs"), "/v2/")
}
