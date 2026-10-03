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

package analytics

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/server/errcode"
	"github.com/go-sigma/sigma/pkg/utils"
)

// GetUserPushHeatmap handles the user push heatmap request.
//
//	@Summary	Get user push activity heatmap
//	@security	BasicAuth
//	@Tags		Analytics
//	@Accept		json
//	@Produce	json
//	@Router		/users/{user_id}/activity/heatmap [get]
//	@Param		user_id	path		string	true	"User ID"
//	@Param		days	query		int		false	"Number of days to look back"	minimum(1)	maximum(400)	default(365)
//	@Success	200		{object}	api.ListDailyCountResponse
//	@Failure	400		{object}	errcode.ErrCode
//	@Failure	401		{object}	errcode.ErrCode
//	@Failure	500		{object}	errcode.ErrCode
func (h *handler) GetUserPushHeatmap(c *gin.Context) {
	ctx := c.Request.Context()
	user, ok := utils.GetFromCtx[*models.User](c, consts.ContextUser)
	if !ok {
		errcode.NewHTTPError(c, errcode.HTTPErrCodeUnauthorized)
		return
	}
	userID := c.Param("user_id")
	if userID == "" {
		errcode.NewHTTPError(c, errcode.HTTPErrCodeBadRequest, "invalid user id")
		return
	}
	if user.ID != userID && user.Role != enums.UserRoleAdmin && user.Role != enums.UserRoleRoot {
		errcode.NewHTTPError(c, errcode.HTTPErrCodeUnauthorized, "no permission with this api")
		return
	}
	days, err := parseDays(c.DefaultQuery("days", "365"), 1, 400)
	if err != nil {
		errcode.NewHTTPError(c, errcode.HTTPErrCodeBadRequest, err.Error())
		return
	}
	items, err := h.AnalyticsSvc.GetUserPushHeatmap(ctx, userID, days)
	if err != nil {
		errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError, err.Error())
		return
	}
	c.JSON(http.StatusOK, api.ListDailyCountResponse{Items: items})
}

func parseDays(value string, minDays, maxDays int) (int, error) {
	days, err := strconv.Atoi(value)
	if err != nil {
		return 0, err
	}
	if days < minDays || days > maxDays {
		return 0, errcode.HTTPErrCodeBadRequest.Detail("days out of range")
	}
	return days, nil
}
