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
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/dao"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
	"github.com/go-sigma/sigma/pkg/xerrors"
)

// SetupBuilder setup builder for code repository
// @Summary Setup builder for code repository
// @security BasicAuth
// @Tags CodeRepository
// @Accept json
// @Produce json
// @Router /coderepos/{id}/setup-builder [post]
// @Param id path string true "code repository id"
// @Param message body types.PostCodeRepositorySetupBuilderSwagger true "Code repository setup builder object"
// @Success 201
// @Failure 401 {object} xerrors.ErrCode
// @Failure 500 {object} xerrors.ErrCode
func (h *handlers) SetupBuilder(c echo.Context) error {
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

	var req types.PostCodeRepositorySetupBuilder
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, err.Error())
	}

	namespaceService := h.namespaceServiceFactory.New()
	namespaceObj, err := namespaceService.Get(ctx, req.ID)
	if err != nil {
		log.Error().Err(err).Msg("Get namespace by id failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, err.Error())
	}

	repositoryService := h.repositoryServiceFactory.New()
	repositoryObj, err := repositoryService.GetByName(ctx, req.RepositoryName)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = query.Q.Transaction(func(tx *query.Query) error {
				repositoryService := h.repositoryServiceFactory.New(tx)
				repositoryObj = &models.Repository{
					NamespaceID: namespaceObj.ID,
					Name:        req.RepositoryName,
					Visibility:  namespaceObj.Visibility,
				}
				err = repositoryService.Create(ctx, repositoryObj, dao.AutoCreateNamespace{AutoCreate: false})
				if err != nil {
					log.Error().Err(err).Msg("Create repository failed")
					return xerrors.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Create repository failed: %v", err))
				}
				auditService := h.auditServiceFactory.New(tx)
				err = auditService.Create(ctx, &models.Audit{
					UserID:       user.ID,
					NamespaceID:  ptr.Of(namespaceObj.ID),
					Action:       enums.AuditActionCreate,
					ResourceType: enums.AuditResourceTypeRepository,
					Resource:     req.RepositoryName,
					ReqRaw:       utils.MustMarshal(req),
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
		}
	}

	builderService := h.builderServiceFactory.New()
	_, err = builderService.GetByRepositoryID(ctx, repositoryObj.ID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Error().Err(err).Msg("Get builder by repository id failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Get builder by repository id failed: %v", err))
	}
	if err == nil {
		log.Error().Msgf("Repository(%s) already have a builder", repositoryObj.Name)
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeConflict, fmt.Sprintf("Repository(%s) already have a builder", repositoryObj.Name))
	}

	builderObj := &models.Builder{
		RepositoryID: repositoryObj.ID,
		Active:       true,
		Source:       enums.BuilderSourceCodeRepository,

		CodeRepositoryID: ptr.Of(req.ID),

		BuildkitContext:    req.BuildkitContext,
		BuildkitDockerfile: req.BuildkitDockerfile,
		BuildkitPlatforms:  utils.StringsJoin(req.BuildkitPlatforms, ","),
	}
	err = query.Q.Transaction(func(tx *query.Query) error {
		builderService := h.builderServiceFactory.New(tx)
		err = builderService.Create(ctx, builderObj)
		if err != nil {
			log.Error().Err(err).Msg("Create builder failed")
			return xerrors.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Create builder failed: %v", err))
		}
		auditService := h.auditServiceFactory.New(tx)
		err = auditService.Create(ctx, &models.Audit{
			UserID:       user.ID,
			NamespaceID:  ptr.Of(namespaceObj.ID),
			Action:       enums.AuditActionCreate,
			ResourceType: enums.AuditResourceTypeBuilder,
			Resource:     req.RepositoryName,
			ReqRaw:       utils.MustMarshal(req),
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
