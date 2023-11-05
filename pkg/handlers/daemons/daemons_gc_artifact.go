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

// UpdateGcArtifactRule ...
func (h *handlers) UpdateGcArtifactRule(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	var req types.UpdateGcArtifactRuleRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, fmt.Sprintf("Bind and validate request body failed: %v", err))
	}
	var namespaceID *int64
	if req.NamespaceID != 0 {
		namespaceID = ptr.Of(req.NamespaceID)
	}
	daemonService := h.daemonServiceFactory.New()
	ruleObj, err := daemonService.GetGcArtifactRule(ctx, namespaceID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Error().Err(err).Msg("Get gc artifact rule failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Get gc artifact rule failed: %v", err))
	}
	if ruleObj != nil && ruleObj.IsRunning {
		log.Error().Int64("NamespaceID", ptr.To(namespaceID)).Msg("The gc artifact rule is running")
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
			updates[query.DaemonGcArtifactRule.CronRule.ColumnName().String()] = ptr.To(req.CronRule)
			updates[query.DaemonGcArtifactRule.CronNextTrigger.ColumnName().String()] = ptr.To(nextTrigger)
		}
	}
	err = query.Q.Transaction(func(tx *query.Query) error {
		if ruleObj == nil { // rule not found, we need create the rule
			err = daemonService.CreateGcArtifactRule(ctx, &models.DaemonGcArtifactRule{
				NamespaceID:     namespaceID,
				CronEnabled:     req.CronEnabled,
				CronRule:        req.CronRule,
				CronNextTrigger: nextTrigger,
			})
			if err != nil {
				log.Error().Err(err).Msg("Create gc artifact rule failed")
				return xerrors.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Create gc artifact rule failed: %v", err))
			}
		}
		err = daemonService.UpdateGcArtifactRule(ctx, ruleObj.ID, updates)
		if err != nil {
			log.Error().Err(err).Msg("Update gc artifact rule failed")
			return xerrors.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Update gc artifact rule failed: %v", err))
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

// GetGcArtifactRule ...
func (h *handlers) GetGcArtifactRule(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	var req types.GetGcArtifactRuleRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, fmt.Sprintf("Bind and validate request body failed: %v", err))
	}
	var namespaceID *int64
	if req.NamespaceID != 0 {
		namespaceID = ptr.Of(req.NamespaceID)
	}
	daemonService := h.daemonServiceFactory.New()
	ruleObj, err := daemonService.GetGcArtifactRule(ctx, namespaceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Int64("namespaceID", req.NamespaceID).Msg("Get gc artifact rule not found")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, fmt.Sprintf("Get gc artifact rule not found: %v", err))
		}
		log.Error().Err(err).Msg("Get gc artifact rule failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Get gc artifact rule failed: %v", err))
	}
	return c.JSON(http.StatusOK, types.GetGcArtifactRuleResponse{
		CronEnabled:     ruleObj.CronEnabled,
		CronRule:        ruleObj.CronRule,
		CronNextTrigger: ptr.Of(""),
		CreatedAt:       ruleObj.CreatedAt.Format(consts.DefaultTimePattern),
		UpdatedAt:       ruleObj.UpdatedAt.Format(consts.DefaultTimePattern),
	})
}

// GetGcArtifactLatestRunner ...
func (h *handlers) GetGcArtifactLatestRunner(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	var req types.GetGcArtifactLatestRunnerRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, fmt.Sprintf("Bind and validate request body failed: %v", err))
	}
	var namespaceID *int64
	if req.NamespaceID != 0 {
		namespaceID = ptr.Of(req.NamespaceID)
	}
	daemonService := h.daemonServiceFactory.New()
	ruleObj, err := daemonService.GetGcArtifactRule(ctx, namespaceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Int64("namespaceID", req.NamespaceID).Msg("Get gc artifact rule not found")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, fmt.Sprintf("Get gc artifact rule not found: %v", err))
		}
		log.Error().Err(err).Msg("Get gc artifact rule failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Get gc artifact rule failed: %v", err))
	}
	runnerObj, err := daemonService.GetGcArtifactLatestRunner(ctx, ruleObj.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Int64("namespaceID", req.NamespaceID).Msg("Get gc artifact runner not found")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, fmt.Sprintf("Get gc artifact runner not found: %v", err))
		}
		log.Error().Err(err).Msg("Get gc artifact runner failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Get gc artifact runner failed: %v", err))
	}
	return c.JSON(http.StatusOK, types.GcArtifactRunnerItem{
		ID:        runnerObj.ID,
		Status:    runnerObj.Status,
		Message:   string(runnerObj.Message),
		CreatedAt: runnerObj.CreatedAt.Format(consts.DefaultTimePattern),
		UpdatedAt: runnerObj.UpdatedAt.Format(consts.DefaultTimePattern),
	})
}

