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

	"github.com/go-sigma/sigma/pkg/modules/workq"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/xerrors"
)

// GetRunnerStop ...
func (h *handlers) GetRunnerStop(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	var req types.GetRunnerStop
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, fmt.Sprintf("Bind and validate request body failed: %v", err))
	}

	builderService := h.builderServiceFactory.New()
	builderObj, err := builderService.GetByRepositoryID(ctx, req.RepositoryID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Error().Err(err).Int64("id", req.RepositoryID).Msg("Get builder by repository id failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Get builder by repository id failed: %v", err))
	}
	if builderObj.ID != req.BuilderID {
		log.Error().Int64("builder_id", req.BuilderID).Int64("builder_id", builderObj.ID).Msg("Get builder by id failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, "Get builder by id failed")
	}

	runnerObj, err := builderService.GetRunner(ctx, req.RunnerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Msgf("Builder runner not found")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, "Builder runner not found")
		}
		log.Error().Err(err).Msgf("Builder runner find failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Builder runner find failed: %v", err))
	}

	if runnerObj.Status != enums.BuildStatusBuilding {
		log.Error().Str("status", runnerObj.Status.String()).Msgf("Builder runner status %s not support stop", runnerObj.Status.String())
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, fmt.Sprintf("Builder runner status %s not support stop", runnerObj.Status.String()))
	}

	err = workq.ProducerClient.Produce(ctx, enums.DaemonBuilder.String(), types.DaemonBuilderPayload{
		Action:       enums.DaemonBuilderActionStop,
		RepositoryID: req.RepositoryID,
		BuilderID:    req.BuilderID,
		RunnerID:     req.RunnerID,
	})
	if err != nil {
		log.Error().Err(err).Msgf("Send topic %s to work queue failed", enums.DaemonBuilder.String())
		return xerrors.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Send topic %s to work queue failed", enums.DaemonBuilder.String()))
	}

	return c.NoContent(http.StatusNoContent)
}
