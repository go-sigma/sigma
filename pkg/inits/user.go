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

package inits

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/dal/dao"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils/password"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

func init() {
	inits["user"] = initUser
}

func initUser(config configs.Configuration) error {
	ctx := log.Logger.WithContext(context.Background())

	passwordService := password.New()
	userServiceFactory := dao.NewUserServiceFactory()
	userService := userServiceFactory.New()
	userCount, err := userService.Count(ctx)
	if err != nil {
		return err
	}
	if userCount > 0 {
		return nil
	}
	internalUserUsername := config.Auth.InternalUser.Username
	if internalUserUsername == "" {
		return fmt.Errorf("the internal user username is not set")
	}
	internalUser := &models.User{
		Username: internalUserUsername,
	}
	err = userService.Create(ctx, internalUser)
	if err != nil {
		return err
	}

	adminUserPassword := config.Auth.Admin.Password
	if adminUserPassword == "" {
		return fmt.Errorf("the admin user password is not set")
	}
	adminUserUsername := config.Auth.Admin.Username
	if adminUserUsername == "" {
		return fmt.Errorf("the admin user username is not set")
	}
	adminUserPasswordHashed, err := passwordService.Hash(adminUserPassword)
	if err != nil {
		return err
	}
	adminUserEmail := config.Auth.Admin.Email
	adminUser := &models.User{
		Username: adminUserUsername,
		Password: ptr.Of(adminUserPasswordHashed),
		Email:    ptr.Of(adminUserEmail),
		Role:     enums.UserRoleRoot,
	}
	err = userService.Create(ctx, adminUser)
	if err != nil {
		return err
	}

	return nil
}
