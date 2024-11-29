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

package namespaces

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"

	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
	"github.com/go-sigma/sigma/pkg/xerrors"
)

// ListNamespaces handles the list namespace request
//
//	@Summary	List namespace
//	@security	BasicAuth
//	@Tags		Namespace
//	@Accept		json
//	@Produce	json
//	@Router		/namespaces/ [get]
//	@Param		limit	query		int64	false	"Limit size"	minimum(10)	maximum(100)	default(10)
//	@Param		page	query		int64	false	"Page number"	minimum(1)	default(1)
//	@Param		sort	query		string	false	"Sort field"
//	@Param		method	query		string	false	"Sort method"	Enums(asc, desc)
//	@Param		name	query		string	false	"Search namespace with name"
//	@Success	200		{object}	types.CommonList{items=[]types.NamespaceItem}
//	@Failure	500		{object}	xerrors.ErrCode
func (h *handler) ListNamespaces(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	var user *models.User
	iuser := c.Get(consts.ContextUser)
	if iuser == nil {
		user = &models.User{ID: 0}
	} else {
		var ok bool
		user, ok = iuser.(*models.User)
		if !ok {
			log.Error().Msg("Convert user from header failed")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeUnauthorized)
		}
	}

	var req types.ListNamespaceRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, err.Error())
	}
	req.Pagination = utils.NormalizePagination(req.Pagination)

	namespaceService := h.NamespaceServiceFactory.New()
	namespaceObjs, total, err := namespaceService.ListNamespaceWithAuth(ctx, user.ID, req.Name, req.Pagination, req.Sortable)
	if err != nil {
		log.Error().Err(err).Msg("List namespace failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, err.Error())
	}

	var namespaceIDs = make([]int64, 0, len(namespaceObjs))

	authService := h.AuthServiceFactory.New()
	namespacesRole, err := authService.NamespacesRole(ptr.To(user), namespaceIDs)
	if err != nil {
		log.Error().Err(err).Msg("Get namespaces role failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Get namespaces role failed: %v", err))
	}

	var resp = make([]any, 0, len(namespaceObjs))
	for _, namespaceObj := range namespaceObjs {
		resp = append(resp, types.NamespaceItem{
			ID:              namespaceObj.ID,
			Name:            namespaceObj.Name,
			Description:     namespaceObj.Description,
			Visibility:      namespaceObj.Visibility,
			Role:            namespacesRole[namespaceObj.ID],
			Size:            namespaceObj.Size,
			SizeLimit:       namespaceObj.SizeLimit,
			RepositoryLimit: namespaceObj.RepositoryLimit,
			RepositoryCount: namespaceObj.RepositoryCount,
			TagLimit:        namespaceObj.TagLimit,
			TagCount:        namespaceObj.TagCount,
			CreatedAt:       time.Unix(0, int64(time.Millisecond)*namespaceObj.CreatedAt).UTC().Format(consts.DefaultTimePattern),
			UpdatedAt:       time.Unix(0, int64(time.Millisecond)*namespaceObj.UpdatedAt).UTC().Format(consts.DefaultTimePattern),
		})
	}

	return c.JSON(http.StatusOK, types.CommonList{Total: total, Items: resp})
}
