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

	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/modules/workq"
	"github.com/go-sigma/sigma/pkg/modules/workq/definition"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/xerrors"
)

// PostRunnerRun ...
func (h *handler) PostRunnerRun(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	var req types.PostRunnerRun
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, fmt.Sprintf("Bind and validate request body failed: %v", err))
	}

	builderService := h.BuilderServiceFactory.New()
	builderObj, err := builderService.GetByRepositoryID(ctx, req.RepositoryID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Int64("id", req.RepositoryID).Msg("Get builder by repository id not found")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Get builder by repository id not found: %v", err))
		}
		log.Error().Err(err).Int64("id", req.RepositoryID).Msg("Get builder by repository id failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Get builder by repository id failed: %v", err))
	}
	if builderObj.ID != req.BuilderID {
		log.Error().Int64("builder_id", req.BuilderID).Int64("builder_id", builderObj.ID).Msg("Get builder by id failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, "Get builder by id failed")
	}

	var runnerObj *models.BuilderRunner
	err = query.Q.Transaction(func(tx *query.Query) error {
		builderService := h.BuilderServiceFactory.New(tx)
		runnerObj = &models.BuilderRunner{
			BuilderID: req.BuilderID,
			RawTag:    req.RawTag,
			ScmBranch: req.ScmBranch,
		}
		err = builderService.CreateRunner(ctx, runnerObj)
		if err != nil {
			log.Error().Err(err).Msg("Create builder runner failed")
			return xerrors.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Create builder runner failed: %v", err))
		}
		err = workq.ProducerClient.Produce(ctx, enums.DaemonBuilder, types.DaemonBuilderPayload{
			Action:       enums.DaemonBuilderActionStart,
			RepositoryID: req.RepositoryID,
			BuilderID:    req.BuilderID,
			RunnerID:     runnerObj.ID,
		}, definition.ProducerOption{Tx: tx})
		if err != nil {
			log.Error().Err(err).Msgf("Send topic %s to work queue failed", enums.DaemonBuilder.String())
			return xerrors.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Send topic %s to work queue failed", enums.DaemonBuilder.String()))
		}
		return nil
	})
	if err != nil {
		return xerrors.NewHTTPError(c, err.(xerrors.ErrCode))
	}
	return c.JSON(http.StatusCreated, types.RunOrRerunRunnerResponse{
		RunnerID: runnerObj.ID,
	})
}
