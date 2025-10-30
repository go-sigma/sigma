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
	"github.com/go-sigma/sigma/pkg/server/errcode"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

// UpdateGcArtifactRule handles the update gc artifact rule request
//
//	@Summary	Update gc artifact rule
//	@security	BasicAuth
//	@Tags		Daemon
//	@Accept		json
//	@Produce	json
//	@Router		/daemons/gc-artifact/{namespace_id}/ [put]
//	@Param		namespace_id	path	int64								true	"Namespace id"
//	@Param		message			body	types.UpdateGcArtifactRuleRequest	true	"Gc artifact rule object"
//	@Success	204
//	@Failure	400	{object}	errcode.ErrCode
//	@Failure	404	{object}	errcode.ErrCode
//	@Failure	500	{object}	errcode.ErrCode
func (h *handler) UpdateGcArtifactRule(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	var req types.UpdateGcArtifactRuleRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return errcode.NewHTTPError(c, errcode.HTTPErrCodeBadRequest, fmt.Sprintf("Bind and validate request body failed: %v", err))
	}
	var namespaceID *int64
	if req.NamespaceID != 0 {
		namespaceID = ptr.Of(req.NamespaceID)
	}
	daemonService := h.DaemonServiceFactory.New()
	ruleObj, err := daemonService.GetGcArtifactRule(ctx, namespaceID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Error().Err(err).Msg("Get gc artifact rule failed")
		return errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError, fmt.Sprintf("Get gc artifact rule failed: %v", err))
	}
	if ruleObj != nil && ruleObj.IsRunning {
		log.Error().Int64("NamespaceID", ptr.To(namespaceID)).Msg("The gc artifact rule is running")
		return errcode.NewHTTPError(c, errcode.HTTPErrCodeBadRequest, "The gc artifact rule is running")
	}
	var nextTrigger *int64
	if req.CronRule != nil {
		schedule, _ := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow).Parse(ptr.To(req.CronRule))
		nextTrigger = ptr.Of(schedule.Next(time.Now()).UnixMilli()) // TODO: set unix
	}
	updates := make(map[string]any, 5)
	updates[query.DaemonGcArtifactRule.RetentionDay.ColumnName().String()] = req.RetentionDay
	updates[query.DaemonGcArtifactRule.CronEnabled.ColumnName().String()] = req.CronEnabled
	if req.CronEnabled {
		updates[query.DaemonGcArtifactRule.CronRule.ColumnName().String()] = ptr.To(req.CronRule)
		updates[query.DaemonGcArtifactRule.CronNextTrigger.ColumnName().String()] = ptr.To(nextTrigger)
	}
	err = query.Q.Transaction(func(tx *query.Query) error {
		daemonService := h.DaemonServiceFactory.New(tx)
		if ruleObj == nil { // rule not found, we need create the rule
			err = daemonService.CreateGcArtifactRule(ctx, &models.DaemonGcArtifactRule{
				NamespaceID:     namespaceID,
				RetentionDay:    req.RetentionDay,
				CronEnabled:     req.CronEnabled,
				CronRule:        req.CronRule,
				CronNextTrigger: nextTrigger,
			})
			if err != nil {
				log.Error().Err(err).Msg("Create gc artifact rule failed")
				return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Create gc artifact rule failed: %v", err))
			}
			return nil
		}
		err = daemonService.UpdateGcArtifactRule(ctx, ruleObj.ID, updates)
		if err != nil {
			log.Error().Err(err).Msg("Update gc artifact rule failed")
			return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Update gc artifact rule failed: %v", err))
		}
		err = h.Producer.Produce(ctx, enums.DaemonWebhook, types.DaemonWebhookPayload{
			NamespaceID:  namespaceID,
			Action:       enums.WebhookActionUpdate,
			ResourceType: enums.WebhookResourceTypeDaemonTaskGcArtifactRule,
			Payload:      utils.MustMarshal(req),
		}, workq.ProducerOption{Tx: tx})
		if err != nil {
			log.Error().Err(err).Msg("Webhook event produce failed")
			return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Webhook event produce failed: %v", err))
		}
		return nil
	})
	if err != nil {
		var e errcode.ErrCode
		if errors.As(err, &e) {
			return errcode.NewHTTPError(c, e)
		}
		return errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError)
	}
	return c.NoContent(http.StatusNoContent)
}

