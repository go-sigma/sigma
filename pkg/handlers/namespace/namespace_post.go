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

package namespace

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/ximager/ximager/pkg/consts"
	"github.com/ximager/ximager/pkg/dal/models"
	"github.com/ximager/ximager/pkg/dal/query"
	"github.com/ximager/ximager/pkg/types"
	"github.com/ximager/ximager/pkg/types/enums"
	"github.com/ximager/ximager/pkg/utils"
	"github.com/ximager/ximager/pkg/utils/ptr"
	"github.com/ximager/ximager/pkg/xerrors"
)

// PostNamespace handles the post namespace request
// @Summary Create namespace
// @Accept json
// @Produce json
// @Router /namespace/ [post]
// @Param name body string true "Namespace name"
// @Param description body string false "Namespace description"
// @Success 201
// @Failure 400
func (h *handlers) PostNamespace(c echo.Context) error {
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

	var req types.CreateNamespaceRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, err.Error())
	}

	namespaceService := h.namespaceServiceFactory.New()
	_, err = namespaceService.GetByName(ctx, req.Name)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Msg("Get namespace by name failed")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, "Get namespace by name failed")
		}
	}
	if err == nil {
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeConflict, "Namespace already exists")
	}

	namespaceObj := &models.Namespace{
		Name:        req.Name,
		Description: req.Description,
		UserID:      user.ID,
		Visibility:  ptr.Of(enums.VisibilityPrivate),
	}
	if req.Visibility != nil {
		namespaceObj.Visibility = req.Visibility
	}
	if ptr.To(req.Limit) > 0 {
		namespaceObj.Limit = ptr.To(req.Limit)
	}
	err = query.Q.Transaction(func(tx *query.Query) error {
		namespaceService := h.namespaceServiceFactory.New(tx)
		err = namespaceService.Create(ctx, namespaceObj)
		if err != nil {
			log.Error().Err(err).Msg("Create namespace failed")
			return xerrors.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Create namespace failed: %v", err))
		}
		return nil
	})
	if err != nil {
		return xerrors.NewHTTPError(c, err.(xerrors.ErrCode))
	}

	return c.JSON(http.StatusCreated, types.CreateNamespaceResponse{ID: namespaceObj.ID})
}
