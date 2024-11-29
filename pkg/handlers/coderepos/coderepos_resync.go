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
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/modules/workq"
	"github.com/go-sigma/sigma/pkg/modules/workq/definition"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/xerrors"
)

// Resync resync all of the code repositories
//
//	@Summary	Resync code repository
//	@security	BasicAuth
//	@Tags		CodeRepository
//	@Accept		json
//	@Produce	json
//	@Router		/coderepos/{provider}/resync [get]
//	@Param		provider	path	string	true	"Search code repository with scm provider"
//	@Success	202
//	@Failure	500	{object}	xerrors.ErrCode
func (h *handler) Resync(c echo.Context) error {
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

	var req types.GetCodeRepositoryResyncRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, err.Error())
	}

	userService := h.UserServiceFactory.New()
	user3rdPartyObj, err := userService.GetUser3rdPartyByProvider(ctx, user.ID, req.Provider)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Int64("userID", user.ID).Str("provider", req.Provider.String()).Msg("Code repository not found")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, "Code repository not found")
		}
		log.Error().Err(err).Int64("userID", user.ID).Str("provider", req.Provider.String()).Msg("Code repository find failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, "Code repository find failed")
	}
	if user3rdPartyObj.CrLastUpdateStatus == enums.TaskCommonStatusDoing {
		log.Error().Str("provider", req.Provider.String()).Msg("Code repository status already is syncing")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeConflict, fmt.Sprintf("Code repository(%s) status already is syncing", req.Provider.String()))
	}
	err = query.Q.Transaction(func(tx *query.Query) error {
		userService := h.UserServiceFactory.New(tx)
		err = userService.UpdateUser3rdParty(ctx, user3rdPartyObj.ID, map[string]any{
			query.User3rdParty.CrLastUpdateTimestamp.ColumnName().String(): time.Now().UnixMilli(),
			query.User3rdParty.CrLastUpdateStatus.ColumnName().String():    enums.TaskCommonStatusDoing,
			query.User3rdParty.CrLastUpdateMessage.ColumnName().String():   "",
		})
		if err != nil {
			return xerrors.HTTPErrCodeInternalError.Detail("Update user status failed")
		}
		err = workq.ProducerClient.Produce(ctx, enums.DaemonCodeRepository,
			types.DaemonCodeRepositoryPayload{User3rdPartyID: user3rdPartyObj.ID}, definition.ProducerOption{Tx: tx})
		if err != nil {
			log.Error().Err(err).Int64("user_id", user3rdPartyObj.UserID).Msg("Publish sync code repository failed")
		}
		return nil
	})
	if err != nil {
		return xerrors.NewHTTPError(c, err.(xerrors.ErrCode))
	}
	return c.NoContent(http.StatusAccepted)
}
