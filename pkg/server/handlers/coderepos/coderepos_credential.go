// Copyright 2026 sigma
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
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/go-sigma/sigma/pkg/api"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/server/errcode"
	"github.com/go-sigma/sigma/pkg/utils"
)

// maskedSecretLength is the number of leading and trailing characters kept when masking a secret.
const maskedSecretLength = 4

// GetCredential gets the SCM credential of a provider
//
//	@Summary	Get code repository credential
//	@security	BasicAuth
//	@Tags		CodeRepository
//	@Accept		json
//	@Produce	json
//	@Router		/coderepos/{provider}/credential [get]
//	@Param		provider	path		string	true	"Get the SCM credential of the provider"
//	@Success	200			{object}	api.GetCodeRepositoryCredentialResponse
//	@Failure	404			{object}	errcode.ErrCode
//	@Failure	500			{object}	errcode.ErrCode
func (h *handler) GetCredential(c *gin.Context, req *api.GetCodeRepositoryCredentialRequest) {
	ctx := c.Request.Context()

	user, ok := utils.GetFromCtx[*models.User](c, consts.ContextUser)
	if !ok {
		slog.Error("get user from context failed")
		errcode.NewHTTPError(c, errcode.HTTPErrCodeUnauthorized)
		return
	}

	credentialObj, err := h.CodeRepoSvc.GetCodeRepositoryCredential(ctx, user.ID, req.Provider)
	if err != nil {
		if e, ok := errcode.AsType[errcode.ErrCode](err); ok {
			errcode.NewHTTPError(c, e)
			return
		}
		errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError)
		return
	}

	c.JSON(http.StatusOK, api.GetCodeRepositoryCredentialResponse{
		Type:        credentialObj.Type,
		Token:       maskSecret(credentialObj.Token),
		Username:    credentialObj.Username,
		HasSshKey:   credentialObj.SshKey != nil,
		HasPassword: credentialObj.Password != nil,
		CreatedAt:   time.UnixMilli(credentialObj.CreatedAt).UTC().Format(consts.DefaultTimePattern),
		UpdatedAt:   time.UnixMilli(credentialObj.UpdatedAt).UTC().Format(consts.DefaultTimePattern),
	})
}

// UpdateCredential creates or replaces the SCM credential of a provider
//
//	@Summary	Update code repository credential
//	@security	BasicAuth
//	@Tags		CodeRepository
//	@Accept		json
//	@Produce	json
//	@Router		/coderepos/{provider}/credential [put]
//	@Param		provider	path	string										true	"Update the SCM credential of the provider"
//	@Param		body		body	api.UpdateCodeRepositoryCredentialRequest	true	"SCM credential"
//	@Success	200
//	@Failure	400	{object}	errcode.ErrCode
//	@Failure	409	{object}	errcode.ErrCode
//	@Failure	500	{object}	errcode.ErrCode
func (h *handler) UpdateCredential(c *gin.Context, req *api.UpdateCodeRepositoryCredentialRequest) {
	ctx := c.Request.Context()

	user, ok := utils.GetFromCtx[*models.User](c, consts.ContextUser)
	if !ok {
		slog.Error("get user from context failed")
		errcode.NewHTTPError(c, errcode.HTTPErrCodeUnauthorized)
		return
	}

	err := h.CodeRepoSvc.UpsertCodeRepositoryCredential(ctx, user.ID, *req)
	if err != nil {
		if e, ok := errcode.AsType[errcode.ErrCode](err); ok {
			errcode.NewHTTPError(c, e)
			return
		}
		errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError)
		return
	}
	c.Status(http.StatusOK)
}

// DeleteCredential clears the SCM credential of a provider
//
//	@Summary	Delete code repository credential
//	@security	BasicAuth
//	@Tags		CodeRepository
//	@Accept		json
//	@Produce	json
//	@Router		/coderepos/{provider}/credential [delete]
//	@Param		provider	path	string	true	"Delete the SCM credential of the provider"
//	@Success	204
//	@Failure	404	{object}	errcode.ErrCode
//	@Failure	500	{object}	errcode.ErrCode
func (h *handler) DeleteCredential(c *gin.Context, req *api.DeleteCodeRepositoryCredentialRequest) {
	ctx := c.Request.Context()

	user, ok := utils.GetFromCtx[*models.User](c, consts.ContextUser)
	if !ok {
		slog.Error("get user from context failed")
		errcode.NewHTTPError(c, errcode.HTTPErrCodeUnauthorized)
		return
	}

	err := h.CodeRepoSvc.DeleteCodeRepositoryCredential(ctx, user.ID, req.Provider)
	if err != nil {
		if e, ok := errcode.AsType[errcode.ErrCode](err); ok {
			errcode.NewHTTPError(c, e)
			return
		}
		errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError)
		return
	}
	c.Status(http.StatusNoContent)
}

// ValidateCredential verifies the SCM credential of a provider without storing it
//
//	@Summary	Validate code repository credential
//	@security	BasicAuth
//	@Tags		CodeRepository
//	@Accept		json
//	@Produce	json
//	@Router		/coderepos/{provider}/credential/validate [post]
//	@Param		provider	path		string										true	"Validate the SCM credential of the provider"
//	@Param		body		body		api.UpdateCodeRepositoryCredentialRequest	true	"SCM credential"
//	@Success	200			{object}	api.ValidateCodeRepositoryCredentialResponse
//	@Failure	400			{object}	errcode.ErrCode
//	@Failure	500			{object}	errcode.ErrCode
func (h *handler) ValidateCredential(c *gin.Context, req *api.UpdateCodeRepositoryCredentialRequest) {
	ctx := c.Request.Context()

	resp, err := h.CodeRepoSvc.ValidateCodeRepositoryCredential(ctx, *req)
	if err != nil {
		if e, ok := errcode.AsType[errcode.ErrCode](err); ok {
			errcode.NewHTTPError(c, e)
			return
		}
		errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// maskSecret returns a masked representation of a secret, keeping only its first and last
// characters, so that the API never echoes a stored secret back.
func maskSecret(secret *string) *string {
	if secret == nil || *secret == "" {
		return nil
	}
	runes := []rune(*secret)
	if len(runes) < 2*maskedSecretLength+1 {
		masked := "****"
		return &masked
	}
	masked := string(runes[:maskedSecretLength]) + "****" + string(runes[len(runes)-maskedSecretLength:])
	return &masked
}
