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
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/xerrors"
)

// CreateCache ...
func (h *handlers) CreateCache(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	urlValues := c.Request().URL.Query()

	builderID, err := strconv.ParseInt(urlValues.Get("builder_id"), 10, 64)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, fmt.Sprintf("Bind and validate request body failed: %v", err))
	}
	runnerID, err := strconv.ParseInt(urlValues.Get("runner_id"), 10, 64)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, fmt.Sprintf("Bind and validate request body failed: %v", err))
	}
	var req = types.CreateCacheRequest{
		BuilderID: builderID,
		RunnerID:  runnerID,
	}

	err = storage.Driver.Upload(ctx, h.genPath(req.BuilderID), c.Request().Body)
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
