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
	"log/slog"
	"strings"
	"unicode"

	passwordgenerator "github.com/sethvargo/go-password/password"
	"go.uber.org/dig"

	"github.com/go-sigma/sigma/pkg/api/enums"
	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/models"
	repouser "github.com/go-sigma/sigma/pkg/dal/repository/user"
	svcpassword "github.com/go-sigma/sigma/pkg/service/password"
	"github.com/go-sigma/sigma/pkg/utils/uuid"
)

type initUserParams struct {
	dig.In

	Config         *config.Configuration
	PasswordSvc    svcpassword.Service
	UserRepository repouser.UserRepository
}

const (
	// adminPasswordLength is the length of the auto-generated administrator password.
	adminPasswordLength = 10
	// adminPasswordDigits is the number of digits in the auto-generated administrator password.
	adminPasswordDigits = 2
	// adminPasswordGenerateAttempts bounds the retries needed to obtain a password
	// that contains both upper and lower case letters.
	adminPasswordGenerateAttempts = 10
)

// ErrAdminUsername is the error of the admin user username is not set
var ErrAdminUsername = fmt.Errorf("the admin user username is not set")

// errAdminPasswordGenerate is the error of failing to generate an admin user password
var errAdminPasswordGenerate = fmt.Errorf("failed to generate the admin user password")

// generateAdminPassword generates a random administrator password. The password
// only contains upper/lower case letters and digits, so it can be copied from
// the logs without shell quoting issues.
func generateAdminPassword() (string, error) {
	for range adminPasswordGenerateAttempts {
		pwd, err := passwordgenerator.Generate(adminPasswordLength, adminPasswordDigits, 0, false, true)
		if err != nil {
			return "", err
		}
		var hasUpper, hasLower bool
		for _, char := range pwd {
			if unicode.IsUpper(char) {
				hasUpper = true
			}
			if unicode.IsLower(char) {
				hasLower = true
			}
		}
		if hasUpper && hasLower {
			return pwd, nil
		}
	}
	return "", fmt.Errorf("%w: no password with upper and lower case letters generated after %d attempts", errAdminPasswordGenerate, adminPasswordGenerateAttempts)
}

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

	adminUserUsername := strings.TrimSpace(config.Auth.Admin.Username)
	if adminUserUsername == "" {
		return ErrAdminUsername
	}
	adminUserPassword := strings.TrimSpace(config.Auth.Admin.Password)
	generatedPassword := false
	if adminUserPassword == "" {
		adminUserPassword, err = generateAdminPassword()
		if err != nil {
			return err
		}
		generatedPassword = true
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
	if generatedPassword {
		slog.Warn("admin user created with an auto-generated password, please record it and change it after login",
			"username", adminUserUsername, "password", adminUserPassword)
	}

	return nil
}
