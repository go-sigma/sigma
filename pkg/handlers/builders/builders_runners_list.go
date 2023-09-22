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

package builders

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
	"github.com/go-sigma/sigma/pkg/xerrors"
)

// ListRunners handles the list builder runners request
func (h *handlers) ListRunners(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	var req types.ListBuilderRunnersRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, fmt.Sprintf("Bind and validate request body failed: %v", err))
	}
	req.Pagination = utils.NormalizePagination(req.Pagination)

	builderService := h.builderServiceFactory.New()
	_, err = builderService.GetByRepositoryID(ctx, req.RepositoryID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Error().Err(err).Int64("id", req.RepositoryID).Msg("Get builder by repository id failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Get builder by repository id failed: %v", err))
	}
	runnerObjs, total, err := builderService.ListRunners(ctx, req.ID, req.Pagination, req.Sortable)
	if err != nil {
		log.Error().Err(err).Msg("List builder runners failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("List builder runners failed: %v", err))
	}
	var resp = make([]any, 0, len(runnerObjs))
	for _, runner := range runnerObjs {
		var startedAt, endedAt string
		if runner.StartedAt != nil {
			startedAt = runner.StartedAt.Format(consts.DefaultTimePattern)
		}
		if runner.EndedAt != nil {
			endedAt = runner.EndedAt.Format(consts.DefaultTimePattern)
		}
		resp = append(resp, types.BuilderRunnerItem{
			ID:          runner.ID,
			BuilderID:   runner.BuilderID,
			Log:         runner.Log,
			Status:      runner.Status,
			Tag:         runner.Tag,
			Description: runner.Description,
			ScmBranch:   runner.ScmBranch,
			StartedAt:   ptr.Of(startedAt),
			EndedAt:     ptr.Of(endedAt),
			CreatedAt:   runner.CreatedAt.Format(consts.DefaultTimePattern),
			UpdatedAt:   runner.UpdatedAt.Format(consts.DefaultTimePattern),
		})
	}
	return c.JSON(http.StatusOK, types.CommonList{Total: total, Items: resp})
}
