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
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"

	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/xerrors"
)

// ListOwner list all of the code repository owner
//
//	@Summary	List code repository owners
//	@security	BasicAuth
//	@Tags		CodeRepository
//	@Accept		json
//	@Produce	json
//	@Router		/coderepos/{provider}/owners [get]
//	@Param		provider	path		string	true	"search code repository with provider"
//	@Param		name		query		string	false	"search code repository with name"
//	@Success	200			{object}	types.CommonList{items=[]types.CodeRepositoryOwnerItem}
//	@Failure	500			{object}	xerrors.ErrCode
func (h *handler) ListOwners(c echo.Context) error {
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

	var req types.ListCodeRepositoryOwnerRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, err.Error())
	}

	codeRepositoryService := h.CodeRepositoryServiceFactory.New()
	codeRepositoryOwnerObjs, total, err := codeRepositoryService.ListOwnerWithoutPagination(ctx, user.ID, req.Provider, req.Name)
	if err != nil {
		log.Error().Err(err).Msg("List code repository owners failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, err.Error())
	}
	resp := make([]any, 0, len(codeRepositoryOwnerObjs))
	for _, codeRepositoryOwnerObj := range codeRepositoryOwnerObjs {
		resp = append(resp, types.CodeRepositoryOwnerItem{
			ID:        codeRepositoryOwnerObj.ID,
			OwnerID:   codeRepositoryOwnerObj.OwnerID,
			Owner:     codeRepositoryOwnerObj.Owner,
			IsOrg:     codeRepositoryOwnerObj.IsOrg,
			CreatedAt: time.Unix(0, int64(time.Millisecond)*codeRepositoryOwnerObj.CreatedAt).UTC().Format(consts.DefaultTimePattern),
			UpdatedAt: time.Unix(0, int64(time.Millisecond)*codeRepositoryOwnerObj.CreatedAt).UTC().Format(consts.DefaultTimePattern),
		})
	}

	return c.JSON(http.StatusOK, types.CommonList{Total: total, Items: resp})
}