// GetGcArtifactRule handles the get gc artifact rule request
//
//	@Summary	Get gc artifact rule
//	@security	BasicAuth
//	@Tags		Daemon
//	@Accept		json
//	@Produce	json
//	@Router		/daemons/gc-artifact/{namespace_id}/ [get]
//	@Param		namespace_id	path		int64	true	"Namespace id"
//	@Success	200				{object}	types.GetGcArtifactRuleResponse
//	@Failure	400				{object}	errcode.ErrCode
//	@Failure	404				{object}	errcode.ErrCode
//	@Failure	500				{object}	errcode.ErrCode
func (h *handler) GetGcArtifactRule(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	var req types.GetGcArtifactRuleRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return errcode.NewHTTPError(c, errcode.HTTPErrCodeBadRequest, fmt.Sprintf("Bind and validate request body failed: %v", err))
	}
	var namespaceID *int64
	if req.NamespaceID != 0 {
		namespaceID = ptr.Of(req.NamespaceID)
	}
	daemonService := h.DaemonServiceFactory.New()
	ruleObj, err := daemonService.GetGcArtifactRule(ctx, namespaceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Int64("NamespaceID", req.NamespaceID).Msg("Get gc artifact rule not found")
			return errcode.NewHTTPError(c, errcode.HTTPErrCodeNotFound, fmt.Sprintf("Get gc artifact rule not found: %v", err))
		}
		log.Error().Err(err).Msg("Get gc artifact rule failed")
		return errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError, fmt.Sprintf("Get gc artifact rule failed: %v", err))
	}
	var nextTrigger *string
	if ruleObj.CronNextTrigger != nil {
		nextTrigger = ptr.Of(time.Unix(0, int64(time.Millisecond)*ptr.To(ruleObj.CronNextTrigger)).UTC().Format(consts.DefaultTimePattern))
	}
	return c.JSON(http.StatusOK, types.GetGcArtifactRuleResponse{
		RetentionDay:    ruleObj.RetentionDay,
		CronEnabled:     ruleObj.CronEnabled,
		CronRule:        ruleObj.CronRule,
		CronNextTrigger: nextTrigger,
		CreatedAt:       time.Unix(0, int64(time.Millisecond)*ruleObj.CreatedAt).UTC().Format(consts.DefaultTimePattern),
		UpdatedAt:       time.Unix(0, int64(time.Millisecond)*ruleObj.CreatedAt).UTC().Format(consts.DefaultTimePattern),
	})
}

