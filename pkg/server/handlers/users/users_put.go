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

package users

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/server/errcode"
)

// Put handles the put request
//
//	@Summary	Update user
//	@Tags		User
//	@Accept		json
//	@Produce	json
//	@Router		/users/{id} [get]
//	@Param		id	path	string	true	"User id"
//	@Success	204
//	@Failure	500	{object}	errcode.ErrCode
func (h *handler) Put(c *gin.Context, req *api.PutUserRequest) {
	ctx := c.Request.Context()

	err := h.UserSvc.UpdateUser(ctx, req.UserID, *req)
	if err != nil {
		if e, ok := errcode.AsType[errcode.ErrCode](err); ok {
			errcode.NewHTTPError(c, e)
			return
		}
		errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError, fmt.Sprintf("Update user failed: %v", err))
		return
	}

	c.Status(http.StatusNoContent)
}
