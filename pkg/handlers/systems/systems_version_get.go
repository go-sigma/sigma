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

package systems

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/version"
)

// GetEndpoint handles the get version request
//
//	@Summary	Get version
//	@Tags		System
//	@Accept		json
//	@Produce	json
//	@Router		/systems/version [get]
//	@Success	200	{object}	types.GetSystemVersionResponse
func (h *handler) GetVersion(c echo.Context) error {
	return c.JSON(http.StatusOK, types.GetSystemVersionResponse{
		Version:   version.Version,
		GitHash:   version.GitHash,
		BuildDate: version.BuildDate,
	})
}
