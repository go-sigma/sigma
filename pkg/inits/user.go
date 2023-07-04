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

package inits

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"

	"github.com/ximager/ximager/pkg/dal/dao"
	"github.com/ximager/ximager/pkg/dal/models"
	"github.com/ximager/ximager/pkg/types/enums"
	"github.com/ximager/ximager/pkg/utils/password"
	"github.com/ximager/ximager/pkg/utils/ptr"
)

func init() {
	inits["user"] = initUser
}

func initUser() error {
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
	internalUserPassword := viper.GetString("auth.internalUser.password")
	if internalUserPassword == "" {
		return fmt.Errorf("the internal user password is not set")
	}
	internalUserUsername := viper.GetString("auth.internalUser.username")
	if internalUserUsername == "" {
		return fmt.Errorf("the internal user username is not set")
	}
	internalUserPasswordHashed, err := passwordService.Hash(internalUserPassword)
	if err != nil {
		return err
	}
	internalUser := &models.User{
		Provider: enums.ProviderLocal,
		Username: internalUserUsername,
		Password: ptr.Of(internalUserPasswordHashed),
		Email:    ptr.Of("internal-fake@gmail.com"),
	}
	err = userService.Create(ctx, internalUser)
	if err != nil {
		return err
	}

	adminUserPassword := viper.GetString("auth.admin.password")
	if adminUserPassword == "" {
		return fmt.Errorf("the admin user password is not set")
	}
	adminUserUsername := viper.GetString("auth.admin.username")
	if adminUserUsername == "" {
		return fmt.Errorf("the admin user username is not set")
	}
	adminUserPasswordHashed, err := passwordService.Hash(adminUserPassword)
	if err != nil {
		return err
	}
	adminUserEmail := viper.GetString("auth.admin.email")
	if adminUserEmail == "" {
		adminUserEmail = "fake@gmail.com"
	}
	adminUser := &models.User{
		Provider: enums.ProviderLocal,
		Username: adminUserUsername,
		Password: ptr.Of(adminUserPasswordHashed),
		Email:    ptr.Of(adminUserEmail),
	}
	err = userService.Create(ctx, adminUser)
	if err != nil {
		return err
	}

	namespaceServiceFactory := dao.NewNamespaceServiceFactory()
	namespaceService := namespaceServiceFactory.New()
	err = namespaceService.Create(ctx, &models.Namespace{
		Name:       "library",
		UserID:     adminUser.ID,
		Visibility: ptr.Of(enums.VisibilityPrivate),
	})
	if err != nil {
		return err
	}

	return nil
}
