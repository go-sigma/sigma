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

package validators

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	pwdvalidate "github.com/wagslane/go-password-validator"

	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/xerrors"
)

// GetPassword handles the validate password request
//
//	@Summary	Validate password
//	@Tags		Validator
//	@security	BasicAuth
//	@Accept		json
//	@Produce	json
//	@Router		/validators/password [get]
//	@Param		message	body	types.ValidatePasswordRequest	true	"Validate password object"
//	@Success	204
//	@Failure	400	{object}	xerrors.ErrCode
func (h *handler) GetPassword(c echo.Context) error {
	var req types.ValidatePasswordRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, fmt.Sprintf("Bind and validate request body failed: %v", err))
	}

	err = pwdvalidate.Validate(req.Password, consts.PwdStrength)
	if err != nil {
		log.Error().Err(err).Msg("Password strength is not enough")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, fmt.Sprintf("Password strength is not enough: %v", err))
	}
	return c.NoContent(http.StatusNoContent)
}
