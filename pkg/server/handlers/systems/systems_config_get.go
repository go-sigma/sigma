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

package systems

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/server/errcode"
)

// GetConfig handles the get config request
//
//	@Summary	Get config
//	@Tags		System
//	@Accept		json
//	@Produce	json
//	@Router		/systems/config [get]
//	@Success	200	{object}	api.GetSystemConfigResponse
func (h *handler) GetConfig(c *gin.Context) {
	ctx := c.Request.Context()
	config, err := h.SystemsSvc.GetConfig(ctx)
	if err != nil {
		if e, ok := errcode.AsType[errcode.ErrCode](err); ok {
			errcode.NewHTTPError(c, e)
			return
		}
		errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError)
		return
	}
	c.JSON(http.StatusOK, api.GetSystemConfigResponse{
		Daemon: api.GetSystemConfigDaemon{
			Builder: config.Daemon.Builder.Enabled,
		},
		Anonymous: config.Auth.Anonymous.Enabled,
		OAuth2: api.GetSystemConfigOAuth2{
			GitHub: config.Auth.Oauth2.Github.Enabled,
			GitLab: config.Auth.Oauth2.Gitlab.Enabled,
		},
	})
}
