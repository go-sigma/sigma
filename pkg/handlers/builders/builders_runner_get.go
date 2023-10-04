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

// GetRunner handles the get builder runner request
//
//	@Summary	Get builder runner by runner id
//	@Tags		Builder
//	@security	BasicAuth
//	@Accept		json
//	@Produce	json
//	@Router		/namespaces/{namespace_id}/repositories/{repository_id}/builders/{builder_id}/runners/{runner_id} [get]
//	@Param		namespace_id	path		string	true	"Namespace ID"
//	@Param		repository_id	path		string	true	"Repository ID"
//	@Param		builder_id		path		string	true	"Builder ID"
//	@Param		runner_id		path		string	true	"Runner ID"
//	@Success	200				{object}	types.BuilderItem
//	@Failure	404				{object}	xerrors.ErrCode
//	@Failure	500				{object}	xerrors.ErrCode
func (h *handlers) GetRunner(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	var req types.GetRunner
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, fmt.Sprintf("Bind and validate request body failed: %v", err))
	}

	builderService := h.builderServiceFactory.New()
	runnerObj, err := builderService.GetRunner(ctx, req.RunnerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Int64("id", req.RepositoryID).Msg("Get builder by repository id not found")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Get builder by repository id not found: %v", err))
		}
		log.Error().Err(err).Int64("id", req.RepositoryID).Msg("Get builder by repository id failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Get builder by repository id failed: %v", err))
	}
	if runnerObj.BuilderID != req.BuilderID {
		log.Error().Int64("builder_id", runnerObj.BuilderID).Int64("builder_id", req.BuilderID).Msg("Get builder by id failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, "Get builder by id failed")
	}

	var startedAt, endedAt string
	if runnerObj.StartedAt != nil {
		startedAt = runnerObj.StartedAt.Format(consts.DefaultTimePattern)
	}
	if runnerObj.EndedAt != nil {
		endedAt = runnerObj.EndedAt.Format(consts.DefaultTimePattern)
	}

	var duration *string
	if runnerObj.Duration != nil {
		duration = ptr.Of(durafmt.ParseShort(time.Millisecond * time.Duration(ptr.To(runnerObj.Duration))).String())
	}

	return c.JSON(http.StatusOK, types.BuilderRunnerItem{
		ID:        runnerObj.ID,
		BuilderID: runnerObj.BuilderID,

		Log: runnerObj.Log,

		Status:      runnerObj.Status,
		Tag:         runnerObj.Tag,
		RawTag:      runnerObj.RawTag,
		Description: runnerObj.Description,
		ScmBranch:   runnerObj.ScmBranch,

		StartedAt:   ptr.Of(startedAt),
		EndedAt:     ptr.Of(endedAt),
		RawDuration: runnerObj.Duration,
		Duration:    duration,

		CreatedAt: runnerObj.CreatedAt.Format(consts.DefaultTimePattern),
		UpdatedAt: runnerObj.UpdatedAt.Format(consts.DefaultTimePattern),
	})
}
