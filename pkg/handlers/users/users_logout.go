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

package user

import (
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"

	"github.com/go-sigma/sigma/pkg/xerrors"
)

// Logout handles the logout request
func (h *handlers) Logout(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	jti, ok := c.Get("jti").(string)
	if !ok || jti == "" {
		log.Error().Str("jti", jti).Msg("Get jti failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeUnauthorized, "Get jti failed")
	}
	err := h.tokenService.Revoke(ctx, jti)
	if err != nil {
		log.Error().Err(err).Msg("Revoke token failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, err.Error())
	}
	return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeOK, "Logout successfully")
}
