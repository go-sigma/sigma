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

package repositories

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/server/errcode"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

// DeleteRepository handles the delete repository request
//
//	@Summary	Delete repository
//	@Tags		Repository
//	@security	BasicAuth
//	@Accept		json
//	@Produce	json
//	@Router		/namespaces/{namespace_id}/repositories/{repository_id} [delete]
//	@Param		namespace_id	path	number	true	"Namespace id"
//	@Param		repository_id	path	number	true	"Repository id"
//	@Success	204
//	@Failure	404	{object}	errcode.ErrCode
//	@Failure	500	{object}	errcode.ErrCode
func (h *handler) DeleteRepository(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	iuser := c.Get(consts.ContextUser)
	if iuser == nil {
		log.Error().Msg("Get user from header failed")
		return errcode.NewHTTPError(c, errcode.HTTPErrCodeUnauthorized)
	}
	user, ok := iuser.(*models.User)
	if !ok {
		log.Error().Msg("Convert user from header failed")
		return errcode.NewHTTPError(c, errcode.HTTPErrCodeUnauthorized)
	}

	var req types.DeleteRepositoryRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return errcode.NewHTTPError(c, errcode.HTTPErrCodeBadRequest, err.Error())
	}

	authChecked, err := h.authServiceFactory.New().Repository(ptr.To(user), req.ID, enums.AuthManage)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(errors.New(utils.UnwrapJoinedErrors(err))).Int64("NamespaceID", req.NamespaceID).Int64("RepositoryID", req.ID).Msg("Resource not found")
			return errcode.NewHTTPError(c, errcode.HTTPErrCodeNotFound, utils.UnwrapJoinedErrors(err))
		}
		log.Error().Err(errors.New(utils.UnwrapJoinedErrors(err))).Int64("NamespaceID", req.NamespaceID).Int64("RepositoryID", req.ID).Msg("Get resource failed")
		return errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError, utils.UnwrapJoinedErrors(err))
	}
	if !authChecked {
		log.Error().Int64("UserID", user.ID).Int64("NamespaceID", req.NamespaceID).Int64("RepositoryID", req.ID).Msg("Auth check failed")
		return errcode.NewHTTPError(c, errcode.HTTPErrCodeUnauthorized, "No permission with this api or resource")
	}

	namespaceService := h.namespaceServiceFactory.New()
	namespaceObj, err := namespaceService.Get(ctx, req.NamespaceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Int64("NamespaceID", req.NamespaceID).Msg("Namespace not found")
			return errcode.NewHTTPError(c, errcode.HTTPErrCodeNotFound, fmt.Sprintf("Namespace(%d) not found: %v", req.NamespaceID, err))
		}
		log.Error().Err(err).Int64("NamespaceID", req.NamespaceID).Msg("Namespace find failed")
		return errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError, fmt.Sprintf("Namespace(%d) find failed: %v", req.NamespaceID, err))
	}

	repositoryService := h.repositoryServiceFactory.New()
	repositoryObj, err := repositoryService.Get(ctx, req.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Int64("id", req.ID).Msg("Repository not found")
			return errcode.NewHTTPError(c, errcode.HTTPErrCodeNotFound, err.Error())
		}
		log.Error().Err(err).Msg("Repository find failed")
		return errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError, err.Error())
	}
	if repositoryObj.NamespaceID != namespaceObj.ID {
		log.Error().Interface("RepositoryObj", repositoryObj).Interface("NamespaceObj", namespaceObj).Msg("Repository's namespace ref id not equal namespace id")
		return errcode.NewHTTPError(c, errcode.HTTPErrCodeNotFound)
	}

	err = query.Q.Transaction(func(tx *query.Query) error {
		repositoryService := h.repositoryServiceFactory.New(tx)
		err = repositoryService.DeleteByID(ctx, req.ID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				log.Error().Err(err).Msg("Delete repository by id not found")
				return errcode.HTTPErrCodeNotFound.Detail(fmt.Sprintf("Delete repository by id not found: %v", err))
			}
			log.Error().Err(err).Msg("Delete repository by id failed")
			return errcode.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Delete repository by id failed: %v", err))
		}
		return nil
	})
	if err != nil {
		var e errcode.ErrCode
		if errors.As(err, &e) {
			return errcode.NewHTTPError(c, e)
		}
		return errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError)
	}

	return c.NoContent(http.StatusNoContent)
}
