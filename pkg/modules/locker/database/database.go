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
	"path"
	"reflect"
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/configs"
	"github.com/go-sigma/sigma/pkg/dal/dao"
	"github.com/go-sigma/sigma/pkg/modules/locker"
)

func init() {
	locker.LockerFactories[path.Base(reflect.TypeOf(lockerFactory{}).PkgPath())] = &lockerFactory{}
}

type lockerDatabase struct {
	lockerServiceFactory dao.LockerServiceFactory
}

type lockerFactory struct{}

func (f lockerFactory) New(config configs.Configuration) (locker.Locker, error) {
	return &lockerDatabase{
		lockerServiceFactory: dao.NewLockerServiceFactory(),
	}, nil
}

type lock struct {
	name                 string
	lockerServiceFactory dao.LockerServiceFactory
}

// Lock ...
func (l lockerDatabase) Lock(ctx context.Context, name string, expire time.Duration) (locker.Lock, error) {
	err := l.lockerServiceFactory.New().Create(ctx, name)
	if err != nil {
		return nil, err
	}
	time.AfterFunc(expire, func() {
		err = l.lockerServiceFactory.New().Delete(context.Background(), name)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error().Err(err).Msgf("Delete locker(%s) failed", name)
		}
	})
	return &lock{
		name:                 name,
		lockerServiceFactory: l.lockerServiceFactory,
	}, nil
}

// Unlock ...
func (l lock) Unlock() error {
	err := l.lockerServiceFactory.New().Delete(context.Background(), l.name)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Error().Err(err).Msgf("Delete locker(%s) failed", l.name)
		return err
	}
	return nil
}
