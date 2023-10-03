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

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"

	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/types"
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
//	@Param		limit		query		int64	false	"limit"	minimum(10)	maximum(100)	default(10)
//	@Param		page		query		int64	false	"page"	minimum(1)	default(1)
//	@Param		sort		query		string	false	"sort field"
//	@Param		method		query		string	false	"sort method"	Enums(asc, desc)
//	@Param		name		query		string	false	"search code repository with name"
//	@Param		owner		query		string	false	"search code repository with owner"
//	@Param		provider	path		string	true	"search code repository with provider"
//	@Success	200			{object}	types.CommonList{items=[]types.CodeRepositoryItem}
//	@Failure	500			{object}	xerrors.ErrCode
func (h *handlers) List(c echo.Context) error {
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

	codeRepositoryService := h.codeRepositoryServiceFactory.New()
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
			Name:         codeRepositoryObj.Name,
			OwnerID:      codeRepositoryObj.OwnerID,
			Owner:        codeRepositoryObj.Owner,
			IsOrg:        codeRepositoryObj.IsOrg,
			CloneUrl:     codeRepositoryObj.CloneUrl,
			SshUrl:       codeRepositoryObj.SshUrl,
			OciRepoCount: codeRepositoryObj.OciRepoCount,
			CreatedAt:    codeRepositoryObj.CreatedAt.Format(consts.DefaultTimePattern),
			UpdatedAt:    codeRepositoryObj.UpdatedAt.Format(consts.DefaultTimePattern),
		})
	}

	return c.JSON(http.StatusOK, types.CommonList{Total: total, Items: resp})
}
