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

// ListRepositories handles the list repositories request
//
//	@Summary	List repositories
//	@Tags		Repository
//	@security	BasicAuth
//	@Accept		json
//	@Produce	json
//	@Router		/namespaces/{namespace_id}/repositories/ [get]
//	@Param		namespace_id	path		number	true	"Namespace id"
//	@Param		limit			query		number	false	"Limit size"	minimum(10)	maximum(100)	default(10)
//	@Param		page			query		number	false	"Page number"	minimum(1)	default(1)
//	@Param		sort			query		string	false	"Sort field"
//	@Param		method			query		string	false	"Sort method"	Enums(asc, desc)
//	@Param		name			query		string	false	"Search repository with name"
//	@Success	200				{object}	types.CommonList{items=[]types.RepositoryItem}
//	@Failure	404				{object}	xerrors.ErrCode
//	@Failure	500				{object}	xerrors.ErrCode
func (h *handler) ListRepositories(c echo.Context) error {
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

	var req types.ListRepositoryRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, fmt.Sprintf("Bind and validate request body failed: %v", err))
	}
	req.Pagination = utils.NormalizePagination(req.Pagination)

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

	repositoryService := h.repositoryServiceFactory.New()
	repositoryObjs, total, err := repositoryService.ListRepositoryWithAuth(ctx, namespaceObj.ID, user.ID, req.Name, req.Pagination, req.Sortable)
	if err != nil {
		log.Error().Err(err).Msg("List repository failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, err.Error())
	}

	var repositoryIDs = make([]int64, 0, len(repositoryObjs))
	for _, repository := range repositoryObjs {
		repositoryIDs = append(repositoryIDs, repository.ID)
	}
	builderService := h.builderServiceFactory.New()
	builderMap, err := builderService.GetByRepositoryIDs(ctx, repositoryIDs)
	if err != nil {
		log.Error().Err(err).Msg("Find builders with repository failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("Find builders with repository failed: %v", err))
	}

	var resp = make([]any, 0, len(repositoryObjs))
	for _, repository := range repositoryObjs {
		repositoryObj := types.RepositoryItem{
			ID:          repository.ID,
			NamespaceID: repository.NamespaceID,
			Name:        repository.Name,
			Description: repository.Description,
			Overview:    ptr.Of(string(repository.Overview)),
			SizeLimit:   ptr.Of(repository.SizeLimit),
			Size:        ptr.Of(repository.Size),
			TagCount:    repository.TagCount,
			TagLimit:    ptr.Of(repository.TagLimit),
			CreatedAt:   time.Unix(0, int64(time.Millisecond)*repository.CreatedAt).UTC().Format(consts.DefaultTimePattern),
			UpdatedAt:   time.Unix(0, int64(time.Millisecond)*repository.CreatedAt).UTC().Format(consts.DefaultTimePattern),
		}
		if builderMap != nil && builderMap[repository.ID] != nil {
			builderObj := builderMap[repository.ID]
			platforms := []enums.OciPlatform{}
			for _, p := range strings.Split(builderObj.BuildkitPlatforms, ",") {
				platforms = append(platforms, enums.OciPlatform(p))
			}
			var scmProvider *enums.ScmProvider
			if repository.Builder.CodeRepository != nil {
				scmProvider = ptr.Of(enums.ScmProvider(repository.Builder.CodeRepository.User3rdParty.Provider.String()))
			}
			repositoryObj.Builder = &types.BuilderItem{
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
				ScmProvider:       scmProvider,

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
			}
		}
		resp = append(resp, repositoryObj)
	}

	return c.JSON(http.StatusOK, types.CommonList{Total: total, Items: resp})
}
