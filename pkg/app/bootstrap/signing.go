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
	"time"

	"github.com/sigstore/cosign/v3/pkg/cosign"
	"go.uber.org/dig"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/consts"
	"github.com/go-sigma/sigma/pkg/dal/query"
	reposetting "github.com/go-sigma/sigma/pkg/dal/repository/setting"
	"github.com/go-sigma/sigma/pkg/infra/lock"
)

const (
	// signingKeyLockKey serializes key generation across replicas.
	signingKeyLockKey  = "sigma:bootstrap:signing-key"
	signingKeyLockTTL  = 30 * time.Second
	signingKeyLockWait = 60 * time.Second
)

type signingParams struct {
	dig.In

	SettingRepository reposetting.SettingRepository
	Locker            lock.Locker
}

func signing(digCon *dig.Container) error {
	return digCon.Invoke(signingWithParams)
}

// signingWithParams makes sure a cosign signing key pair exists. The pair is
// generated once and persisted in the settings table; a distributed lock plus a
// second read keeps concurrent first starts from writing different pairs.
func signingWithParams(params signingParams) error {
	ctx := context.Background()

	lockObj, err := params.Locker.Acquire(ctx, signingKeyLockKey, signingKeyLockTTL, signingKeyLockWait)
	if err != nil {
		slog.Error("acquire signing key lock failed", "err", err)
		return fmt.Errorf("acquire signing key lock failed: %w", err)
	}
	defer func() {
		if e := lockObj.Unlock(ctx); e != nil {
			slog.Error("release signing key lock failed", "err", e)
		}
	}()

	settingRepository := params.SettingRepository
	_, err = settingRepository.Get(ctx, consts.SettingSignPrivateKey)
	if err == nil {
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		slog.Error("get signing key failed", "err", err)
		return err
	}

	keyBytes, err := cosign.GenerateKeyPair(nil)
	if err != nil {
		slog.Error("generate key failed", "err", err)
		return fmt.Errorf("generate key failed: %w", err)
	}

	err = query.Q.Transaction(func(tx *query.Query) error {
		txSettingRepository := reposetting.NewSettingRepository(tx)
		if err := txSettingRepository.Create(ctx, consts.SettingSignPrivateKey, keyBytes.PrivateBytes); err != nil {
			return err
		}
		return txSettingRepository.Create(ctx, consts.SettingSignPublicKey, keyBytes.PublicBytes)
	})
	if err != nil {
		// The in-memory locker does not span replicas, so two first starts can
		// both insert; when the pair exists now the other replica won.
		if tolerateConcurrentCreate(err, func() error {
			_, getErr := settingRepository.Get(ctx, consts.SettingSignPrivateKey)
			return getErr
		}) {
			slog.Info("signing key already created by another replica")
			return nil
		}
		slog.Error("create signing key failed", "err", err)
		return err
	}

	slog.Info("signing key pair generated and stored in settings")
	return nil
}
