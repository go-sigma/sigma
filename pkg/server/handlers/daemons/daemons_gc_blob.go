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

// UpdateGcBlobRule handles the update gc blob rule request
//
//	@Summary	Update gc blob rule
//	@security	BasicAuth
//	@Tags		Daemon
//	@Accept		json
//	@Produce	json
//	@Router		/daemons/gc-blob/{namespace_id}/ [put]
//	@Param		namespace_id	path	int64							true	"Namespace id"
//	@Param		message			body	types.UpdateGcBlobRuleRequest	true	"Gc blob rule object"
//	@Success	204
//	@Failure	400	{object}	xerrors.ErrCode
//	@Failure	404	{object}	xerrors.ErrCode
//	@Failure	500	{object}	xerrors.ErrCode
func (h *handler) UpdateGcBlobRule(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	// We still keep the "NamespaceID" field because we want to ensure API consistency and it can also be used for permission verification.
	var req types.UpdateGcBlobRuleRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, fmt.Sprintf("Bind and validate request body failed: %v", err))
	}
	if req.NamespaceID != 0 {
		log.Error().Msg("NamespaceID should always be 0 in action UpdateGcBlobRule")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeUnauthorized)
	}

	daemonService := h.DaemonServiceFactory.New()
	ruleObj, err := daemonService.GetGcBlobRule(ctx)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Error().Err(err).Msg("Get gc tag rule failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Get gc tag rule failed: %v", err))
	}
	if ruleObj != nil && ruleObj.IsRunning {
		log.Error().Msg("The gc blob rule is running")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, "The gc tag rule is running")
	}
	var nextTrigger *int64
	if req.CronRule != nil {
		schedule, _ := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow).Parse(ptr.To(req.CronRule))
		nextTrigger = ptr.Of(schedule.Next(time.Now()).UnixMilli())
	}
	updates := make(map[string]any, 5)
	updates[query.DaemonGcBlobRule.RetentionDay.ColumnName().String()] = req.RetentionDay
	updates[query.DaemonGcBlobRule.CronEnabled.ColumnName().String()] = req.CronEnabled
	if req.CronEnabled {
		updates[query.DaemonGcBlobRule.CronRule.ColumnName().String()] = ptr.To(req.CronRule)
		updates[query.DaemonGcBlobRule.CronNextTrigger.ColumnName().String()] = ptr.To(nextTrigger)
	}
	err = query.Q.Transaction(func(tx *query.Query) error {
		daemonService := h.DaemonServiceFactory.New(tx)
		if ruleObj == nil { // rule not found, we need create the rule
			err = daemonService.CreateGcBlobRule(ctx, &models.DaemonGcBlobRule{
				CronEnabled:     req.CronEnabled,
				RetentionDay:    req.RetentionDay,
				CronRule:        req.CronRule,
				CronNextTrigger: nextTrigger,
			})
			if err != nil {
				log.Error().Err(err).Msg("Create gc blob rule failed")
				return xerrors.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Create gc blob rule failed: %v", err))
			}
			return nil
		}
		err = daemonService.UpdateGcBlobRule(ctx, ruleObj.ID, updates)
		if err != nil {
			log.Error().Err(err).Msg("Update gc blob rule failed")
			return xerrors.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Update gc blob rule failed: %v", err))
		}
		err = h.ProducerClient.Produce(ctx, enums.DaemonWebhook, types.DaemonWebhookPayload{
			Action:       enums.WebhookActionUpdate,
			ResourceType: enums.WebhookResourceTypeDaemonTaskGcBlobRule,
			Payload:      utils.MustMarshal(req),
		}, definition.ProducerOption{Tx: tx})
		if err != nil {
			log.Error().Err(err).Msg("Webhook event produce failed")
			return xerrors.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Webhook event produce failed: %v", err))
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

// GetGcBlobRule handles the get gc blob rule request
//
//	@Summary	Get gc blob rule
//	@security	BasicAuth
//	@Tags		Daemon
//	@Accept		json
//	@Produce	json
//	@Router		/daemons/gc-blob/{namespace_id}/ [get]
//	@Param		namespace_id	path		int64	true	"Namespace id"
//	@Success	200				{object}	types.GetGcBlobRuleResponse
//	@Failure	400				{object}	xerrors.ErrCode
//	@Failure	404				{object}	xerrors.ErrCode
//	@Failure	500				{object}	xerrors.ErrCode
func (h *handler) GetGcBlobRule(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	var req types.GetGcBlobRuleRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, fmt.Sprintf("Bind and validate request body failed: %v", err))
	}

	daemonService := h.DaemonServiceFactory.New()
	ruleObj, err := daemonService.GetGcBlobRule(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Int64("NamespaceID", req.NamespaceID).Msg("Get gc blob rule not found")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, fmt.Sprintf("Get gc blob rule not found: %v", err))
		}
		log.Error().Err(err).Msg("Get gc blob rule failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Get gc blob rule failed: %v", err))
	}
	var nextTrigger *string
	if ruleObj.CronNextTrigger != nil {
		nextTrigger = ptr.Of(time.Unix(0, int64(time.Millisecond)*ptr.To(ruleObj.CronNextTrigger)).UTC().Format(consts.DefaultTimePattern))
	}
	return c.JSON(http.StatusOK, types.GetGcBlobRuleResponse{
		RetentionDay:    ruleObj.RetentionDay,
		CronEnabled:     ruleObj.CronEnabled,
		CronRule:        ruleObj.CronRule,
		CronNextTrigger: nextTrigger,
		CreatedAt:       time.Unix(0, int64(time.Millisecond)*ruleObj.CreatedAt).UTC().Format(consts.DefaultTimePattern),
		UpdatedAt:       time.Unix(0, int64(time.Millisecond)*ruleObj.CreatedAt).UTC().Format(consts.DefaultTimePattern),
	})
}

