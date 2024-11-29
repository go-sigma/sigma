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
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"

	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/xerrors"
)

// List list all of the code repositories
//
//	@Summary	List code repositories
//	@security	BasicAuth
//	@Tags		CodeRepository
//	@Accept		json
//	@Produce	json
//	@Router		/coderepos/{provider} [get]
//	@Param		limit		query		int64	false	"Limit size"	minimum(10)	maximum(100)	default(10)
//	@Param		page		query		int64	false	"Page number"	minimum(1)	default(1)
//	@Param		sort		query		string	false	"Sort field"
//	@Param		method		query		string	false	"Sort method"	Enums(asc, desc)
//	@Param		name		query		string	false	"Search code repository with name"
//	@Param		owner		query		string	false	"Search code repository with owner"
//	@Param		provider	path		string	true	"search code repository with provider"
//	@Success	200			{object}	types.CommonList{items=[]types.CodeRepositoryItem}
//	@Failure	500			{object}	xerrors.ErrCode
func (h *handler) List(c echo.Context) error {
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

	var req types.ListCodeRepositoryRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, err.Error())
	}
	req.Pagination = utils.NormalizePagination(req.Pagination)

	userService := h.UserServiceFactory.New()
	user3rdPartyObj, err := userService.GetUser3rdPartyByProvider(ctx, user.ID, req.Provider)
	if err != nil {
		log.Error().Err(err).Str("Provider", req.Provider.String()).Msg("Get user 3rdParty by provider failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Get user 3rdParty by provider failed: %v", err))
	}

	codeRepositoryService := h.CodeRepositoryServiceFactory.New()
	ownerObjs, err := codeRepositoryService.ListOwnersAll(ctx, user3rdPartyObj.ID)
	if err != nil {
		log.Error().Err(err).Msg("List all owners failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("List all owners failed: %v", err))
	}

	codeRepositoryObjs, total, err := codeRepositoryService.ListWithPagination(ctx, user.ID, req.Provider, req.Owner, req.Name, req.Pagination, req.Sortable)
	if err != nil {
		log.Error().Err(err).Msg("List code repositories failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, err.Error())
	}
	resp := make([]any, 0, len(codeRepositoryObjs))
	for _, codeRepositoryObj := range codeRepositoryObjs {
		resp = append(resp, types.CodeRepositoryItem{
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

	return c.JSON(http.StatusOK, types.CommonList{Total: total, Items: resp})
}

func (h *handler) getOwnerID(ownerObjs []*models.CodeRepositoryOwner, owner string) int64 {
	for _, ownerObj := range ownerObjs {
		if ownerObj.Owner == owner {
			return ownerObj.ID
		}
	}
	return 0
}
