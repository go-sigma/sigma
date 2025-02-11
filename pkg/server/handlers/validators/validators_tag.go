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
	"bytes"
	"fmt"
	"net/http"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"

	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/xerrors"
)

// GetTag handles the validate tag request
//
//	@Summary	Validate tag
//	@Tags		Validator
//	@security	BasicAuth
//	@Accept		json
//	@Produce	json
//	@Router		/validators/tag [get]
//	@Param		tag	query	string	true	"Reference"
//	@Success	204
//	@Failure	400	{object}	xerrors.ErrCode
func (h *handler) GetTag(c echo.Context) error {
	var req types.GetValidatorTagRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, err.Error())
	}
	t, err := template.New("tag").Funcs(sprig.FuncMap()).Parse(req.Tag)
	if err != nil {
		log.Error().Err(err).Str("template", req.Tag).Msg("Parse tag template failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, fmt.Sprintf("Parse tag template failed: %v", err))
	}
	var sample = types.BuildTagOption{ScmBranch: "main", ScmTag: "v0.1", ScmRef: "581758eb7d96ae4d113649668fa96acc74d46e7f"}
	var buffer bytes.Buffer
	err = t.Execute(&buffer, sample)
	if err != nil {
		log.Error().Err(err).Str("template", req.Tag).Msg("Render tag template failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, fmt.Sprintf("Render tag template failed: %v", err))
	}
	var tag = buffer.String()
	if !consts.TagRegexp.MatchString(tag) {
		log.Error().Str("template", req.Tag).Str("tag", tag).Msg("Tag is invalid")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, "Tag is invalid")
	}
	return c.NoContent(http.StatusNoContent)
}
