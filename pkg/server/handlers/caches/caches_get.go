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
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"

	"github.com/go-sigma/sigma/pkg/storage"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/xerrors"
)

// GetCache handles the get cache request
//
//	@Summary	Get cache
//	@Tags		Cache
//	@security	BasicAuth
//	@Accept		json
//	@Produce	application/octet-stream
//	@Router		/caches/{builder_id} [get]
//	@Param		builder_id	path		string	true	"Builder ID"
//	@Success	200			{string}	file	"Cache content"
//	@Failure	404			{object}	xerrors.ErrCode
//	@Failure	500			{object}	xerrors.ErrCode
func (h *handler) GetCache(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	var req types.GetCacheRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, fmt.Sprintf("Bind and validate request body failed: %v", err))
	}

	var path = h.genPath(req.BuilderID)
	reader, err := storage.Driver.Reader(ctx, path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Error().Err(err).Str("cache", path).Msg("Cache not found")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, fmt.Sprintf("Cache not found: %v", err))
		}
		log.Error().Err(err).Str("cache", path).Msg("Get cache failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Get cache failed: %v", err))
	}

	return c.Stream(http.StatusOK, "application/gzip", reader)
}
