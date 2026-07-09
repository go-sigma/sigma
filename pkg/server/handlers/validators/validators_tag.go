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
	"github.com/gin-gonic/gin"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/server/errcode"
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
//	@Failure	400	{object}	errcode.ErrCode
func (h *handler) GetTag(c *gin.Context, req *api.GetValidatorTagRequest) {
	t, err := template.New("tag").Funcs(sprig.FuncMap()).Parse(req.Tag)
	if err != nil {
		errcode.NewHTTPError(c, errcode.HTTPErrCodeBadRequest, fmt.Sprintf("Parse tag template failed: %v", err))
		return
	}
	var sample = api.BuildTagOption{ScmBranch: "main", ScmTag: "v0.1", ScmRef: "581758eb7d96ae4d113649668fa96acc74d46e7f"}
	var buffer bytes.Buffer
	err = t.Execute(&buffer, sample)
	if err != nil {
		errcode.NewHTTPError(c, errcode.HTTPErrCodeBadRequest, fmt.Sprintf("Render tag template failed: %v", err))
		return
	}
	var tag = buffer.String()
	if !consts.TagRegexp.MatchString(tag) {
		errcode.NewHTTPError(c, errcode.HTTPErrCodeBadRequest, "Tag is invalid")
		return
	}
	c.Status(http.StatusNoContent)
}
