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

package coderepos

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/server/errcode"
	"github.com/go-sigma/sigma/pkg/utils"
)

// Providers list providers
//
//	@Summary	List code repository providers
//	@security	BasicAuth
//	@Tags		CodeRepository
//	@Accept		json
//	@Produce	json
//	@Router		/coderepos/providers [get]
//	@Success	200	{object}	api.CommonList{items=[]api.ListCodeRepositoryProvidersResponse}
//	@Failure	401	{object}	errcode.ErrCode
//	@Failure	500	{object}	errcode.ErrCode
func (h *handler) Providers(c *gin.Context) {
	ctx := c.Request.Context()

	user, ok := utils.GetFromCtx[*models.User](c, consts.ContextUser)
	if !ok {
		slog.Error("get user from context failed")
		errcode.NewHTTPError(c, errcode.HTTPErrCodeUnauthorized)
		return
	}

	user3rdPartyObjs, err := h.CodeRepoSvc.ListCodeRepositoryProviders(ctx, user.ID)
	if err != nil {
		if e, ok := errcode.AsType[errcode.ErrCode](err); ok {
			errcode.NewHTTPError(c, e)
			return
		}
		errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError, fmt.Sprintf("List providers failed: %v", err))
		return
	}
	resp := make([]any, 0, len(user3rdPartyObjs))
	for _, user3rdPartyObj := range user3rdPartyObjs {
		resp = append(resp, api.ListCodeRepositoryProvidersResponse{
			Provider: user3rdPartyObj.Provider,
		})
	}
	c.JSON(http.StatusOK, api.CommonList{Total: int64(len(user3rdPartyObjs)), Items: resp})
}
