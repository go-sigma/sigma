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
	"time"

	"github.com/labstack/echo/v4"
	"github.com/robfig/cron/v3"
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
	"github.com/go-sigma/sigma/pkg/utils/ptr"
	"github.com/go-sigma/sigma/pkg/xerrors"
)

// UpdateGcBlobRule ...
func (h *handlers) UpdateGcBlobRule(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	var req types.UpdateGcBlobRuleRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, fmt.Sprintf("Bind and validate request body failed: %v", err))
	}

	daemonService := h.daemonServiceFactory.New()
	ruleObj, err := daemonService.GetGcBlobRule(ctx)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Error().Err(err).Msg("Get gc artifact rule failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Get gc artifact rule failed: %v", err))
	}
	if ruleObj != nil && ruleObj.IsRunning {
		log.Error().Msg("The gc blob rule is running")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, "The gc artifact rule is running")
	}
	var nextTrigger *time.Time
	if req.CronRule != nil {
		schedule, _ := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow).Parse(ptr.To(req.CronRule))
		nextTrigger = ptr.Of(schedule.Next(time.Now()))
	}
	updates := make(map[string]any, 5)
	if req.CronEnabled {
		if req.CronRule != nil {
			updates[query.DaemonGcBlobRule.CronRule.ColumnName().String()] = ptr.To(req.CronRule)
			updates[query.DaemonGcBlobRule.CronNextTrigger.ColumnName().String()] = ptr.To(nextTrigger)
		}
	}
	err = query.Q.Transaction(func(tx *query.Query) error {
		if ruleObj == nil { // rule not found, we need create the rule
			err = daemonService.CreateGcBlobRule(ctx, &models.DaemonGcBlobRule{
				CronEnabled:     req.CronEnabled,
				CronRule:        req.CronRule,
				CronNextTrigger: nextTrigger,
			})
			if err != nil {
				log.Error().Err(err).Msg("Create gc blob rule failed")
				return xerrors.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Create gc blob rule failed: %v", err))
			}
		}
		err = daemonService.UpdateGcBlobRule(ctx, ruleObj.ID, updates)
		if err != nil {
			log.Error().Err(err).Msg("Update gc blob rule failed")
			return xerrors.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Update gc blob rule failed: %v", err))
		}
		return nil
	})
	if err != nil {
		var e xerrors.ErrCode
		if errors.As(err, &e) {
			return xerrors.NewHTTPError(c, e)
		}
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError)
	}
	return c.NoContent(http.StatusNoContent)
}

// GetGcBlobRule ...
func (h *handlers) GetGcBlobRule(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	var req types.GetGcBlobRuleRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, fmt.Sprintf("Bind and validate request body failed: %v", err))
	}

	daemonService := h.daemonServiceFactory.New()
	ruleObj, err := daemonService.GetGcBlobRule(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Int64("namespaceID", req.NamespaceID).Msg("Get gc blob rule not found")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, fmt.Sprintf("Get gc blob rule not found: %v", err))
		}
		log.Error().Err(err).Msg("Get gc blob rule failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Get gc blob rule failed: %v", err))
	}
	return c.JSON(http.StatusOK, types.GetGcBlobRuleResponse{
		CronEnabled:     ruleObj.CronEnabled,
		CronRule:        ruleObj.CronRule,
		CronNextTrigger: ptr.Of(""),
		CreatedAt:       ruleObj.CreatedAt.Format(consts.DefaultTimePattern),
		UpdatedAt:       ruleObj.UpdatedAt.Format(consts.DefaultTimePattern),
	})
}

