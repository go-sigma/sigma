// Copyright 2026 sigma
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

	"github.com/gin-gonic/gin"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/server/errcode"
)

// GetArtifactLayers handles the get artifact layers request
//
//	@Summary	Get artifact layers
//	@security	BasicAuth
//	@Tags		Artifact
//	@Accept		json
//	@Produce	json
//	@Router		/namespaces/{namespace_id}/repositories/{repository_id}/artifacts/{digest}/layers [get]
//	@Param		namespace_id	path		string	true	"Namespace ID"
//	@Param		repository_id	path		string	true	"Repository ID"
//	@Param		digest			path		string	true	"Artifact digest"
//	@Success	200				{object}	api.ArtifactLayers
//	@Failure	400				{object}	errcode.ErrCode
//	@Failure	404				{object}	errcode.ErrCode
//	@Failure	500				{object}	errcode.ErrCode
func (h *handler) GetArtifactLayers(c *gin.Context, req *api.GetArtifactLayersRequest) {
	ctx := c.Request.Context()

	layers, err := h.ArtifactSvc.GetArtifactLayers(ctx, req.NamespaceID, req.RepositoryID, req.Digest)
	if err != nil {
		if e, ok := errcode.AsType[errcode.ErrCode](err); ok {
			errcode.NewHTTPError(c, e)
			return
		}
		errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError, err.Error())
		return
	}

	c.JSON(http.StatusOK, layers)
}