// GetGcArtifactLatestRunner handles the get gc artifact latest runner request
//
//	@Summary	Get gc artifact latest runner
//	@security	BasicAuth
//	@Tags		Daemon
//	@Accept		json
//	@Produce	json
//	@Router		/daemons/gc-artifact/{namespace_id}/runners/latest [get]
//	@Param		namespace_id	path		int64	true	"Namespace id"
//	@Success	200				{object}	types.GcArtifactRunnerItem
//	@Failure	400				{object}	errcode.ErrCode
//	@Failure	404				{object}	errcode.ErrCode
//	@Failure	500				{object}	errcode.ErrCode
func (h *handler) GetGcArtifactLatestRunner(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	var req types.GetGcArtifactLatestRunnerRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return errcode.NewHTTPError(c, errcode.HTTPErrCodeBadRequest, fmt.Sprintf("Bind and validate request body failed: %v", err))
	}
	var namespaceID *int64
	if req.NamespaceID != 0 {
		namespaceID = ptr.Of(req.NamespaceID)
	}
	daemonService := h.DaemonServiceFactory.New()
	ruleObj, err := daemonService.GetGcArtifactRule(ctx, namespaceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Int64("NamespaceID", req.NamespaceID).Msg("Get gc artifact rule not found")
			return errcode.NewHTTPError(c, errcode.HTTPErrCodeNotFound, fmt.Sprintf("Get gc artifact rule not found: %v", err))
		}
		log.Error().Err(err).Msg("Get gc artifact rule failed")
		return errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError, fmt.Sprintf("Get gc artifact rule failed: %v", err))
	}
	runnerObj, err := daemonService.GetGcArtifactLatestRunner(ctx, ruleObj.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Int64("NamespaceID", req.NamespaceID).Msg("Get gc artifact runner not found")
			return errcode.NewHTTPError(c, errcode.HTTPErrCodeNotFound, fmt.Sprintf("Get gc artifact runner not found: %v", err))
		}
		log.Error().Err(err).Msg("Get gc artifact runner failed")
		return errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError, fmt.Sprintf("Get gc artifact runner failed: %v", err))
	}
	base := buildRunnerItemBase(runnerObj)
	return c.JSON(http.StatusOK, types.GcArtifactRunnerItem{
		ID:           base.ID,
		Status:       base.Status,
		Message:      base.Message,
		FailedCount:  base.FailedCount,
		SuccessCount: base.SuccessCount,
		RawDuration:  base.RawDuration,
		Duration:     base.Duration,
		StartedAt:    base.StartedAt,
		EndedAt:      base.EndedAt,
		CreatedAt:    base.CreatedAt,
		UpdatedAt:    base.UpdatedAt,
	})
}

// CreateGcArtifactRunner handles the create gc artifact runner request
//
//	@Summary	Create gc artifact runner
//	@security	BasicAuth
//	@Tags		Daemon
//	@Accept		json
//	@Produce	json
//	@Router		/daemons/gc-artifact/{namespace_id}/runners/ [post]
//	@Param		namespace_id	path	int64								true	"Namespace id"
//	@Param		message			body	types.CreateGcArtifactRunnerRequest	true	"Gc artifact runner object"
//	@Success	201
//	@Failure	400	{object}	errcode.ErrCode
//	@Failure	404	{object}	errcode.ErrCode
//	@Failure	500	{object}	errcode.ErrCode
func (h *handler) CreateGcArtifactRunner(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	var req types.CreateGcArtifactRunnerRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return errcode.NewHTTPError(c, errcode.HTTPErrCodeBadRequest, fmt.Sprintf("Bind and validate request body failed: %v", err))
	}
	var namespaceID *int64
	if req.NamespaceID != 0 {
		namespaceID = ptr.Of(req.NamespaceID)
	}
	daemonService := h.DaemonServiceFactory.New()
	ruleObj, err := daemonService.GetGcArtifactRule(ctx, namespaceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Int64("namespaceID", req.NamespaceID).Msg("Get gc artifact rule not found")
			return errcode.NewHTTPError(c, errcode.HTTPErrCodeNotFound, fmt.Sprintf("Get gc artifact rule not found: %v", err))
		}
		log.Error().Err(err).Msg("Get gc artifact rule failed")
		return errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError, fmt.Sprintf("Get gc artifact rule failed: %v", err))
	}
	if ruleObj != nil && ruleObj.IsRunning {
		log.Error().Int64("NamespaceID", ptr.To(namespaceID)).Msg("The gc artifact rule is running")
		return errcode.NewHTTPError(c, errcode.HTTPErrCodeBadRequest, "The gc artifact rule is running")
	}
	err = query.Q.Transaction(func(tx *query.Query) error {
		runnerObj := &models.DaemonGcArtifactRunner{RuleID: ruleObj.ID, Status: enums.TaskCommonStatusPending, OperateType: enums.OperateTypeManual}
		err = daemonService.CreateGcArtifactRunner(ctx, runnerObj)
		if err != nil {
			log.Error().Int64("ruleID", ruleObj.ID).Msgf("Create gc artifact runner failed: %v", err)
			return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Create gc artifact runner failed: %v", err))
		}
		err = h.Producer.Produce(ctx, enums.DaemonGcArtifact,
			types.DaemonGcPayload{RunnerID: runnerObj.ID}, workq.ProducerOption{Tx: tx})
		if err != nil {
			log.Error().Err(err).Msgf("Send topic %s to work queue failed", enums.DaemonGcArtifact.String())
			return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Send topic %s to work queue failed", enums.DaemonGcArtifact.String()))
		}
		err = h.Producer.Produce(ctx, enums.DaemonWebhook, types.DaemonWebhookPayload{
			NamespaceID:  namespaceID,
			Action:       enums.WebhookActionCreate,
			ResourceType: enums.WebhookResourceTypeDaemonTaskGcArtifactRunner,
			Payload:      utils.MustMarshal(req),
		}, workq.ProducerOption{Tx: tx})
		if err != nil {
			log.Error().Err(err).Msg("Webhook event produce failed")
			return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Webhook event produce failed: %v", err))
		}
		return nil
	})
	if err != nil {
		var e errcode.ErrCode
		if errors.As(err, &e) {
			return errcode.NewHTTPError(c, e)
		}
		return errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError)
	}
	return c.NoContent(http.StatusCreated)
}

