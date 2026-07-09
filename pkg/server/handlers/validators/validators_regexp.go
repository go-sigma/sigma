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
	"log/slog"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/server/errcode"
)

// ValidateRegexp handles the validate regexp request
//
//	@Summary	Validate regexp
//	@Tags		Validator
//	@security	BasicAuth
//	@Accept		json
//	@Produce	json
//	@Router		/validators/regexp [post]
//	@Param		message	body	api.ValidateCronRequest	true	"Validate regexp object"
//	@Success	204
//	@Failure	400	{object}	errcode.ErrCode
func (h *handler) ValidateRegexp(c *gin.Context, req *api.ValidateRegexpRequest) {
	_, err := regexp.Compile(req.Regexp)
	if err != nil {
		slog.Error("parse regex failed", "err", err)
		errcode.NewHTTPError(c, errcode.HTTPErrCodeBadRequest, fmt.Sprintf("Parse regex failed: %v", err))
		return
	}

	c.Status(http.StatusNoContent)
}