// GetGcBlobLatestRunner handles the get gc blob latest runner request
//
//	@Summary	Get gc blob latest runner
//	@security	BasicAuth
//	@Tags		Daemon
//	@Accept		json
//	@Produce	json
//	@Router		/daemons/gc-blob/{namespace_id}/runners/latest [get]
//	@Param		namespace_id	path		int64	true	"Namespace id"
//	@Success	200				{object}	types.GcBlobRunnerItem
//	@Failure	400				{object}	xerrors.ErrCode
//	@Failure	404				{object}	xerrors.ErrCode
//	@Failure	500				{object}	xerrors.ErrCode
func (h *handler) GetGcBlobLatestRunner(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	var req types.GetGcBlobLatestRunnerRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, fmt.Sprintf("Bind and validate request body failed: %v", err))
	}
	if req.NamespaceID != 0 {
		log.Error().Msg("NamespaceID should always be 0 in action GetGcBlobLatestRunner")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeUnauthorized)
	}
	daemonService := h.DaemonServiceFactory.New()
	ruleObj, err := daemonService.GetGcBlobRule(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Int64("NamespaceID", req.NamespaceID).Msg("Get gc blob rule not found")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, fmt.Sprintf("Get gc blob rule not found: %v", err))
		}
		log.Error().Err(err).Msg("Get gc blob rule failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Get gc blob rule failed: %v", err))
	}
	runnerObj, err := daemonService.GetGcBlobLatestRunner(ctx, ruleObj.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Msg("Get gc blob latest runner not found")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, fmt.Sprintf("Get gc blob latest runner not found: %v", err))
		}
		log.Error().Err(err).Msg("Get gc blob latest runner failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Get gc blob latest runner failed: %v", err))
	}
	var startedAt, endedAt *string
	if runnerObj.StartedAt != nil {
		startedAt = ptr.Of(time.Unix(0, int64(time.Millisecond)*ptr.To(runnerObj.StartedAt)).UTC().Format(consts.DefaultTimePattern))
	}
	if runnerObj.EndedAt != nil {
		endedAt = ptr.Of(time.Unix(0, int64(time.Millisecond)*ptr.To(runnerObj.EndedAt)).UTC().Format(consts.DefaultTimePattern))
	}
	var duration *string
	if runnerObj.Duration != nil {
		duration = ptr.Of(durafmt.ParseShort(time.Millisecond * time.Duration(ptr.To(runnerObj.Duration))).String())
	}
	return c.JSON(http.StatusOK, types.GcBlobRunnerItem{
		ID:           runnerObj.ID,
		Status:       runnerObj.Status,
		Message:      string(runnerObj.Message),
		FailedCount:  runnerObj.FailedCount,
		SuccessCount: runnerObj.SuccessCount,
		RawDuration:  runnerObj.Duration,
		Duration:     duration,
		StartedAt:    startedAt,
		EndedAt:      endedAt,
		CreatedAt:    time.Unix(0, int64(time.Millisecond)*ruleObj.CreatedAt).UTC().Format(consts.DefaultTimePattern),
		UpdatedAt:    time.Unix(0, int64(time.Millisecond)*ruleObj.CreatedAt).UTC().Format(consts.DefaultTimePattern),
	})
}

