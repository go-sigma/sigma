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
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"unicode"

	passwordgenerator "github.com/sethvargo/go-password/password"
	"go.uber.org/dig"
	"gorm.io/gorm"

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

	// Do not stop when the users table is not empty: a replica can die after
	// creating some of the bootstrap accounts, and the missing ones must still
	// be created on the next start. ensureUser is idempotent per username.
	var err error
	if _, err = ensureUser(ctx, userSvc, &models.User{
		ID:       uuid.NewV7String(),
		Username: consts.UserInternal,
		Role:     enums.UserRoleRoot,
	}); err != nil {
		return err
	}
	if _, err = ensureUser(ctx, userSvc, &models.User{
		ID:       uuid.NewV7String(),
		Username: consts.UserAnonymous,
		Role:     enums.UserRoleAnonymous,
	}); err != nil {
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
		// A generated password is only known through the logs, so the login page asks for a new
		// one before the account enters the app.
		PasswordChangeRequired: generatedPassword,
	}
	created, err := ensureUser(ctx, userSvc, adminUser)
	if err != nil {
		return err
	}
	// Only the replica that inserted the admin user knows the password in use.
	// A tolerated create error still may mean the insert was applied and only
	// its acknowledgement was lost, so compare the stored row instead of
	// trusting the error alone.
	if generatedPassword && !created {
		if stored, getErr := userSvc.GetByUsername(ctx, adminUserUsername); getErr == nil {
			created = stored.ID == adminUser.ID
		}
	}
	if generatedPassword && created {
		slog.Warn("admin user created with an auto-generated password, please record it and change it after login",
			"username", adminUserUsername, "password", adminUserPassword)
	}

	return nil
}

// ensureUser creates user unless another replica created the same username
// first, and reports whether this replica performed the insert.
func ensureUser(ctx context.Context, userRepository repouser.UserRepository, user *models.User) (bool, error) {
	_, err := userRepository.GetByUsername(ctx, user.Username)
	if err == nil {
		return false, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return false, err
	}

	err = userRepository.Create(ctx, user)
	if err == nil {
		return true, nil
	}
	if tolerateConcurrentCreate(err, func() error {
		_, getErr := userRepository.GetByUsername(ctx, user.Username)
		return getErr
	}) {
		slog.Info("user already created by another replica", "username", user.Username)
		return false, nil
	}
	return false, err
}
