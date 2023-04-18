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

	"github.com/spf13/viper"

	"github.com/ximager/ximager/pkg/dal/dao"
	"github.com/ximager/ximager/pkg/dal/models"
)

func init() {
	inits["user"] = initUser
}

func initUser() error {
	userService := dao.NewUserService()
	userCount, err := userService.Count(context.Background())
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
	internalUser := &models.User{
		Username: internalUserUsername,
		Password: internalUserPassword,
		Email:    "internal-fake@gmail.com",
		Role:     "",
	}
	err = userService.Create(context.Background(), internalUser)
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
	adminUser := &models.User{
		Username: adminUserUsername,
		Password: adminUserPassword,
		Email:    "fake@gmail.com",
		Role:     "admin",
	}
	err = userService.Create(context.Background(), adminUser)
	if err != nil {
		return err
	}

	return nil
}
