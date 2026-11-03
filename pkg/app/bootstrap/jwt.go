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
	svctoken "github.com/go-sigma/sigma/pkg/service/token"
)

const (
	// jwtPrivateKeyLockKey serializes key generation across replicas.
	jwtPrivateKeyLockKey  = "sigma:bootstrap:jwt-private-key"
	jwtPrivateKeyLockTTL  = 30 * time.Second
	jwtPrivateKeyLockWait = 60 * time.Second
)

type jwtParams struct {
	dig.In

	Config            *config.Configuration
	SettingRepository reposetting.SettingRepository
	Locker            lock.Locker
}

func jwtKey(digCon *dig.Container) error {
	return digCon.Invoke(jwtWithParams)
}

// jwtWithParams makes sure a JWT signing key is available.
//
// A configured auth.jwt.privateKey is authoritative and is only validated.
// Otherwise the key stored in the settings table is used, and when neither
// exists a fresh Ed25519 key is generated and persisted so that restarts and
// replicas keep issuing tokens that verify against each other. A distributed
// lock plus a second read keeps concurrent first starts from writing different
// keys.
func jwtWithParams(params jwtParams) error {
	if configured := params.Config.Auth.Jwt.PrivateKey; configured != "" {
		if _, err := svctoken.ParsePrivateKey(configured); err != nil {
			return fmt.Errorf("invalid auth.jwt.privateKey: %w", err)
		}
		slog.Info("jwt private key loaded from config")
		return nil
	}

	ctx := context.Background()

	lockObj, err := params.Locker.Acquire(ctx, jwtPrivateKeyLockKey, jwtPrivateKeyLockTTL, jwtPrivateKeyLockWait)
	if err != nil {
		slog.Error("acquire jwt private key lock failed", "err", err)
		return fmt.Errorf("acquire jwt private key lock failed: %v", err)
	}
	defer func() {
		if e := lockObj.Unlock(ctx); e != nil {
			slog.Error("release jwt private key lock failed", "err", e)
		}
	}()

	settingObj, err := params.SettingRepository.Get(ctx, consts.SettingJwtPrivateKey)
	if err == nil {
		key := string(settingObj.Val)
		if _, verr := svctoken.ParsePrivateKey(key); verr != nil {
			return fmt.Errorf("invalid %s in settings: %w", consts.SettingJwtPrivateKey, verr)
		}
		params.Config.Auth.Jwt.PrivateKey = key
		slog.Info("jwt private key loaded from settings")
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		slog.Error("get jwt private key failed", "err", err)
		return err
	}

	key, err := svctoken.GeneratePrivateKey()
	if err != nil {
		slog.Error("generate jwt private key failed", "err", err)
		return err
	}

	err = query.Q.Transaction(func(tx *query.Query) error {
		return reposetting.NewSettingRepository(tx).Create(ctx, consts.SettingJwtPrivateKey, []byte(key))
	})
	if err != nil {
		slog.Error("create jwt private key failed", "err", err)
		return err
	}

	params.Config.Auth.Jwt.PrivateKey = key
	slog.Info("jwt private key generated and stored in settings")
	return nil
}
