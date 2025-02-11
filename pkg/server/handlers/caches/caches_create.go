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

package caches

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"

	"github.com/go-sigma/sigma/pkg/storage"
	"github.com/go-sigma/sigma/pkg/xerrors"
)

// CreateCache handles the create cache request
//
//	@Summary	Create cache
//	@Tags		Cache
//	@security	BasicAuth
//	@Accept		application/octet-stream
//	@Produce	json
//	@Router		/caches/{builder_id} [post]
//	@Param		builder_id	path	string	true	"Builder ID"
//	@Param		file		body	string	true	"Cache file"
//	@Success	201
//	@Failure	404	{object}	xerrors.ErrCode
//	@Failure	500	{object}	xerrors.ErrCode
func (h *handler) CreateCache(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	builderIDStr := c.Param("builder_id")
	if builderIDStr == "" {
		log.Error().Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, "Bind and validate request body failed")
	}
	builderID, err := strconv.ParseInt(builderIDStr, 10, 64)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, fmt.Sprintf("Bind and validate request body failed: %v", err))
	}

	err = storage.Driver.Upload(ctx, h.genPath(builderID), c.Request().Body)
	if err != nil {
		log.Error().Err(err).Msg("Upload file failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Upload file failed: %v", err))
	}

	err = c.Request().Body.Close()
	if err != nil {
		log.Error().Err(err).Msg("Close body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Close body failed: %v", err))
	}

	return c.NoContent(http.StatusCreated)
}
