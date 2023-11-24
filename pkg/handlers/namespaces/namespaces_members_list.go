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
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"

	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/types"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/xerrors"
)

// ListNamespaceMembers handles the list namespace members request
func (h *handler) ListNamespaceMembers(c echo.Context) error {
	ctx := log.Logger.WithContext(c.Request().Context())

	var req types.ListNamespaceMemberRequest
	err := utils.BindValidate(c, &req)
	if err != nil {
		log.Error().Err(err).Msg("Bind and validate request body failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeBadRequest, fmt.Sprintf("Bind and validate request body failed: %v", err))
	}

	namespaceMemberService := h.namespaceMemberServiceFactory.New()
	namespaceMemberObjs, total, err := namespaceMemberService.ListNamespaceMembers(ctx, req.ID, req.Name, req.Pagination, req.Sortable)
	if err != nil {
		log.Error().Err(err).Msg("List namespace role failed")
		return xerrors.NewHTTPError(c, xerrors.HTTPErrCodeInternalError, fmt.Sprintf("List namespace role failed: %v", err))
	}

	var resp = make([]any, 0, len(namespaceMemberObjs))
	for _, namespaceMemberObj := range namespaceMemberObjs {
		resp = append(resp, types.NamespaceRoleItem{
			ID:        namespaceMemberObj.ID,
			Username:  namespaceMemberObj.User.Username,
			UserID:    namespaceMemberObj.User.ID,
			Role:      namespaceMemberObj.Role,
			CreatedAt: time.Unix(0, int64(time.Millisecond)*namespaceMemberObj.CreatedAt).UTC().Format(consts.DefaultTimePattern),
			UpdatedAt: time.Unix(0, int64(time.Millisecond)*namespaceMemberObj.UpdatedAt).UTC().Format(consts.DefaultTimePattern),
		})
	}

	return c.JSON(http.StatusOK, types.CommonList{Total: total, Items: resp})
}
