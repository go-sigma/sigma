// Copyright 2023 XImager
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

// PostRepository handles the post repository request
// @Summary Create repository
// @Tags Repository
// @security BasicAuth
// @Accept json
// @Produce json
// @Router /namespaces/{namespace}/repositories/ [post]
// @Param namespace path string true "Namespace name"
// @Param message body types.PostRepositoryRequestSwagger true "Repository object"
// @Success 201 {object} types.PostRepositoryResponse
// @Failure 400 {object} xerrors.ErrCode
// @Failure 404 {object} xerrors.ErrCode
// @Failure 500 {object} xerrors.ErrCode
func (h *handlers) PostRepository(c echo.Context) error {
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

	var req types.PostRepositoryRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, err.Error())
	}

	namespaceService := h.namespaceServiceFactory.New()
	namespaceObj, err := namespaceService.GetByName(ctx, req.Namespace)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Str("namespace", req.Namespace).Msg("Namespace not found")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, fmt.Sprintf("Namespace(%s) not found: %v", req.Namespace, err))
		}
		log.Error().Err(err).Str("namespace", req.Namespace).Msg("Namespace find failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Namespace(%s) find failed: %v", req.Namespace, err))
	}

	log.Info().Interface("req", req).Send()
	repositoryObj := &models.Repository{
		NamespaceID: namespaceObj.ID,
		Name:        req.Name,
		Description: req.Description,
		Overview:    []byte(ptr.To(req.Overview)),
		Visibility:  ptr.To(req.Visibility),
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
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, err.Error())
	}

	return c.JSON(http.StatusCreated, types.PostRepositoryResponse{ID: repositoryObj.ID})
}
