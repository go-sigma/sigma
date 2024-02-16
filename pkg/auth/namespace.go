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
	"context"
	"errors"
	"fmt"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

// Namespace ...
func (s authService) Namespace(user models.User, namespaceID int64, auth enums.Auth) (bool, error) {
	ctx := log.Logger.WithContext(context.Background())

	// 1. check user is admin or not
	if user.Role == enums.UserRoleAdmin || user.Role == enums.UserRoleRoot {
		return true, nil
	}

	// 2. check namespace visibility
	namespaceService := s.namespaceServiceFactory.New()
	namespaceObj, err := namespaceService.Get(ctx, namespaceID)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Msg("Get namespace by id failed")
			return false, errors.Join(err, fmt.Errorf("Get namespace by id(%d) failed", namespaceID))
		}
		log.Error().Err(err).Msg("Get namespace by id not found")
		return false, errors.Join(err, fmt.Errorf("Get namespace by id(%d) not found", namespaceID))
	}
	if namespaceObj.Visibility == enums.VisibilityPublic && auth == enums.AuthRead {
		return true, nil
	}

	// 3. check user is member of the namespace
	roleService := s.namespaceMemberServiceFactory.New()
	namespaceMemberObj, err := roleService.GetNamespaceMember(ctx, namespaceID, user.ID)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) { // check user's role in this namespace
			log.Error().Err(err).Msg("Get namespace member by namespace id and user id failed")
		}
		return false, nil
	}
	if namespaceMemberObj.Role == enums.NamespaceRoleReader && auth == enums.AuthRead {
		return true, nil
	}
	if namespaceMemberObj.Role == enums.NamespaceRoleManager && (auth == enums.AuthManage || auth == enums.AuthRead) {
		return true, nil
	}
	if namespaceMemberObj.Role == enums.NamespaceRoleAdmin && (auth == enums.AuthAdmin || auth == enums.AuthManage || auth == enums.AuthRead) {
		return true, nil
	}
	return false, nil
}

// NamespaceRole ...
func (s authService) NamespaceRole(user models.User, namespaceID int64) (*enums.NamespaceRole, error) {
	ctx := log.Logger.WithContext(context.Background())

	roleService := s.namespaceMemberServiceFactory.New()
	namespaceMemberObj, err := roleService.GetNamespaceMember(ctx, namespaceID, user.ID)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) { // check user's role in this namespace
			log.Error().Err(err).Msg("Get namespace member by namespace id and user id failed")
		}
		return nil, err
	}
	return ptr.Of(namespaceMemberObj.Role), nil
}

// NamespaceRole ...
func (s authService) NamespacesRole(user models.User, namespaceIDs []int64) (map[int64]*enums.NamespaceRole, error) {
	ctx := log.Logger.WithContext(context.Background())

	roleService := s.namespaceMemberServiceFactory.New()
	namespaceMemberObjs, err := roleService.GetNamespacesMember(ctx, namespaceIDs, user.ID)
	if err != nil {
		return nil, err
	}

	var result = make(map[int64]*enums.NamespaceRole, len(namespaceIDs))
	for _, o := range namespaceMemberObjs {
		result[o.NamespaceID] = ptr.Of(o.Role)
	}

	return result, nil
}
