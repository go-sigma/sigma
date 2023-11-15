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

// UpdateGcTagRule handles the update gc tag rule request
//
//	@Summary	Update gc tag rule
//	@security	BasicAuth
//	@Tags		Daemon
//	@Accept		json
//	@Produce	json
//	@Router		/daemons/gc-tag/{namespace_id}/ [put]
//	@Param		namespace_id	path	int64							true	"Namespace id"
//	@Param		message			body	types.UpdateGcTagRuleRequest	true	"Gc tag rule object"
//	@Success	204
//	@Failure	400	{object}	xerrors.ErrCode
//	@Failure	404	{object}	xerrors.ErrCode
//	@Failure	500	{object}	xerrors.ErrCode
func (h *handlers) UpdateGcTagRule(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	var req types.UpdateGcTagRuleRequest
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
	ruleObj, err := daemonService.GetGcTagRule(ctx, namespaceID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Error().Err(err).Msg("Get gc tag rule failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Get gc tag rule failed: %v", err))
	}
	if ruleObj != nil && ruleObj.IsRunning {
		log.Error().Int64("NamespaceID", ptr.To(namespaceID)).Msg("The gc tag rule is running")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, "The gc tag rule is running")
	}
	var nextTrigger *time.Time
	if req.CronRule != nil {
		schedule, _ := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow).Parse(ptr.To(req.CronRule))
		nextTrigger = ptr.Of(schedule.Next(time.Now()))
	}
	updates := make(map[string]any, 6)
	updates[query.DaemonGcTagRule.CronEnabled.ColumnName().String()] = req.CronEnabled
	updates[query.DaemonGcTagRule.RetentionRuleType.ColumnName().String()] = req.RetentionRuleType
	updates[query.DaemonGcTagRule.RetentionRuleAmount.ColumnName().String()] = req.RetentionRuleAmount
	if req.CronEnabled {
		if req.CronRule != nil {
			updates[query.DaemonGcTagRule.CronRule.ColumnName().String()] = ptr.To(req.CronRule)
			updates[query.DaemonGcTagRule.CronNextTrigger.ColumnName().String()] = ptr.To(nextTrigger)
		}
	}
	if req.RetentionPattern != nil {
		updates[query.DaemonGcTagRule.RetentionPattern.ColumnName().String()] = ptr.To(req.RetentionPattern)
	}
	err = query.Q.Transaction(func(tx *query.Query) error {
		if ruleObj == nil { // rule not found, we need create the rule
			err = daemonService.CreateGcTagRule(ctx, &models.DaemonGcTagRule{
				NamespaceID:         namespaceID,
				CronEnabled:         req.CronEnabled,
				CronRule:            req.CronRule,
				CronNextTrigger:     nextTrigger,
				RetentionPattern:    req.RetentionPattern,
				RetentionRuleType:   req.RetentionRuleType,
				RetentionRuleAmount: req.RetentionRuleAmount,
			})
			if err != nil {
				log.Error().Err(err).Msg("Create gc tag rule failed")
				return xerrors.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Create gc tag rule failed: %v", err))
			}
			return nil
		}
		err = daemonService.UpdateGcTagRule(ctx, ruleObj.ID, updates)
		if err != nil {
			log.Error().Err(err).Msg("Update gc tag rule failed")
			return xerrors.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Update gc tag rule failed: %v", err))
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

// GetGcTagRule handles the get gc tag rule request
//
//	@Summary	Get gc tag rule
//	@security	BasicAuth
//	@Tags		Daemon
//	@Accept		json
//	@Produce	json
//	@Router		/daemons/gc-tag/{namespace_id}/ [get]
//	@Param		namespace_id	path		int64	true	"Namespace id"
//	@Success	200				{object}	types.GetGcTagRuleResponse
//	@Failure	400				{object}	xerrors.ErrCode
//	@Failure	404				{object}	xerrors.ErrCode
//	@Failure	500				{object}	xerrors.ErrCode
func (h *handlers) GetGcTagRule(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	var req types.GetGcTagRuleRequest
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
	ruleObj, err := daemonService.GetGcTagRule(ctx, namespaceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Int64("namespaceID", req.NamespaceID).Msg("Get gc tag rule not found")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, fmt.Sprintf("Get gc tag rule not found: %v", err))
		}
		log.Error().Err(err).Msg("Get gc tag rule failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Get gc tag rule failed: %v", err))
	}
	return c.JSON(http.StatusOK, types.GetGcTagRuleResponse{
		CronEnabled:         ruleObj.CronEnabled,
		CronRule:            ruleObj.CronRule,
		CronNextTrigger:     ptr.Of(""), // response utc time, fe format with tz
		RetentionRuleType:   ruleObj.RetentionRuleType,
		RetentionRuleAmount: ruleObj.RetentionRuleAmount,
		RetentionPattern:    ruleObj.RetentionPattern,
		CreatedAt:           ruleObj.CreatedAt.Format(consts.DefaultTimePattern),
		UpdatedAt:           ruleObj.UpdatedAt.Format(consts.DefaultTimePattern),
	})
}

