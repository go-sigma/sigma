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
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hako/durafmt"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/server/errcode"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

// UpdateGcTagRule handles the update gc tag rule request
//
//	@Summary	Update gc tag rule
//	@security	BasicAuth
//	@Tags		Daemon
//	@Accept		json
//	@Produce	json
//	@Router		/daemons/gc-tag/{namespace_id}/ [put]
//	@Param		namespace_id	path	int64						true	"Namespace id"
//	@Param		message			body	api.UpdateGcTagRuleRequest	true	"Gc tag rule object"
//	@Success	204
//	@Failure	400	{object}	errcode.ErrCode
//	@Failure	404	{object}	errcode.ErrCode
//	@Failure	500	{object}	errcode.ErrCode
func (h *handler) UpdateGcTagRule(c *gin.Context, req *api.UpdateGcTagRuleRequest) {
	ctx := c.Request.Context()

	err := h.DaemonSvc.UpdateGcTagRule(ctx, *req)
	if err != nil {
		if e, ok := errcode.AsType[errcode.ErrCode](err); ok {
			errcode.NewHTTPError(c, e)
			return
		}
		errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError, fmt.Sprintf("Update gc tag rule failed: %v", err))
		return
	}
	c.Status(http.StatusNoContent)
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
//	@Success	200				{object}	api.GetGcTagRuleResponse
//	@Failure	400				{object}	errcode.ErrCode
//	@Failure	404				{object}	errcode.ErrCode
//	@Failure	500				{object}	errcode.ErrCode
func (h *handler) GetGcTagRule(c *gin.Context, req *api.GetGcTagRuleRequest) {
	ctx := c.Request.Context()

	ruleObj, err := h.DaemonSvc.GetGcTagRule(ctx, req.NamespaceID)
	if err != nil {
		if e, ok := errcode.AsType[errcode.ErrCode](err); ok {
			errcode.NewHTTPError(c, e)
			return
		}
		errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError, fmt.Sprintf("Get gc tag rule failed: %v", err))
		return
	}
	var nextTrigger *string
	if ruleObj.CronNextTrigger != nil {
		nextTrigger = new(time.Unix(0, int64(time.Millisecond)*ptr.To(ruleObj.CronNextTrigger)).UTC().Format(consts.DefaultTimePattern))
	}
	c.JSON(http.StatusOK, api.GetGcTagRuleResponse{
		CronEnabled:         ruleObj.CronEnabled,
		CronRule:            ruleObj.CronRule,
		CronNextTrigger:     nextTrigger,
		RetentionRuleType:   ruleObj.RetentionRuleType,
		RetentionRuleAmount: ruleObj.RetentionRuleAmount,
		RetentionPattern:    ruleObj.RetentionPattern,
		CreatedAt:           time.Unix(0, int64(time.Millisecond)*ruleObj.CreatedAt).UTC().Format(consts.DefaultTimePattern),
		UpdatedAt:           time.Unix(0, int64(time.Millisecond)*ruleObj.CreatedAt).UTC().Format(consts.DefaultTimePattern),
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
//	@Success	200				{object}	api.GcTagRunnerItem
//	@Failure	400				{object}	errcode.ErrCode
//	@Failure	404				{object}	errcode.ErrCode
//	@Failure	500				{object}	errcode.ErrCode
func (h *handler) GetGcTagLatestRunner(c *gin.Context, req *api.GetGcTagLatestRunnerRequest) {
	ctx := c.Request.Context()

	runnerObj, err := h.DaemonSvc.GetGcTagLatestRunner(ctx, req.NamespaceID)
	if err != nil {
		if e, ok := errcode.AsType[errcode.ErrCode](err); ok {
			errcode.NewHTTPError(c, e)
			return
		}
		errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError, fmt.Sprintf("Get gc tag latest runner failed: %v", err))
		return
	}
	var startedAt, endedAt *string
	if runnerObj.StartedAt != nil {
		startedAt = new(time.Unix(0, int64(time.Millisecond)*ptr.To(runnerObj.StartedAt)).UTC().Format(consts.DefaultTimePattern))
	}
	if runnerObj.EndedAt != nil {
		endedAt = new(time.Unix(0, int64(time.Millisecond)*ptr.To(runnerObj.EndedAt)).UTC().Format(consts.DefaultTimePattern))
	}
	var duration *string
	if runnerObj.Duration != nil {
		duration = new(durafmt.ParseShort(time.Millisecond * time.Duration(ptr.To(runnerObj.Duration))).String())
	}
	c.JSON(http.StatusOK, api.GcTagRunnerItem{
		ID:           runnerObj.ID,
		Status:       runnerObj.Status,
		Message:      string(runnerObj.Message),
		FailedCount:  runnerObj.FailedCount,
		SuccessCount: runnerObj.SuccessCount,
		RawDuration:  runnerObj.Duration,
		Duration:     duration,
		StartedAt:    startedAt,
		EndedAt:      endedAt,
		CreatedAt:    time.Unix(0, int64(time.Millisecond)*runnerObj.CreatedAt).UTC().Format(consts.DefaultTimePattern),
		UpdatedAt:    time.Unix(0, int64(time.Millisecond)*runnerObj.CreatedAt).UTC().Format(consts.DefaultTimePattern),
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
//	@Param		message			body	api.CreateGcTagRunnerRequest	true	"Gc tag runner object"
//	@Success	201
//	@Failure	400	{object}	errcode.ErrCode
//	@Failure	404	{object}	errcode.ErrCode
//	@Failure	500	{object}	errcode.ErrCode
func (h *handler) CreateGcTagRunner(c *gin.Context, req *api.CreateGcTagRunnerRequest) {
	ctx := c.Request.Context()

	err := h.DaemonSvc.CreateGcTagRunner(ctx, *req)
	if err != nil {
		if e, ok := errcode.AsType[errcode.ErrCode](err); ok {
			errcode.NewHTTPError(c, e)
			return
		}
		errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError, fmt.Sprintf("Create gc tag runner failed: %v", err))
		return
	}
	c.Status(http.StatusCreated)
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
//	@Success	200				{object}	api.CommonList{items=[]api.GcTagRunnerItem}
//	@Failure	400				{object}	errcode.ErrCode
//	@Failure	404				{object}	errcode.ErrCode
//	@Failure	500				{object}	errcode.ErrCode
func (h *handler) ListGcTagRunners(c *gin.Context, req *api.ListGcTagRunnersRequest) {
	ctx := c.Request.Context()

	runnerObjs, total, err := h.DaemonSvc.ListGcTagRunners(ctx, req.NamespaceID, req.Pagination, req.Sortable)
	if err != nil {
		if e, ok := errcode.AsType[errcode.ErrCode](err); ok {
			errcode.NewHTTPError(c, e)
			return
		}
		errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError, fmt.Sprintf("List gc tag runners failed: %v", err))
		return
	}
	var resp = make([]any, 0, len(runnerObjs))
	for _, runnerObj := range runnerObjs {
		var startedAt, endedAt *string
		if runnerObj.StartedAt != nil {
			startedAt = new(time.Unix(0, int64(time.Millisecond)*ptr.To(runnerObj.StartedAt)).UTC().Format(consts.DefaultTimePattern))
		}
		if runnerObj.EndedAt != nil {
			endedAt = new(time.Unix(0, int64(time.Millisecond)*ptr.To(runnerObj.EndedAt)).UTC().Format(consts.DefaultTimePattern))
		}
		var duration *string
		if runnerObj.Duration != nil {
			duration = new(durafmt.ParseShort(time.Millisecond * time.Duration(ptr.To(runnerObj.Duration))).String())
		}
		resp = append(resp, api.GcTagRunnerItem{
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
	c.JSON(http.StatusOK, api.CommonList{Total: total, Items: resp})
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
//	@Success	200				{object}	api.GcTagRunnerItem
//	@Failure	400				{object}	errcode.ErrCode
//	@Failure	404				{object}	errcode.ErrCode
//	@Failure	500				{object}	errcode.ErrCode
func (h *handler) GetGcTagRunner(c *gin.Context, req *api.GetGcTagRunnerRequest) {
	ctx := c.Request.Context()

	runnerObj, err := h.DaemonSvc.GetGcTagRunner(ctx, req.NamespaceID, req.RunnerID)
	if err != nil {
		if e, ok := errcode.AsType[errcode.ErrCode](err); ok {
			errcode.NewHTTPError(c, e)
			return
		}
		errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError, fmt.Sprintf("Get gc tag runner failed: %v", err))
		return
	}
	var startedAt, endedAt *string
	if runnerObj.StartedAt != nil {
		startedAt = new(time.Unix(0, int64(time.Millisecond)*ptr.To(runnerObj.StartedAt)).UTC().Format(consts.DefaultTimePattern))
	}
	if runnerObj.EndedAt != nil {
		endedAt = new(time.Unix(0, int64(time.Millisecond)*ptr.To(runnerObj.EndedAt)).UTC().Format(consts.DefaultTimePattern))
	}
	var duration *string
	if runnerObj.Duration != nil {
		duration = new(durafmt.ParseShort(time.Millisecond * time.Duration(ptr.To(runnerObj.Duration))).String())
	}
	c.JSON(http.StatusOK, api.GcTagRunnerItem{
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
//	@Success	200				{object}	api.CommonList{items=[]api.GcTagRecordItem}
//	@Failure	400				{object}	errcode.ErrCode
//	@Failure	404				{object}	errcode.ErrCode
//	@Failure	500				{object}	errcode.ErrCode
func (h *handler) ListGcTagRecords(c *gin.Context, req *api.ListGcTagRecordsRequest) {
	ctx := c.Request.Context()

	recordObjs, total, err := h.DaemonSvc.ListGcTagRecords(ctx, req.RunnerID, req.Pagination, req.Sortable)
	if err != nil {
		if e, ok := errcode.AsType[errcode.ErrCode](err); ok {
			errcode.NewHTTPError(c, e)
			return
		}
		errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError, fmt.Sprintf("List gc tag records failed: %v", err))
		return
	}
	var resp = make([]any, 0, len(recordObjs))
	for _, recordObj := range recordObjs {
		resp = append(resp, api.GcTagRecordItem{
			ID:        recordObj.ID,
			Tag:       recordObj.Resource,
			Status:    recordObj.Status,
			Message:   string(recordObj.Message),
			CreatedAt: time.Unix(0, int64(time.Millisecond)*recordObj.CreatedAt).UTC().Format(consts.DefaultTimePattern),
			UpdatedAt: time.Unix(0, int64(time.Millisecond)*recordObj.CreatedAt).UTC().Format(consts.DefaultTimePattern),
		})
	}
	c.JSON(http.StatusOK, api.CommonList{Total: total, Items: resp})
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
//	@Success	200				{object}	api.GcTagRecordItem
//	@Failure	400				{object}	errcode.ErrCode
//	@Failure	404				{object}	errcode.ErrCode
//	@Failure	500				{object}	errcode.ErrCode
func (h *handler) GetGcTagRecord(c *gin.Context, req *api.GetGcTagRecordRequest) {
	ctx := c.Request.Context()

	recordObj, err := h.DaemonSvc.GetGcTagRecord(ctx, req.NamespaceID, req.RunnerID, req.RecordID)
	if err != nil {
		if e, ok := errcode.AsType[errcode.ErrCode](err); ok {
			errcode.NewHTTPError(c, e)
			return
		}
		errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError, fmt.Sprintf("Get gc tag record failed: %v", err))
		return
	}
	c.JSON(http.StatusOK, api.GcTagRecordItem{
		ID:        recordObj.ID,
		Tag:       recordObj.Resource,
		Status:    recordObj.Status,
		Message:   string(recordObj.Message),
		CreatedAt: time.Unix(0, int64(time.Millisecond)*recordObj.CreatedAt).UTC().Format(consts.DefaultTimePattern),
		UpdatedAt: time.Unix(0, int64(time.Millisecond)*recordObj.CreatedAt).UTC().Format(consts.DefaultTimePattern),
	})
}
