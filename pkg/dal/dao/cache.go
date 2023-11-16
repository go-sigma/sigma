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

	"gorm.io/gorm"

	"github.com/go-sigma/sigma/pkg/dal/models"
	"github.com/go-sigma/sigma/pkg/dal/query"
)

//go:generate mockgen -destination=mocks/cache.go -package=mocks github.com/go-sigma/sigma/pkg/dal/dao CacheService
//go:generate mockgen -destination=mocks/cache_factory.go -package=mocks github.com/go-sigma/sigma/pkg/dal/dao CacheServiceFactory

// CacheService is the interface that provides methods to operate on cache model
type CacheService interface {
	// Save save a new cache record in the database
	Save(ctx context.Context, key string, val []byte, size int64, threshold float64) error
	// Delete get a cache record
	Delete(ctx context.Context, key string) error
	// Get get a cache record
	Get(ctx context.Context, key string) (*models.Cache, error)
}

type cacheService struct {
	tx *query.Query
}

// CacheServiceFactory is the interface that provides the cache service factory methods.
type CacheServiceFactory interface {
	New(txs ...*query.Query) CacheService
}

type cacheServiceFactory struct{}

// NewCacheServiceFactory creates a new cache service factory.
func NewCacheServiceFactory() CacheServiceFactory {
	return &cacheServiceFactory{}
}

func (s *cacheServiceFactory) New(txs ...*query.Query) CacheService {
	tx := query.Q
	if len(txs) > 0 {
		tx = txs[0]
	}
	return &cacheService{
		tx: tx,
	}
}

// Create creates a new cache record in the database
func (s cacheService) Save(ctx context.Context, key string, val []byte, size int64, threshold float64) error {
	total, err := s.tx.Cache.WithContext(ctx).Count()
	if err != nil {
		return err
	}
	err = query.Q.Transaction(func(tx *query.Query) error {
		if total > int64((float64(total) * (1 + threshold))) {
			err = tx.Cache.WithContext(ctx).DeleteOutsideThreshold(size, threshold)
			if err != nil {
				return err
			}
		}
		err = tx.Cache.WithContext(ctx).Save(&models.Cache{Key: key, Val: val})
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

// Delete get a cache record
func (s cacheService) Delete(ctx context.Context, key string) error {
	matched, err := s.tx.Cache.WithContext(ctx).Unscoped().Where(s.tx.Cache.Key.Eq(key)).Delete()
	if err != nil {
		return err
	}
	if matched.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// Get get a cache record
func (s cacheService) Get(ctx context.Context, key string) (*models.Cache, error) {
	return s.tx.Cache.WithContext(ctx).Where(s.tx.Cache.Key.Eq(key)).First()
}
