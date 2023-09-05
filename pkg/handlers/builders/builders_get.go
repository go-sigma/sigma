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
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
	"github.com/go-sigma/sigma/pkg/xerrors"
)

// GetBuilder handles the get builder request
// @Summary Get a builder by repository id
// @Tags Builder
// @security BasicAuth
// @Accept json
// @Produce json
// @Router /builders/{id} [get]
// @Param id path string true "Builder ID"
// @Success 201
// @Failure 400 {object} xerrors.ErrCode
// @Failure 404 {object} xerrors.ErrCode
// @Failure 500 {object} xerrors.ErrCode
func (h *handlers) GetBuilder(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	var req types.GetBuilderRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, err.Error())
	}

	builderService := h.builderServiceFactory.New()
	builderObj, err := builderService.GetByRepositoryID(ctx, req.RepositoryID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Int64("repositoryID", req.RepositoryID).Msg("Builder not found")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, err.Error())
		}
		log.Error().Err(err).Msg("Get builder failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, err.Error())
	}

	platforms := []enums.OciPlatform{}
	for _, p := range strings.Split(builderObj.BuildkitPlatforms, ",") {
		platforms = append(platforms, enums.OciPlatform(p))
	}

	return c.JSON(http.StatusOK, types.GetBuilderResponse{
		ID:           builderObj.ID,
		RepositoryID: builderObj.RepositoryID,

		Source: builderObj.Source,

		CodeRepositoryID: builderObj.CodeRepositoryID,

		Dockerfile: ptr.Of(string(builderObj.Dockerfile)),

		ScmRepository:     builderObj.ScmRepository,
		ScmCredentialType: builderObj.ScmCredentialType,
		ScmSshKey:         builderObj.ScmSshKey,
		ScmToken:          builderObj.ScmToken,
		ScmUsername:       builderObj.ScmUsername,
		ScmPassword:       builderObj.ScmPassword,

		ScmBranch: builderObj.ScmBranch,

		ScmDepth:     builderObj.ScmDepth,
		ScmSubmodule: builderObj.ScmSubmodule,

		CronRule:        builderObj.CronRule,
		CronBranch:      builderObj.CronBranch,
		CronTagTemplate: builderObj.CronTagTemplate,

		WebhookBranchName:        builderObj.WebhookBranchName,
		WebhookBranchTagTemplate: builderObj.WebhookBranchTagTemplate,
		WebhookTagTagTemplate:    builderObj.WebhookTagTagTemplate,

		BuildkitInsecureRegistries: strings.Split(builderObj.BuildkitInsecureRegistries, ","),
		BuildkitContext:            builderObj.BuildkitContext,
		BuildkitDockerfile:         builderObj.BuildkitDockerfile,
		BuildkitPlatforms:          platforms,
	})
}