// ListGcArtifactRunners handles the list gc artifact runners request
//
//	@Summary	List gc artifact runners
//	@security	BasicAuth
//	@Tags		Daemon
//	@Accept		json
//	@Produce	json
//	@Router		/daemons/gc-artifact/{namespace_id}/runners/ [get]
//	@Param		namespace_id	path		int64	true	"Namespace id"
//	@Param		limit			query		int64	false	"limit"	minimum(10)	maximum(100)	default(10)
//	@Param		page			query		int64	false	"page"	minimum(1)	default(1)
//	@Param		sort			query		string	false	"sort field"
//	@Param		method			query		string	false	"sort method"	Enums(asc, desc)
//	@Success	200				{object}	types.CommonList{items=[]types.GcArtifactRunnerItem}
//	@Failure	400				{object}	errcode.ErrCode
//	@Failure	404				{object}	errcode.ErrCode
//	@Failure	500				{object}	errcode.ErrCode
func (h *handler) ListGcArtifactRunners(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	var req types.ListGcArtifactRunnersRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return errcode.NewHTTPError(c, errcode.HTTPErrCodeBadRequest, fmt.Sprintf("Bind and validate request body failed: %v", err))
	}
	var namespaceID *int64
	if req.NamespaceID != 0 {
		namespaceID = ptr.Of(req.NamespaceID)
	}
	daemonService := h.DaemonServiceFactory.New()
	ruleObj, err := daemonService.GetGcArtifactRule(ctx, namespaceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Int64("namespaceID", req.NamespaceID).Msg("Get gc artifact rule not found")
			return errcode.NewHTTPError(c, errcode.HTTPErrCodeNotFound, fmt.Sprintf("Get gc artifact rule not found: %v", err))
		}
		log.Error().Err(err).Msg("Get gc artifact rule failed")
		return errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError, fmt.Sprintf("Get gc artifact rule failed: %v", err))
	}
	runnerObjs, total, err := daemonService.ListGcArtifactRunners(ctx, ruleObj.ID, req.Pagination, req.Sortable)
	if err != nil {
		log.Error().Err(err).Msg("List gc artifact rules failed")
		return errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError, fmt.Sprintf("List gc artifact rules failed: %v", err))
	}
	var resp = make([]any, 0, len(runnerObjs))
	for _, runnerObj := range runnerObjs {
		base := buildRunnerItemBase(runnerObj)
		resp = append(resp, types.GcArtifactRunnerItem{
			ID:           base.ID,
			Status:       base.Status,
			Message:      base.Message,
			SuccessCount: base.SuccessCount,
			FailedCount:  base.FailedCount,
			RawDuration:  base.RawDuration,
			Duration:     base.Duration,
			StartedAt:    base.StartedAt,
			EndedAt:      base.EndedAt,
			CreatedAt:    base.CreatedAt,
			UpdatedAt:    base.UpdatedAt,
		})
	}
	return c.JSON(http.StatusOK, types.CommonList{Total: total, Items: resp})
}

