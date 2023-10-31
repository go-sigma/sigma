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

package daemons

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/modules/workq"
	"github.com/go-sigma/sigma/pkg/modules/workq/definition"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/xerrors"
)

// GcArtifactRun ...
func (h *handlers) GcArtifactRun(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	var req types.GcArtifactRunRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, fmt.Sprintf("Bind and validate request body failed: %v", err))
	}
	daemonService := h.daemonServiceFactory.New()
	lastRunnerObj, err := daemonService.GetLastGcRepositoryRunner(ctx, req.NamespaceID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Error().Err(err).Msg("Get Last runner failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Get last runner failed: %v", err))
	}
	if lastRunnerObj.Status == enums.TaskCommonStatusDoing || lastRunnerObj.Status == enums.TaskCommonStatusPending {
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, fmt.Sprintf("Last runner is still running: %v", lastRunnerObj.Status))
	}
	err = query.Q.Transaction(func(tx *query.Query) error {
		daemonService := h.daemonServiceFactory.New(tx)
		runnerObj := models.DaemonGcRepositoryRunner{
			Status:      enums.TaskCommonStatusPending,
			NamespaceID: req.NamespaceID,
		}
		err = daemonService.CreateGcRepositoryRunner(ctx, &runnerObj)
		if err != nil {
			log.Error().Err(err).Msg("Create gc repository runner failed")
			return xerrors.DSErrCodeUnknown.Detail(fmt.Sprintf("Create gc repository runner failed: %v", err))
		}
		err = workq.ProducerClient.Produce(ctx, enums.DaemonGcRepository.String(), "nil", definition.ProducerOption{Tx: tx})
		if err != nil {
			log.Error().Err(err).Msg("Publish topic failed")
			return xerrors.DSErrCodeUnknown.Detail(fmt.Sprintf("Publish the topic failed: %v", err))
		}
		return nil
	})
	if err != nil {
		e, ok := err.(xerrors.ErrCode)
		if ok {
			return xerrors.NewDSError(c, e)
		}
		return xerrors.NewHTTPError(c, xerrors.DSErrCodeUnknown)
	}
	return c.NoContent(http.StatusCreated)
}

// GcRepositoryRunners ...
func (h *handlers) GcRepositoryRunners(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	var req types.GcArtifactRunnersRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, fmt.Sprintf("Bind and validate request body failed: %v", err))
	}

	daemonService := h.daemonServiceFactory.New()
	recordObjs, total, err := daemonService.ListGcRepositoryRunners(ctx, req.NamespaceId, req.Pagination, req.Sortable)
	if err != nil {
		log.Error().Err(err).Msg("List gc artifact failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("List gc artifact failed: %v", err))
	}
	var resp = make([]any, 0, len(recordObjs))
	for _, record := range recordObjs {
		resp = append(resp, types.GcArtifactRunnerItem{
			ID:        record.ID,
			Status:    record.Status,
			Message:   string(record.Message),
			CreatedAt: record.CreatedAt.Format(consts.DefaultTimePattern),
			UpdatedAt: record.UpdatedAt.Format(consts.DefaultTimePattern),
		})
	}

	return c.JSON(http.StatusOK, types.CommonList{Total: total, Items: resp})
}

// GcArtifactGet get the last runner
func (h *handlers) GcArtifactGet(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	var req types.GcArtifactRunRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, fmt.Sprintf("Bind and validate request body failed: %v", err))
	}

	daemonService := h.daemonServiceFactory.New()
	lastRunnerObj, err := daemonService.GetLastGcRepositoryRunner(ctx, req.NamespaceID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Error().Err(err).Msg("Get Last runner failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Get last runner failed: %v", err))
	}
	return c.JSON(http.StatusOK, types.GcArtifactGetResponse{
		ID:        lastRunnerObj.ID,
		Status:    lastRunnerObj.Status,
		Message:   string(lastRunnerObj.Message),
		CreatedAt: lastRunnerObj.CreatedAt.Format(consts.DefaultTimePattern),
		UpdatedAt: lastRunnerObj.UpdatedAt.Format(consts.DefaultTimePattern),
	})
}

// GcArtifactRecords ...
func (h *handlers) GcArtifactRecords(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	var req types.GcArtifactRecordsRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, fmt.Sprintf("Bind and validate request body failed: %v", err))
	}

	daemonService := h.daemonServiceFactory.New()
	recordObjs, total, err := daemonService.ListGcArtifactRecords(ctx, req.RunnerID, req.Pagination, req.Sortable)
	if err != nil {
		log.Error().Err(err).Msg("List gc artifact failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("List gc artifact failed: %v", err))
	}
	var resp = make([]any, 0, len(recordObjs))
	for _, record := range recordObjs {
		resp = append(resp, types.GcArtifactRecordItem{
			ID:        record.ID,
			Digest:    record.Digest,
			CreatedAt: record.CreatedAt.Format(consts.DefaultTimePattern),
			UpdatedAt: record.UpdatedAt.Format(consts.DefaultTimePattern),
		})
	}

	return c.JSON(http.StatusOK, types.CommonList{Total: total, Items: resp})
}

// GcArtifactRecord ...
func (h *handlers) GcArtifactRecord(c echo.Context) error {
	return nil
}
