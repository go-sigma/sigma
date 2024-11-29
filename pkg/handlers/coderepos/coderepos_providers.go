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

package coderepos

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"

	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/xerrors"
)

// Providers list providers
//
//	@Summary	List code repository providers
//	@security	BasicAuth
//	@Tags		CodeRepository
//	@Accept		json
//	@Produce	json
//	@Router		/coderepos/providers [get]
//	@Success	200	{object}	types.CommonList{items=[]types.ListCodeRepositoryProvidersResponse}
//	@Failure	401	{object}	xerrors.ErrCode
//	@Failure	500	{object}	xerrors.ErrCode
func (h *handler) Providers(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	iuser := c.Get(consts.ContextUser)
	if iuser == nil {
		log.Error().Msg("Get user from header failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeUnauthorized)
	}
	user, ok := iuser.(*models.User)
	if !ok {
		log.Error().Msg("Convert user from header failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeUnauthorized)
	}

	userService := h.UserServiceFactory.New()
	user3rdPartyObjs, err := userService.ListUser3rdParty(ctx, user.ID)
	if err != nil {
		log.Error().Err(err).Msg("List providers failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("List providers failed: %v", err))
	}
	resp := make([]any, 0, len(user3rdPartyObjs))
	for _, user3rdPartyObj := range user3rdPartyObjs {
		resp = append(resp, types.ListCodeRepositoryProvidersResponse{
			Provider: user3rdPartyObj.Provider,
		})
	}
	return c.JSON(http.StatusOK, types.CommonList{Total: int64(len(user3rdPartyObjs)), Items: resp})
}
