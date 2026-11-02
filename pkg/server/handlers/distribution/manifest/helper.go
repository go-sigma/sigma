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

package manifest

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/opencontainers/go-digest"

	"github.com/go-sigma/sigma/pkg/server/errcode"
)

// Refs image tag and digest
type Refs struct {
	Tag    string
	Digest digest.Digest
}

func (h *handler) parseRef(ref string) Refs {
	var refs = Refs{}

	d, err := digest.Parse(ref)
	if err != nil {
		refs.Tag = ref
	} else {
		refs.Digest = d
	}

	return refs
}

// dsError maps a service error (typically an errcode.ErrCode) into a
// distribution-spec HTTP error response.
//
// A canceled request context means the client closed the connection (docker push
// abandons the in-flight requests when it retries), so the failure is reported
// with the client-closed status instead of a misleading 500.
func (h *handler) dsError(c *gin.Context, err error) {
	if err == nil {
		return
	}
	if ctxErr := c.Request.Context().Err(); ctxErr != nil {
		slog.Debug("client closed the request before it completed", "err", err, "ctxErr", ctxErr)
		c.AbortWithStatus(errcode.StatusClientClosedRequest)
		return
	}
	if e, ok := errcode.AsType[errcode.ErrCode](err); ok {
		errcode.NewDSError(c, e)
		return
	}
	errcode.NewDSError(c, errcode.DSErrCodeUnknown)
}
