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

// GetBranch get branch by name
//
//	@Summary	Get specific name code repository branch
//	@security	BasicAuth
//	@Tags		CodeRepository
//	@Accept		json
//	@Produce	json
//	@Router		/{provider}/repos/coderepos/{id}/branches/{name} [get]
//	@Param		provider	path		string	true	"code repository provider"
//	@Param		id			path		number	true	"Code repository id"
//	@Param		name		path		string	true	"Branch name"
//	@Success	200			{object}	types.CodeRepositoryBranchItem
//	@Failure	500			{object}	xerrors.ErrCode
func (h *handler) GetBranch(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	var req types.GetCodeRepositoryBranchRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, fmt.Sprintf("Bind and validate request body failed: %v", err))
	}

	codeRepositoryService := h.CodeRepositoryServiceFactory.New()
	branchObj, err := codeRepositoryService.GetBranchByName(ctx, req.ID, req.Name)
	if err != nil {
		log.Error().Err(err).Msg("Get branch by id failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("List branches failed: %v", err))
	}
	return c.JSON(http.StatusOK, types.CodeRepositoryBranchItem{
		ID:        branchObj.ID,
		Name:      branchObj.Name,
		CreatedAt: time.Unix(0, int64(time.Millisecond)*branchObj.CreatedAt).UTC().Format(consts.DefaultTimePattern),
		UpdatedAt: time.Unix(0, int64(time.Millisecond)*branchObj.CreatedAt).UTC().Format(consts.DefaultTimePattern),
	})
}
