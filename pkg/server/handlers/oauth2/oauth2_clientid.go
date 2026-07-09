// Copyright 2024 sigma
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

package oauth2

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/server/errcode"
)

// ClientID Get oauth2 provider client id
//
//	@Summary	Get oauth2 provider client id
//	@security	BasicAuth
//	@Tags		OAuth2
//	@Accept		json
//	@Produce	json
//	@Router		/oauth2/{provider}/client_id [get]
//	@Param		provider	path		string	true	"oauth2 provider"
//	@Success	200			{object}	api.Oauth2ClientIDResponse
//	@Failure	500			{object}	errcode.ErrCode
func (h *handler) ClientID(c *gin.Context, req *api.Oauth2ClientIDRequest) {
	switch req.Provider {
	case enums.ProviderGithub:
		c.JSON(http.StatusOK, api.Oauth2ClientIDResponse{
			ClientID: h.Config.Auth.Oauth2.Github.ClientID,
		})
		return
	case enums.ProviderGitlab:
		c.JSON(http.StatusOK, api.Oauth2ClientIDResponse{
			ClientID: h.Config.Auth.Oauth2.Gitlab.ClientID,
		})
		return
	case enums.ProviderGitea:
		c.JSON(http.StatusOK, api.Oauth2ClientIDResponse{
			ClientID: h.Config.Auth.Oauth2.Gitea.ClientID,
		})
		return
	default:
		errcode.NewHTTPError(c, errcode.HTTPErrCodeBadRequest, fmt.Sprintf("invalid provider %s", req.Provider))
		return
	}
}
