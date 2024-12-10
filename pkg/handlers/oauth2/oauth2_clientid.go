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

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"

	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/xerrors"
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
//	@Success	200			{object}	types.Oauth2ClientIDResponse
//	@Failure	500			{object}	xerrors.ErrCode
func (h *handler) ClientID(c echo.Context) error {
	var req types.Oauth2ClientIDRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, err.Error())
	}
	switch req.Provider {
	case enums.ProviderGithub:
		return c.JSON(http.StatusOK, types.Oauth2ClientIDResponse{
			ClientID: h.Config.Auth.Oauth2.Github.ClientID,
		})
	case enums.ProviderGitlab:
		return c.JSON(http.StatusOK, types.Oauth2ClientIDResponse{
			ClientID: h.Config.Auth.Oauth2.Gitlab.ClientID,
		})
	case enums.ProviderGitea:
		return c.JSON(http.StatusOK, types.Oauth2ClientIDResponse{
			ClientID: h.Config.Auth.Oauth2.Gitea.ClientID,
		})
	default:
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, fmt.Sprintf("invalid provider %s", req.Provider))
	}
}