// GetGcTagLatestRunner handles the get gc tag latest runner request
//
//	@Summary	Get gc tag latest runner
//	@security	BasicAuth
//	@Tags		Daemon
//	@Accept		json
//	@Produce	json
//	@Router		/daemons/gc-tag/{namespace_id}/runners/latest [get]
//	@Param		namespace_id	path		int64	true	"Namespace id"
//	@Success	200				{object}	types.GcTagRunnerItem
//	@Failure	400				{object}	xerrors.ErrCode
//	@Failure	404				{object}	xerrors.ErrCode
//	@Failure	500				{object}	xerrors.ErrCode
func (h *handlers) GetGcTagLatestRunner(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	var req types.GetGcTagLatestRunnerRequest
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
	ruleObj, err := daemonService.GetGcTagRule(ctx, namespaceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Int64("namespaceID", req.NamespaceID).Msg("Get gc tag rule not found")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, fmt.Sprintf("Get gc tag rule not found: %v", err))
		}
		log.Error().Err(err).Msg("Get gc tag rule failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Get gc tag rule failed: %v", err))
	}
	runnerObj, err := daemonService.GetGcTagLatestRunner(ctx, ruleObj.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Int64("namespaceID", req.NamespaceID).Msg("Get gc tag rule not found")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, fmt.Sprintf("Get gc tag rule not found: %v", err))
		}
		log.Error().Err(err).Msg("Get gc tag rule failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Get gc tag rule failed: %v", err))
	}
	return c.JSON(http.StatusOK, types.GcTagRunnerItem{
		ID:        runnerObj.ID,
		Status:    runnerObj.Status,
		Message:   string(runnerObj.Message),
		CreatedAt: ruleObj.CreatedAt.Format(consts.DefaultTimePattern),
		UpdatedAt: ruleObj.UpdatedAt.Format(consts.DefaultTimePattern),
	})
}

