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

package upload

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/server/errcode"
)

// GetUpload handles the get upload request
func (h *handler) GetUpload(c *gin.Context) {
	ctx := c.Request.Context()

	uri := c.Request.URL.Path
	uploadID := strings.TrimPrefix(uri[strings.LastIndex(uri, "/"):], "/")

	c.Writer.Header().Set(consts.UploadUUID, uploadID)
	c.Writer.Header().Set(consts.HeaderLocation, uri)

	_, err := h.UploadSvc.GetUpload(ctx, uploadID)
	if err != nil {
		if e, ok := errcode.AsType[errcode.ErrCode](err); ok {
			errcode.NewDSError(c, e)
			return
		}
		errcode.NewDSError(c, errcode.DSErrCodeUnknown)
		return
	}

	c.Status(http.StatusNoContent)
}