// GetGcBlobLatestRunner ...
func (h *handlers) GetGcBlobLatestRunner(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	var req types.GetGcBlobLatestRunnerRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, fmt.Sprintf("Bind and validate request body failed: %v", err))
	}
	// TODO: check namespaceID is 0
	daemonService := h.daemonServiceFactory.New()
	ruleObj, err := daemonService.GetGcBlobRule(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Int64("namespaceID", req.NamespaceID).Msg("Get gc blob rule not found")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, fmt.Sprintf("Get gc blob rule not found: %v", err))
		}
		log.Error().Err(err).Msg("Get gc blob rule failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Get gc blob rule failed: %v", err))
	}
	runnerObj, err := daemonService.GetGcTagLatestRunner(ctx, ruleObj.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Msg("Get gc blob latest runner not found")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, fmt.Sprintf("Get gc blob latest runner not found: %v", err))
		}
		log.Error().Err(err).Msg("Get gc blob latest runner failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Get gc blob latest runner failed: %v", err))
	}
	return c.JSON(http.StatusOK, types.GcBlobRunnerItem{
		ID:        runnerObj.ID,
		Status:    runnerObj.Status,
		Message:   string(runnerObj.Message),
		CreatedAt: ruleObj.CreatedAt.Format(consts.DefaultTimePattern),
		UpdatedAt: ruleObj.UpdatedAt.Format(consts.DefaultTimePattern),
	})
}

// CreateGcBlobRunner ...
func (h *handlers) CreateGcBlobRunner(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	var req types.CreateGcBlobRunnerRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, fmt.Sprintf("Bind and validate request body failed: %v", err))
	}
	// TODO: check namespace id
	daemonService := h.daemonServiceFactory.New()
	ruleObj, err := daemonService.GetGcBlobRule(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Int64("namespaceID", req.NamespaceID).Msg("Get gc blob rule not found")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, fmt.Sprintf("Get gc blob rule not found: %v", err))
		}
		log.Error().Err(err).Msg("Get gc blob rule failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Get gc blob rule failed: %v", err))
	}
	if ruleObj != nil && ruleObj.IsRunning {
		log.Error().Int64("NamespaceID", req.NamespaceID).Msg("The gc blob rule is running")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, "The gc blob rule is running")
	}
	err = query.Q.Transaction(func(tx *query.Query) error {
		runnerObj := &models.DaemonGcBlobRunner{RuleID: ruleObj.ID, Status: enums.TaskCommonStatusPending}
		err = daemonService.CreateGcBlobRunner(ctx, runnerObj)
		if err != nil {
			log.Error().Int64("ruleID", ruleObj.ID).Msgf("Create gc blob runner failed: %v", err)
			return xerrors.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Create gc blob runner failed: %v", err))
		}
		err = workq.ProducerClient.Produce(ctx, enums.DaemonGcBlob.String(),
			types.DaemonGcPayload{RunnerID: runnerObj.ID}, definition.ProducerOption{Tx: tx})
		if err != nil {
			log.Error().Err(err).Msgf("Send topic %s to work queue failed", enums.DaemonGcBlob.String())
			return xerrors.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Send topic %s to work queue failed", enums.DaemonGcBlob.String()))
		}
		if err != nil {
			var e xerrors.ErrCode
			if errors.As(err, &e) {
				return xerrors.NewHTTPError(c, e)
			}
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError)
		}
		return nil
	})

	return c.NoContent(http.StatusCreated)
}

// ListGcBlobRunners ...
func (h *handlers) ListGcBlobRunners(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	var req types.ListGcBlobRunnersRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, fmt.Sprintf("Bind and validate request body failed: %v", err))
	}
	daemonService := h.daemonServiceFactory.New()
	ruleObj, err := daemonService.GetGcBlobRule(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Int64("namespaceID", req.NamespaceID).Msg("Get gc blob rule not found")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, fmt.Sprintf("Get gc blob rule not found: %v", err))
		}
		log.Error().Err(err).Msg("Get gc blob rule failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Get gc blob rule failed: %v", err))
	}
	runnerObjs, total, err := daemonService.ListGcBlobRunners(ctx, ruleObj.ID, req.Pagination, req.Sortable)
	if err != nil {
		log.Error().Err(err).Msg("List gc blob rule failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("List gc blob rule failed: %v", err))
	}
	var resp = make([]any, 0, len(runnerObjs))
	for _, runnerObj := range runnerObjs {
		resp = append(resp, types.GcBlobRunnerItem{
			ID:        runnerObj.ID,
			Status:    runnerObj.Status,
			Message:   string(runnerObj.Message),
			CreatedAt: runnerObj.CreatedAt.Format(consts.DefaultTimePattern),
			UpdatedAt: runnerObj.UpdatedAt.Format(consts.DefaultTimePattern),
		})
	}
	return c.JSON(http.StatusOK, types.CommonList{Total: total, Items: resp})
}