// CreateGcTagRunner handles the create gc tag runner request
//
//	@Summary	Create gc tag runner
//	@security	BasicAuth
//	@Tags		Daemon
//	@Accept		json
//	@Produce	json
//	@Router		/daemons/gc-tag/{namespace_id}/runners/ [post]
//	@Param		namespace_id	path	int64							true	"Namespace id"
//	@Param		message			body	types.CreateGcTagRunnerRequest	true	"Gc tag runner object"
//	@Success	201
//	@Failure	400	{object}	xerrors.ErrCode
//	@Failure	404	{object}	xerrors.ErrCode
//	@Failure	500	{object}	xerrors.ErrCode
func (h *handlers) CreateGcTagRunner(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	var req types.CreateGcTagRunnerRequest
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
	ruleObj, err := daemonService.GetGcTagRule(ctx, namespaceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Int64("namespaceID", req.NamespaceID).Msg("Get gc tag rule not found")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, fmt.Sprintf("Get gc tag rule not found: %v", err))
		}
		log.Error().Err(err).Msg("Get gc tag rule failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Get gc tag rule failed: %v", err))
	}
	if ruleObj != nil && ruleObj.IsRunning {
		log.Error().Int64("NamespaceID", req.NamespaceID).Msg("The gc tag rule is running")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, "The gc tag rule is running")
	}
	err = query.Q.Transaction(func(tx *query.Query) error {
		runnerObj := &models.DaemonGcTagRunner{RuleID: ruleObj.ID, Status: enums.TaskCommonStatusPending}
		err = daemonService.CreateGcTagRunner(ctx, runnerObj)
		if err != nil {
			log.Error().Int64("ruleID", ruleObj.ID).Msgf("Create gc tag runner failed: %v", err)
			return xerrors.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Create gc tag runner failed: %v", err))
		}
		err = workq.ProducerClient.Produce(ctx, enums.DaemonGcTag.String(),
			types.DaemonGcPayload{RunnerID: runnerObj.ID}, definition.ProducerOption{Tx: tx})
		if err != nil {
			log.Error().Err(err).Msgf("Send topic %s to work queue failed", enums.DaemonGcTag.String())
			return xerrors.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Send topic %s to work queue failed", enums.DaemonGcTag.String()))
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

// ListGcTagRunners handles the list gc tag runners request
//
//	@Summary	List gc tag runners
//	@security	BasicAuth
//	@Tags		Daemon
//	@Accept		json
//	@Produce	json
//	@Router		/daemons/gc-tag/{namespace_id}/runners/ [get]
//	@Param		namespace_id	path		int64	true	"Namespace id"
//	@Param		limit			query		int64	false	"limit"	minimum(10)	maximum(100)	default(10)
//	@Param		page			query		int64	false	"page"	minimum(1)	default(1)
//	@Param		sort			query		string	false	"sort field"
//	@Param		method			query		string	false	"sort method"	Enums(asc, desc)
//	@Success	200				{object}	types.CommonList{items=[]types.GcTagRunnerItem}
//	@Failure	400				{object}	xerrors.ErrCode
//	@Failure	404				{object}	xerrors.ErrCode
//	@Failure	500				{object}	xerrors.ErrCode
func (h *handlers) ListGcTagRunners(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	var req types.ListGcTagRunnersRequest
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
	ruleObj, err := daemonService.GetGcTagRule(ctx, namespaceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Int64("namespaceID", req.NamespaceID).Msg("Get gc tag rule not found")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, fmt.Sprintf("Get gc tag rule not found: %v", err))
		}
		log.Error().Err(err).Msg("Get gc artifact rule failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Get gc artifact rule failed: %v", err))
	}
	runnerObjs, total, err := daemonService.ListGcTagRunners(ctx, ruleObj.ID, req.Pagination, req.Sortable)
	if err != nil {
		log.Error().Err(err).Msg("List gc artifact rule failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("List gc artifact rule failed: %v", err))
	}
	var resp = make([]any, 0, len(runnerObjs))
	for _, runnerObj := range runnerObjs {
		resp = append(resp, types.GcTagRunnerItem{
			ID:        runnerObj.ID,
			Status:    runnerObj.Status,
			Message:   string(runnerObj.Message),
			CreatedAt: runnerObj.CreatedAt.Format(consts.DefaultTimePattern),
			UpdatedAt: runnerObj.UpdatedAt.Format(consts.DefaultTimePattern),
		})
	}
	return c.JSON(http.StatusOK, types.CommonList{Total: total, Items: resp})
}

// GetGcTagRunner handles the get gc tag runner request
//
//	@Summary	List gc tag runners
//	@security	BasicAuth
//	@Tags		Daemon
//	@Accept		json
//	@Produce	json
//	@Router		/daemons/gc-tag/{namespace_id}/runners/{runner_id} [get]
//	@Param		namespace_id	path		int64	true	"Namespace id"
//	@Param		runner_id		path		int64	true	"Runner id"
//	@Success	200				{object}	types.GcTagRunnerItem
//	@Failure	400				{object}	xerrors.ErrCode
//	@Failure	404				{object}	xerrors.ErrCode
//	@Failure	500				{object}	xerrors.ErrCode
func (h *handlers) GetGcTagRunner(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	var req types.GetGcTagRunnerRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, fmt.Sprintf("Bind and validate request body failed: %v", err))
	}
	daemonService := h.daemonServiceFactory.New()
	runnerObj, err := daemonService.GetGcTagRunner(ctx, req.RunnerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Int64("namespaceID", req.NamespaceID).Int64("runnerID", req.RunnerID).Msg("Get gc tag runner not found")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, fmt.Sprintf("Get gc tag runner not found: %v", err))
		}
		log.Error().Err(err).Msg("Get gc tag runner failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Get gc tag runner failed: %v", err))
	}
	if ptr.To(runnerObj.Rule.NamespaceID) != req.NamespaceID {
		log.Error().Err(err).Int64("namespaceID", req.NamespaceID).Int64("runnerID", req.RunnerID).Msg("Get gc tag runner not found")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, fmt.Sprintf("Get gc tag runner not found: %v", err))
	}
	return c.JSON(http.StatusOK, types.GcTagRunnerItem{
		ID:        runnerObj.ID,
		Status:    runnerObj.Status,
		Message:   string(runnerObj.Message),
		CreatedAt: runnerObj.CreatedAt.Format(consts.DefaultTimePattern),
		UpdatedAt: runnerObj.UpdatedAt.Format(consts.DefaultTimePattern),
	})
}

