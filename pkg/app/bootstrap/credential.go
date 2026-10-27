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

package bootstrap

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"go.uber.org/dig"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/config"
	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/query"
	reposetting "github.com/go-sigma/sigma/pkg/dal/repository/setting"
	"github.com/go-sigma/sigma/pkg/infra/lock"
	"github.com/go-sigma/sigma/pkg/service/credential"
)

const (
	// credentialMasterKeyLockKey serializes key generation across replicas.
	credentialMasterKeyLockKey  = "sigma:bootstrap:credential-master-key"
	credentialMasterKeyLockTTL  = 30 * time.Second
	credentialMasterKeyLockWait = 60 * time.Second
)

type credentialParams struct {
	dig.In

	Config            *config.Configuration
	SettingRepository reposetting.SettingRepository
	Locker            lock.Locker
}

func credentialBootstrap(digCon *dig.Container) error {
	return digCon.Invoke(credentialWithParams)
}

// credentialWithParams makes sure a credential master key is available.
//
// When auth.credentialKey is configured it only validates the key. Otherwise it
// generates a random key and persists it in the settings table, so that replicas
// and restarts keep decrypting the credentials written before. A distributed lock
// plus a second read keeps concurrent first starts from writing different keys.
func credentialWithParams(params credentialParams) error {
	if key := params.Config.Auth.CredentialKey; key != "" {
		if len(key) < credential.MinMasterKeyLength {
			return fmt.Errorf("auth.credentialKey must be at least %d bytes", credential.MinMasterKeyLength)
		}
		slog.Info("credential master key loaded from config")
		return nil
	}

	ctx := context.Background()

	lockObj, err := params.Locker.Acquire(ctx, credentialMasterKeyLockKey, credentialMasterKeyLockTTL, credentialMasterKeyLockWait)
	if err != nil {
		slog.Error("acquire credential master key lock failed", "err", err)
		return fmt.Errorf("acquire credential master key lock failed: %v", err)
	}
	defer func() {
		if e := lockObj.Unlock(ctx); e != nil {
			slog.Error("release credential master key lock failed", "err", e)
		}
	}()

	_, err = params.SettingRepository.Get(ctx, consts.SettingCredentialMasterKey)
	if err == nil {
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		slog.Error("get credential master key failed", "err", err)
		return err
	}

	key := make([]byte, credential.MinMasterKeyLength)
	if _, err = rand.Read(key); err != nil {
		slog.Error("generate credential master key failed", "err", err)
		return fmt.Errorf("generate credential master key failed: %v", err)
	}

	err = query.Q.Transaction(func(tx *query.Query) error {
		txSettingRepository := reposetting.NewSettingRepository(tx)
		return txSettingRepository.Create(ctx, consts.SettingCredentialMasterKey, key)
	})
	if err != nil {
		slog.Error("create credential master key failed", "err", err)
		return err
	}
	slog.Info("credential master key generated and stored in settings")
	return nil
}
