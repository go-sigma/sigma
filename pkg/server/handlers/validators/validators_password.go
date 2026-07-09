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

	"github.com/gin-gonic/gin"
	pwdvalidate "github.com/wagslane/go-password-validator"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/server/errcode"
)

// GetPassword handles the validate password request
//
//	@Summary	Validate password
//	@Tags		Validator
//	@security	BasicAuth
//	@Accept		json
//	@Produce	json
//	@Router		/validators/password [get]
//	@Param		message	body	api.ValidatePasswordRequest	true	"Validate password object"
//	@Success	204
//	@Failure	400	{object}	errcode.ErrCode
func (h *handler) GetPassword(c *gin.Context, req *api.ValidatePasswordRequest) {
	err := pwdvalidate.Validate(req.Password, consts.PwdStrength)
	if err != nil {
		slog.Error("password strength is not enough", "err", err)
		errcode.NewHTTPError(c, errcode.HTTPErrCodeBadRequest, fmt.Sprintf("Password strength is not enough: %v", err))
		return
	}
	c.Status(http.StatusNoContent)
}