// CreateGcBlobRunner handles the create gc blob runner request
//
//	@Summary	Create gc blob runner
//	@security	BasicAuth
//	@Tags		Daemon
//	@Accept		json
//	@Produce	json
//	@Router		/daemons/gc-blob/{namespace_id}/runners/ [post]
//	@Param		namespace_id	path	int64							true	"Namespace id"
//	@Param		message			body	types.CreateGcBlobRunnerRequest	true	"Gc blob runner object"
//	@Success	201
//	@Failure	400	{object}	xerrors.ErrCode
//	@Failure	404	{object}	xerrors.ErrCode
//	@Failure	500	{object}	xerrors.ErrCode
func (h *handler) CreateGcBlobRunner(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	iuser := c.Get(consts.ContextUser)
	if iuser == nil {
		log.Error().Msg("Get user from header failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeUnauthorized)
	}
	user, ok := iuser.(*models.User)
	if !ok {
		log.Error().Msg("Convert user from header failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeUnauthorized)
	}

	var req types.CreateGcBlobRunnerRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, fmt.Sprintf("Bind and validate request body failed: %v", err))
	}
	daemonService := h.DaemonServiceFactory.New()
	ruleObj, err := daemonService.GetGcBlobRule(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Msg("Get gc blob rule not found")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, fmt.Sprintf("Get gc blob rule not found: %v", err))
		}
		log.Error().Err(err).Msg("Get gc blob rule failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Get gc blob rule failed: %v", err))
	}
	if ruleObj != nil && ruleObj.IsRunning {
		log.Error().Msg("The gc blob rule is running")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, "The gc blob rule is running")
	}
	err = query.Q.Transaction(func(tx *query.Query) error {
		runnerObj := &models.DaemonGcBlobRunner{RuleID: ruleObj.ID, Status: enums.TaskCommonStatusPending,
			OperateType:   enums.OperateTypeManual,
			OperateUserID: ptr.Of(user.ID)}
		err = daemonService.CreateGcBlobRunner(ctx, runnerObj)
		if err != nil {
			log.Error().Int64("RuleID", ruleObj.ID).Msgf("Create gc blob runner failed: %v", err)
			return xerrors.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Create gc blob runner failed: %v", err))
		}
		err = workq.ProducerClient.Produce(ctx, enums.DaemonGcBlob,
			types.DaemonGcPayload{RunnerID: runnerObj.ID}, definition.ProducerOption{Tx: tx})
		if err != nil {
			log.Error().Err(err).Msgf("Send topic %s to work queue failed", enums.DaemonGcBlob.String())
			return xerrors.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Send topic %s to work queue failed", enums.DaemonGcBlob.String()))
		}
		err = h.ProducerClient.Produce(ctx, enums.DaemonWebhook, types.DaemonWebhookPayload{
			Action:       enums.WebhookActionCreate,
			ResourceType: enums.WebhookResourceTypeDaemonTaskGcBlobRunner,
			Payload:      utils.MustMarshal(req),
		}, definition.ProducerOption{Tx: tx})
		if err != nil {
			log.Error().Err(err).Msg("Webhook event produce failed")
			return xerrors.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Webhook event produce failed: %v", err))
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

// ListGcBlobRunners handles the list gc blob runners request
//
//	@Summary	List gc blob runners
//	@security	BasicAuth
//	@Tags		Daemon
//	@Accept		json
//	@Produce	json
//	@Router		/daemons/gc-blob/{namespace_id}/runners/ [get]
//	@Param		namespace_id	path		int64	true	"Namespace id"
//	@Param		limit			query		int64	false	"limit"	minimum(10)	maximum(100)	default(10)
//	@Param		page			query		int64	false	"page"	minimum(1)	default(1)
//	@Param		sort			query		string	false	"sort field"
//	@Param		method			query		string	false	"sort method"	Enums(asc, desc)
//	@Success	200				{object}	types.CommonList{items=[]types.GcBlobRunnerItem}
//	@Failure	400				{object}	xerrors.ErrCode
//	@Failure	404				{object}	xerrors.ErrCode
//	@Failure	500				{object}	xerrors.ErrCode
func (h *handler) ListGcBlobRunners(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	var req types.ListGcBlobRunnersRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, fmt.Sprintf("Bind and validate request body failed: %v", err))
	}
	daemonService := h.DaemonServiceFactory.New()
	ruleObj, err := daemonService.GetGcBlobRule(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Int64("NamespaceID", req.NamespaceID).Msg("Get gc blob rule not found")
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
		var startedAt, endedAt *string
		if runnerObj.StartedAt != nil {
			startedAt = ptr.Of(time.Unix(0, int64(time.Millisecond)*ptr.To(runnerObj.StartedAt)).UTC().Format(consts.DefaultTimePattern))
		}
		if runnerObj.EndedAt != nil {
			endedAt = ptr.Of(time.Unix(0, int64(time.Millisecond)*ptr.To(runnerObj.EndedAt)).UTC().Format(consts.DefaultTimePattern))
		}
		var duration *string
		if runnerObj.Duration != nil {
			duration = ptr.Of(durafmt.ParseShort(time.Millisecond * time.Duration(ptr.To(runnerObj.Duration))).String())
		}
		resp = append(resp, types.GcBlobRunnerItem{
			ID:           runnerObj.ID,
			Status:       runnerObj.Status,
			Message:      string(runnerObj.Message),
			SuccessCount: runnerObj.SuccessCount,
			FailedCount:  runnerObj.FailedCount,
			RawDuration:  runnerObj.Duration,
			Duration:     duration,
			StartedAt:    startedAt,
			EndedAt:      endedAt,
			CreatedAt:    time.Unix(0, int64(time.Millisecond)*runnerObj.CreatedAt).UTC().Format(consts.DefaultTimePattern),
			UpdatedAt:    time.Unix(0, int64(time.Millisecond)*runnerObj.CreatedAt).UTC().Format(consts.DefaultTimePattern),
		})
	}
	return c.JSON(http.StatusOK, types.CommonList{Total: total, Items: resp})
}

// GetGcBlobRunner handles the get gc blob runner request
//
//	@Summary	List gc blob runners
//	@security	BasicAuth
//	@Tags		Daemon
//	@Accept		json
//	@Produce	json
//	@Router		/daemons/gc-blob/{namespace_id}/runners/{runner_id} [get]
//	@Param		namespace_id	path		int64	true	"Namespace id"
//	@Param		runner_id		path		int64	true	"Runner id"
//	@Success	200				{object}	types.GcBlobRunnerItem
//	@Failure	400				{object}	xerrors.ErrCode
//	@Failure	404				{object}	xerrors.ErrCode
//	@Failure	500				{object}	xerrors.ErrCode
func (h *handler) GetGcBlobRunner(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	var req types.GetGcBlobRunnerRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, fmt.Sprintf("Bind and validate request body failed: %v", err))
	}
	daemonService := h.DaemonServiceFactory.New()
	runnerObj, err := daemonService.GetGcBlobRunner(ctx, req.RunnerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Int64("NamespaceID", req.NamespaceID).Int64("runnerID", req.RunnerID).Msg("Get gc tag runner not found")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, fmt.Sprintf("Get gc tag runner not found: %v", err))
		}
		log.Error().Err(err).Msg("Get gc tag runner failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Get gc tag runner failed: %v", err))
	}
	var startedAt, endedAt *string
	if runnerObj.StartedAt != nil {
		startedAt = ptr.Of(time.Unix(0, int64(time.Millisecond)*ptr.To(runnerObj.StartedAt)).UTC().Format(consts.DefaultTimePattern))
	}
	if runnerObj.EndedAt != nil {
		endedAt = ptr.Of(time.Unix(0, int64(time.Millisecond)*ptr.To(runnerObj.EndedAt)).UTC().Format(consts.DefaultTimePattern))
	}
	var duration *string
	if runnerObj.Duration != nil {
		duration = ptr.Of(durafmt.ParseShort(time.Millisecond * time.Duration(ptr.To(runnerObj.Duration))).String())
	}
	return c.JSON(http.StatusOK, types.GcBlobRunnerItem{
		ID:           runnerObj.ID,
		Status:       runnerObj.Status,
		Message:      string(runnerObj.Message),
		SuccessCount: runnerObj.SuccessCount,
		FailedCount:  runnerObj.FailedCount,
		RawDuration:  runnerObj.Duration,
		Duration:     duration,
		StartedAt:    startedAt,
		EndedAt:      endedAt,
		CreatedAt:    time.Unix(0, int64(time.Millisecond)*runnerObj.CreatedAt).UTC().Format(consts.DefaultTimePattern),
		UpdatedAt:    time.Unix(0, int64(time.Millisecond)*runnerObj.CreatedAt).UTC().Format(consts.DefaultTimePattern),
	})
}

