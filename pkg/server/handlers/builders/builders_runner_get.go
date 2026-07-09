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

package builders

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hako/durafmt"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/server/errcode"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

// GetRunner handles the get builder runner request
//
//	@Summary	Get builder runner by runner id
//	@Tags		Builder
//	@security	BasicAuth
//	@Accept		json
//	@Produce	json
//	@Router		/namespaces/{namespace_id}/repositories/{repository_id}/builders/{builder_id}/runners/{runner_id} [get]
//	@Param		namespace_id	path		string	true	"Namespace ID"
//	@Param		repository_id	path		string	true	"Repository ID"
//	@Param		builder_id		path		string	true	"Builder ID"
//	@Param		runner_id		path		string	true	"Runner ID"
//	@Success	200				{object}	api.BuilderItem
//	@Failure	404				{object}	errcode.ErrCode
//	@Failure	500				{object}	errcode.ErrCode
func (h *handler) GetRunner(c *gin.Context, req *api.GetRunner) {
	ctx := c.Request.Context()

	runnerObj, err := h.BuilderSvc.GetRunner(ctx, req.RunnerID)
	if err != nil {
		if e, ok := errcode.AsType[errcode.ErrCode](err); ok {
			errcode.NewHTTPError(c, e)
			return
		}
		errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError, fmt.Sprintf("Get builder runner failed: %v", err))
		return
	}
	if runnerObj.BuilderID != req.BuilderID {
		slog.Error("get builder by id failed", "builder_id", runnerObj.BuilderID, "builder_id", req.BuilderID)
		errcode.NewHTTPError(c, errcode.HTTPErrCodeForbidden, "Get builder by id failed")
		return
	}

	var duration *string
	if runnerObj.Duration != nil {
		duration = new(durafmt.ParseShort(time.Millisecond * time.Duration(ptr.To(runnerObj.Duration))).String())
	}

	c.JSON(http.StatusOK, api.BuilderRunnerItem{
		ID:        runnerObj.ID,
		BuilderID: runnerObj.BuilderID,

		Log: runnerObj.Log,

		Status:        runnerObj.Status,
		StatusMessage: runnerObj.StatusMessage,
		Tag:           runnerObj.Tag,
		RawTag:        runnerObj.RawTag,
		Description:   runnerObj.Description,
		ScmBranch:     runnerObj.ScmBranch,

		StartedAt:   runnerObj.StartedAt,
		EndedAt:     runnerObj.EndedAt,
		RawDuration: runnerObj.Duration,
		Duration:    duration,

		CreatedAt: time.Unix(0, int64(time.Millisecond)*runnerObj.CreatedAt).UTC().Format(consts.DefaultTimePattern),
		UpdatedAt: time.Unix(0, int64(time.Millisecond)*runnerObj.CreatedAt).UTC().Format(consts.DefaultTimePattern),
	})
}
