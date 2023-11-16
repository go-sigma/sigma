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

	"github.com/hako/durafmt"
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

// UpdateGcRepositoryRule handles the update gc repository rule request
//
//	@Summary	Update gc repository rule
//	@security	BasicAuth
//	@Tags		Daemon
//	@Accept		json
//	@Produce	json
//	@Router		/daemons/gc-repository/{namespace_id}/ [put]
//	@Param		namespace_id	path	int64								true	"Namespace id"
//	@Param		message			body	types.UpdateGcRepositoryRuleRequest	true	"Gc repository rule object"
//	@Success	204
//	@Failure	400	{object}	xerrors.ErrCode
//	@Failure	404	{object}	xerrors.ErrCode
//	@Failure	500	{object}	xerrors.ErrCode
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
		log.Error().Err(err).Msg("Get gc repository rule failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Get gc repository rule failed: %v", err))
	}
	if ruleObj != nil && ruleObj.IsRunning {
		log.Error().Int64("NamespaceID", ptr.To(namespaceID)).Msg("The gc repository rule is running")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, "The gc repository rule is running")
	}
	var nextTrigger *time.Time
	if req.CronRule != nil {
		schedule, _ := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow).Parse(ptr.To(req.CronRule))
		nextTrigger = ptr.Of(schedule.Next(time.Now()))
	}
	updates := make(map[string]any, 5)
	updates[query.DaemonGcRepositoryRule.RetentionDay.ColumnName().String()] = req.RetentionDay
	updates[query.DaemonGcRepositoryRule.CronEnabled.ColumnName().String()] = req.CronEnabled
	if ptr.To(req.CronEnabled) {
		updates[query.DaemonGcRepositoryRule.CronRule.ColumnName().String()] = ptr.To(req.CronRule)
		updates[query.DaemonGcRepositoryRule.CronNextTrigger.ColumnName().String()] = ptr.To(nextTrigger)
	}
	err = query.Q.Transaction(func(tx *query.Query) error {
		if ruleObj == nil { // rule not found, we need create the rule
			err = daemonService.CreateGcRepositoryRule(ctx, &models.DaemonGcRepositoryRule{
				NamespaceID:     namespaceID,
				RetentionDay:    req.RetentionDay,
				CronEnabled:     ptr.To(req.CronEnabled),
				CronRule:        req.CronRule,
				CronNextTrigger: nextTrigger,
			})
			if err != nil {
				log.Error().Err(err).Msg("Create gc repository rule failed")
				return xerrors.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Create gc repository rule failed: %v", err))
			}
			return nil
		}
		err = daemonService.UpdateGcRepositoryRule(ctx, ruleObj.ID, updates)
		if err != nil {
			log.Error().Err(err).Msg("Update gc repository rule failed")
			return xerrors.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Update gc repository rule failed: %v", err))
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

// GetGcRepositoryRule handles the get gc repository rule request
//
//	@Summary	Get gc repository rule
//	@security	BasicAuth
//	@Tags		Daemon
//	@Accept		json
//	@Produce	json
//	@Router		/daemons/gc-repository/{namespace_id}/ [get]
//	@Param		namespace_id	path		int64	true	"Namespace id"
//	@Success	200				{object}	types.GetGcRepositoryRuleResponse
//	@Failure	400				{object}	xerrors.ErrCode
//	@Failure	404				{object}	xerrors.ErrCode
//	@Failure	500				{object}	xerrors.ErrCode
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
			log.Error().Err(err).Int64("NamespaceID", req.NamespaceID).Msg("Get gc repository rule not found")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, fmt.Sprintf("Get gc repository rule not found: %v", err))
		}
		log.Error().Err(err).Msg("Get gc repository rule failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Get gc repository rule failed: %v", err))
	}
	return c.JSON(http.StatusOK, types.GetGcRepositoryRuleResponse{
		RetentionDay:    ruleObj.RetentionDay,
		CronEnabled:     ruleObj.CronEnabled,
		CronRule:        ruleObj.CronRule,
		CronNextTrigger: ptr.Of(ptr.To(ruleObj.CronNextTrigger).Format(consts.DefaultTimePattern)),
		CreatedAt:       ruleObj.CreatedAt.Format(consts.DefaultTimePattern),
		UpdatedAt:       ruleObj.UpdatedAt.Format(consts.DefaultTimePattern),
	})
}

