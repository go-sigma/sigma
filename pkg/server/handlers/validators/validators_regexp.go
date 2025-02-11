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
	"regexp"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"

	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/xerrors"
)

// ValidateRegexp handles the validate regexp request
//
//	@Summary	Validate regexp
//	@Tags		Validator
//	@security	BasicAuth
//	@Accept		json
//	@Produce	json
//	@Router		/validators/regexp [post]
//	@Param		message	body	types.ValidateCronRequest	true	"Validate regexp object"
//	@Success	204
//	@Failure	400	{object}	xerrors.ErrCode
func (h *handler) ValidateRegexp(c echo.Context) error {
	var req types.ValidateRegexpRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, fmt.Sprintf("Bind and validate request body failed: %v", err))
	}

	_, err = regexp.Compile(req.Regexp)
	if err != nil {
		log.Error().Err(err).Msg("Parse regex failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, fmt.Sprintf("Parse regex failed: %v", err))
	}

	return c.NoContent(http.StatusNoContent)
}
