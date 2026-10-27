// Copyright 2026 sigma
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

package credential

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"go.uber.org/dig"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/consts"
	reposetting "github.com/go-sigma/sigma/pkg/dal/repository/setting"
	"github.com/go-sigma/sigma/pkg/utils/crypt"
)

//go:generate go tool mockgen -mock_names Service=MockCredentialService -destination=mocks/credential_mocks.go -package=mocks github.com/go-sigma/sigma/pkg/service/credential Service

// MinMasterKeyLength is the minimum accepted length of the credential master key.
const MinMasterKeyLength = 32

// Service seals and opens values (SCM credentials) that are stored at rest.
type Service interface {
	// Seal encrypts plaintext with the process-wide credential master key.
	Seal(plaintext string) (string, error)
	// Open decrypts a value produced by Seal.
	Open(ciphertext string) (string, error)
}

// Params declares the dependencies needed to construct the credential service.
type Params struct {
	dig.In

	Config            *config.Configuration
	SettingRepository reposetting.SettingRepository
}

// New returns a Service bound to the configured credential master key.
//
// The key is resolved once, at construction time: auth.credentialKey when set,
// otherwise the value the credential bootstrap task persisted in the settings
// table. It fails when neither is available, because storing credentials that
// cannot be decrypted later is worse than refusing to start.
func New(params Params) (Service, error) {
	key, err := resolveMasterKey(context.Background(), params)
	if err != nil {
		return nil, err
	}
	return &service{key: key}, nil
}

// resolveMasterKey returns the credential master key from config or the settings table.
func resolveMasterKey(ctx context.Context, params Params) (string, error) {
	if key := params.Config.Auth.CredentialKey; key != "" {
		if len(key) < MinMasterKeyLength {
			return "", fmt.Errorf("auth.credentialKey must be at least %d bytes", MinMasterKeyLength)
		}
		return key, nil
	}

	settingObj, err := params.SettingRepository.Get(ctx, consts.SettingCredentialMasterKey)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", fmt.Errorf("credential master key is not initialized, configure auth.credentialKey or restart sigma to generate one")
		}
		slog.Error("get credential master key failed", "err", err)
		return "", err
	}
	if len(settingObj.Val) < MinMasterKeyLength {
		return "", fmt.Errorf("credential master key in settings must be at least %d bytes", MinMasterKeyLength)
	}
	return string(settingObj.Val), nil
}

type service struct {
	key string
}

// Seal encrypts plaintext with the process-wide credential master key.
func (s *service) Seal(plaintext string) (string, error) {
	return crypt.Seal(s.key, plaintext)
}

// Open decrypts a value produced by Seal.
func (s *service) Open(ciphertext string) (string, error) {
	return crypt.Open(s.key, ciphertext)
}