// GetGcRepositoryLatestRunner handles the get gc repository latest runner request
//
//	@Summary	Get gc repository latest runner
//	@security	BasicAuth
//	@Tags		Daemon
//	@Accept		json
//	@Produce	json
//	@Router		/daemons/gc-repository/{namespace_id}/runners/latest [get]
//	@Param		namespace_id	path		int64	true	"Namespace id"
//	@Success	200				{object}	types.GcRepositoryRunnerItem
//	@Failure	400				{object}	xerrors.ErrCode
//	@Failure	404				{object}	xerrors.ErrCode
//	@Failure	500				{object}	xerrors.ErrCode
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
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Int64("NamespaceID", req.NamespaceID).Msg("Get gc repository rule not found")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, fmt.Sprintf("Get gc repository rule not found: %v", err))
		}
		log.Error().Err(err).Msg("Get gc repository rule failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Get gc repository rule failed: %v", err))
	}
	runnerObj, err := daemonService.GetGcRepositoryLatestRunner(ctx, ruleObj.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Int64("NamespaceID", req.NamespaceID).Msg("Get gc repository rule not found")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, fmt.Sprintf("Get gc repository rule not found: %v", err))
		}
		log.Error().Err(err).Msg("Get gc repository rule failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Get gc repository rule failed: %v", err))
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
	return c.JSON(http.StatusOK, types.GcRepositoryRunnerItem{
		ID:           runnerObj.ID,
		Status:       runnerObj.Status,
		Message:      string(runnerObj.Message),
		FailedCount:  runnerObj.FailedCount,
		SuccessCount: runnerObj.SuccessCount,
		RawDuration:  runnerObj.Duration,
		Duration:     duration,
		StartedAt:    ptr.Of(startedAt),
		EndedAt:      ptr.Of(endedAt),
		CreatedAt:    ruleObj.CreatedAt.Format(consts.DefaultTimePattern),
		UpdatedAt:    ruleObj.UpdatedAt.Format(consts.DefaultTimePattern),
	})
}

