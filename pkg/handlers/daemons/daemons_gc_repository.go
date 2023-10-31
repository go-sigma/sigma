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
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
	"github.com/go-sigma/sigma/pkg/xerrors"
)

// UpdateGcRepositoryRule ...
func (h *handlers) UpdateGcRepositoryRule(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	var req types.UpdateGcRepositoryRuleRequest
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
	ruleObj, err := daemonService.GetGcRepositoryRule(ctx, namespaceID)
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
			updates[query.DaemonGcRepositoryRule.CronRule.ColumnName().String()] = ptr.To(req.CronRule)
			updates[query.DaemonGcRepositoryRule.CronNextTrigger.ColumnName().String()] = ptr.To(nextTrigger)
		}
	}
	err = query.Q.Transaction(func(tx *query.Query) error {
		if ruleObj == nil { // rule not found, we need create the rule
			err = daemonService.CreateGcRepositoryRule(ctx, &models.DaemonGcRepositoryRule{
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
		err = daemonService.UpdateGcRepositoryRule(ctx, ruleObj.ID, updates)
		if err != nil {
			log.Error().Err(err).Msg("Update gc artifact rule failed")
			return xerrors.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Update gc artifact rule failed: %v", err))
		}
		return nil
	})
	if err != nil {
		e, ok := err.(xerrors.ErrCode) // maybe got exceed tag quota limit error
		if ok {
			return xerrors.NewHTTPError(c, e)
		}
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError)
	}
	// err = daemonService.UpdateGcRepositoryRule(ctx, &models.DaemonGcRepositoryRule{
	// 	NamespaceID: namespaceID,
	// 	CronEnabled: req.CronEnabled,
	// 	CronRule:    req.CronRule, // TODO: next trigger
	// })
	// if err != nil {
	// 	log.Error().Err(err).Msg("Update gc artifact rule failed")
	// 	return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Update gc artifact rule failed: %v", err))
	// }
	return c.NoContent(http.StatusNoContent)
}

// GetGcRepositoryRule
func (h *handlers) GetGcRepositoryRule(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	var req types.GetGcRepositoryRuleRequest
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
	ruleObj, err := daemonService.GetGcRepositoryRule(ctx, namespaceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Int64("namespaceID", req.NamespaceID).Msg("Get gc repository rule not found")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, fmt.Sprintf("Get gc repository rule not found: %v", err))
		}
		log.Error().Err(err).Msg("Get gc repository rule failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Get gc repository rule failed: %v", err))
	}
	return c.JSON(http.StatusOK, types.GetGcRepositoryRuleResponse{
		CronEnabled:     ruleObj.CronEnabled,
		CronRule:        ruleObj.CronRule,
		CronNextTrigger: ptr.Of(""), // response utc time, fe format with tz
		CreatedAt:       ruleObj.CreatedAt.Format(consts.DefaultTimePattern),
		UpdatedAt:       ruleObj.UpdatedAt.Format(consts.DefaultTimePattern),
	})
}

// GetGcRepositoryLatestRunner ...
func (h *handlers) GetGcRepositoryLatestRunner(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	var req types.GetGcRepositoryLatestRunnerRequest
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
	ruleObj, err := daemonService.GetGcRepositoryRule(ctx, namespaceID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Error().Err(err).Msg("Get gc repository rule failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Get gc repository rule failed: %v", err))
	}
	runnerObj, err := daemonService.GetGcRepositoryLatestRunner(ctx, ruleObj.ID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Error().Err(err).Msg("Get gc repository rule failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Get gc repository rule failed: %v", err))
	}
	return c.JSON(http.StatusOK, types.GcRepositoryRunnerItem{
		ID:        runnerObj.ID,
		Status:    runnerObj.Status,
		Message:   string(runnerObj.Message),
		CreatedAt: ruleObj.CreatedAt.Format(consts.DefaultTimePattern),
		UpdatedAt: ruleObj.UpdatedAt.Format(consts.DefaultTimePattern),
	})
}