// ListGcTagRecords handles the list gc tag records request
//
//	@Summary	List gc tag records
//	@security	BasicAuth
//	@Tags		Daemon
//	@Accept		json
//	@Produce	json
//	@Router		/daemons/gc-tag/{namespace_id}/runners/{runner_id}/records/ [get]
//	@Param		namespace_id	path		int64	true	"Namespace id"
//	@Param		runner_id		path		int64	true	"Runner id"
//	@Param		limit			query		int64	false	"limit"	minimum(10)	maximum(100)	default(10)
//	@Param		page			query		int64	false	"page"	minimum(1)	default(1)
//	@Param		sort			query		string	false	"sort field"
//	@Param		method			query		string	false	"sort method"	Enums(asc, desc)
//	@Success	200				{object}	types.CommonList{items=[]types.GcTagRecordItem}
//	@Failure	400				{object}	xerrors.ErrCode
//	@Failure	404				{object}	xerrors.ErrCode
//	@Failure	500				{object}	xerrors.ErrCode
func (h *handlers) ListGcTagRecords(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	var req types.ListGcTagRecordsRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, fmt.Sprintf("Bind and validate request body failed: %v", err))
	}
	daemonService := h.daemonServiceFactory.New()
	recordObjs, total, err := daemonService.ListGcTagRecords(ctx, req.RunnerID, req.Pagination, req.Sortable)
	if err != nil {
		log.Error().Err(err).Int64("ruleID", req.RunnerID).Msgf("List gc tag records failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("List gc tag records failed: %v", err))
	}
	var resp = make([]any, 0, len(recordObjs))
	for _, recordObj := range recordObjs {
		resp = append(resp, types.GcTagRecordItem{
			ID:        recordObj.ID,
			Tag:       recordObj.Tag,
			CreatedAt: recordObj.CreatedAt.Format(consts.DefaultTimePattern),
			UpdatedAt: recordObj.UpdatedAt.Format(consts.DefaultTimePattern),
		})
	}
	return c.JSON(http.StatusOK, types.CommonList{Total: total, Items: resp})
}

// GetGcTagRecord handles the get gc tag record request
//
//	@Summary	Get gc tag record
//	@security	BasicAuth
//	@Tags		Daemon
//	@Accept		json
//	@Produce	json
//	@Router		/daemons/gc-tag/{namespace_id}/runners/{runner_id}/records/{record_id} [get]
//	@Param		namespace_id	path		int64	true	"Namespace id"
//	@Param		runner_id		path		int64	true	"Runner id"
//	@Param		record_id		path		int64	true	"Record id"
//	@Success	200				{object}	types.GcTagRecordItem
//	@Failure	400				{object}	xerrors.ErrCode
//	@Failure	404				{object}	xerrors.ErrCode
//	@Failure	500				{object}	xerrors.ErrCode
func (h *handlers) GetGcTagRecord(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	var req types.GetGcTagRecordRequest
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
	ruleObj, err := daemonService.GetGcTagRule(ctx, namespaceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Int64("namespaceID", req.NamespaceID).Msg("Get gc tag rule not found")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, fmt.Sprintf("Get gc tag rule not found: %v", err))
		}
		log.Error().Err(err).Msg("Get gc tag rule failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Get gc tag rule failed: %v", err))
	}
	recordObj, err := daemonService.GetGcTagRecord(ctx, req.RecordID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Int64("namespaceID", req.NamespaceID).Int64("runnerID", req.RunnerID).Msg("Get gc tag record not found")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, fmt.Sprintf("Get gc tag record not found: %v", err))
		}
		log.Error().Err(err).Msg("Get gc tag record failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Get gc tag record failed: %v", err))
	}
	if recordObj.Runner.ID != req.RunnerID || recordObj.Runner.Rule.ID != ruleObj.ID {
		log.Error().Err(err).Int64("namespaceID", req.NamespaceID).Int64("runnerID", req.RunnerID).Msg("Get gc tag record not found")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, fmt.Sprintf("Get gc tag record not found: %v", err))
	}
	return c.JSON(http.StatusOK, types.GcTagRecordItem{
		ID:        recordObj.ID,
		Tag:       recordObj.Tag,
		CreatedAt: recordObj.CreatedAt.Format(consts.DefaultTimePattern),
		UpdatedAt: recordObj.UpdatedAt.Format(consts.DefaultTimePattern),
	})
}