// GetGcArtifactRunner handles the get gc artifact runner request
//
//	@Summary	List gc artifact runners
//	@security	BasicAuth
//	@Tags		Daemon
//	@Accept		json
//	@Produce	json
//	@Router		/daemons/gc-artifact/{namespace_id}/runners/{runner_id} [get]
//	@Param		namespace_id	path		int64	true	"Namespace id"
//	@Param		runner_id		path		int64	true	"Runner id"
//	@Success	200				{object}	types.GcArtifactRunnerItem
//	@Failure	400				{object}	errcode.ErrCode
//	@Failure	404				{object}	errcode.ErrCode
//	@Failure	500				{object}	errcode.ErrCode
func (h *handler) GetGcArtifactRunner(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	var req types.GetGcArtifactRunnerRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return errcode.NewHTTPError(c, errcode.HTTPErrCodeBadRequest, fmt.Sprintf("Bind and validate request body failed: %v", err))
	}
	daemonService := h.DaemonServiceFactory.New()
	runnerObj, err := daemonService.GetGcArtifactRunner(ctx, req.RunnerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Int64("namespaceID", req.NamespaceID).Int64("runnerID", req.RunnerID).Msg("Get gc artifact runner not found")
			return errcode.NewHTTPError(c, errcode.HTTPErrCodeNotFound, fmt.Sprintf("Get gc artifact runner not found: %v", err))
		}
		log.Error().Err(err).Msg("Get gc artifact runner failed")
		return errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError, fmt.Sprintf("Get gc artifact runner failed: %v", err))
	}
	if ptr.To(runnerObj.Rule.NamespaceID) != req.NamespaceID {
		log.Error().Err(err).Int64("namespaceID", req.NamespaceID).Int64("runnerID", req.RunnerID).Msg("Get gc artifact runner not found")
		return errcode.NewHTTPError(c, errcode.HTTPErrCodeNotFound, fmt.Sprintf("Get gc artifact runner not found: %v", err))
	}
	base := buildRunnerItemBase(runnerObj)
	return c.JSON(http.StatusOK, types.GcArtifactRunnerItem{
		ID:           base.ID,
		Status:       base.Status,
		Message:      base.Message,
		SuccessCount: base.SuccessCount,
		FailedCount:  base.FailedCount,
		RawDuration:  base.RawDuration,
		Duration:     base.Duration,
		StartedAt:    base.StartedAt,
		EndedAt:      base.EndedAt,
		CreatedAt:    base.CreatedAt,
		UpdatedAt:    base.UpdatedAt,
	})
}

