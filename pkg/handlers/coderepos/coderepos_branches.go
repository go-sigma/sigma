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
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/xerrors"
)

// ListBranches list all of the branches
//
//	@Summary	List code repository branches
//	@security	BasicAuth
//	@Tags		CodeRepository
//	@Accept		json
//	@Produce	json
//	@Router		/{provider}/repos/coderepos/{id}/branches [get]
//	@Param		provider	path		string	true	"code repository provider"
//	@Param		id			path		string	true	"code repository id"
//	@Success	200			{object}	types.CommonList{items=[]types.CodeRepositoryBranchItem}
//	@Failure	500			{object}	xerrors.ErrCode
func (h *handler) ListBranches(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	var req types.ListCodeRepositoryBranchesRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, err.Error())
	}

	codeRepositoryService := h.CodeRepositoryServiceFactory.New()
	branchObjs, total, err := codeRepositoryService.ListBranchesWithoutPagination(ctx, req.ID)
	if err != nil {
		log.Error().Err(err).Msg("List branches failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("List branches failed: %v", err))
	}
	resp := make([]any, 0, len(branchObjs))
	for _, branchObj := range branchObjs {
		resp = append(resp, types.CodeRepositoryBranchItem{
			ID:        branchObj.ID,
			Name:      branchObj.Name,
			CreatedAt: time.Unix(0, int64(time.Millisecond)*branchObj.CreatedAt).UTC().Format(consts.DefaultTimePattern),
			UpdatedAt: time.Unix(0, int64(time.Millisecond)*branchObj.CreatedAt).UTC().Format(consts.DefaultTimePattern),
		})
	}

	return c.JSON(http.StatusOK, types.CommonList{Total: total, Items: resp})
}
