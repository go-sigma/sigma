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

package database

import (
	"context"
	"errors"
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/dal/dao"
	"github.com/go-sigma/sigma/pkg/modules/locker/definition"
)

type lockerDatabase struct {
	lockerServiceFactory dao.LockerServiceFactory
}

func New(config configs.Configuration) (definition.Locker, error) {
	return &lockerDatabase{
		lockerServiceFactory: dao.NewLockerServiceFactory(),
	}, nil
}

type lock struct {
	name                 string
	release              bool
	lockerServiceFactory dao.LockerServiceFactory
}

// Lock ...
func (l lockerDatabase) Lock(ctx context.Context, name string, expire time.Duration) (definition.Lock, error) {
	err := l.lockerServiceFactory.New().Create(ctx, name)
	if err != nil {
		return nil, err
	}
	locker := &lock{
		name:                 name,
		lockerServiceFactory: l.lockerServiceFactory,
	}
	time.AfterFunc(expire, func() {
		err = locker.Unlock()
		if err != nil {
			log.Error().Err(err).Msgf("Delete locker(%s) failed", name)
		}
	})
	return locker, nil
}

// Unlock ...
func (l *lock) Unlock() error {
	if !l.release {
		err := l.lockerServiceFactory.New().Delete(context.Background(), l.name)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Msgf("Delete locker(%s) failed", l.name)
			return err
		}
		l.release = true
		return nil
	}
	return nil
}
