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

// UpdateGcRepositoryRule handles the update gc repository rule request
//
//	@Summary	Update gc repository rule
//	@security	BasicAuth
//	@Tags		Daemon
//	@Accept		json
//	@Produce	json
//	@Router		/daemons/gc-repository/{namespace_id}/ [put]
//	@Param		namespace_id	path	int64								true	"Namespace id"
//	@Param		message			body	api.UpdateGcRepositoryRuleRequest	true	"Gc repository rule object"
//	@Success	204
//	@Failure	400	{object}	errcode.ErrCode
//	@Failure	404	{object}	errcode.ErrCode
//	@Failure	500	{object}	errcode.ErrCode
func (h *handler) UpdateGcRepositoryRule(c *gin.Context, req *api.UpdateGcRepositoryRuleRequest) {
	ctx := c.Request.Context()

	err := h.DaemonSvc.UpdateGcRepositoryRule(ctx, *req)
	if err != nil {
		if e, ok := errcode.AsType[errcode.ErrCode](err); ok {
			errcode.NewHTTPError(c, e)
			return
		}
		errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError, fmt.Sprintf("Update gc repository rule failed: %v", err))
		return
	}
	c.Status(http.StatusNoContent)
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
//	@Success	200				{object}	api.GetGcRepositoryRuleResponse
//	@Failure	400				{object}	errcode.ErrCode
//	@Failure	404				{object}	errcode.ErrCode
//	@Failure	500				{object}	errcode.ErrCode
func (h *handler) GetGcRepositoryRule(c *gin.Context, req *api.GetGcRepositoryRuleRequest) {
	ctx := c.Request.Context()

	ruleObj, err := h.DaemonSvc.GetGcRepositoryRule(ctx, req.NamespaceID)
	if err != nil {
		if e, ok := errcode.AsType[errcode.ErrCode](err); ok {
			errcode.NewHTTPError(c, e)
			return
		}
		errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError, fmt.Sprintf("Get gc repository rule failed: %v", err))
		return
	}
	var nextTrigger *string
	if ruleObj.CronNextTrigger != nil {
		nextTrigger = new(time.Unix(0, int64(time.Millisecond)*ptr.To(ruleObj.CronNextTrigger)).UTC().Format(consts.DefaultTimePattern))
	}
	c.JSON(http.StatusOK, api.GetGcRepositoryRuleResponse{
		RetentionDay:    ruleObj.RetentionDay,
		CronEnabled:     ruleObj.CronEnabled,
		CronRule:        ruleObj.CronRule,
		CronNextTrigger: nextTrigger,
		CreatedAt:       time.Unix(0, int64(time.Millisecond)*ruleObj.CreatedAt).UTC().Format(consts.DefaultTimePattern),
		UpdatedAt:       time.Unix(0, int64(time.Millisecond)*ruleObj.CreatedAt).UTC().Format(consts.DefaultTimePattern),
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
//	@Success	200				{object}	api.GcRepositoryRunnerItem
//	@Failure	400				{object}	errcode.ErrCode
//	@Failure	404				{object}	errcode.ErrCode
//	@Failure	500				{object}	errcode.ErrCode
func (h *handler) GetGcRepositoryLatestRunner(c *gin.Context, req *api.GetGcRepositoryLatestRunnerRequest) {
	ctx := c.Request.Context()

	runnerObj, err := h.DaemonSvc.GetGcRepositoryLatestRunner(ctx, req.NamespaceID)
	if err != nil {
		if e, ok := errcode.AsType[errcode.ErrCode](err); ok {
			errcode.NewHTTPError(c, e)
			return
		}
		errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError, fmt.Sprintf("Get gc repository latest runner failed: %v", err))
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
	c.JSON(http.StatusOK, api.GcRepositoryRunnerItem{
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

// CreateGcRepositoryRunner handles the create gc repository runner request
//
//	@Summary	Create gc repository runner
//	@security	BasicAuth
//	@Tags		Daemon
//	@Accept		json
//	@Produce	json
//	@Router		/daemons/gc-repository/{namespace_id}/runners/ [post]
//	@Param		namespace_id	path	int64								true	"Namespace id"
//	@Param		message			body	api.CreateGcRepositoryRunnerRequest	true	"Gc repository runner object"
//	@Success	201
//	@Failure	400	{object}	errcode.ErrCode
//	@Failure	404	{object}	errcode.ErrCode
//	@Failure	500	{object}	errcode.ErrCode
func (h *handler) CreateGcRepositoryRunner(c *gin.Context, req *api.CreateGcRepositoryRunnerRequest) {
	ctx := c.Request.Context()

	err := h.DaemonSvc.CreateGcRepositoryRunner(ctx, *req)
	if err != nil {
		if e, ok := errcode.AsType[errcode.ErrCode](err); ok {
			errcode.NewHTTPError(c, e)
			return
		}
		errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError, fmt.Sprintf("Create gc repository runner failed: %v", err))
		return
	}
	c.Status(http.StatusCreated)
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
//	@Success	200				{object}	api.CommonList{items=[]api.GcRepositoryRunnerItem}
//	@Failure	400				{object}	errcode.ErrCode
//	@Failure	404				{object}	errcode.ErrCode
//	@Failure	500				{object}	errcode.ErrCode
func (h *handler) ListGcRepositoryRunners(c *gin.Context, req *api.ListGcRepositoryRunnersRequest) {
	ctx := c.Request.Context()

	runnerObjs, total, err := h.DaemonSvc.ListGcRepositoryRunners(ctx, req.NamespaceID, req.Pagination, req.Sortable)
	if err != nil {
		if e, ok := errcode.AsType[errcode.ErrCode](err); ok {
			errcode.NewHTTPError(c, e)
			return
		}
		errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError, fmt.Sprintf("List gc repository runners failed: %v", err))
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
		resp = append(resp, api.GcRepositoryRunnerItem{
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
//	@Success	200				{object}	api.GcRepositoryRunnerItem
//	@Failure	400				{object}	errcode.ErrCode
//	@Failure	404				{object}	errcode.ErrCode
//	@Failure	500				{object}	errcode.ErrCode
func (h *handler) GetGcRepositoryRunner(c *gin.Context, req *api.GetGcRepositoryRunnerRequest) {
	ctx := c.Request.Context()

	runnerObj, err := h.DaemonSvc.GetGcRepositoryRunner(ctx, req.NamespaceID, req.RunnerID)
	if err != nil {
		if e, ok := errcode.AsType[errcode.ErrCode](err); ok {
			errcode.NewHTTPError(c, e)
			return
		}
		errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError, fmt.Sprintf("Get gc repository runner failed: %v", err))
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
	c.JSON(http.StatusOK, api.GcRepositoryRunnerItem{
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
//	@Success	200				{object}	api.CommonList{items=[]api.GcRepositoryRecordItem}
//	@Failure	400				{object}	errcode.ErrCode
//	@Failure	404				{object}	errcode.ErrCode
//	@Failure	500				{object}	errcode.ErrCode
func (h *handler) ListGcRepositoryRecords(c *gin.Context, req *api.ListGcRepositoryRecordsRequest) {
	ctx := c.Request.Context()

	recordObjs, total, err := h.DaemonSvc.ListGcRepositoryRecords(ctx, req.RunnerID, req.Pagination, req.Sortable)
	if err != nil {
		if e, ok := errcode.AsType[errcode.ErrCode](err); ok {
			errcode.NewHTTPError(c, e)
			return
		}
		errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError, fmt.Sprintf("List gc repository records failed: %v", err))
		return
	}
	var resp = make([]any, 0, len(recordObjs))
	for _, recordObj := range recordObjs {
		resp = append(resp, api.GcRepositoryRecordItem{
			ID:         recordObj.ID,
			Repository: recordObj.Resource,
			Status:     recordObj.Status,
			Message:    string(recordObj.Message),
			CreatedAt:  time.Unix(0, int64(time.Millisecond)*recordObj.CreatedAt).UTC().Format(consts.DefaultTimePattern),
			UpdatedAt:  time.Unix(0, int64(time.Millisecond)*recordObj.CreatedAt).UTC().Format(consts.DefaultTimePattern),
		})
	}
	c.JSON(http.StatusOK, api.CommonList{Total: total, Items: resp})
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
//	@Success	200				{object}	api.GcRepositoryRecordItem
//	@Failure	400				{object}	errcode.ErrCode
//	@Failure	404				{object}	errcode.ErrCode
//	@Failure	500				{object}	errcode.ErrCode
func (h *handler) GetGcRepositoryRecord(c *gin.Context, req *api.GetGcRepositoryRecordRequest) {
	ctx := c.Request.Context()

	recordObj, err := h.DaemonSvc.GetGcRepositoryRecord(ctx, req.NamespaceID, req.RunnerID, req.RecordID)
	if err != nil {
		if e, ok := errcode.AsType[errcode.ErrCode](err); ok {
			errcode.NewHTTPError(c, e)
			return
		}
		errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError, fmt.Sprintf("Get gc repository record failed: %v", err))
		return
	}
	c.JSON(http.StatusOK, api.GcRepositoryRecordItem{
		ID:         recordObj.ID,
		Repository: recordObj.Resource,
		Status:     recordObj.Status,
		Message:    string(recordObj.Message),
		CreatedAt:  time.Unix(0, int64(time.Millisecond)*recordObj.CreatedAt).UTC().Format(consts.DefaultTimePattern),
		UpdatedAt:  time.Unix(0, int64(time.Millisecond)*recordObj.CreatedAt).UTC().Format(consts.DefaultTimePattern),
	})
}
