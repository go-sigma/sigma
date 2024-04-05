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

package dao

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
	"github.com/go-sigma/sigma/pkg/modules/locker/definition"
)

//go:generate mockgen -destination=mocks/locker.go -package=mocks github.com/go-sigma/sigma/pkg/dal/dao LockerService
//go:generate mockgen -destination=mocks/locker_factory.go -package=mocks github.com/go-sigma/sigma/pkg/dal/dao LockerServiceFactory

// LockerService is the interface that provides methods to operate on locker model
type LockerService interface {
	// Create creates a new work queue record in the database
	Create(ctx context.Context, key, val string, expire int64) error
	// Delete get a locker record
	Delete(ctx context.Context, key, val string) error
	// Renew renew a locker record
	Renew(ctx context.Context, key, val string, expire int64) error
}

type lockerService struct {
	tx *query.Query
}

// LockerServiceFactory is the interface that provides the locker service factory methods.
type LockerServiceFactory interface {
	New(txs ...*query.Query) LockerService
}

type lockerServiceFactory struct{}

// NewLockerServiceFactory creates a new locker service factory.
func NewLockerServiceFactory() LockerServiceFactory {
	return &lockerServiceFactory{}
}

func (s *lockerServiceFactory) New(txs ...*query.Query) LockerService {
	tx := query.Q
	if len(txs) > 0 {
		tx = txs[0]
	}
	return &lockerService{
		tx: tx,
	}
}

// Create creates a new work queue record in the database
func (s lockerService) Create(ctx context.Context, key, value string, expire int64) error {
	lock, err := s.tx.Locker.WithContext(ctx).Where(s.tx.Locker.Key.Eq(key)).First()
	if err == nil {
		if lock.Expire < time.Now().UnixMilli() {
			_, err = s.tx.Locker.WithContext(ctx).Where(s.tx.Locker.Key.Eq(key)).Delete()
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("Locker %s already exists", key)
		}
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return s.tx.Locker.WithContext(ctx).Create(&models.Locker{Key: key, Value: value, Expire: expire})
}

// Delete get a locker record
func (s lockerService) Delete(ctx context.Context, key, value string) error {
	_, err := s.tx.Locker.WithContext(ctx).Unscoped().Where(
		s.tx.Locker.Key.Eq(key), s.tx.Locker.Value.Eq(value)).Delete()
	return err
}

// Renew renew a locker record
func (s lockerService) Renew(ctx context.Context, key, value string, expire int64) error {
	lock, err := s.tx.Locker.WithContext(ctx).Where(s.tx.Locker.Key.Eq(key)).First()
	if err == nil {
		if lock.Value != value {
			return definition.ErrLockNotHeld
		}
		if lock.Expire < time.Now().UnixMilli() {
			_, err = s.tx.Locker.WithContext(ctx).Where(s.tx.Locker.Key.Eq(key)).Delete()
			if err != nil {
				return err
			}
			return definition.ErrLockAlreadyExpired
		} else {
			_, err := s.tx.Locker.WithContext(ctx).Where(s.tx.Locker.Key.Eq(key)).UpdateColumns(map[string]any{
				query.Locker.Expire.ColumnName().String(): expire,
			})
			if err != nil {
				return err
			}
		}
	}
	return err
}
