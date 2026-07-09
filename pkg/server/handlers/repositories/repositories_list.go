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
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/server/errcode"
	"github.com/go-sigma/sigma/pkg/utils"
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
//	@Success	200				{object}	api.CommonList{items=[]api.RepositoryItem}
//	@Failure	404				{object}	errcode.ErrCode
//	@Failure	500				{object}	errcode.ErrCode
func (h *handler) ListRepositories(c *gin.Context, req *api.ListRepositoryRequest) {
	ctx := c.Request.Context()

	user, ok := utils.GetFromCtx[*models.User](c, consts.ContextUser)
	if !ok {
		user = &models.User{}
	}

	req.Pagination = utils.NormalizePagination(req.Pagination)

	repositoryObjs, builderMap, total, err := h.RepoSvc.ListRepositories(ctx, user.ID, req.NamespaceID, req.Name, req.Pagination, req.Sortable)
	if err != nil {
		if e, ok := errcode.AsType[errcode.ErrCode](err); ok {
			errcode.NewHTTPError(c, e)
			return
		}
		errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError, err.Error())
		return
	}

	var resp = make([]any, 0, len(repositoryObjs))
	for _, repository := range repositoryObjs {
		repositoryObj := api.RepositoryItem{
			ID:          repository.ID,
			NamespaceID: repository.NamespaceID,
			Name:        repository.Name,
			Description: repository.Description,
			Overview:    new(string(repository.Overview)),
			SizeLimit:   new(repository.SizeLimit),
			Size:        new(repository.Size),
			TagCount:    repository.TagCount,
			TagLimit:    new(repository.TagLimit),
			CreatedAt:   time.Unix(0, int64(time.Millisecond)*repository.CreatedAt).UTC().Format(consts.DefaultTimePattern),
			UpdatedAt:   time.Unix(0, int64(time.Millisecond)*repository.CreatedAt).UTC().Format(consts.DefaultTimePattern),
		}
		if builderMap != nil && builderMap[repository.ID] != nil {
			builderObj := builderMap[repository.ID]
			platforms := []enums.OciPlatform{}
			for p := range strings.SplitSeq(builderObj.BuildkitPlatforms, ",") {
				platforms = append(platforms, enums.OciPlatform(p))
			}
			var scmProvider *enums.ScmProvider
			if repository.Builder.CodeRepository != nil {
				scmProvider = new(enums.ScmProvider(repository.Builder.CodeRepository.User3rdParty.Provider.String()))
			}
			repositoryObj.Builder = &api.BuilderItem{
				ID:           builderObj.ID,
				RepositoryID: builderObj.RepositoryID,

				Source: builderObj.Source,

				CodeRepositoryID: builderObj.CodeRepositoryID,

				Dockerfile: new(string(builderObj.Dockerfile)),

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

	c.JSON(http.StatusOK, api.CommonList{Total: total, Items: resp})
}