// CreateGcArtifactRunner ...
func (h *handlers) CreateGcArtifactRunner(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	var req types.CreateGcArtifactRunnerRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, fmt.Sprintf("Bind and validate request body failed: %v", err))
	}
	var namespaceID *int64
	if req.NamespaceID != 0 {
		namespaceID = ptr.Of(req.NamespaceID)
	}
	daemonService := h.daemonServiceFactory.New()
	ruleObj, err := daemonService.GetGcArtifactRule(ctx, namespaceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Int64("namespaceID", req.NamespaceID).Msg("Get gc artifact rule not found")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, fmt.Sprintf("Get gc artifact rule not found: %v", err))
		}
		log.Error().Err(err).Msg("Get gc artifact rule failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Get gc artifact rule failed: %v", err))
	}
	if ruleObj != nil && ruleObj.IsRunning {
		log.Error().Int64("NamespaceID", ptr.To(namespaceID)).Msg("The gc artifact rule is running")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, "The gc artifact rule is running")
	}
	err = query.Q.Transaction(func(tx *query.Query) error {
		runnerObj := &models.DaemonGcArtifactRunner{RuleID: ruleObj.ID, Status: enums.TaskCommonStatusPending}
		err = daemonService.CreateGcArtifactRunner(ctx, runnerObj)
		if err != nil {
			log.Error().Int64("ruleID", ruleObj.ID).Msgf("Create gc artifact runner failed: %v", err)
			return xerrors.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Create gc artifact runner failed: %v", err))
		}
		err = workq.ProducerClient.Produce(ctx, enums.DaemonGcArtifact.String(),
			types.DaemonGcPayload{RunnerID: runnerObj.ID}, definition.ProducerOption{Tx: tx})
		if err != nil {
			log.Error().Err(err).Msgf("Send topic %s to work queue failed", enums.DaemonGcArtifact.String())
			return xerrors.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Send topic %s to work queue failed", enums.DaemonGcArtifact.String()))
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
	return c.NoContent(http.StatusCreated)
}

// ListGcArtifactRunners ...
func (h *handlers) ListGcArtifactRunners(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	var req types.ListGcArtifactRunnersRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, fmt.Sprintf("Bind and validate request body failed: %v", err))
	}
	var namespaceID *int64
	if req.NamespaceID != 0 {
		namespaceID = ptr.Of(req.NamespaceID)
	}
	daemonService := h.daemonServiceFactory.New()
	ruleObj, err := daemonService.GetGcArtifactRule(ctx, namespaceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Int64("namespaceID", req.NamespaceID).Msg("Get gc artifact rule not found")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, fmt.Sprintf("Get gc artifact rule not found: %v", err))
		}
		log.Error().Err(err).Msg("Get gc artifact rule failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Get gc artifact rule failed: %v", err))
	}
	runnerObjs, total, err := daemonService.ListGcArtifactRunners(ctx, ruleObj.ID, req.Pagination, req.Sortable)
	if err != nil {
		log.Error().Err(err).Msg("List gc artifact rules failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("List gc artifact rules failed: %v", err))
	}
	var resp = make([]any, 0, len(runnerObjs))
	for _, runnerObj := range runnerObjs {
		resp = append(resp, types.GcArtifactRunnerItem{
			ID:        runnerObj.ID,
			Status:    runnerObj.Status,
			Message:   string(runnerObj.Message),
			CreatedAt: runnerObj.CreatedAt.Format(consts.DefaultTimePattern),
			UpdatedAt: runnerObj.UpdatedAt.Format(consts.DefaultTimePattern),
		})
	}
	return c.JSON(http.StatusOK, types.CommonList{Total: total, Items: resp})
}

