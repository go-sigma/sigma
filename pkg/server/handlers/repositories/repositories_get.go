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
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/server/errcode"
	"github.com/go-sigma/sigma/pkg/utils"
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
//	@Success	200				{object}	api.RepositoryItem
//	@Failure	404				{object}	errcode.ErrCode
//	@Failure	500				{object}	errcode.ErrCode
func (h *handler) GetRepository(c *gin.Context, req *api.GetRepositoryRequest) {
	ctx := c.Request.Context()

	user, ok := utils.GetFromCtx[*models.User](c, consts.ContextUser)
	if !ok {
		slog.Error("get user from context failed")
		errcode.NewHTTPError(c, errcode.HTTPErrCodeUnauthorized)
		return
	}

	authChecked, err := h.Authorizer.Repository(ctx, *user, req.ID, enums.AuthRead)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Error("resource not found", "err", errors.New(utils.UnwrapJoinedErrors(err)), "NamespaceID", req.NamespaceID, "RepositoryID", req.ID)
			errcode.NewHTTPError(c, errcode.HTTPErrCodeNotFound, utils.UnwrapJoinedErrors(err))
			return
		}
		slog.Error("get resource failed", "err", errors.New(utils.UnwrapJoinedErrors(err)), "NamespaceID", req.NamespaceID, "RepositoryID", req.ID)
		errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError, utils.UnwrapJoinedErrors(err))
		return
	}
	if !authChecked {
		slog.Error("auth check failed", "UserID", user.ID, "NamespaceID", req.NamespaceID, "RepositoryID", req.ID)
		errcode.NewHTTPError(c, errcode.HTTPErrCodeUnauthorized, "No permission with this api or resource")
		return
	}

	repositoryObj, err := h.RepoSvc.GetRepository(ctx, req.ID)
	if err != nil {
		if e, ok := errcode.AsType[errcode.ErrCode](err); ok {
			errcode.NewHTTPError(c, e)
			return
		}
		errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError, fmt.Sprintf("Get repository failed: %v", err))
		return
	}
	if repositoryObj.NamespaceID != req.NamespaceID {
		slog.Error("repository's namespace ref id not equal namespace id", "RepositoryObj", repositoryObj, "NamespaceID", req.NamespaceID)
		errcode.NewHTTPError(c, errcode.HTTPErrCodeNotFound)
		return
	}

	var builderItemObj *api.BuilderItem
	if repositoryObj.Builder != nil {
		platforms := []enums.OciPlatform{}
		for p := range strings.SplitSeq(repositoryObj.Builder.BuildkitPlatforms, ",") {
			platforms = append(platforms, enums.OciPlatform(p))
		}

		var scmProvider *enums.ScmProvider
		if repositoryObj.Builder.CodeRepository != nil {
			scmProvider = new(enums.ScmProvider(repositoryObj.Builder.CodeRepository.User3rdParty.Provider.String()))
		}
		builderItemObj = &api.BuilderItem{
			ID:           repositoryObj.Builder.ID,
			RepositoryID: repositoryObj.Builder.RepositoryID,

			Source: repositoryObj.Builder.Source,

			CodeRepositoryID: repositoryObj.Builder.CodeRepositoryID,

			Dockerfile: new(string(repositoryObj.Builder.Dockerfile)),

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

	c.JSON(http.StatusOK, api.RepositoryItem{
		ID:          repositoryObj.ID,
		NamespaceID: repositoryObj.NamespaceID,
		Name:        repositoryObj.Name,
		Description: repositoryObj.Description,
		Overview:    new(string(repositoryObj.Overview)),
		SizeLimit:   new(repositoryObj.SizeLimit),
		Size:        new(repositoryObj.Size),
		TagCount:    repositoryObj.TagCount,
		Builder:     builderItemObj,
		CreatedAt:   time.Unix(0, int64(time.Millisecond)*repositoryObj.CreatedAt).UTC().Format(consts.DefaultTimePattern),
		UpdatedAt:   time.Unix(0, int64(time.Millisecond)*repositoryObj.CreatedAt).UTC().Format(consts.DefaultTimePattern),
	})
}
