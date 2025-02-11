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
	"net/http"
	"strings"

	"github.com/distribution/reference"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"

	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/xerrors"
)

// GetReference handles the validate reference request
//
//	@Summary	Validate reference
//	@Tags		Validator
//	@security	BasicAuth
//	@Accept		json
//	@Produce	json
//	@Router		/validators/reference [get]
//	@Param		reference	query	string	true	"Reference"
//	@Success	204
//	@Failure	400	{object}	xerrors.ErrCode
func (h *handler) GetReference(c echo.Context) error {
	var req types.GetValidatorReferenceRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, err.Error())
	}
	if len(strings.Split(req.Reference, "/")) < 2 {
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, "reference name should have one slash at last")
	}
	_, err = reference.ParseNormalizedNamed(req.Reference)
	if err != nil {
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}
