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

package builders

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
	"github.com/go-sigma/sigma/pkg/xerrors"
)

// PostBuilder handles the post builder request
// @Summary Create a builder for repository
// @Tags Builder
// @security BasicAuth
// @Accept json
// @Produce json
// @Router /builders [post]
// @Param repository_id query int64 true "create builder for repository"
// @Param message body types.PostBuilderRequestSwagger true "Builder object"
// @Success 201
// @Failure 400 {object} xerrors.ErrCode
// @Failure 404 {object} xerrors.ErrCode
// @Failure 500 {object} xerrors.ErrCode
func (h *handlers) PostBuilder(c echo.Context) error {
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

	var req types.PostBuilderRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, err.Error())
	}

	repositoryService := h.repositoryServiceFactory.New()
	repositoryObj, err := repositoryService.Get(ctx, req.RepositoryID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Int64("id", req.RepositoryID).Msg("Repository not found")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, "Repository not found")
		}
		log.Error().Err(err).Int64("id", req.RepositoryID).Msg("Repository find failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, "Repository find failed")
	}

	builderService := h.builderServiceFactory.New()
	_, err = builderService.GetByRepositoryID(ctx, req.RepositoryID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Error().Err(err).Int64("id", req.RepositoryID).Msg("Get builder by repository id failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, err.Error())
	}
	if err == nil {
		log.Error().Err(err).Int64("id", req.RepositoryID).Msg("Repository has been already create builder")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeConflict, "Repository has been already create builder")
	}
	err = query.Q.Transaction(func(tx *query.Query) error {
		builderService := h.builderServiceFactory.New(tx)
		builderObj := &models.Builder{
			RepositoryID:      req.RepositoryID,
			ScmCredentialType: req.ScmCredentialType,
			ScmToken:          req.ScmToken,
			ScmSshKey:         req.ScmSshKey,
			ScmUsername:       req.ScmUsername,
			ScmPassword:       req.ScmPassword, // should encrypt the password
			ScmRepository:     req.ScmRepository,
			ScmDepth:          req.ScmDepth,
			ScmSubmodule:      req.ScmSubmodule,
		}
		err = builderService.Create(ctx, builderObj)
		if err != nil {
			log.Error().Err(err).Int64("id", req.RepositoryID).Msg("Create builder for repository failed")
			return xerrors.HTTPErrCodeInternalError.Detail("Create builder for repository failed")
		}
		auditService := h.auditServiceFactory.New(tx)
		err = auditService.Create(ctx, &models.Audit{
			UserID:       user.ID,
			NamespaceID:  ptr.Of(repositoryObj.NamespaceID),
			Action:       enums.AuditActionCreate,
			ResourceType: enums.AuditResourceTypeBuilder,
			Resource:     strconv.FormatInt(builderObj.ID, 10),
			ReqRaw:       utils.MustMarshal(builderObj),
		})
		if err != nil {
			log.Error().Err(err).Msg("Create audit failed")
			return xerrors.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Create audit failed: %v", err))
		}
		return nil
	})
	if err != nil {
		return xerrors.NewHTTPError(c, err.(xerrors.ErrCode))
	}
	return c.NoContent(http.StatusCreated)
}
