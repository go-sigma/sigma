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

package users

import (
	"errors"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/token"
	"github.com/go-sigma/sigma/pkg/xerrors"
)

// Logout handles the logout request
//
//	@Summary	Logout user
//	@security	BasicAuth
//	@Tags		User
//	@Accept		json
//	@Produce	json
//	@Param		message	body	types.PostUserLogoutRequest	true	"Logout user object"
//	@Router		/users/logout [post]
//	@Failure	500	{object}	xerrors.ErrCode
//	@Failure	401	{object}	xerrors.ErrCode
//	@Success	204
func (h *handler) Logout(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	var req types.PostUserLogoutRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, err.Error())
	}

	var ids = sets.New[string]()
	for _, t := range req.Tokens {
		if t == "" {
			continue
		}
		id, _, err := h.TokenService.Validate(ctx, t)
		if err != nil {
			if errors.Is(err, token.ErrRevoked) || errors.Is(err, jwt.ErrTokenExpired) {
				continue
			}
			log.Error().Err(err).Str("token", t).Msg("Revoke token failed")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, err.Error())
		}
		ids.Insert(id)
	}

	jti, ok := c.Get("jti").(string)
	if !ok || jti == "" {
		log.Error().Str("jti", jti).Msg("Get jti failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeUnauthorized, "Get jti failed")
	}

	ids.Insert(jti)

	for {
		id, ok := ids.PopAny()
		if !ok {
			break
		}

		err = h.TokenService.Revoke(ctx, id)
		if err != nil {
			log.Error().Err(err).Msg("Revoke token failed")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, err.Error())
		}
	}

	return c.NoContent(http.StatusNoContent)
}
