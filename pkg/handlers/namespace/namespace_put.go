// Copyright 2023 XImager
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

package namespace

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/ximager/ximager/pkg/dal/query"
	"github.com/ximager/ximager/pkg/types"
	"github.com/ximager/ximager/pkg/utils"
	"github.com/ximager/ximager/pkg/utils/ptr"
	"github.com/ximager/ximager/pkg/xerrors"
)

// PutNamespace handles the put namespace request
func (h *handlers) PutNamespace(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	var req types.PutNamespaceRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, err.Error())
	}

	namespaceService := h.namespaceServiceFactory.New()
	namespaceObj, err := namespaceService.Get(ctx, req.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Msg("Namespace not found")
			return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeNotFound, err.Error())
		}
		log.Error().Err(err).Msg("Find namespace failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, err.Error())
	}

	if req.Quota != nil && namespaceObj.Quota.Limit > ptr.To(req.Quota) {
		log.Error().Err(err).Msg("Namespace quota is less than the before limit")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, "Namespace quota is less than the before limit")
	}

	err = query.Q.Transaction(func(tx *query.Query) error {
		namespaceService := h.namespaceServiceFactory.New(tx)
		err = namespaceService.UpdateByID(ctx, namespaceObj.ID, req)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				log.Error().Err(err).Msg("Update namespace failed")
				return xerrors.HTTPErrCodeNotFound.Detail("Namespace not found")
			}
			log.Error().Err(err).Msg("Update namespace failed")
			return xerrors.HTTPErrCodeInternalError.Detail(fmt.Sprintf("Update namespace failed: %v", err))
		}

		if req.Quota != nil {
			err = namespaceService.UpdateQuota(ctx, namespaceObj.ID, ptr.To(req.Quota))
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					log.Error().Err(err).Msg("Namespace quota not found")
					return xerrors.HTTPErrCodeNotFound.Detail("Namespace quota not found")
				}
				log.Error().Err(err).Msg("Update namespace quota failed")
				return xerrors.HTTPErrCodeNotFound.Detail(fmt.Sprintf("Update namespace quota failed: %v", err))
			}
		}
		return nil
	})
	if err != nil {
		return xerrors.NewHTTPError(c, err.(xerrors.ErrCode))
	}
	return c.NoContent(http.StatusNoContent)
}