// CreateGcRepositoryRunner ...
func (h *handlers) CreateGcRepositoryRunner(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	var req types.CreateGcRepositoryRunnerRequest
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
	ruleObj, err := daemonService.GetGcRepositoryRule(ctx, namespaceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Int64("namespaceID", req.NamespaceID).Msg("Get gc repository rule not found")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, fmt.Sprintf("Get gc repository rule not found: %v", err))
		}
		log.Error().Err(err).Msg("Get gc repository rule failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Get gc repository rule failed: %v", err))
	}
	if ruleObj != nil && ruleObj.IsRunning {
		log.Error().Int64("NamespaceID", ptr.To(namespaceID)).Msg("The gc repository rule is running")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, "The gc repository rule is running")
	}
	err = daemonService.CreateGcRepositoryRunner(ctx, &models.DaemonGcRepositoryRunner{RuleID: ruleObj.ID, Status: enums.TaskCommonStatusPending})
	if err != nil {
		log.Error().Int64("ruleID", ruleObj.ID).Msgf("Create gc repository runner failed: %v", err)
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Create gc repository runner failed: %v", err))
	}
	return c.NoContent(http.StatusCreated)
}

// ListGcRepositoryRunners ...
func (h *handlers) ListGcRepositoryRunners(c echo.Context) error {
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
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Error().Err(err).Msg("Get gc artifact rule failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Get gc artifact rule failed: %v", err))
	}
	runnerObjs, total, err := daemonService.ListGcArtifactRunners(ctx, ruleObj.ID, req.Pagination, req.Sortable)
	if err != nil {
		log.Error().Err(err).Msg("List gc artifact rule failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("List gc artifact rule failed: %v", err))
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

// GetGcRepositoryRunner ...
func (h *handlers) GetGcRepositoryRunner(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	var req types.GetGcRepositoryRunnerRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, fmt.Sprintf("Bind and validate request body failed: %v", err))
	}
	daemonService := h.daemonServiceFactory.New()
	runnerObj, err := daemonService.GetGcRepositoryRunner(ctx, req.RunnerID)
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
	return c.JSON(http.StatusOK, types.GcRepositoryRunnerItem{
		ID:        runnerObj.ID,
		Status:    runnerObj.Status,
		Message:   string(runnerObj.Message),
		CreatedAt: runnerObj.CreatedAt.Format(consts.DefaultTimePattern),
		UpdatedAt: runnerObj.UpdatedAt.Format(consts.DefaultTimePattern),
	})
}

// ListGcRepositoryRecords ...
func (h *handlers) ListGcRepositoryRecords(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	var req types.ListGcRepositoryRecordsRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, fmt.Sprintf("Bind and validate request body failed: %v", err))
	}
	daemonService := h.daemonServiceFactory.New()
	recordObjs, total, err := daemonService.ListGcRepositoryRecords(ctx, req.RunnerID, req.Pagination, req.Sortable)
	if err != nil {
		log.Error().Err(err).Int64("ruleID", req.RunnerID).Msgf("List gc artifact records failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("List gc artifact records failed: %v", err))
	}
	var resp = make([]any, 0, len(recordObjs))
	for _, recordObj := range recordObjs {
		resp = append(resp, types.GcRepositoryRecordItem{
			ID:         recordObj.ID,
			Repository: recordObj.Repository,
			CreatedAt:  recordObj.CreatedAt.Format(consts.DefaultTimePattern),
			UpdatedAt:  recordObj.UpdatedAt.Format(consts.DefaultTimePattern),
		})
	}
	return c.JSON(http.StatusOK, types.CommonList{Total: total, Items: resp})
}

// GetGcRepositoryRecord ...
func (h *handlers) GetGcRepositoryRecord(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	var req types.GetGcRepositoryRecordRequest
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
	ruleObj, err := daemonService.GetGcRepositoryRule(ctx, namespaceID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Error().Err(err).Msg("Get gc repository rule failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Get gc repository rule failed: %v", err))
	}
	recordObj, err := daemonService.GetGcRepositoryRecord(ctx, req.RecordID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Int64("namespaceID", req.NamespaceID).Int64("runnerID", req.RunnerID).Msg("Get gc repository record not found")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, fmt.Sprintf("Get gc repository record not found: %v", err))
		}
		log.Error().Err(err).Msg("Get gc repository record failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Get gc repository record failed: %v", err))
	}
	if recordObj.Runner.ID != req.RunnerID || recordObj.Runner.Rule.ID != ruleObj.ID {
		log.Error().Err(err).Int64("namespaceID", req.NamespaceID).Int64("runnerID", req.RunnerID).Msg("Get gc repository record not found")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, fmt.Sprintf("Get gc repository record not found: %v", err))
	}
	return c.JSON(http.StatusOK, types.GcRepositoryRecordItem{
		ID:         recordObj.ID,
		Repository: recordObj.Repository,
		CreatedAt:  recordObj.CreatedAt.Format(consts.DefaultTimePattern),
		UpdatedAt:  recordObj.UpdatedAt.Format(consts.DefaultTimePattern),
	})
}
