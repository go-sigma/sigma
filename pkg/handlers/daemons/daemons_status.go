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

package daemons

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"

	"github.com/go-sigma/sigma/pkg/types"
)

// Status get the specific daemon task status
// @Summary Get specific daemon task status
// @Tags Daemon
// @Accept json
// @Produce json
// @Router /daemons/{daemon}/ [get]
// @Param daemon path string true "Daemon name"
// @Param namespace_id query string false "Namespace ID"
// @Success 202
// @Failure 404 {object} xerrors.ErrCode
// @Failure 500 {object} xerrors.ErrCode
func (h *handlers) Status(c echo.Context) error {
	return c.JSON(http.StatusOK, types.GetEndpointResponse{
		Endpoint: viper.GetString("server.endpoint"),
	})
}
