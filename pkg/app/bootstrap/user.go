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

package bootstrap

import (
	"context"
	"fmt"
	"strings"

	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	repouser "github.com/go-sigma/sigma/pkg/dal/repository/user"
	"github.com/go-sigma/sigma/pkg/service/password"
	"github.com/go-sigma/sigma/pkg/utils/uuid"
)

type initUserParams struct {
	dig.In

	Config         *config.Configuration
	PasswordSvc    password.Service
	UserRepository repouser.UserRepository
}

var (
	// ErrAdminUsername is the error of the admin user username is not set
	ErrAdminUsername = fmt.Errorf("the admin user username is not set")
	// ErrAdminPassword is the error of the admin user password is not set
	ErrAdminPassword = fmt.Errorf("the admin user password is not set")
)

func initUser(digCon *dig.Container) error {
	return digCon.Invoke(initUserWithParams)
}

func initUserWithParams(params initUserParams) error {
	ctx := context.Background()
	config := params.Config
	pwdSvc := params.PasswordSvc
	userSvc := params.UserRepository
	userCount, err := userSvc.Count(ctx)
	if err != nil {
		return err
	}
	if userCount > 0 {
		return nil
	}
	err = userSvc.Create(ctx, &models.User{
		ID:       uuid.NewV7String(),
		Username: consts.UserInternal,
		Role:     enums.UserRoleRoot,
	})
	if err != nil {
		return err
	}
	err = userSvc.Create(ctx, &models.User{
		ID:       uuid.NewV7String(),
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
		ID:       uuid.NewV7String(),
		Username: adminUserUsername,
		Password: new(adminUserPasswordHashed),
		Email:    new(adminUserEmail),
		Role:     enums.UserRoleRoot,
	}
	err = userSvc.Create(ctx, adminUser)
	if err != nil {
		return err
	}

	return nil
}
