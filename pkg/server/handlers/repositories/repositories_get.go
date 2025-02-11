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
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
	"github.com/go-sigma/sigma/pkg/xerrors"
)

// GetRepository handles the get repository request
//
//	@Summary	Get repository
//	@Tags		Repository
//	@security	BasicAuth
//	@Accept		json
//	@Produce	json
//	@Router		/namespaces/{namespace_id}/repositories/{repository_id} [get]
//	@Param		namespace_id	path		number	true	"Namespace id"
//	@Param		repository_id	path		number	true	"Repository id"
//	@Success	200				{object}	types.RepositoryItem
//	@Failure	404				{object}	xerrors.ErrCode
//	@Failure	500				{object}	xerrors.ErrCode
func (h *handler) GetRepository(c echo.Context) error {
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

	var req types.GetRepositoryRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, fmt.Sprintf("Bind and validate request body failed: %v", err))
	}

	authChecked, err := h.authServiceFactory.New().Repository(ptr.To(user), req.ID, enums.AuthRead)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(errors.New(utils.UnwrapJoinedErrors(err))).Int64("NamespaceID", req.NamespaceID).Int64("RepositoryID", req.ID).Msg("Resource not found")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, utils.UnwrapJoinedErrors(err))
		}
		log.Error().Err(errors.New(utils.UnwrapJoinedErrors(err))).Int64("NamespaceID", req.NamespaceID).Int64("RepositoryID", req.ID).Msg("Get resource failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, utils.UnwrapJoinedErrors(err))
	}
	if !authChecked {
		log.Error().Int64("UserID", user.ID).Int64("NamespaceID", req.NamespaceID).Int64("RepositoryID", req.ID).Msg("Auth check failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeUnauthorized, "No permission with this api or resource")
	}

	repositoryService := h.repositoryServiceFactory.New()
	repositoryObj, err := repositoryService.Get(ctx, req.ID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Error().Err(err).Msg("Get repository by id not found")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, fmt.Sprintf("Get repository by id not found: %v", err))
		}
		log.Error().Err(err).Msg("Get repository by id failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Get repository by id failed: %v", err))
	}
	if repositoryObj.NamespaceID != req.NamespaceID {
		log.Error().Interface("RepositoryObj", repositoryObj).Int64("NamespaceID", req.NamespaceID).Msg("Repository's namespace ref id not equal namespace id")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound)
	}

	var builderItemObj *types.BuilderItem
	if repositoryObj.Builder != nil {
		platforms := []enums.OciPlatform{}
		for _, p := range strings.Split(repositoryObj.Builder.BuildkitPlatforms, ",") {
			platforms = append(platforms, enums.OciPlatform(p))
		}

		var scmProvider *enums.ScmProvider
		if repositoryObj.Builder.CodeRepository != nil {
			scmProvider = ptr.Of(enums.ScmProvider(repositoryObj.Builder.CodeRepository.User3rdParty.Provider.String()))
		}
		builderItemObj = &types.BuilderItem{
			ID:           repositoryObj.Builder.ID,
			RepositoryID: repositoryObj.Builder.RepositoryID,

			Source: repositoryObj.Builder.Source,

			CodeRepositoryID: repositoryObj.Builder.CodeRepositoryID,

			Dockerfile: ptr.Of(string(repositoryObj.Builder.Dockerfile)),

			ScmRepository:     repositoryObj.Builder.ScmRepository,
			ScmCredentialType: repositoryObj.Builder.ScmCredentialType,
			ScmSshKey:         repositoryObj.Builder.ScmSshKey,
			ScmToken:          repositoryObj.Builder.ScmToken,
			ScmUsername:       repositoryObj.Builder.ScmUsername,
			ScmPassword:       repositoryObj.Builder.ScmPassword,
			ScmProvider:       scmProvider,

			ScmBranch: repositoryObj.Builder.ScmBranch,

			ScmDepth:     repositoryObj.Builder.ScmDepth,
			ScmSubmodule: repositoryObj.Builder.ScmSubmodule,

			CronRule:        repositoryObj.Builder.CronRule,
			CronBranch:      repositoryObj.Builder.CronBranch,
			CronTagTemplate: repositoryObj.Builder.CronTagTemplate,

			WebhookBranchName:        repositoryObj.Builder.WebhookBranchName,
			WebhookBranchTagTemplate: repositoryObj.Builder.WebhookBranchTagTemplate,
			WebhookTagTagTemplate:    repositoryObj.Builder.WebhookTagTagTemplate,

			BuildkitInsecureRegistries: strings.Split(repositoryObj.Builder.BuildkitInsecureRegistries, ","),
			BuildkitContext:            repositoryObj.Builder.BuildkitContext,
			BuildkitDockerfile:         repositoryObj.Builder.BuildkitDockerfile,
			BuildkitPlatforms:          platforms,
			BuildkitBuildArgs:          repositoryObj.Builder.BuildkitBuildArgs,
		}
	}

	return c.JSON(http.StatusOK, types.RepositoryItem{
		ID:          repositoryObj.ID,
		NamespaceID: repositoryObj.NamespaceID,
		Name:        repositoryObj.Name,
		Description: repositoryObj.Description,
		Overview:    ptr.Of(string(repositoryObj.Overview)),
		SizeLimit:   ptr.Of(repositoryObj.SizeLimit),
		Size:        ptr.Of(repositoryObj.Size),
		TagCount:    repositoryObj.TagCount,
		Builder:     builderItemObj,
		CreatedAt:   time.Unix(0, int64(time.Millisecond)*repositoryObj.CreatedAt).UTC().Format(consts.DefaultTimePattern),
		UpdatedAt:   time.Unix(0, int64(time.Millisecond)*repositoryObj.CreatedAt).UTC().Format(consts.DefaultTimePattern),
	})
}
