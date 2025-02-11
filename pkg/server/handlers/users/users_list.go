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
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"

	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
	"github.com/go-sigma/sigma/pkg/xerrors"
)

// List handles the list user request
//
//	@Summary	List users with pagination
//	@Tags		User
//	@Accept		json
//	@Produce	json
//	@Router		/users/ [get]
//	@Param		limit			query		int64	false	"limit"	minimum(10)	maximum(100)	default(10)
//	@Param		page			query		int64	false	"page"	minimum(1)	default(1)
//	@Param		sort			query		string	false	"sort field"
//	@Param		method			query		string	false	"sort method"	Enums(asc, desc)
//	@Param		name			query		string	false	"Username"
//	@Param		without_admin	query		boolean	false	"Response with admin"
//	@Success	200				{object}	types.CommonList{items=[]types.UserItem}
//	@Failure	500				{object}	xerrors.ErrCode
func (h *handler) List(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	var req types.GetUserListRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, err.Error())
	}
	req.Pagination = utils.NormalizePagination(req.Pagination)

	userService := h.UserServiceFactory.New()

	var exceptUsername = []string{consts.UserInternal, consts.UserAnonymous}
	userObjs, total, err := userService.ListWithoutUsername(ctx, exceptUsername, req.WithoutAdmin, req.Name, req.Pagination, req.Sortable)
	if err != nil {
		log.Error().Err(err).Msg("List user failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, err.Error())
	}
	var resp = make([]any, 0, len(userObjs))
	for _, userObj := range userObjs {
		resp = append(resp, types.UserItem{
			ID:             userObj.ID,
			Username:       userObj.Username,
			Email:          ptr.To(userObj.Email),
			Status:         userObj.Status,
			LastLogin:      time.Unix(0, int64(time.Millisecond)*userObj.LastLogin).UTC().Format(consts.DefaultTimePattern),
			NamespaceLimit: userObj.NamespaceLimit,
			NamespaceCount: userObj.NamespaceCount,
			CreatedAt:      time.Unix(0, int64(time.Millisecond)*userObj.CreatedAt).UTC().Format(consts.DefaultTimePattern),
			UpdatedAt:      time.Unix(0, int64(time.Millisecond)*userObj.CreatedAt).UTC().Format(consts.DefaultTimePattern),
		})
	}

	return c.JSON(http.StatusOK, types.CommonList{Total: total, Items: resp})
}
