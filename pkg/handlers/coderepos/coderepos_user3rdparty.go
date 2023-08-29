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
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
	"github.com/go-sigma/sigma/pkg/xerrors"
)

// User3rdParty get user 3rdparty
func (h *handlers) User3rdParty(c echo.Context) error {
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

	var req types.GetCodeRepositoryUser3rdPartyRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, err.Error())
	}

	userService := h.userServiceFactory.New()
	user3rdPartyObj, err := userService.GetUser3rdPartyByProvider(ctx, user.ID, req.Provider)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Int64("userID", user.ID).Str("provider", req.Provider.String()).Msg("Code repository not found")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, "Code repository not found")
		}
		log.Error().Err(err).Int64("userID", user.ID).Str("provider", req.Provider.String()).Msg("Code repository find failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, "Code repository find failed")
	}

	return c.JSON(http.StatusOK, types.GetCodeRepositoryUser3rdPartyResponse{
		ID:                    user3rdPartyObj.ID,
		AccountID:             ptr.To(user3rdPartyObj.AccountID),
		CrLastUpdateTimestamp: user3rdPartyObj.CrLastUpdateTimestamp.Format(consts.DefaultTimePattern),
		CrLastUpdateStatus:    user3rdPartyObj.CrLastUpdateStatus,
		CrLastUpdateMessage:   user3rdPartyObj.CrLastUpdateMessage,

		CreatedAt: user3rdPartyObj.CreatedAt.Format(consts.DefaultTimePattern),
		UpdatedAt: user3rdPartyObj.UpdatedAt.Format(consts.DefaultTimePattern),
	})
}
