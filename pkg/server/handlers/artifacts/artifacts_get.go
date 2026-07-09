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
	"time"

	"github.com/gin-gonic/gin"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/server/errcode"
)

// GetArtifact handles the get artifact request
func (h *handler) GetArtifact(c *gin.Context, req *api.GetArtifactRequest) {
	ctx := c.Request.Context()

	artifactObj, err := h.ArtifactSvc.GetArtifact(ctx, req.Repository, req.Digest)
	if err != nil {
		if e, ok := errcode.AsType[errcode.ErrCode](err); ok {
			errcode.NewHTTPError(c, e)
			return
		}
		errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError, err.Error())
		return
	}

	c.JSON(200, api.ArtifactItem{
		ID:        artifactObj.ID,
		Digest:    artifactObj.Digest,
		ConfigRaw: string(artifactObj.ConfigRaw),
		Size:      artifactObj.Size,
		BlobSize:  artifactObj.BlobsSize,
		PullTimes: artifactObj.PullTimes,
		LastPull:  time.Unix(0, int64(time.Millisecond)*artifactObj.LastPull).UTC().Format(consts.DefaultTimePattern),
		PushedAt:  time.Unix(0, int64(time.Millisecond)*artifactObj.PushedAt).UTC().Format(consts.DefaultTimePattern),
		CreatedAt: time.Unix(0, int64(time.Millisecond)*artifactObj.CreatedAt).UTC().Format(consts.DefaultTimePattern),
		UpdatedAt: time.Unix(0, int64(time.Millisecond)*artifactObj.CreatedAt).UTC().Format(consts.DefaultTimePattern),
	})
}
