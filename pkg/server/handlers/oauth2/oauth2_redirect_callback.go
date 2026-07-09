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
)

// RedirectCallback Redirect oauth2 provider callback
//
//	@Summary	Redirect oauth2 provider callback
//	@security	BasicAuth
//	@Tags		OAuth2
//	@Accept		json
//	@Produce	json
//	@Router		/oauth2/{provider}/redirect_callback [get]
//	@Param		provider	path	string	true	"oauth2 provider"
//	@Success	301
//	@Failure	500	{object}	errcode.ErrCode
func (h *handler) RedirectCallback(c *gin.Context, req *api.Oauth2CallbackRequest) {
	c.Redirect(http.StatusMovedPermanently, fmt.Sprintf("%s/#/login/callback/%s?code=%s", req.Endpoint, req.Provider, req.Code))
}
