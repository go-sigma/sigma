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
	"strings"

	"github.com/rs/zerolog/log"
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/dao"
	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/types/enums"
	"github.com/go-sigma/sigma/pkg/utils"
	"github.com/go-sigma/sigma/pkg/utils/password"
	"github.com/go-sigma/sigma/pkg/utils/ptr"
)

func init() {
	inits["user"] = initUser
}

var (
	// ErrAdminUsername is the error of the admin user username is not set
	ErrAdminUsername = fmt.Errorf("the admin user username is not set")
	// ErrAdminPassword is the error of the admin user password is not set
	ErrAdminPassword = fmt.Errorf("the admin user password is not set")
)

func initUser(digCon *dig.Container) error {
	ctx := log.Logger.WithContext(context.Background())

	config := utils.MustGetObjFromDigCon[configs.Configuration](digCon)
	pwdSvc := utils.MustGetObjFromDigCon[password.Service](digCon)
	userSvcFactory := utils.MustGetObjFromDigCon[dao.UserServiceFactory](digCon)

	userSvc := userSvcFactory.New()
	userCount, err := userSvc.Count(ctx)
	if err != nil {
		return err
	}
	if userCount > 0 {
		return nil
	}
	err = userSvc.Create(ctx, &models.User{
		Username: consts.UserInternal,
		Role:     enums.UserRoleRoot,
	})
	if err != nil {
		return err
	}
	err = userSvc.Create(ctx, &models.User{
		Username: consts.UserAnonymous,
		Role:     enums.UserRoleAnonymous,
	})
	if err != nil {
		return err
	}

	adminUserPassword := strings.TrimSpace(config.Auth.Admin.Password)
	if adminUserPassword == "" {
		return ErrAdminPassword
	}
	adminUserUsername := strings.TrimSpace(config.Auth.Admin.Username)
	if adminUserUsername == "" {
		return ErrAdminUsername
	}
	adminUserPasswordHashed, err := pwdSvc.Hash(adminUserPassword)
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
	err = userSvc.Create(ctx, adminUser)
	if err != nil {
		return err
	}

	return nil
}