// ListGcBlobRecords handles the list gc blob records request
//
//	@Summary	List gc blob records
//	@security	BasicAuth
//	@Tags		Daemon
//	@Accept		json
//	@Produce	json
//	@Router		/daemons/gc-blob/{namespace_id}/runners/{runner_id}/records/ [get]
//	@Param		namespace_id	path		int64	true	"Namespace id"
//	@Param		runner_id		path		int64	true	"Runner id"
//	@Param		limit			query		int64	false	"limit"	minimum(10)	maximum(100)	default(10)
//	@Param		page			query		int64	false	"page"	minimum(1)	default(1)
//	@Param		sort			query		string	false	"sort field"
//	@Param		method			query		string	false	"sort method"	Enums(asc, desc)
//	@Success	200				{object}	types.CommonList{items=[]types.GcBlobRecordItem}
//	@Failure	400				{object}	xerrors.ErrCode
//	@Failure	404				{object}	xerrors.ErrCode
//	@Failure	500				{object}	xerrors.ErrCode
func (h *handler) ListGcBlobRecords(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	var req types.ListGcBlobRecordsRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, fmt.Sprintf("Bind and validate request body failed: %v", err))
	}
	daemonService := h.DaemonServiceFactory.New()
	recordObjs, total, err := daemonService.ListGcBlobRecords(ctx, req.RunnerID, req.Pagination, req.Sortable)
	if err != nil {
		log.Error().Err(err).Int64("RuleID", req.RunnerID).Msgf("List gc blob records failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("List gc blob records failed: %v", err))
	}
	var resp = make([]any, 0, len(recordObjs))
	for _, recordObj := range recordObjs {
		resp = append(resp, types.GcBlobRecordItem{
			ID:        recordObj.ID,
			Digest:    recordObj.Digest,
			Status:    recordObj.Status,
			Message:   string(recordObj.Message),
			CreatedAt: time.Unix(0, int64(time.Millisecond)*recordObj.CreatedAt).UTC().Format(consts.DefaultTimePattern),
			UpdatedAt: time.Unix(0, int64(time.Millisecond)*recordObj.CreatedAt).UTC().Format(consts.DefaultTimePattern),
		})
	}
	return c.JSON(http.StatusOK, types.CommonList{Total: total, Items: resp})
}

