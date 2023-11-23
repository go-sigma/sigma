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

package auth

import (
	"errors"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/types/enums"
)

// Namespace ...
func (s service) Namespace(c echo.Context, namespaceID int64, auth enums.Auth) bool {
	ctx := log.Logger.WithContext(c.Request().Context())

	iuser := c.Get(consts.ContextUser)
	if iuser == nil {
		log.Error().Msg("Get user from header failed")
		return false
	}
	user, ok := iuser.(*models.User)
	if !ok {
		log.Error().Msg("Convert user from header failed")
		return false
	}

	// 1. check user is admin or not
	if user.Role == enums.UserRoleAdmin || user.Role == enums.UserRoleRoot {
		return true
	}

	// 2. check namespace visibility
	namespaceService := s.namespaceServiceFactory.New()
	namespaceObj, err := namespaceService.Get(ctx, namespaceID)
	if err != nil {
		log.Error().Err(err).Msg("Get namespace by id failed")
		return false
	}
	if namespaceObj.Visibility == enums.VisibilityPublic && auth == enums.AuthRead {
		return true
	}

	// 3. check user is member of the namespace
	roleService := s.roleServiceFactory.New()
	namespaceMemberObj, err := roleService.GetNamespaceMember(ctx, namespaceID, user.ID)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Msg("Get namespace member failed")
		}
		return false
	}
	if namespaceMemberObj.Role == enums.NamespaceRoleReader && auth == enums.AuthRead {
		return true
	}
	if namespaceMemberObj.Role == enums.NamespaceRoleManager && (auth == enums.AuthManage || auth == enums.AuthRead) {
		return true
	}
	if namespaceMemberObj.Role == enums.NamespaceRoleAdmin && (auth == enums.AuthAdmin || auth == enums.AuthManage || auth == enums.AuthRead) {
		return true
	}
	return false
}