// ListGcArtifactRecords handles the list gc artifact records request
//
//	@Summary	List gc artifact records
//	@security	BasicAuth
//	@Tags		Daemon
//	@Accept		json
//	@Produce	json
//	@Router		/daemons/gc-artifact/{namespace_id}/runners/{runner_id}/records/ [get]
//	@Param		namespace_id	path		int64	true	"Namespace id"
//	@Param		runner_id		path		int64	true	"Runner id"
//	@Param		limit			query		int64	false	"limit"	minimum(10)	maximum(100)	default(10)
//	@Param		page			query		int64	false	"page"	minimum(1)	default(1)
//	@Param		sort			query		string	false	"sort field"
//	@Param		method			query		string	false	"sort method"	Enums(asc, desc)
//	@Success	200				{object}	types.CommonList{items=[]types.GcArtifactRecordItem}
//	@Failure	400				{object}	errcode.ErrCode
//	@Failure	404				{object}	errcode.ErrCode
//	@Failure	500				{object}	errcode.ErrCode
func (h *handler) ListGcArtifactRecords(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	var req types.ListGcArtifactRecordsRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return errcode.NewHTTPError(c, errcode.HTTPErrCodeBadRequest, fmt.Sprintf("Bind and validate request body failed: %v", err))
	}
	daemonService := h.DaemonServiceFactory.New()
	recordObjs, total, err := daemonService.ListGcArtifactRecords(ctx, req.RunnerID, req.Pagination, req.Sortable)
	if err != nil {
		log.Error().Err(err).Int64("ruleID", req.RunnerID).Msgf("List gc artifact records failed")
		return errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError, fmt.Sprintf("List gc artifact records failed: %v", err))
	}
	var resp = make([]any, 0, len(recordObjs))
	for _, recordObj := range recordObjs {
		resp = append(resp, types.GcArtifactRecordItem{
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

// GetGcArtifactRecord handles the get gc artifact record request
//
//	@Summary	Get gc artifact record
//	@security	BasicAuth
//	@Tags		Daemon
//	@Accept		json
//	@Produce	json
//	@Router		/daemons/gc-artifact/{namespace_id}/runners/{runner_id}/records/{record_id} [get]
//	@Param		namespace_id	path		int64	true	"Namespace id"
//	@Param		runner_id		path		int64	true	"Runner id"
//	@Param		record_id		path		int64	true	"Record id"
//	@Success	200				{object}	types.GcArtifactRecordItem
//	@Failure	400				{object}	errcode.ErrCode
//	@Failure	404				{object}	errcode.ErrCode
//	@Failure	500				{object}	errcode.ErrCode
func (h *handler) GetGcArtifactRecord(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	var req types.GetGcArtifactRecordRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return errcode.NewHTTPError(c, errcode.HTTPErrCodeBadRequest, fmt.Sprintf("Bind and validate request body failed: %v", err))
	}
	var namespaceID *int64
	if req.NamespaceID != 0 {
		namespaceID = ptr.Of(req.NamespaceID)
	}
	daemonService := h.DaemonServiceFactory.New()
	ruleObj, err := daemonService.GetGcArtifactRule(ctx, namespaceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Int64("namespaceID", req.NamespaceID).Msg("Get gc artifact rule not found")
			return errcode.NewHTTPError(c, errcode.HTTPErrCodeNotFound, fmt.Sprintf("Get gc artifact rule not found: %v", err))
		}
		log.Error().Err(err).Msg("Get gc artifact rule failed")
		return errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError, fmt.Sprintf("Get gc artifact rule failed: %v", err))
	}
	recordObj, err := daemonService.GetGcArtifactRecord(ctx, req.RecordID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Int64("namespaceID", req.NamespaceID).Int64("runnerID", req.RunnerID).Msg("Get gc artifact record not found")
			return errcode.NewHTTPError(c, errcode.HTTPErrCodeNotFound, fmt.Sprintf("Get gc artifact record not found: %v", err))
		}
		log.Error().Err(err).Msg("Get gc artifact record failed")
		return errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError, fmt.Sprintf("Get gc artifact record failed: %v", err))
	}
	if recordObj.Runner.ID != req.RunnerID || recordObj.Runner.Rule.ID != ruleObj.ID {
		log.Error().Err(err).Int64("namespaceID", req.NamespaceID).Int64("runnerID", req.RunnerID).Msg("Get gc artifact record not found")
		return errcode.NewHTTPError(c, errcode.HTTPErrCodeNotFound, fmt.Sprintf("Get gc artifact record not found: %v", err))
	}
	return c.JSON(http.StatusOK, types.GcArtifactRecordItem{
		ID:        recordObj.ID,
		Digest:    recordObj.Digest,
		Status:    recordObj.Status,
		Message:   string(recordObj.Message),
		CreatedAt: time.Unix(0, int64(time.Millisecond)*recordObj.CreatedAt).UTC().Format(consts.DefaultTimePattern),
		UpdatedAt: time.Unix(0, int64(time.Millisecond)*recordObj.CreatedAt).UTC().Format(consts.DefaultTimePattern),
	})
}