// GetGcBlobRecord handles the get gc blob record request
//
//	@Summary	Get gc blob record
//	@security	BasicAuth
//	@Tags		Daemon
//	@Accept		json
//	@Produce	json
//	@Router		/daemons/gc-blob/{namespace_id}/runners/{runner_id}/records/{record_id} [get]
//	@Param		namespace_id	path		int64	true	"Namespace id"
//	@Param		runner_id		path		int64	true	"Runner id"
//	@Param		record_id		path		int64	true	"Record id"
//	@Success	200				{object}	types.GcBlobRecordItem
//	@Failure	400				{object}	xerrors.ErrCode
//	@Failure	404				{object}	xerrors.ErrCode
//	@Failure	500				{object}	xerrors.ErrCode
func (h *handler) GetGcBlobRecord(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	var req types.GetGcBlobRecordRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, fmt.Sprintf("Bind and validate request body failed: %v", err))
	}
	daemonService := h.DaemonServiceFactory.New()
	ruleObj, err := daemonService.GetGcBlobRule(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Int64("NamespaceID", req.NamespaceID).Msg("Get gc blob rule not found")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, fmt.Sprintf("Get gc blob rule not found: %v", err))
		}
		log.Error().Err(err).Msg("Get gc blob rule failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Get gc blob rule failed: %v", err))
	}
	recordObj, err := daemonService.GetGcBlobRecord(ctx, req.RecordID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Int64("NamespaceID", req.NamespaceID).Int64("runnerID", req.RunnerID).Msg("Get gc blob record not found")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, fmt.Sprintf("Get gc blob record not found: %v", err))
		}
		log.Error().Err(err).Msg("Get gc blob record failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Get gc blob record failed: %v", err))
	}
	if recordObj.Runner.ID != req.RunnerID || recordObj.Runner.Rule.ID != ruleObj.ID {
		log.Error().Err(err).Int64("NamespaceID", req.NamespaceID).Int64("runnerID", req.RunnerID).Msg("Get gc blob record not found")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, fmt.Sprintf("Get gc blob record not found: %v", err))
	}
	return c.JSON(http.StatusOK, types.GcBlobRecordItem{
		ID:        recordObj.ID,
		Digest:    recordObj.Digest,
		Status:    recordObj.Status,
		Message:   string(recordObj.Message),
		CreatedAt: time.Unix(0, int64(time.Millisecond)*recordObj.CreatedAt).UTC().Format(consts.DefaultTimePattern),
		UpdatedAt: time.Unix(0, int64(time.Millisecond)*recordObj.CreatedAt).UTC().Format(consts.DefaultTimePattern),
	})
}
