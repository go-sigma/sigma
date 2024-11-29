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
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/xerrors"
)

// Get get code repository by id
//
//	@Summary	Get code repository by id
//	@security	BasicAuth
//	@Tags		CodeRepository
//	@Accept		json
//	@Produce	json
//	@Router		/coderepos/{provider}/repos/{id} [get]
//	@Param		provider	path		string	true	"Search code repository with provider"
//	@Param		id			path		string	true	"Code repository id"
//	@Success	200			{object}	types.CodeRepositoryItem
//	@Failure	500			{object}	xerrors.ErrCode
func (h *handler) Get(c echo.Context) error {
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

	var req types.GetCodeRepositoryRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, fmt.Sprintf("Bind and validate request body failed: %v", err))
	}

	codeRepositoryService := h.CodeRepositoryServiceFactory.New()
	userService := h.UserServiceFactory.New()
	user3rdPartyObj, err := userService.GetUser3rdPartyByProvider(ctx, user.ID, req.Provider)
	if err != nil {
		log.Error().Err(err).Str("Provider", req.Provider.String()).Msg("Get user 3rdParty by provider failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Get user 3rdParty by provider failed: %v", err))
	}

	ownerObjs, err := codeRepositoryService.ListOwnersAll(ctx, user3rdPartyObj.ID)
	if err != nil {
		log.Error().Err(err).Msg("List all owners failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("List all owners failed: %v", err))
	}

	codeRepositoryObj, err := codeRepositoryService.Get(ctx, req.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Str("provider", req.Provider.String()).Int64("id", req.ID).Msg("Code repository not found")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, fmt.Sprintf("Code repository(%d) not found: %s", req.ID, err))
		}
		log.Error().Err(err).Int64("repositoryID", req.ID).Int64("id", req.ID).Msg("Get code repository failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Code repository(%d) not found: %s", req.ID, err))
	}
	if codeRepositoryObj.User3rdParty.Provider != req.Provider {
		log.Error().Err(err).Str("provider", req.Provider.String()).Int64("id", req.ID).Msg("Code repository not found")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, fmt.Sprintf("Code repository(%d) not found", req.ID))
	}

	return c.JSON(http.StatusOK, types.CodeRepositoryItem{
		ID:           codeRepositoryObj.ID,
		RepositoryID: codeRepositoryObj.RepositoryID,
		Provider:     enums.ScmProvider(user3rdPartyObj.Provider),
		Name:         codeRepositoryObj.Name,
		OwnerID:      h.getOwnerID(ownerObjs, codeRepositoryObj.Owner),
		Owner:        codeRepositoryObj.Owner,
		IsOrg:        codeRepositoryObj.IsOrg,
		CloneUrl:     codeRepositoryObj.CloneUrl,
		SshUrl:       codeRepositoryObj.SshUrl,
		OciRepoCount: codeRepositoryObj.OciRepoCount,
		CreatedAt:    time.Unix(0, int64(time.Millisecond)*codeRepositoryObj.CreatedAt).UTC().Format(consts.DefaultTimePattern),
		UpdatedAt:    time.Unix(0, int64(time.Millisecond)*codeRepositoryObj.CreatedAt).UTC().Format(consts.DefaultTimePattern),
	})
}
