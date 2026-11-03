// Copyright 2026 sigma
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

package overview

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/server/errcode"
	"github.com/go-sigma/sigma/pkg/utils"
)

// GetOverview handles the home dashboard request
//
//	@Summary	Get the home dashboard of the current user
//	@security	BasicAuth
//	@Tags		Overview
//	@Accept		json
//	@Produce	json
//	@Success	200	{object}	api.OverviewResponse
//	@Failure	401	{object}	errcode.ErrCode
//	@Failure	500	{object}	errcode.ErrCode
//	@Router		/overview [get]
func (h *handler) GetOverview(c *gin.Context) {
	ctx := c.Request.Context()

	user, ok := utils.GetFromCtx[*models.User](c, consts.ContextUser)
	if !ok {
		errcode.NewHTTPError(c, errcode.HTTPErrCodeUnauthorized)
		return
	}

	resp, err := h.OverviewSvc.GetOverview(ctx, user)
	if err != nil {
		if e, ok := errcode.AsType[errcode.ErrCode](err); ok {
			errcode.NewHTTPError(c, e)
			return
		}
		errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError, err.Error())
		return
	}
	c.JSON(http.StatusOK, resp)
}
