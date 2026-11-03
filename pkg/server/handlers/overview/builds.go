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

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/server/errcode"
	"github.com/go-sigma/sigma/pkg/utils"
)

// GetDailyBuilds handles the daily build request
//
//	@Summary	Get daily build results
//	@security	BasicAuth
//	@Tags		Overview
//	@Accept		json
//	@Produce	json
//	@Param		days	query		int		false	"Number of days to look back"	minimum(1)	maximum(90)	default(30)
//	@Param		scope	query		string	false	"Data scope: self for the current user, global for the whole instance (administrators only)"	Enums(self, global)
//	@Success	200		{object}	api.ListDailyBuildResponse
//	@Failure	400		{object}	errcode.ErrCode
//	@Failure	401		{object}	errcode.ErrCode
//	@Failure	403		{object}	errcode.ErrCode
//	@Router		/builds/daily [get]
func (h *handler) GetDailyBuilds(c *gin.Context) {
	ctx := c.Request.Context()

	user, ok := utils.GetFromCtx[*models.User](c, consts.ContextUser)
	if !ok {
		errcode.NewHTTPError(c, errcode.HTTPErrCodeUnauthorized)
		return
	}

	days, err := queryInt(c, "days", 30)
	if err != nil {
		errcode.NewHTTPError(c, errcode.HTTPErrCodeBadRequest, err.Error())
		return
	}
	items, summary, err := h.OverviewSvc.GetDailyBuilds(ctx, user, days, api.OverviewScope(c.Query("scope")))
	if err != nil {
		if e, ok := errcode.AsType[errcode.ErrCode](err); ok {
			errcode.NewHTTPError(c, e)
			return
		}
		errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError, err.Error())
		return
	}
	c.JSON(http.StatusOK, api.ListDailyBuildResponse{Items: items, Summary: summary})
}
