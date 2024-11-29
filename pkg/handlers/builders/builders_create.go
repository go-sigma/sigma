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
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/compress"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
	"github.com/go-sigma/sigma/pkg/xerrors"
)

// CreateBuilder handles the create builder request
//
//	@Summary	Create a builder for repository
//	@Tags		Builder
//	@security	BasicAuth
//	@Accept		json
//	@Produce	json
//	@Router		/namespaces/{namespace_id}/repositories/{repository_id}/builders/ [post]
//	@Param		namespace_id	path	string						true	"Namespace ID"
//	@Param		repository_id	path	string						true	"Repository ID"
//	@Param		message			body	types.CreateBuilderRequest	true	"Builder object"
//	@Success	201
//	@Failure	400	{object}	xerrors.ErrCode
//	@Failure	404	{object}	xerrors.ErrCode
//	@Failure	500	{object}	xerrors.ErrCode
func (h *handler) CreateBuilder(c echo.Context) error {
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

	var req types.CreateBuilderRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, err.Error())
	}

	err = h.CreateBuilderValidator(req)
	if err != nil {
		return xerrors.NewHTTPError(c, err.(xerrors.ErrCode))
	}

	repositoryService := h.RepositoryServiceFactory.New()
	repositoryObj, err := repositoryService.Get(ctx, req.RepositoryID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Int64("id", req.RepositoryID).Msg("Repository not found")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, "Repository not found")
		}
		log.Error().Err(err).Int64("id", req.RepositoryID).Msg("Repository find failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, "Repository find failed")
	}

	builderService := h.BuilderServiceFactory.New()
	_, err = builderService.GetByRepositoryID(ctx, req.RepositoryID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Error().Err(err).Int64("id", req.RepositoryID).Msg("Get builder by repository id failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, err.Error())
	}
	if err == nil {
		log.Error().Err(err).Int64("id", req.RepositoryID).Msg("Repository has been already create builder")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeConflict, "Repository has been already create builder")
	}

	compressedDockerfile, err := h.CompressDockerfile(req.Dockerfile)
	if err != nil {
		log.Error().Err(err).Msg("Dockerfile base64 decode failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, fmt.Sprintf("Dockerfile base64 decode failed: %v", err))
	}

	builderObj := &models.Builder{
		RepositoryID: req.RepositoryID,

		Source: req.Source,

		CodeRepositoryID: req.CodeRepositoryID,

		Dockerfile: compressedDockerfile,

		ScmRepository:     req.ScmRepository,
		ScmCredentialType: req.ScmCredentialType,
		ScmToken:          req.ScmToken,
		ScmSshKey:         req.ScmSshKey,
		ScmUsername:       req.ScmUsername,
		ScmPassword:       req.ScmPassword, // should encrypt the password

		ScmBranch: req.ScmBranch,

		ScmDepth:     req.ScmDepth,
		ScmSubmodule: req.ScmSubmodule,

		CronRule:        req.CronRule,
		CronBranch:      req.CronBranch,
		CronTagTemplate: req.CronTagTemplate,

		WebhookBranchName:        req.WebhookBranchName,
		WebhookBranchTagTemplate: req.WebhookBranchTagTemplate,
		WebhookTagTagTemplate:    req.WebhookTagTagTemplate,

		BuildkitInsecureRegistries: strings.Join(req.BuildkitInsecureRegistries, ","),
		BuildkitContext:            req.BuildkitContext,
		BuildkitDockerfile:         req.BuildkitDockerfile,
		BuildkitPlatforms:          utils.StringsJoin(req.BuildkitPlatforms, ","),
	}
	if builderObj.Source == enums.BuilderSourceCodeRepository && req.ScmCredentialType == nil {
		codeRepositoryService := h.CodeRepositoryServiceFactory.New()
		codeRepositoryObj, err := codeRepositoryService.Get(ctx, ptr.To(req.CodeRepositoryID))
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				log.Error().Err(err).Int64("CodeRepositoryID", ptr.To(req.CodeRepositoryID)).Msg("Get code repository by id not found")
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
				builderObj.ScmCredentialType = ptr.Of(enums.ScmCredentialTypeToken)
				builderObj.ScmToken = codeRepositoryObj.User3rdParty.Token
			}
		} else {
			builderObj.ScmCredentialType = ptr.Of(cloneCredentialObj.Type)
			builderObj.ScmUsername = cloneCredentialObj.Username
			builderObj.ScmPassword = cloneCredentialObj.Password
			builderObj.ScmSshKey = cloneCredentialObj.SshKey
			builderObj.ScmToken = cloneCredentialObj.Token
		}
	}
	err = query.Q.Transaction(func(tx *query.Query) error {
		builderService := h.BuilderServiceFactory.New(tx)
		err = builderService.Create(ctx, builderObj)
		if err != nil {
			log.Error().Err(err).Int64("id", req.RepositoryID).Msg("Create builder for repository failed")
			return xerrors.HTTPErrCodeInternalError.Detail("Create builder for repository failed")
		}
		auditService := h.AuditServiceFactory.New(tx)
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

// CreateBuilderValidator ...
func (h *handler) CreateBuilderValidator(req types.CreateBuilderRequest) error {
	switch req.Source {
	case enums.BuilderSourceSelfCodeRepository:
		if req.ScmCredentialType == nil {
			log.Error().Interface("ScmCredentialType", ptr.To(req.ScmCredentialType)).Msg("ScmCredentialType cannot be nil")
			return xerrors.HTTPErrCodeBadRequest.Detail("parameter 'scm credential_type' is invalid")
		}
		switch ptr.To(req.ScmCredentialType) {
		case enums.ScmCredentialTypeNone:
		case enums.ScmCredentialTypeUsername:
			if len(ptr.To(req.ScmUsername)) == 0 || len(ptr.To(req.ScmPassword)) == 0 {
				log.Error().Str("ScmUsername", ptr.To(req.ScmUsername)).Str("ScmPassword", ptr.To(req.ScmPassword)).Msg("ScmUsername and ScmPassword cannot be nil")
				return xerrors.HTTPErrCodeBadRequest.Detail("parameter 'scm_username' or 'scm_password' is invalid")
			}
		case enums.ScmCredentialTypeToken:
			if len(ptr.To(req.ScmToken)) == 0 {
				log.Error().Str("ScmToken", ptr.To(req.ScmToken)).Msg("ScmToken cannot be nil")
				return xerrors.HTTPErrCodeBadRequest.Detail("parameter 'scm_token' is invalid")
			}
		case enums.ScmCredentialTypeSsh:
			if len(ptr.To(req.ScmSshKey)) == 0 {
				log.Error().Str("ScmSshKey", ptr.To(req.ScmSshKey)).Msg("ScmSshKey cannot be nil")
				return xerrors.HTTPErrCodeBadRequest.Detail("parameter 'scm_ssh_key' is invalid")
			}
		default:
			log.Error().Interface("ScmCredentialType", ptr.To(req.ScmCredentialType)).Msg("ScmCredentialType cannot be nil")
			return xerrors.HTTPErrCodeBadRequest.Detail("parameter 'scm credential_type' is invalid")
		}
	case enums.BuilderSourceCodeRepository:
		if req.CodeRepositoryID == nil {
			log.Error().Msg("CodeRepositoryID cannot be nil")
			return xerrors.HTTPErrCodeBadRequest.Detail("parameter 'code_repository_id' is invalid")
		}
	case enums.BuilderSourceDockerfile:
		if req.Dockerfile == nil || len(ptr.To(req.Dockerfile)) == 0 {
			log.Error().Str("Dockerfile", ptr.To(req.Dockerfile)).Msg("Dockerfile cannot be nil")
			return xerrors.HTTPErrCodeBadRequest.Detail("parameter 'dockerfile' is invalid")
		}
	default:
		log.Error().Str("Source", string(req.Source)).Msg("Source is invalid")
		return xerrors.HTTPErrCodeBadRequest.Detail("parameter 'source' is invalid")
	}
	return nil
}

// CompressDockerfile ...
func (h *handler) CompressDockerfile(str *string) ([]byte, error) {
	if str == nil {
		return nil, nil
	}
	data, err := base64.StdEncoding.DecodeString(ptr.To(str))
	if err != nil {
		return nil, err
	}
	return compress.CompressBytes(data)
}
