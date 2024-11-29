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
	"time"

	"github.com/hako/durafmt"
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
//
//	@Summary	Get builder runners by builder id
//	@Tags		Builder
//	@security	BasicAuth
//	@Accept		json
//	@Produce	json
//	@Router		/namespaces/{namespace_id}/repositories/{repository_id}/builders/{builder_id}/runners/ [get]
//	@Param		namespace_id	path		string	true	"Namespace ID"
//	@Param		repository_id	path		string	true	"Repository ID"
//	@Param		builder_id		path		string	true	"Builder ID"
//	@Success	200				{object}	types.BuilderItem
//	@Failure	404				{object}	xerrors.ErrCode
//	@Failure	500				{object}	xerrors.ErrCode
func (h *handler) ListRunners(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	var req types.ListBuilderRunnersRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, fmt.Sprintf("Bind and validate request body failed: %v", err))
	}
	req.Pagination = utils.NormalizePagination(req.Pagination)

	builderService := h.BuilderServiceFactory.New()
	_, err = builderService.GetByRepositoryID(ctx, req.RepositoryID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Error().Err(err).Int64("id", req.RepositoryID).Msg("Get builder by repository id failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Get builder by repository id failed: %v", err))
	}
	runnerObjs, total, err := builderService.ListRunners(ctx, req.BuilderID, req.Pagination, req.Sortable)
	if err != nil {
		log.Error().Err(err).Msg("List builder runners failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("List builder runners failed: %v", err))
	}
	var resp = make([]any, 0, len(runnerObjs))
	for _, runnerObj := range runnerObjs {
		var duration *string
		if runnerObj.Duration != nil {
			duration = ptr.Of(durafmt.ParseShort(time.Millisecond * time.Duration(ptr.To(runnerObj.Duration))).String())
		}

		resp = append(resp, types.BuilderRunnerItem{
			ID:        runnerObj.ID,
			BuilderID: runnerObj.BuilderID,

			Log: runnerObj.Log,

			Status:        runnerObj.Status,
			StatusMessage: runnerObj.StatusMessage,
			Tag:           runnerObj.Tag,
			RawTag:        runnerObj.RawTag,
			Description:   runnerObj.Description,
			ScmBranch:     runnerObj.ScmBranch,

			StartedAt:   runnerObj.StartedAt,
			EndedAt:     runnerObj.EndedAt,
			RawDuration: runnerObj.Duration,
			Duration:    duration,

			CreatedAt: time.Unix(0, int64(time.Millisecond)*runnerObj.CreatedAt).UTC().Format(consts.DefaultTimePattern),
			UpdatedAt: time.Unix(0, int64(time.Millisecond)*runnerObj.CreatedAt).UTC().Format(consts.DefaultTimePattern),
		})
	}
	return c.JSON(http.StatusOK, types.CommonList{Total: total, Items: resp})
}
