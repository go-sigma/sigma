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

package artifacts

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/server/errcode"
)

// ListArtifact handles the list artifact request
func (h *handler) ListArtifact(c *gin.Context, req *api.ListArtifactRequest) {
	ctx := c.Request.Context()

	artifactObjs, total, err := h.ArtifactSvc.ListArtifacts(ctx, *req)
	if err != nil {
		if e, ok := errcode.AsType[errcode.ErrCode](err); ok {
			errcode.NewHTTPError(c, e)
			return
		}
		errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError, err.Error())
		return
	}

	var resp = make([]any, 0, len(artifactObjs))
	for _, artifactObj := range artifactObjs {
		resp = append(resp, api.ArtifactItem{
			ID:        artifactObj.ID,
			Digest:    artifactObj.Digest,
			Size:      artifactObj.Size,
			CreatedAt: time.Unix(0, int64(time.Millisecond)*artifactObj.CreatedAt).UTC().Format(consts.DefaultTimePattern),
			UpdatedAt: time.Unix(0, int64(time.Millisecond)*artifactObj.CreatedAt).UTC().Format(consts.DefaultTimePattern),
		})
	}

	c.JSON(http.StatusOK, api.CommonList{Total: total, Items: resp})
}
