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
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"

	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/xerrors"
)

// HotNamespace handles the hot namespace request
//
//	@Summary	Hot namespace
//	@security	BasicAuth
//	@Tags		Namespace
//	@Accept		json
//	@Produce	json
//	@Router		/namespaces/hot [get]
//	@Success	200	{object}	types.CommonList{items=[]types.NamespaceItem}
//	@Failure	500	{object}	xerrors.ErrCode
//	@Failure	401	{object}	xerrors.ErrCode
func (h *handler) HotNamespace(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	iuser := c.Get(consts.ContextUser)
	if iuser == nil {
		log.Error().Msg("Get user from header failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeUnauthorized)
	}
	user, ok := iuser.(*models.User)
	if !ok {
		log.Error().Msg("Convert user from header failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeUnauthorized)
	}

	auditService := h.AuditServiceFactory.New()
	namespaceObjs, err := auditService.HotNamespace(ctx, user.ID, consts.HotNamespace) // TODO: remove the namespace that user not have permission
	if err != nil {
		log.Error().Err(err).Msg("Get hot namespaces failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, err.Error())
	}

	var resp = make([]any, 0, len(namespaceObjs))
	for _, namespaceObj := range namespaceObjs {
		resp = append(resp, types.NamespaceItem{
			ID:              namespaceObj.ID,
			Name:            namespaceObj.Name,
			Description:     namespaceObj.Description,
			Visibility:      namespaceObj.Visibility,
			Size:            namespaceObj.Size,
			SizeLimit:       namespaceObj.SizeLimit,
			RepositoryLimit: namespaceObj.RepositoryLimit,
			RepositoryCount: namespaceObj.RepositoryCount,
			TagLimit:        namespaceObj.TagLimit,
			TagCount:        namespaceObj.TagCount,
			CreatedAt:       time.Unix(namespaceObj.CreatedAt, 0).UTC().Format(consts.DefaultTimePattern),
			UpdatedAt:       time.Unix(namespaceObj.UpdatedAt, 0).UTC().Format(consts.DefaultTimePattern),
		})
	}

	return c.JSON(http.StatusOK, types.CommonList{Total: int64(len(namespaceObjs)), Items: resp})
}
