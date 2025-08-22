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

package namespaces

import (
	"errors"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/server/errcode"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

// GetNamespace handles the get namespace request
//
//	@Summary	Get namespace
//	@security	BasicAuth
//	@Tags		Namespace
//	@Accept		json
//	@Produce	json
//	@Router		/namespaces/{namespace_id} [get]
//	@Param		namespace_id	path		number	true	"Namespace id"
//	@Success	200				{object}	types.NamespaceItem
//	@Failure	404				{object}	errcode.ErrCode
//	@Failure	500				{object}	errcode.ErrCode
func (h *handler) GetNamespace(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	var req types.GetNamespaceRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return errcode.NewHTTPError(c, errcode.HTTPErrCodeBadRequest, err.Error())
	}

	namespaceService := h.NamespaceServiceFactory.New()
	namespaceObj, err := namespaceService.Get(ctx, req.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Msg("Get namespace from db failed")
			return errcode.NewHTTPError(c, errcode.HTTPErrCodeNotFound, err.Error())
		}
		log.Error().Err(err).Msg("Get namespace from db failed")
		return errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError, err.Error())
	}

	repositoryService := h.RepositoryServiceFactory.New()
	repositoryMapCount, err := repositoryService.CountByNamespace(ctx, []int64{namespaceObj.ID})
	if err != nil {
		log.Error().Err(err).Msg("Count repository failed")
		return errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError, err.Error())
	}

	tagService := h.TagServiceFactory.New()
	tagMapCount, err := tagService.CountByNamespace(ctx, []int64{namespaceObj.ID})
	if err != nil {
		log.Error().Err(err).Msg("Count tag failed")
		return errcode.NewHTTPError(c, errcode.HTTPErrCodeInternalError, err.Error())
	}

	return c.JSON(http.StatusOK, types.NamespaceItem{
		ID:              namespaceObj.ID,
		Name:            namespaceObj.Name,
		Description:     namespaceObj.Description,
		Overview:        ptr.Of(string(namespaceObj.Overview)),
		Visibility:      namespaceObj.Visibility,
		Size:            namespaceObj.Size,
		SizeLimit:       namespaceObj.SizeLimit,
		RepositoryCount: repositoryMapCount[namespaceObj.ID],
		RepositoryLimit: namespaceObj.RepositoryLimit,
		TagCount:        tagMapCount[namespaceObj.ID],
		TagLimit:        namespaceObj.TagLimit,
		CreatedAt:       time.Unix(0, int64(time.Millisecond)*namespaceObj.CreatedAt).UTC().Format(consts.DefaultTimePattern),
		UpdatedAt:       time.Unix(0, int64(time.Millisecond)*namespaceObj.CreatedAt).UTC().Format(consts.DefaultTimePattern),
	})
}
