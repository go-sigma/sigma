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
	"github.com/spf13/viper"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/dao"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
	"github.com/go-sigma/sigma/pkg/xerrors"
)

// CreateRepository handles the create repository request
//
//	@Summary	Create repository
//	@Tags		Repository
//	@security	BasicAuth
//	@Accept		json
//	@Produce	json
//	@Router		/namespaces/{namespace_id}/repositories/ [post]
//	@Param		namespace_id	path		number							true	"Namespace id"
//	@Param		message			body		types.CreateRepositoryRequest	true	"Repository object"
//	@Success	201				{object}	types.CreateRepositoryResponse
//	@Failure	400				{object}	xerrors.ErrCode
//	@Failure	404				{object}	xerrors.ErrCode
//	@Failure	500				{object}	xerrors.ErrCode
func (h *handler) CreateRepository(c echo.Context) error {
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

	var req types.CreateRepositoryRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, err.Error())
	}

	authChecked, err := h.authServiceFactory.New().Namespace(c, req.NamespaceID, enums.AuthManage)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(errors.New(utils.UnwrapJoinedErrors(err))).Int64("NamespaceID", req.NamespaceID).Msg("Resource not found")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, utils.UnwrapJoinedErrors(err))
		}
		log.Error().Err(errors.New(utils.UnwrapJoinedErrors(err))).Int64("NamespaceID", req.NamespaceID).Msg("Get resource failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, utils.UnwrapJoinedErrors(err))
	}
	if !authChecked {
		log.Error().Int64("UserID", user.ID).Int64("NamespaceID", req.NamespaceID).Msg("Auth check failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeUnauthorized, "No permission with this api or resource")
	}

	namespaceService := h.namespaceServiceFactory.New()
	namespaceObj, err := namespaceService.Get(ctx, req.NamespaceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Int64("NamespaceID", req.NamespaceID).Msg("Namespace not found")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, fmt.Sprintf("Namespace(%d) not found: %v", req.NamespaceID, err))
		}
		log.Error().Err(err).Int64("NamespaceID", req.NamespaceID).Msg("Namespace find failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Namespace(%d) find failed: %v", req.NamespaceID, err))
	}

	repositoryObj := &models.Repository{
		NamespaceID: namespaceObj.ID,
		Name:        req.Name,
		Description: req.Description,
		Overview:    []byte(ptr.To(req.Overview)),
		TagLimit:    ptr.To(req.TagLimit),
		SizeLimit:   ptr.To(req.SizeLimit),
	}
	repositoryService := h.repositoryServiceFactory.New()
	err = repositoryService.Create(ctx, repositoryObj, dao.AutoCreateNamespace{
		AutoCreate: viper.GetBool("namespace.autoCreate"),
		Visibility: enums.MustParseVisibility(viper.GetString("namespace.visibility")),
		UserID:     user.ID,
	})
	if err != nil {
		log.Error().Err(err).Interface("repositoryObj", repositoryObj).Msg("Repository create failed")
		e, ok := err.(xerrors.ErrCode) // maybe got exceed repository count quota limit error
		if ok {
			return xerrors.NewHTTPError(c, e)
		}
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Create repository failed: %v", err)))
	}

	return c.JSON(http.StatusCreated, types.CreateRepositoryResponse{ID: repositoryObj.ID})
}
