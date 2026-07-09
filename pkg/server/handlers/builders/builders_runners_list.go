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
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hako/durafmt"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/server/errcode"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

// ListRunners handles the list builder runners request
//
//	@Summary	Get builder runners by builder id
//	@Tags		Builder
//	@security	BasicAuth
//	@Accept		json
//	@Produce	json
//	@Router		/namespaces/{namespace_id}/repositories/{repository_id}/builders/{builder_id}/runners/ [get]
//	@Param		namespace_id	path		string	true	"Namespace ID"
//	@Param		repository_id	path		string	true	"Repository ID"
//	@Param		builder_id		path		string	true	"Builder ID"
//	@Success	200				{object}	api.BuilderItem
//	@Failure	404				{object}	errcode.ErrCode
//	@Failure	500				{object}	errcode.ErrCode
func (h *handler) ListRunners(c *gin.Context, req *api.ListBuilderRunnersRequest) {
	ctx := c.Request.Context()

	req.Pagination = utils.NormalizePagination(req.Pagination)

	runnerObjs, total, err := h.BuilderSvc.ListRunners(ctx, req.BuilderID, req.Pagination, req.Sortable)
	if err != nil {
		if e, ok := errcode.AsType[errcode.ErrCode](err); ok {
			errcode.NewHTTPError(c, e)
			return
		}
		errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError, fmt.Sprintf("List builder runners failed: %v", err))
		return
	}
	var resp = make([]any, 0, len(runnerObjs))
	for _, runnerObj := range runnerObjs {
		var duration *string
		if runnerObj.Duration != nil {
			duration = new(durafmt.ParseShort(time.Millisecond * time.Duration(ptr.To(runnerObj.Duration))).String())
		}

		resp = append(resp, api.BuilderRunnerItem{
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
	c.JSON(http.StatusOK, api.CommonList{Total: total, Items: resp})
}