// GetGcBlobRunner ...
func (h *handlers) GetGcBlobRunner(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	var req types.GetGcBlobRunnerRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, fmt.Sprintf("Bind and validate request body failed: %v", err))
	}
	daemonService := h.daemonServiceFactory.New()
	runnerObj, err := daemonService.GetGcBlobRunner(ctx, req.RunnerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Int64("namespaceID", req.NamespaceID).Int64("runnerID", req.RunnerID).Msg("Get gc artifact runner not found")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, fmt.Sprintf("Get gc artifact runner not found: %v", err))
		}
		log.Error().Err(err).Msg("Get gc artifact runner failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Get gc artifact runner failed: %v", err))
	}
	return c.JSON(http.StatusOK, types.GcBlobRunnerItem{
		ID:        runnerObj.ID,
		Status:    runnerObj.Status,
		Message:   string(runnerObj.Message),
		CreatedAt: runnerObj.CreatedAt.Format(consts.DefaultTimePattern),
		UpdatedAt: runnerObj.UpdatedAt.Format(consts.DefaultTimePattern),
	})
}

// ListGcBlobRecords ...
func (h *handlers) ListGcBlobRecords(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	var req types.ListGcBlobRecordsRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, fmt.Sprintf("Bind and validate request body failed: %v", err))
	}
	daemonService := h.daemonServiceFactory.New()
	recordObjs, total, err := daemonService.ListGcBlobRecords(ctx, req.RunnerID, req.Pagination, req.Sortable)
	if err != nil {
		log.Error().Err(err).Int64("ruleID", req.RunnerID).Msgf("List gc blob records failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("List gc blob records failed: %v", err))
	}
	var resp = make([]any, 0, len(recordObjs))
	for _, recordObj := range recordObjs {
		resp = append(resp, types.GcBlobRecordItem{
			ID:        recordObj.ID,
			Digest:    recordObj.Digest,
			CreatedAt: recordObj.CreatedAt.Format(consts.DefaultTimePattern),
			UpdatedAt: recordObj.UpdatedAt.Format(consts.DefaultTimePattern),
		})
	}
	return c.JSON(http.StatusOK, types.CommonList{Total: total, Items: resp})
}

// GetGcBlobRecord ...
func (h *handlers) GetGcBlobRecord(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	var req types.GetGcBlobRecordRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, fmt.Sprintf("Bind and validate request body failed: %v", err))
	}
	daemonService := h.daemonServiceFactory.New()
	ruleObj, err := daemonService.GetGcBlobRule(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Int64("namespaceID", req.NamespaceID).Msg("Get gc blob rule not found")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, fmt.Sprintf("Get gc blob rule not found: %v", err))
		}
		log.Error().Err(err).Msg("Get gc blob rule failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Get gc blob rule failed: %v", err))
	}
	recordObj, err := daemonService.GetGcBlobRecord(ctx, req.RecordID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Int64("namespaceID", req.NamespaceID).Int64("runnerID", req.RunnerID).Msg("Get gc blob record not found")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, fmt.Sprintf("Get gc blob record not found: %v", err))
		}
		log.Error().Err(err).Msg("Get gc blob record failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Get gc blob record failed: %v", err))
	}
	if recordObj.Runner.ID != req.RunnerID || recordObj.Runner.Rule.ID != ruleObj.ID {
		log.Error().Err(err).Int64("namespaceID", req.NamespaceID).Int64("runnerID", req.RunnerID).Msg("Get gc blob record not found")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, fmt.Sprintf("Get gc blob record not found: %v", err))
	}
	return c.JSON(http.StatusOK, types.GcBlobRecordItem{
		ID:        recordObj.ID,
		Digest:    recordObj.Digest,
		CreatedAt: recordObj.CreatedAt.Format(consts.DefaultTimePattern),
		UpdatedAt: recordObj.UpdatedAt.Format(consts.DefaultTimePattern),
	})
}
