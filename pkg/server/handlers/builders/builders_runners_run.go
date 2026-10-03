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

	"github.com/gin-gonic/gin"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/server/errcode"
)

// PostRunnerRun handles POST .../builders/:builder_id/runners/run; it starts a new builder runner from the request spec and returns the new runner id with 201, or an error code on failure.
//
//	@Summary	Run a builder runner
//	@security	BasicAuth
//	@Tags		Builder
//	@Accept		json
//	@Produce	json
//	@Router		/namespaces/{namespace_id}/repositories/{repository_id}/builders/{builder_id}/runners/run [post]
//	@Param		namespace_id	path		string				true	"Namespace ID"
//	@Param		repository_id	path		string				true	"Repository ID"
//	@Param		builder_id		path		string				true	"Builder ID"
//	@Param		message			body		api.PostRunnerRun	true	"Runner object"
//	@Success	201				{object}	api.RunOrRerunRunnerResponse
//	@Failure	400				{object}	errcode.ErrCode
//	@Failure	404				{object}	errcode.ErrCode
//	@Failure	500				{object}	errcode.ErrCode
func (h *handler) PostRunnerRun(c *gin.Context, req *api.PostRunnerRun) {
	ctx := c.Request.Context()

	runnerID, err := h.BuilderSvc.RunRunner(ctx, *req)
	if err != nil {
		if e, ok := errcode.AsType[errcode.ErrCode](err); ok {
			errcode.NewHTTPError(c, e)
			return
		}
		errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError, fmt.Sprintf("Run builder runner failed: %v", err))
		return
	}
	c.JSON(http.StatusCreated, api.RunOrRerunRunnerResponse{
		RunnerID: runnerID,
	})
}