// CreateGcRepositoryRunner handles the create gc repository runner request
//
//	@Summary	Create gc repository runner
//	@security	BasicAuth
//	@Tags		Daemon
//	@Accept		json
//	@Produce	json
//	@Router		/daemons/gc-repository/{namespace_id}/runners/ [post]
//	@Param		namespace_id	path	int64									true	"Namespace id"
//	@Param		message			body	types.CreateGcRepositoryRunnerRequest	true	"Gc repository runner object"
//	@Success	201
//	@Failure	400	{object}	xerrors.ErrCode
//	@Failure	404	{object}	xerrors.ErrCode
//	@Failure	500	{object}	xerrors.ErrCode
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
	err = query.Q.Transaction(func(tx *query.Query) error {
		runnerObj := &models.DaemonGcRepositoryRunner{RuleID: ruleObj.ID, Status: enums.TaskCommonStatusPending}
		err = daemonService.CreateGcRepositoryRunner(ctx, runnerObj)
		if err != nil {
			log.Error().Int64("ruleID", ruleObj.ID).Msgf("Create gc repository runner failed: %v", err)
			return xerrors.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Create gc repository runner failed: %v", err))
		}
		err = workq.ProducerClient.Produce(ctx, enums.DaemonGcRepository.String(),
			types.DaemonGcPayload{RunnerID: runnerObj.ID}, definition.ProducerOption{Tx: tx})
		if err != nil {
			log.Error().Err(err).Msgf("Send topic %s to work queue failed", enums.DaemonGcRepository.String())
			return xerrors.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Send topic %s to work queue failed", enums.DaemonGcRepository.String()))
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

// ListGcRepositoryRunners handles the list gc repository runners request
//
//	@Summary	List gc repository runners
//	@security	BasicAuth
//	@Tags		Daemon
//	@Accept		json
//	@Produce	json
//	@Router		/daemons/gc-repository/{namespace_id}/runners/ [get]
//	@Param		namespace_id	path		int64	true	"Namespace id"
//	@Param		limit			query		int64	false	"limit"	minimum(10)	maximum(100)	default(10)
//	@Param		page			query		int64	false	"page"	minimum(1)	default(1)
//	@Param		sort			query		string	false	"sort field"
//	@Param		method			query		string	false	"sort method"	Enums(asc, desc)
//	@Success	200				{object}	types.CommonList{items=[]types.GcRepositoryRunnerItem}
//	@Failure	400				{object}	xerrors.ErrCode
//	@Failure	404				{object}	xerrors.ErrCode
//	@Failure	500				{object}	xerrors.ErrCode
func (h *handlers) ListGcRepositoryRunners(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	var req types.ListGcRepositoryRunnersRequest
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
			log.Error().Err(err).Int64("NamespaceID", req.NamespaceID).Msg("Get gc repository rule not found")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, fmt.Sprintf("Get gc repository rule not found: %v", err))
		}
		log.Error().Err(err).Msg("Get gc repository rule failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Get gc repository rule failed: %v", err))
	}
	runnerObjs, total, err := daemonService.ListGcRepositoryRunners(ctx, ruleObj.ID, req.Pagination, req.Sortable)
	if err != nil {
		log.Error().Err(err).Msg("List gc repository rule failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("List gc repository rule failed: %v", err))
	}
	var resp = make([]any, 0, len(runnerObjs))
	for _, runnerObj := range runnerObjs {
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
		resp = append(resp, types.GcRepositoryRunnerItem{
			ID:           runnerObj.ID,
			Status:       runnerObj.Status,
			Message:      string(runnerObj.Message),
			SuccessCount: runnerObj.SuccessCount,
			FailedCount:  runnerObj.FailedCount,
			RawDuration:  runnerObj.Duration,
			Duration:     duration,
			StartedAt:    ptr.Of(startedAt),
			EndedAt:      ptr.Of(endedAt),
			CreatedAt:    runnerObj.CreatedAt.Format(consts.DefaultTimePattern),
			UpdatedAt:    runnerObj.UpdatedAt.Format(consts.DefaultTimePattern),
		})
	}
	return c.JSON(http.StatusOK, types.CommonList{Total: total, Items: resp})
}

// GetGcRepositoryRunner handles the get gc repository runner request
//
//	@Summary	List gc repository runners
//	@security	BasicAuth
//	@Tags		Daemon
//	@Accept		json
//	@Produce	json
//	@Router		/daemons/gc-repository/{namespace_id}/runners/{runner_id} [get]
//	@Param		namespace_id	path		int64	true	"Namespace id"
//	@Param		runner_id		path		int64	true	"Runner id"
//	@Success	200				{object}	types.GcRepositoryRunnerItem
//	@Failure	400				{object}	xerrors.ErrCode
//	@Failure	404				{object}	xerrors.ErrCode
//	@Failure	500				{object}	xerrors.ErrCode
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
			log.Error().Err(err).Int64("namespaceID", req.NamespaceID).Int64("runnerID", req.RunnerID).Msg("Get gc repository runner not found")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, fmt.Sprintf("Get gc repository runner not found: %v", err))
		}
		log.Error().Err(err).Msg("Get gc repository runner failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Get gc repository runner failed: %v", err))
	}
	if ptr.To(runnerObj.Rule.NamespaceID) != req.NamespaceID {
		log.Error().Err(err).Int64("namespaceID", req.NamespaceID).Int64("runnerID", req.RunnerID).Msg("Get gc repository runner not found")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, fmt.Sprintf("Get gc repository runner not found: %v", err))
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
	return c.JSON(http.StatusOK, types.GcRepositoryRunnerItem{
		ID:           runnerObj.ID,
		Status:       runnerObj.Status,
		Message:      string(runnerObj.Message),
		SuccessCount: runnerObj.SuccessCount,
		FailedCount:  runnerObj.FailedCount,
		RawDuration:  runnerObj.Duration,
		Duration:     duration,
		StartedAt:    ptr.Of(startedAt),
		EndedAt:      ptr.Of(endedAt),
		CreatedAt:    runnerObj.CreatedAt.Format(consts.DefaultTimePattern),
		UpdatedAt:    runnerObj.UpdatedAt.Format(consts.DefaultTimePattern),
	})
}

