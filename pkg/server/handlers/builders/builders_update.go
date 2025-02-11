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
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
	"github.com/go-sigma/sigma/pkg/xerrors"
)

// UpdateBuilder handles the update builder request
//
//	@Summary	Update a builder by id
//	@Tags		Builder
//	@security	BasicAuth
//	@Accept		json
//	@Produce	json
//	@Router		/namespace/{namespace_id}/repositories/{repository_id}/builders/{builder_id} [put]
//	@Param		namespace_id	path	string						true	"Namespace id"
//	@Param		repository_id	path	string						true	"Repository id"
//	@Param		builder_id		path	string						true	"Builder id"
//	@Param		message			body	types.UpdateBuilderRequest	true	"Builder object"
//	@Success	201
//	@Failure	400	{object}	xerrors.ErrCode
//	@Failure	404	{object}	xerrors.ErrCode
//	@Failure	500	{object}	xerrors.ErrCode
func (h *handler) UpdateBuilder(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	var req types.UpdateBuilderRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, err.Error())
	}

	compressedDockerfile, err := h.CompressDockerfile(req.Dockerfile)
	if err != nil {
		log.Error().Err(err).Msg("Dockerfile base64 decode failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, fmt.Sprintf("Dockerfile base64 decode failed: %v", err))
	}

	updates := map[string]any{
		query.Builder.Source.ColumnName().String():                     req.Source,
		query.Builder.CodeRepositoryID.ColumnName().String():           req.CodeRepositoryID,
		query.Builder.Dockerfile.ColumnName().String():                 compressedDockerfile,
		query.Builder.ScmRepository.ColumnName().String():              req.ScmRepository,
		query.Builder.ScmCredentialType.ColumnName().String():          req.ScmCredentialType,
		query.Builder.ScmSshKey.ColumnName().String():                  req.ScmSshKey,
		query.Builder.ScmToken.ColumnName().String():                   req.ScmToken,
		query.Builder.ScmUsername.ColumnName().String():                req.ScmUsername,
		query.Builder.ScmPassword.ColumnName().String():                req.ScmPassword,
		query.Builder.ScmBranch.ColumnName().String():                  req.ScmBranch,
		query.Builder.ScmDepth.ColumnName().String():                   req.ScmDepth,
		query.Builder.ScmSubmodule.ColumnName().String():               req.ScmSubmodule,
		query.Builder.CronRule.ColumnName().String():                   req.CronRule,
		query.Builder.CronBranch.ColumnName().String():                 req.CronBranch,
		query.Builder.CronTagTemplate.ColumnName().String():            req.CronTagTemplate,
		query.Builder.WebhookBranchName.ColumnName().String():          req.WebhookBranchName,
		query.Builder.WebhookBranchTagTemplate.ColumnName().String():   req.WebhookBranchTagTemplate,
		query.Builder.WebhookTagTagTemplate.ColumnName().String():      req.WebhookTagTagTemplate,
		query.Builder.BuildkitInsecureRegistries.ColumnName().String(): strings.Join(req.BuildkitInsecureRegistries, ","),
		query.Builder.BuildkitContext.ColumnName().String():            req.BuildkitContext,
		query.Builder.BuildkitDockerfile.ColumnName().String():         req.BuildkitDockerfile,
		query.Builder.BuildkitPlatforms.ColumnName().String():          utils.StringsJoin(req.BuildkitPlatforms, ","),
	}
	if req.Source == enums.BuilderSourceCodeRepository && req.ScmCredentialType == nil {
		codeRepositoryService := h.CodeRepositoryServiceFactory.New()
		codeRepositoryObj, err := codeRepositoryService.Get(ctx, ptr.To(req.CodeRepositoryID))
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				log.Error().Err(err).Int64("CodeRepositoryID", ptr.To(req.CodeRepositoryID)).
					Msg("Get code repository by id not found")
				return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound.Detail(fmt.Sprintf("Get code repository by id(%d) not found: %v", ptr.To(req.CodeRepositoryID), err)))
			}
			log.Error().Err(err).Int64("CodeRepositoryID", ptr.To(req.CodeRepositoryID)).Msg("Get code repository by id failed")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Get code repository by id(%d) failed: %v", ptr.To(req.CodeRepositoryID), err)))
		}
		cloneCredentialObj, err := codeRepositoryService.GetCloneCredential(ctx, codeRepositoryObj.User3rdPartyID)
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				log.Error().Err(err).Msg("Get code repository clone credential failed")
			} else {
				updates[query.Builder.ScmCredentialType.ColumnName().String()] = ptr.Of(enums.ScmCredentialTypeToken)
				updates[query.Builder.ScmToken.ColumnName().String()] = codeRepositoryObj.User3rdParty.Token
			}
		} else {
			updates[query.Builder.ScmCredentialType.ColumnName().String()] = ptr.Of(cloneCredentialObj.Type)
			updates[query.Builder.ScmToken.ColumnName().String()] = cloneCredentialObj.Token
			updates[query.Builder.ScmUsername.ColumnName().String()] = cloneCredentialObj.Username
			updates[query.Builder.ScmPassword.ColumnName().String()] = cloneCredentialObj.Password
			updates[query.Builder.ScmSshKey.ColumnName().String()] = cloneCredentialObj.SshKey
		}
	}
	err = query.Q.Transaction(func(tx *query.Query) error {
		builderService := h.BuilderServiceFactory.New(tx)
		err = builderService.Update(ctx, req.BuilderID, updates)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				log.Error().Err(err).Int64("builder_id", req.BuilderID).Msg("Builder id not found")
				return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, fmt.Sprintf("Builder id(%d) not found", req.BuilderID))
			}
			log.Error().Err(err).Int64("builder_id", req.BuilderID).Msg("Builder find failed")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, fmt.Sprintf("Builder id(%d) find failed: %v", req.BuilderID, err))
		}
		return nil
	})
	if err != nil {
		return xerrors.NewHTTPError(c, err.(xerrors.ErrCode))
	}
	return c.NoContent(http.StatusNoContent)
}