// GetGcArtifactRunner ...
func (h *handlers) GetGcArtifactRunner(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	var req types.GetGcArtifactRunnerRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, fmt.Sprintf("Bind and validate request body failed: %v", err))
	}
	daemonService := h.daemonServiceFactory.New()
	runnerObj, err := daemonService.GetGcArtifactRunner(ctx, req.RunnerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Int64("namespaceID", req.NamespaceID).Int64("runnerID", req.RunnerID).Msg("Get gc artifact runner not found")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, fmt.Sprintf("Get gc artifact runner not found: %v", err))
		}
		log.Error().Err(err).Msg("Get gc artifact runner failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Get gc artifact runner failed: %v", err))
	}
	if ptr.To(runnerObj.Rule.NamespaceID) != req.NamespaceID {
		log.Error().Err(err).Int64("namespaceID", req.NamespaceID).Int64("runnerID", req.RunnerID).Msg("Get gc artifact runner not found")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, fmt.Sprintf("Get gc artifact runner not found: %v", err))
	}
	return c.JSON(http.StatusOK, types.GcArtifactRunnerItem{
		ID:        runnerObj.ID,
		Status:    runnerObj.Status,
		Message:   string(runnerObj.Message),
		CreatedAt: runnerObj.CreatedAt.Format(consts.DefaultTimePattern),
		UpdatedAt: runnerObj.UpdatedAt.Format(consts.DefaultTimePattern),
	})
}

// ListGcArtifactRecords ...
func (h *handlers) ListGcArtifactRecords(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	var req types.ListGcArtifactRecordsRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, fmt.Sprintf("Bind and validate request body failed: %v", err))
	}
	daemonService := h.daemonServiceFactory.New()
	recordObjs, total, err := daemonService.ListGcArtifactRecords(ctx, req.RunnerID, req.Pagination, req.Sortable)
	if err != nil {
		log.Error().Err(err).Int64("ruleID", req.RunnerID).Msgf("List gc artifact records failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("List gc artifact records failed: %v", err))
	}
	var resp = make([]any, 0, len(recordObjs))
	for _, recordObj := range recordObjs {
		resp = append(resp, types.GcArtifactRecordItem{
			ID:        recordObj.ID,
			Digest:    recordObj.Digest,
			CreatedAt: recordObj.CreatedAt.Format(consts.DefaultTimePattern),
			UpdatedAt: recordObj.UpdatedAt.Format(consts.DefaultTimePattern),
		})
	}
	return c.JSON(http.StatusOK, types.CommonList{Total: total, Items: resp})
}

// GetGcArtifactRecord ...
func (h *handlers) GetGcArtifactRecord(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	var req types.GetGcArtifactRecordRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, fmt.Sprintf("Bind and validate request body failed: %v", err))
	}
	var namespaceID *int64
	if req.NamespaceID != 0 {
		namespaceID = ptr.Of(req.NamespaceID)
	}
	daemonService := h.daemonServiceFactory.New()
	ruleObj, err := daemonService.GetGcArtifactRule(ctx, namespaceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Int64("namespaceID", req.NamespaceID).Msg("Get gc artifact rule not found")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, fmt.Sprintf("Get gc artifact rule not found: %v", err))
		}
		log.Error().Err(err).Msg("Get gc artifact rule failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Get gc artifact rule failed: %v", err))
	}
	recordObj, err := daemonService.GetGcArtifactRecord(ctx, req.RecordID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Int64("namespaceID", req.NamespaceID).Int64("runnerID", req.RunnerID).Msg("Get gc artifact record not found")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, fmt.Sprintf("Get gc artifact record not found: %v", err))
		}
		log.Error().Err(err).Msg("Get gc artifact record failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Get gc artifact record failed: %v", err))
	}
	if recordObj.Runner.ID != req.RunnerID || recordObj.Runner.Rule.ID != ruleObj.ID {
		log.Error().Err(err).Int64("namespaceID", req.NamespaceID).Int64("runnerID", req.RunnerID).Msg("Get gc artifact record not found")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, fmt.Sprintf("Get gc artifact record not found: %v", err))
	}
	return c.JSON(http.StatusOK, types.GcArtifactRecordItem{
		ID:        recordObj.ID,
		Digest:    recordObj.Digest,
		CreatedAt: recordObj.CreatedAt.Format(consts.DefaultTimePattern),
		UpdatedAt: recordObj.UpdatedAt.Format(consts.DefaultTimePattern),
	})
}
