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
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/server/errcode"
	"github.com/go-sigma/sigma/pkg/utils"
)

// GetNamespaceTrends handles the namespace trend request.
func (h *handler) GetNamespaceTrends(c *gin.Context) {
	ctx := c.Request.Context()
	user, ok := utils.GetFromCtx[*models.User](c, consts.ContextUser)
	if !ok {
		errcode.NewHTTPError(c, errcode.HTTPErrCodeUnauthorized)
		return
	}
	namespaceID := c.Param("namespace_id")
	if namespaceID == "" {
		errcode.NewHTTPError(c, errcode.HTTPErrCodeBadRequest, "invalid namespace id")
		return
	}
	authChecked, err := h.Authorizer.Namespace(ctx, *user, namespaceID, enums.AuthRead)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			errcode.NewHTTPError(c, errcode.HTTPErrCodeNotFound, "namespace not found")
			return
		}
		errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError, err.Error())
		return
	}
	if !authChecked {
		errcode.NewHTTPError(c, errcode.HTTPErrCodeUnauthorized, "no permission with this api")
		return
	}
	days, err := parseDays(c.DefaultQuery("days", "30"), 1, 30)
	if err != nil {
		errcode.NewHTTPError(c, errcode.HTTPErrCodeBadRequest, err.Error())
		return
	}
	items, err := h.AnalyticsSvc.GetNamespaceTrends(ctx, namespaceID, days)
	if err != nil {
		errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError, err.Error())
		return
	}
	c.JSON(http.StatusOK, api.ListNamespaceHourlyMetricResponse{Items: items})
}