// ListGcRepositoryRecords handles the list gc repository records request
//
//	@Summary	List gc repository records
//	@security	BasicAuth
//	@Tags		Daemon
//	@Accept		json
//	@Produce	json
//	@Router		/daemons/gc-repository/{namespace_id}/runners/{runner_id}/records/ [get]
//	@Param		namespace_id	path		int64	true	"Namespace id"
//	@Param		runner_id		path		int64	true	"Runner id"
//	@Param		limit			query		int64	false	"limit"	minimum(10)	maximum(100)	default(10)
//	@Param		page			query		int64	false	"page"	minimum(1)	default(1)
//	@Param		sort			query		string	false	"sort field"
//	@Param		method			query		string	false	"sort method"	Enums(asc, desc)
//	@Success	200				{object}	types.CommonList{items=[]types.GcRepositoryRecordItem}
//	@Failure	400				{object}	xerrors.ErrCode
//	@Failure	404				{object}	xerrors.ErrCode
//	@Failure	500				{object}	xerrors.ErrCode
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
		log.Error().Err(err).Int64("ruleID", req.RunnerID).Msgf("List gc repository records failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("List gc repository records failed: %v", err))
	}
	var resp = make([]any, 0, len(recordObjs))
	for _, recordObj := range recordObjs {
		resp = append(resp, types.GcRepositoryRecordItem{
			ID:         recordObj.ID,
			Repository: recordObj.Repository,
			Status:     recordObj.Status,
			Message:    string(recordObj.Message),
			CreatedAt:  recordObj.CreatedAt.Format(consts.DefaultTimePattern),
			UpdatedAt:  recordObj.UpdatedAt.Format(consts.DefaultTimePattern),
		})
	}
	return c.JSON(http.StatusOK, types.CommonList{Total: total, Items: resp})
}

// GetGcRepositoryRecord handles the get gc repository record request
//
//	@Summary	Get gc repository record
//	@security	BasicAuth
//	@Tags		Daemon
//	@Accept		json
//	@Produce	json
//	@Router		/daemons/gc-repository/{namespace_id}/runners/{runner_id}/records/{record_id} [get]
//	@Param		namespace_id	path		int64	true	"Namespace id"
//	@Param		runner_id		path		int64	true	"Runner id"
//	@Param		record_id		path		int64	true	"Record id"
//	@Success	200				{object}	types.GcRepositoryRecordItem
//	@Failure	400				{object}	xerrors.ErrCode
//	@Failure	404				{object}	xerrors.ErrCode
//	@Failure	500				{object}	xerrors.ErrCode
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
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Int64("namespaceID", req.NamespaceID).Msg("Get gc repository rule not found")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, fmt.Sprintf("Get gc repository rule not found: %v", err))
		}
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
		Status:     recordObj.Status,
		Message:    string(recordObj.Message),
		CreatedAt:  recordObj.CreatedAt.Format(consts.DefaultTimePattern),
		UpdatedAt:  recordObj.UpdatedAt.Format(consts.DefaultTimePattern),
	})
}
